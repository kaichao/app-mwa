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

func fromPullUnpack(msg string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromPullUnpack()")
	}()
	common.AddTimeStamp("enter-fromPullUnpack()")
	// input message: 1257617424/p00001_00096/1257617426_1257617465_ch112.dat.tar.zst
	// - target_dir:1257617424/t1257617426_1257617505/ch111
	// semaphore: dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+_[0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+).dat.tar.zst$`)
	ss := re.FindStringSubmatch(msg)
	if len(ss) == 0 {
		logrus.Errorf("Invalid Message Format, body=%s\n", msg)
		return 1
	}
	prefix := ss[1]
	t, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])
	cube := datacube.NewDataCube(prefix)
	t0, t1 := cube.GetTimeRange(t)

	cubeID := fmt.Sprintf("%s/t%d_%d", prefix, t0, t1)
	sema := fmt.Sprintf(`dat-ready:%s/ch%d`, cubeID, ch)
	v, err := semaphore.AddValue(sema, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s\n", sema)
		return 2
	}
	semaVal, _ := strconv.Atoi(v)
	if semaVal > 0 {
		return 0
	}
	common.AddTimeStamp("prepare-messages")
	return toBeamMake(cubeID, ch, headers)
}

func toPullUnpack(body string, fromHeaders map[string]string) int {
	cube := datacube.NewDataCube(body)
	fmt.Println(cube.ToCubeString())
	trs := cube.GetTimeRanges()
	if len(trs) != 2 {
		logrus.Errorf("Only one time-range allowed for cube-id:%s\n", body)
		return 1
	}

	prefix := fmt.Sprintf("%s/p%05d_%05d", cube.ObsID, cube.PointingBegin, cube.PointingEnd)
	numGroups := len(node.Nodes) / 24

	trBegin := trs[0]
	trEnd := trs[1]
	messages := []string{}
	tus := cube.GetTimeUnitsWithinInterval(trBegin, trEnd)
	nTimeUnits := len(tus) / 2

	prs := cube.GetPointingRangesByInterval(cube.PointingBegin, cube.PointingEnd)
	nPRanges := len(prs) / 2

	semaDatReady := ""
	semaDatDone := ""

	for j := 0; j < cube.NumOfChannels; j++ {
		ch := cube.ChannelBegin + j
		id := fmt.Sprintf("%s/t%d_%d/ch%d", prefix, trBegin, trEnd, ch)
		semaPair := fmt.Sprintf(`"dat-ready:%s":%d`, id, nTimeUnits)
		semaDatReady += semaPair + "\n"

		semaPair = fmt.Sprintf(`"dat-done:%s":%d`, id, nPRanges)
		semaDatDone += semaPair + "\n"

		targetSubDir := fmt.Sprintf("%s/t%d_%d/ch%d", cube.ObsID, trBegin, trEnd, ch)
		headers := fmt.Sprintf(`{"target_subdir":"%s"}`, targetSubDir)
		cubeIndex, err := strconv.Atoi(fromHeaders["_cube_index"])
		if err == nil {
			cubeIndex--
			if cubeIndex < numGroups {
				// > 24节点，首次加载设置更高的带宽
				headers = common.SetJSONAttribute(headers, "bw_limit", os.Getenv("FIRST_BW_LIMIT"))
			}
		}
		toHost := node.GetNodeNameByIndexChannel(cube, cubeIndex, ch)
		headers = common.SetJSONAttribute(headers, "to_host", toHost)

		if iopath.IsPreloadMode() {
			headers = common.SetJSONAttribute(headers, "source_url", iopath.GetPreloadRoot(j))
		}

		// if os.Getenv("PRELOAD_MODE") == "multi-account-relay" {
		// 	varName := `cube-stor-index:` + id
		// 	storIndex, err := variable.Get(varName, appID)
		// 	if err != nil {
		// 		logrus.Errorf("variable get, var-name:%s, err-info:%v\n", varName, err)
		// 		return 2
		// 	}
		// 	sourceURL := fmt.Sprintf("cstu00%s@10.100.1.104/public/home/cstu00%s",
		// 		storIndex, storIndex)
		// 	headers = common.SetJSONAttribute(headers, "source_url", sourceURL)
		// } else {
		// 	sourceURL := os.Getenv("SOURCE_TAR_ROOT")
		// 	if sourceURL == "" {
		// 		sourceURL = sourcePicker.GetNext()
		// 	}
		// 	headers = common.SetJSONAttribute(headers, "source_url", sourceURL)
		// }

		for k := 0; k < len(tus); k += 2 {
			m := fmt.Sprintf("%s/%d_%d_ch%d.dat.tar.zst", prefix, tus[k], tus[k+1], ch)
			messages = append(messages, m+","+headers)
		}
	}
	// 信号量dat-ready、dat-done、fits-done、pointing-done
	semaFitsDone := ""
	// fits-done:1257010784/p00001/t1257010786_1257010985
	for k := 0; k < len(prs); k += 2 {
		id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d`, cube.ObsID, prs[k], prs[k+1], trBegin, trEnd)
		semaPair := fmt.Sprintf(`"fits-done:%s":%d`, id, 24)
		semaFitsDone += semaPair + "\n"
	}

	semaphores := semaDatReady + semaDatDone + semaFitsDone
	common.AppendToFile("my-sema.txt", semaphores)
	err := semaphore.CreateFileSemaphores("my-sema.txt", appID, 500)
	if err != nil {
		logrus.Errorf("create sema, err-info:%v\n", err)
		return 1
	}
	// 消息
	cubeID := fmt.Sprintf("%s/p%05d_%05d/t%d_%d", cube.ObsID,
		cube.PointingBegin, cube.PointingEnd, cube.TimeBegin, cube.TimeEnd)
	targetURL := fmt.Sprintf("%s/mydata/mwa/dat", os.Getenv("LOCAL_TMPDIR"))
	headers := map[string]string{
		"_cube_id":    cubeID,
		"_cube_index": fromHeaders["_cube_index"],
		"target_url":  targetURL,
	}
	envs := map[string]string{
		"SINK_JOB": "pull-unpack",
	}

	return task.AddTasksWithMapHeaders(messages, headers, envs)
}
