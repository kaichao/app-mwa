package main

import (
	"beamform/internal/pkg/message"
	"fmt"
	"os"
	"strings"

	"github.com/kaichao/scalebox/pkg/misc"
)

func defaultFunc(msg string, headers map[string]string) int {
	// input message:
	// 	1257010784
	// 	1257010784/p00001_00960
	// 	1257010784/p00001_00960/t1257012766_1257012965
	messages, semaFitsDone, semaPointingDone := message.ProcessForBeamMake(msg)
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
