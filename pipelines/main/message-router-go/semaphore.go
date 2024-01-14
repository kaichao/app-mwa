package main

import (
	"fmt"
	"os"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

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

// func getPointingRangeX(dataset *DataSet) []int {
// 	var ret []int

// 	for i := pBegin; i <= pEnd; i += pStep {
// 		j := i + pStep - 1
// 		if j > pEnd {
// 			j = pEnd
// 		}
// 		ret = append(ret, i, j)
// 	}
// 	return ret
// }

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
