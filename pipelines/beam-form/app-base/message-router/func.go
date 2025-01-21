package main

import (
	"beamform/internal/pkg/datacube"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/misc"
)

func defaultFunc(message string, headers map[string]string) int {
	// input message:
	// 	1257010784
	// 	1257010784/p00001_00960
	// 	1257010784/p00001_00960/t1257012766_1257012965
	messages, semaFitsDone, semaPointingDone := processMessage(message)
	misc.AppendToFile("/work/custom-out.txt",
		fmt.Sprintf("n_messages:%d,nsemaFitsDone:%d,nsemaPointingDone:%d\n",
			len(messages), len(semaFitsDone), len(semaPointingDone)))
	fmtSemaCreate := `scalebox semaphore create '{"semaphores":{%s}}'`
	batchSize := 120
	for i := 0; i < len(semaFitsDone); i += batchSize {
		end := i + batchSize
		if end > len(semaFitsDone) {
			end = len(semaFitsDone)
		}
		cmd := fmt.Sprintf(fmtSemaCreate, strings.Join(semaFitsDone[i:end], ","))
		if code := misc.ExecCommandReturnExitCode(cmd, 600); code != 0 {
			return code
		}
	}
	for i := 0; i < len(semaPointingDone); i += batchSize {
		end := i + batchSize
		if end > len(semaPointingDone) {
			end = len(semaPointingDone)
		}
		cmd := fmt.Sprintf(fmtSemaCreate, strings.Join(semaPointingDone[i:end], ","))
		if code := misc.ExecCommandReturnExitCode(cmd, 600); code != 0 {
			return code
		}
	}

	for _, m := range messages {
		misc.AppendToFile(os.Getenv("WORK_DIR")+"/task-body.txt", m)
	}

	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	cmd := "scalebox task add --sink-job=beam-make"
	return misc.ExecCommandReturnExitCode(cmd, 600)
}

// return:
//
//	messages, sema fits-done, sema pointing-done
func processMessage(m string) ([]string, []string, []string) {
	re := regexp.MustCompile("^([0-9]+)((/p([0-9]+)_([0-9]+))(/t([0-9]+)_([0-9]+))?)?$")
	ss := re.FindStringSubmatch(m)
	dataset := ss[1]
	cube := datacube.GetDataCube(dataset)
	var (
		pBegin, pEnd int
		ts           []int
	)
	if ss[7] != "" {
		// 	1257010784/p00001_00960/t1257012766_1257012965
		t0, _ := strconv.Atoi(ss[7])
		t1, _ := strconv.Atoi(ss[8])
		ts = append(ts, t0, t1)
	} else {
		// 	1257010784/p00001_00960
		// 	1257010784
		ts = cube.GetTimeRanges()
	}
	if ss[4] != "" {
		// 	1257010784/p00001_00960/t1257012766_1257012965
		// 	1257010784/p00001_00960
		pBegin, _ = strconv.Atoi(ss[4])
		pEnd, _ = strconv.Atoi(ss[5])
	} else {
		// 	1257010784
		pBegin = cube.PointingBegin
		pEnd = cube.PointingEnd
	}
	ps := cube.GetPointingRangesByInterval(pBegin, pEnd)

	messages := []string{}
	for k := 0; k < len(ps); k += 2 {
		for j := 0; j < len(ts); j += 2 {
			for i := 0; i < cube.NumOfChannels; i++ {
				messages = append(messages, fmt.Sprintf("%s/p%05d_%05d/t%d_%d/ch%03d",
					dataset, ps[k], ps[k+1], ts[j], ts[j+1], cube.ChannelBegin+i))
			}
		}
	}
	semaFitsDone := []string{}
	semaPointingDone := []string{}
	// fits-done:1257010784/p00001/t1257010786_1257010985
	// pointing-done:1257010784/p00001
	nTimeRanges := len(ts) / 2
	for k := pBegin; k <= pEnd; k++ {
		for j := 0; j < len(ts); j += 2 {
			sema := fmt.Sprintf(`"fits-done:%s/p%05d/t%d_%d":%d`,
				dataset, k, ts[j], ts[j+1], 24)
			semaFitsDone = append(semaFitsDone, sema)
		}
		if ss[7] == "" {
			sema := fmt.Sprintf(`"pointing-done:%s/p%05d":%d`,
				dataset, k, nTimeRanges)
			semaPointingDone = append(semaPointingDone, sema)
		}
	}
	return messages, semaFitsDone, semaPointingDone
}

func fromMessageRouter(message string, headers map[string]string) int {
	return 0
}
func fromDownSample(message string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// output message: 1257010784/p00023/t1257010786_1257010965

	// semaphore: fits-ready:1257010784/p00001/t1257010786_1257010985
	return 0
}
func fromFitsMerge(message string, headers map[string]string) int {
	// semaphore: pointing-ready:1257010784/p00001

	return 0
}

var (
	fromFuncs = map[string]func(string, map[string]string) int{
		"":               defaultFunc,
		"message-router": fromMessageRouter,
		"down-sample":    fromDownSample,
		"fits-merge":     fromFitsMerge,
	}
)
