package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"mr/datacube"

	"github.com/kaichao/scalebox/pkg/exec"
)

var (
	fromFuncs = map[string]func(string, map[string]string) int{
		"":             defaultFunc,
		"dir-list":     fromDirList,
		"cluster-dist": fromClusterDist,
		"pull-unpack":  fromPullUnpack,
		"beam-maker":   fromBeamMaker,
		"down-sampler": fromDownSampler,
		"fits-redist":  fromFitsRedist,
		"fits-merger":  fromFitsMerger,
	}
)

func main() {
	logger.Infoln("00, Entering message-router")
	if len(os.Args) < 3 {
		logger.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
		os.Exit(1)
	}

	logger.Infof("01, after number of arguments verification, message-body:%s,message-header:%s.\n",
		os.Args[1], os.Args[2])
	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &headers); err != nil {
		logger.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	logger.Infoln("02, after JSON format verification of headers")

	doMessageRoute := fromFuncs[headers["from_job"]]
	if doMessageRoute == nil {
		logger.Warnf("from_job not set/not existed in message-router, from_job=%s ,message=%s\n",
			headers["from_job"], os.Args[1])
		os.Exit(4)
	}

	AddTimeStamp("before-mr")
	logger.Infoln("03, message-router not null")
	exitCode := doMessageRoute(os.Args[1], headers)
	if exitCode != 0 {
		logger.Errorf("error found, error-code=%d\n", exitCode)
	}
	AddTimeStamp("before-exit")
	os.Exit(exitCode)
}

func defaultFunc(message string, headers map[string]string) int {
	// 初始的启动消息（数据集ID）
	// /raid0/scalebox/mydata/mwa/tar~1257010784
	// <user>@<remote-ip>/raid0/scalebox/mydata/mwa/tar~1257010784
	ss := strings.Split(message, "~")
	if len(ss) != 2 {
		fmt.Fprintf(os.Stderr, "Invalid message format, msg-body:%s\n", message)
		return 3
	}
	cube := datacube.GetDataCube(ss[1])
	if cube == nil {
		fmt.Fprintf(os.Stderr, "Invalid datacube format, metadata:%s\n", ss[2])
		return 4
	}

	// first one
	createDatReadySemaphores(cube)

	createPointingBatchLeftSemaphores(cube)
	createDatProcessedSemaphores(cube)

	createFits24chReadySemaphores(cube)

	createPullUnpackProgressCountSemaphores(cube)
	createBeamMakerProgressCountSemaphores(cube)

	cmd := fmt.Sprintf("scalebox task add --header prefix_url=%s %s", ss[0], ss[1])
	code, _ := exec.ExecCommandReturnExitCode(cmd, 0)
	return code
}
