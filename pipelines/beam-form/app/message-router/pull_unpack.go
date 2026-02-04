/*

 */

package main

import (
	"beamform/app/message-router/iopath"
	"beamform/internal/datacube"
	"beamform/internal/node"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromPullUnpack(body string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromPullUnpack()")
	}()
	common.AddTimeStamp("enter-fromPullUnpack()")
	// input message: 1257617424/p00001_00096/1257617426_1257617465_ch112.dat.tar.zst
	// - target_dir:1257617424/t1257617426_1257617505/ch111
	// semaphore: dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+_[0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+).dat.tar.zst$`)
	ss := re.FindStringSubmatch(body)
	if len(ss) == 0 {
		logrus.Errorf("Invalid Message Format, body=%s\n", body)
		return 1
	}
	prefix := ss[1]
	t, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])
	cube := datacube.NewDataCube(prefix)
	t0, t1 := cube.GetTimeRange(t)

	cubeID := fmt.Sprintf("%s/t%d_%d", prefix, t0, t1)
	sema := fmt.Sprintf(`dat-ready:%s/ch%d`, cubeID, ch)
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	semaVal, err := semaphore.AddValue(sema, vtaskID, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s\n", sema)
		return 2
	}
	if semaVal > 0 {
		return 0
	}
	common.AddTimeStamp("prepare-messages")
	return toBeamMake(cubeID, ch, headers)
}

func toPullUnpack(body string, fromHeaders map[string]string) int {
	cube := datacube.NewDataCube(body)
	trs := cube.GetTimeRanges()
	if len(trs) != 2 {
		logrus.Errorf("Only one time-range allowed for cube-id:%s\n", body)
		return 1
	}

	prefix := fmt.Sprintf("%s/p%05d_%05d", cube.ObsID, cube.PointingBegin, cube.PointingEnd)
	// numGroups := len(node.Nodes) / 24

	trBegin := trs[0]
	trEnd := trs[1]
	tasks := []string{}
	tus := cube.GetTimeUnitsWithinInterval(trBegin, trEnd)
	nTimeUnits := len(tus) / 2

	prs := cube.GetPointingRangesByInterval(cube.PointingBegin, cube.PointingEnd)
	nPRanges := len(prs) / 2

	semaphores := []string{}
	slotSeq, _ := strconv.Atoi(fromHeaders["_slot_seq"])
	hostExpr := os.Getenv("NODES")
	hostPrefix := hostExpr[0:1]
	if hostPrefix == "^" {
		hostPrefix = hostExpr[1:2]
	}
	cube0 := datacube.NewDataCube(cube.ObsID)
	for j := 0; j < cube.NumOfChannels; j++ {
		ch := cube.ChannelBegin + j
		id := fmt.Sprintf("%s/t%d_%d/ch%d", prefix, trBegin, trEnd, ch)
		semaPair := fmt.Sprintf(`"dat-ready:%s":%d`, id, nTimeUnits)
		semaphores = append(semaphores, semaPair)

		semaPair = fmt.Sprintf(`"dat-done:%s":%d`, id, nPRanges)
		semaphores = append(semaphores, semaPair)

		targetSubDir := fmt.Sprintf("%s/t%d_%d/ch%d", cube.ObsID, trBegin, trEnd, ch)
		headers := fmt.Sprintf(`{"target_subdir":"%s"}`, targetSubDir)
		// cubeIndex, err := strconv.Atoi(fromHeaders["_cube_index"])
		// if err == nil {
		// 	cubeIndex--
		// 	if cubeIndex < numGroups {
		// 		// > 24节点，首次加载设置更高的带宽
		// 		headers = common.SetJSONAttribute(headers, "bw_limit", os.Getenv("FIRST_BW_LIMIT"))
		// 	}
		// }
		// 如果节点数少于24，纠正index
		index := j % len(node.Nodes)
		toHost := fmt.Sprintf("%s%02d-%02d", hostPrefix, slotSeq, index)
		// toHost := node.GetNodeNameByIndexChannel(cube, cubeIndex, ch)
		headers, _ = common.SetJSONAttribute(headers, "to_host", toHost)

		for k := 0; k < len(tus); k += 2 {
			if iopath.IsPreloadMode() {
				index := cube0.GetTimeChannelIndex(tus[k], ch)
				headers, _ = common.SetJSONAttribute(headers, "source_url", iopath.GetPreloadRoot(index))
			}
			m := fmt.Sprintf("%s/%d_%d_ch%d.dat.tar.zst", prefix, tus[k], tus[k+1], ch)
			tasks = append(tasks, m+","+headers)
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
	err := semaphore.CreateSemaphores(semaphores, vtaskID, appID, 500)
	if err != nil {
		logrus.Errorf("create sema, err-info:%v\n", err)
		return 1
	}
	// 消息
	// cubeID := fmt.Sprintf("%s/p%05d_%05d/t%d_%d", cube.ObsID,
	// 	cube.PointingBegin, cube.PointingEnd, cube.TimeBegin, cube.TimeEnd)
	targetURL := fmt.Sprintf("%s/mydata/mwa/dat", os.Getenv("LOCAL_TMPDIR"))
	headers := map[string]string{
		"target_url": targetURL,
	}
	envs := map[string]string{
		"SINK_MODULE": "pull-unpack",
	}

	_, err = task.AddTasksWithMapHeaders(tasks, headers, envs)
	if err != nil {
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	return 0
}
