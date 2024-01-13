package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

var (
	// counters = make(map[string]int)
	workDir string

	db *sql.DB

	pBegin, pEnd, pStep int

	tStep int
)

func init() {
	var err error

	workDir = os.Getenv("WORD_DIR")
	if workDir == "" {
		workDir = "/work"
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
}

func createDatReadySemaphores(dataset *DataSet) {
	// TARGET: beam-maker
	// 1257010784/1257010786_1257010815/112/00001_00024
	// 1257010784/1257010786_1257010815/112
	arr := dataset.getTimeRanges()
	for i := 0; i < len(arr); i += 2 {
		// all dat files in current range
		initValue := arr[i+1] - arr[i] + 1
		for ch := 109; ch <= 132; ch++ {
			sema := fmt.Sprintf("dat-ready:%s/%d_%d/%d",
				dataset.DatasetID, arr[i], arr[i+1], ch)
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			addSemaphore(sema, initValue)
		}
	}
}

func createFits24chReadySemaphores(dataset *DataSet) {
	// TARGET: fits-merger
	// 1257010784/1257010786_1257010815/00024

	// 24-channel
	initValue := 24

	arr := dataset.getTimeRanges()

	for p := pBegin; p <= pEnd; p++ {
		for i := 0; i < len(arr); i += 2 {
			sema := fmt.Sprintf("fits-24ch-ready:%s/%d_%d/%05d", dataset.DatasetID, arr[i], arr[i+1], p)
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			addSemaphore(sema, initValue)
		}
	}
}

func createDatUsedSemaphores(dataset *DataSet) {
	// all pointing
	initValue := pEnd - pBegin + 1

	arr := dataset.getTimeRanges()
	for i := 0; i < len(arr); i += 2 {
		for ch := 109; ch <= 132; ch++ {
			sema := fmt.Sprintf("dat-used:%s/%d_%d/ch%d", dataset.DatasetID, arr[i], arr[i+1], ch)
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			addSemaphore(sema, initValue)
		}
	}
}

func (dataset *DataSet) getTimeRanges() []int {
	var ret []int

	for y := 0; y < dataset.VerticalHeight; y += tStep {
		y0 := dataset.VerticalStart + y
		y1 := y0 + tStep - 1
		if y1 > dataset.VerticalStart+dataset.VerticalHeight-1 {
			y1 = dataset.VerticalStart + dataset.VerticalHeight - 1
		}
		ret = append(ret, y0, y1)
	}
	return ret
}

func (dataset *DataSet) getTimeRange(t int) (int, int) {
	for y := 0; y < dataset.VerticalHeight; y += tStep {
		y0 := dataset.VerticalStart + y
		y1 := y0 + tStep - 1
		if y1 > dataset.VerticalStart+dataset.VerticalHeight-1 {
			y1 = dataset.VerticalStart + dataset.VerticalHeight - 1
		}
		if y0 <= t && t <= y1 {
			return y0, y1
		}
	}
	fmt.Fprintf(os.Stderr, "timestamp %d is out of range [%d..%d]\n",
		t, dataset.VerticalStart, dataset.VerticalStart+dataset.VerticalHeight-1)
	return -2, -1
}

func getPointingRangeX(dataset *DataSet) []int {
	var ret []int

	for i := pBegin; i <= pEnd; i += pStep {
		j := i + pStep - 1
		if j > pEnd {
			j = pEnd
		}
		ret = append(ret, i, j)
	}
	return ret
}

func addSemaphore(semaName string, defaultValue int) int {
	cmdText := fmt.Sprintf("scalebox semaphore create %s %d", semaName, defaultValue)
	// scalebox.ExecShellCommand(cmdText)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func countDown(semaName string) int {
	cmdText := fmt.Sprintf("scalebox semaphore countdown %s", semaName)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

type dataSet struct {
	// prefix ':' type ':' sub-id
	DatasetID string

	// for type "H", x-coord
	HorizontalWidth string

	// for type "V", y-coord
	VerticalStart  string
	VerticalHeight string
}

func parseDataSetX(t string) *DataSet {
	var ds dataSet
	if err := json.Unmarshal([]byte(t), &ds); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-dataset definition
		return nil
	}
	var dataset DataSet
	dataset.HorizontalWidth, _ = strconv.Atoi(ds.HorizontalWidth)
	dataset.VerticalStart, _ = strconv.Atoi(ds.VerticalStart)
	dataset.VerticalHeight, _ = strconv.Atoi(ds.VerticalHeight)
	return &dataset
}
