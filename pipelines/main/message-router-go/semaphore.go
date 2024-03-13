package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

func createDatReadySemaphores(datacube *DataCube) {
	// TARGET: beam-maker
	// 1257010784/1257010786_1257010815/112/00001_00024
	// 1257010784/1257010786_1257010815/112
	arr := datacube.getTimeRanges()
	for i := 0; i < len(arr); i += 2 {
		// all dat files in current range
		initValue := arr[i+1] - arr[i] + 1
		for ch := 109; ch <= 132; ch++ {
			sema := fmt.Sprintf("dat-ready:%s/t%d_%d/ch%d",
				datacube.DatasetID, arr[i], arr[i+1], ch)
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			addSemaphore(sema, initValue)
		}
	}
}

func createFits24chReadySemaphores(datacube *DataCube) {
	// TARGET: fits-merger
	// 1257010784/p00024/t1257010786_1257010815
	// 24-channel
	initValue := 24

	arr := datacube.getTimeRanges()

	for p := datacube.PointingBegin; p <= datacube.PointingEnd; p++ {
		for i := 0; i < len(arr); i += 2 {
			sema := fmt.Sprintf("fits-24ch-ready:%s/p%05d/t%d_%d", datacube.DatasetID, p, arr[i], arr[i+1])
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			addSemaphore(sema, initValue)
		}
	}
}

func createDatProcessedSemaphores(datacube *DataCube) {
	// all pointing
	initValue := datacube.PointingEnd - datacube.PointingBegin + 1

	arr := datacube.getTimeRanges()
	for i := 0; i < len(arr); i += 2 {
		for ch := 109; ch <= 132; ch++ {
			sema := fmt.Sprintf("dat-processed:%s/t%d_%d/ch%d", datacube.DatasetID, arr[i], arr[i+1], ch)
			fmt.Printf("sema:%s,init-value:%d\n", sema, initValue)
			addSemaphore(sema, initValue)
		}
	}
}

func addSemaphore(semaName string, defaultValue int) int {
	cmdText := fmt.Sprintf("scalebox semaphore create %s %d", semaName, defaultValue)
	// scalebox.ExecShellCommand(cmdText)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 15)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func countDown(semaName string) int {
	cmdText := fmt.Sprintf("scalebox semaphore countdown %s", semaName)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 15)
	fmt.Printf("exit-code for semaphore countdown:\n%d\n", code)
	fmt.Printf("stdout for semaphore countdown:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for semaphore countdown:\n%s\n", stderr)
	if code > 0 {
		return -1
	}
	code, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		fmt.Fprintf(os.Stderr, "stderr for convert to code in semaphore countdown:\n%v\n", err)
		return -2
	}

	return code
}
