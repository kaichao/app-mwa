package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

var (
	// counters = make(map[string]int)
	workDir string

	db *sql.DB
)

func init() {
	// var err error

	workDir = os.Getenv("WORD_DIR")
	if workDir == "" {
		workDir = "/work"
	}
}

func createDatReadySemaphores(dataset *DataSet) {
}

func createFits1chReadySemaphores(dataset *DataSet) {
}

func createDatUsedSemaphores(dataset *DataSet) {
	begin, err := strconv.Atoi(os.Getenv("POINTING_BEGIN"))
	if err != nil || begin == 0 {
		begin = 1
	}
	end, err := strconv.Atoi(os.Getenv("POINTING_END"))
	if err != nil || end == 0 {
		end = 144
	}
	initValue := end - begin + 1

	arr := getTimeRange(dataset)
	for ch := 109; ch <= 132; ch++ {
		for i := 0; i < len(arr); i += 2 {
			sema := fmt.Sprintf("dat-used:%s/%d_%d/ch%d", dataset.DatasetID, arr[i], arr[i+1], ch)
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			// cmdText := fmt.Sprintf("scalebox semaphore create %s %d", uri, initValue)
			// scalebox.ExecShellCommand(cmdText)
			addSemaphore(sema, initValue)
		}
	}
}

func getTimeRange(dataset *DataSet) []int {
	var ret []int

	step, err := strconv.Atoi(os.Getenv("NUM_SECONDS_PER_CALC"))
	if err != nil || step == 0 {
		step = 30
	}
	for y := 0; y < dataset.VerticalHeight; y += step {
		y0 := dataset.VerticalStart + y
		y1 := y0 + step - 1
		if y1 > dataset.VerticalStart+dataset.VerticalHeight-1 {
			y1 = dataset.VerticalStart + dataset.VerticalHeight - 1
		}
		ret = append(ret, y0, y1)
	}
	return ret
}

func getPointingRangeX(dataset *DataSet) []int {
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

	var ret []int

	for i := begin; i <= end; i += step {
		j := i + step - 1
		if j > end {
			j = end
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
