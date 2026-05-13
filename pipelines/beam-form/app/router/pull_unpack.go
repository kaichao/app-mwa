/*

 */

package main

import (
	"beamform/internal/datacube"
	"beamform/internal/node"
	"database/sql"

	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/kaichao/scalebox/pkg/vtask"
)

func fromPullUnpack(body string, headers map[string]string) error {
	defer func() {
		common.AddTimeStamp("leave-fromPullUnpack()")
	}()
	common.AddTimeStamp("enter-fromPullUnpack()")
	// task-body: 1257617424/p00001_00096/1257617426_1257617465_ch112.dat.tar.zst
	// - target_dir:1257617424/t1257617426_1257617505/ch111
	// semaphore: dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+_[0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+).dat.tar.zst$`)
	ss := re.FindStringSubmatch(body)
	if len(ss) == 0 {
		return errors.E("invalid task-body Format", "task-body", body)
	}
	prefix := ss[1]
	t, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])
	cube := datacube.NewDataCube(prefix)
	t0, t1 := cube.GetTimeRange(t)

	cubeID := fmt.Sprintf("%s/t%d_%d", prefix, t0, t1)
	sema := fmt.Sprintf(`dat-ready:%s/ch%d`, cubeID, ch)
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	// semaVal, err := semaphore.AddValue(sema, vtaskID, appID, -1)
	semaVal, err := vtask.AddSemaphoreValue(sema, -1, vtaskID, appID)

	if err != nil {
		return errors.WrapE(err, 2, "semaphore-decrement",
			"sema-name", sema, "app-id", appID, "vtask-id", vtaskID)
	}
	if semaVal > 0 {
		return nil
	}
	common.AddTimeStamp("prepare-tasks")
	return errors.WrapE(toBeamMake(cubeID, ch, headers), "toBeamMake()",
		"cube-id", cubeID, "ch", ch, "headers", headers)
}

func toPullUnpack(body string, fromHeaders map[string]string) error {
	// 节点组的数量
	numGroup, err := strconv.Atoi(os.Getenv("NUM_GROUPS"))
	if err != nil || numGroup <= 0 {
		numGroup = 1
	}

	cube := datacube.NewDataCube(body)
	trs := cube.GetTimeRanges()
	if len(trs) != 2 {
		return errors.E("only one time-range allowed", "cube-id", body)
	}

	trBegin := trs[0]
	trEnd := trs[1]

	tus := cube.GetTimeUnitsWithinInterval(trBegin, trEnd)
	nTimeUnits := len(tus) / 2

	prs := cube.GetPointingRangesByInterval(cube.PointingBegin, cube.PointingEnd)
	nPRanges := len(prs) / 2

	// slotSeq, _ := strconv.Atoi(fromHeaders["_slot_seq"])
	hostExpr := os.Getenv("NODES")
	hostPrefix := hostExpr[0:1]
	if hostPrefix == "^" {
		hostPrefix = hostExpr[1:2]
	}
	tasks := []string{}
	semaphores := []string{}
	pointingPrefix := fmt.Sprintf("%s/p%05d_%05d", cube.ObsID, cube.PointingBegin, cube.PointingEnd)
	for j := 0; j < cube.NumOfChannels; j++ {
		ch := cube.ChannelBegin + j
		id := fmt.Sprintf("%s/t%d_%d/ch%d", pointingPrefix, trBegin, trEnd, ch)
		semaphores = append(semaphores, fmt.Sprintf(`"dat-ready:%s":%d`, id, nTimeUnits))
		semaphores = append(semaphores, fmt.Sprintf(`"dat-done:%s":%d`, id, nPRanges))

		targetSubDir := fmt.Sprintf("%s/t%d_%d/ch%d", cube.ObsID, trBegin, trEnd, ch)
		headers := fmt.Sprintf(`{"target_subdir":"%s"}`, targetSubDir)
		// 如果节点数少于24，纠正index
		index := (j + (numGroup-1)*24) % len(node.Nodes)
		// host的格式：d00-23
		// toHost := fmt.Sprintf("%s%02d-%02d", hostPrefix, slotSeq, index)
		toHost := node.Nodes[index].Name
		headers, _ = common.SetJSONAttribute(headers, "to_host", toHost)

		if bwLimitMB := getOptBandwidthMB(toHost); bwLimitMB != "" {
			headers, _ = common.SetJSONAttribute(headers, "bw_limit", bwLimitMB)
		}

		for k := 0; k < len(tus); k += 2 {
			fileName := fmt.Sprintf("%d_%d_ch%d.dat.tar.zst", tus[k], tus[k+1], ch)
			// sourceURL, err := iopath.GetPreloadRoot(cube.ObsID + "/" + fileName)
			sourceURL := os.Getenv("PRELOAD_ROOT")
			if sourceURL == "" {
				if v := os.Getenv("ORIGIN_ROOT"); v != "" {
					sourceURL = v
				} else {
					key := cube.ObsID + "/" + fileName
					vpath, err := vPath.GetPath("preload-tar", key)
					if err != nil {
						return errors.WrapE(err, "vPath.GetPath()",
							"category", "preload-tar", "key", key)
					}
					sourceURL = vpath
				}
			}
			if sourceURL == "" {
				return errors.E("null source_url for preload")
			}
			headers, _ = common.SetJSONAttribute(headers, "source_url", sourceURL)

			// 增加"_global_dat_dir"
			// path: 1302282040/t1302282041_1302282200/ch126
			if os.Getenv("GROUP_NODES") != "" {
				globalDatDir, err := vPath.GetPath("global-dat", targetSubDir)
				if err != nil {
					logger.LogError(err, logEntry)
				} else {
					headers, _ = common.SetJSONAttribute(headers,
						"_global_dat_dir", globalDatDir)
				}
			}

			body := pointingPrefix + "/" + fileName
			tasks = append(tasks, body+","+headers)
		}
	}

	// 信号量dat-ready、dat-done、fits-done、pointing-done
	for k := 0; k < len(prs); k += 2 {
		id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d`, cube.ObsID, prs[k], prs[k+1], trBegin, trEnd)
		// fits-done:1257010784/p00001/t1257010786_1257010985
		semaPair := fmt.Sprintf(`"fits-done:%s":%d`, id, 24)
		semaphores = append(semaphores, semaPair)
	}

	vtaskID, _ := strconv.ParseInt(fromHeaders["_vtask_id"], 10, 64)
	err = vtask.CreateSemaphores(semaphores, vtaskID, appID, 500)
	if err != nil {
		return errors.WrapE(err, "semaphore.CreateSemaphores()",
			"sema-lines", semaphores, "app-id", appID, "vtask-id", vtaskID)
	}

	targetURL := os.Getenv("LOCAL_TMPDIR")
	headers := map[string]string{
		"target_url": targetURL,
	}
	envs := map[string]string{
		"SINK_MODULE": "pull-unpack",
	}
	_, err = task.AddTasksWithMapHeaders(tasks, headers, envs)
	return errors.WrapE(err, "add-tasks",
		"task-lines", tasks, "headers", headers, "envs", envs)
}

// 获取优化的带宽，以MB/s计，返回字符串，'100m'/'1000k'
func getOptBandwidthMB(toHost string) string {
	// 共享变量是否存在，若不存在，若为空，则为缺省值 yes
	varName := "first_load:pull_unpack:" + toHost
	val, err := variable.GetValue(varName, appID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			logger.LogError(err, logEntry)
			return ""
		}
		// 未定义first_load:pull_unpack
	}
	if val != "" {
		return os.Getenv("BW_LIMIT")
	}
	err = variable.Set(varName, "no", appID)
	if err != nil {
		logger.LogError(err, logEntry)
		return ""
	}
	return os.Getenv("FIRST_BW_LIMIT")
}
