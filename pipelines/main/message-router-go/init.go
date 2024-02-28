package main

import (
	"encoding/json"
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

	hosts = []string{"10.11.16.80", "10.11.16.75"}
	// hosts            = []string{"10.11.16.79", "10.11.16.80", "10.11.16.76", "10.11.16.75"}
	numNodesPerGroup int

	localMode bool

	workDir string

	pBegin, pEnd, pStep int

	tStep int
)

func init() {
	var err error

	workDir = os.Getenv("WORD_DIR")
	if workDir == "" {
		workDir = "/work"
	}

	logger = logrus.New()
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.WarnLevel
	}
	logger.SetLevel(level)
	logger.SetReportCaller(true)

	numNodesPerGroup, err = strconv.Atoi(os.Getenv("NUM_NODES_PER_GROUP"))
	if err != nil || numNodesPerGroup == 0 {
		// numNodesPerGroup = 24
		numNodesPerGroup = len(hosts)
	}

	pBegin, err = strconv.Atoi(os.Getenv("POINTING_BEGIN"))
	if err != nil || pBegin == 0 {
		pBegin = 1
	}
	pEnd, err = strconv.Atoi(os.Getenv("POINTING_END"))
	if err != nil || pEnd == 0 {
		pEnd = 144
	}
	pStep, err := strconv.Atoi(os.Getenv("NUM_POINTINGS_PER_CALC"))
	if err != nil || pStep == 0 {
		pStep = 24
	}

	tStep, err = strconv.Atoi(os.Getenv("NUM_SECONDS_PER_CALC"))
	if err != nil || tStep == 0 {
		tStep = 30
	}

	localMode = os.Getenv("LOCAL_MODE") == "yes"
}

func sendNodeAwareMessage(message string, headers map[string]string, sinkJob string, num int) int {
	if !localMode {
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+message)
		return 0
	}

	toHost := hosts[num%numNodesPerGroup]
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, message)
	if len(headers) > 0 {
		h, err := json.Marshal(headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "headers:%v,JSON marshaling failed:%v\n", headers, err)
		} else {
			cmdTxt = fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s --headers '%s' %s", sinkJob, toHost, h, message)
		}
	}

	fmt.Printf("cmd-text:%s\n", cmdTxt)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func getPointingRanges() map[int]int {
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
