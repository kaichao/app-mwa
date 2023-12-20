package main

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	scalebox "github.com/kaichao/scalebox/golang/misc"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger

	// 每n个fits文件合并为1个fil文件
	// numPerGroup int
	// 每次观测最大序列号
	// maxSequence int

	hosts            = []string{"10.11.16.79", "10.11.16.80"}
	numNodesPerGroup int

	numPointingsPerCalc int

	localMode bool

	numSecondsPerCalc int

	datasetFile = "/work/.scalebox/dataset-v.txt"
)

func init() {
	var (
		err error
	)
	logger = logrus.New()
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.WarnLevel
	}
	logger.SetLevel(level)
	logger.SetReportCaller(true)

	numNodesPerGroup, err = strconv.Atoi(os.Getenv("NUM_NODES_PER_GROUP"))
	if err != nil || numNodesPerGroup == 0 {
		numNodesPerGroup = 24
	}
	numPointingsPerCalc, err = strconv.Atoi(os.Getenv("NUM_POINTINGS_PER_CALC"))
	if err != nil || numPointingsPerCalc == 0 {
		numPointingsPerCalc = 24
	}
	numSecondsPerCalc, err = strconv.Atoi(os.Getenv("NUM_SECONDS_PER_CALC"))
	if err != nil || numSecondsPerCalc == 0 {
		numSecondsPerCalc = 120
	}

	localMode = os.Getenv("LOCAL_MODE") == "yes"
}

func sendChannelAwareMessage(message string, sinkJob string, channel int) int {
	if !localMode {
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+message)
		return 0
	}
	toHost := hosts[(channel-109)%numNodesPerGroup]
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, message)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func getPointingRange() map[int]int {
	begin, err := strconv.Atoi(os.Getenv("POINTING_BEGIN"))
	if err != nil || begin == 0 {
		begin = 1
	}
	end, err := strconv.Atoi(os.Getenv("POINTING_END"))
	if err != nil || end == 0 {
		end = 144
	}
	step, err := strconv.Atoi(os.Getenv("NUM_POINTINGS_PER_CALC"))
	if err != nil || step == 0 {
		step = 24
	}

	ret := make(map[int]int)
	for i := begin; i <= end; i += step {
		j := i + step - 1
		if j > end {
			j = end
		}
		ret[i] = j
	}
	return ret
}
