package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
	"github.com/sirupsen/logrus"
)

func (cube *DataCube) createDatReadySemaphores() {
	// TARGET: beam-maker
	// 1257010784/1257010786_1257010815/112/00001_00024
	// 1257010784/1257010786_1257010815/112
	ts := cube.getTimeRanges()
	semaArr := []Sema{}
	for i := 0; i < len(ts); i += 2 {
		// all dat files in current range
		initValue := ts[i+1] - ts[i] + 1
		for ch := 109; ch <= 132; ch++ {
			sema := fmt.Sprintf("dat-ready:%s/t%d_%d/ch%d",
				cube.DatasetID, ts[i], ts[i+1], ch)
			fmt.Printf("In createDatReadySemaphores(),sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}
func (cube *DataCube) createPointingBatchLeftSemaphores() {
	// pointing-batch-left:1257010784/t1257010786_1257010845/ch119
	initValue := cube.getNumOfPointingBatch()
	semaArr := []Sema{}

	ts := cube.getTimeRangesWithinInterval(cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
	for i := 0; i < len(ts); i += 2 {
		// all dat files in current range
		for ch := 109; ch <= 132; ch++ {
			sema := fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%d",
				cube.DatasetID, ts[i], ts[i+1], ch)
			fmt.Printf("In createPointingBatchLeftSemaphores(), sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}
func (cube *DataCube) createFits24chReadySemaphores() {
	// TARGET: fits-merger
	// 1257010784/p00024/t1257010786_1257010815
	// 24-channel
	initValue := 24

	ts := cube.getTimeRanges()
	semaArr := []Sema{}

	for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
		for i := 0; i < len(ts); i += 2 {
			sema := fmt.Sprintf("fits-24ch-ready:%s/p%05d/t%d_%d",
				cube.DatasetID, p, ts[i], ts[i+1])
			fmt.Printf("In createFits24chReadySemaphores(), sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}

func (cube *DataCube) createDatProcessedSemaphores() {
	// dat-processed:1257010784/p00001_00096/t1257010846_1257010905/ch111
	// first batch
	ts := cube.getTimeRanges()
	semaArr := []Sema{}

	fmt.Printf("cube:%v\n", cube)

	for i := 0; i < len(ts); i += 2 {
		for ch := 109; ch <= 132; ch++ {
			for pIndex := 0; pIndex < cube.getNumOfPointingBatch(); pIndex++ {
				p0, p1 := cube.getPointingBatchRange(cube.PointingBegin + pIndex*cube.PointingStep*cube.NumPerBatch)

				fmt.Printf("p-index:%d,p0=%d,p1=%d\n", pIndex, p0, p1)
				sema := fmt.Sprintf("dat-processed:%s/p%05d_%05d/t%d_%d/ch%d",
					cube.DatasetID, p0, p1, ts[i], ts[i+1], ch)
				fmt.Printf("In createDatProcessedSemaphores(), sema:%s,init-value:%d\n", sema, p1-p0+1)
				semaArr = append(semaArr, Sema{name: sema, value: p1 - p0 + 1})
			}
		}
	}
	doInsert(semaArr)
}

func (cube *DataCube) getSemaPointingBatchIndex(t int, ch int) int {
	return doPointingBatchIndex(cube, t, ch, getSemaphore)
}

func (cube *DataCube) countDownSemaPointingBatchIndex(t int, ch int) int {
	return doPointingBatchIndex(cube, t, ch, countDown)
}

func doPointingBatchIndex(cube *DataCube, t int, ch int, op func(string) int) int {
	t0, t1 := cube.getTimeRange(t)
	sema := fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%d",
		cube.DatasetID, t0, t1, ch)
	n := op(sema)
	index := cube.getNumOfPointingBatch() - n
	fmt.Printf("In doGetPointingBatchIndex(), sema:%s, num-of-batch=%d, op(sema)=%d,index=%d \n",
		sema, cube.getNumOfPointingBatch(), n, index)
	return index
}

func createSemaphore(semaName string, defaultValue int) int {
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
		return -10
	}
	code, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		fmt.Fprintf(os.Stderr, "stderr for convert to code in semaphore countdown:\n%v\n", err)
		return -11
	}

	return code
}

func getSemaphore(semaName string) int {
	cmdText := fmt.Sprintf("scalebox semaphore get %s", semaName)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 15)
	fmt.Printf("exit-code for semaphore get:\n%d\n", code)
	fmt.Printf("stdout for semaphore get:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for semaphore get:\n%s\n", stderr)
	if code > 0 {
		return -10
	}
	code, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		fmt.Fprintf(os.Stderr, "stderr for convert to code in semaphore get:\n%v\n", err)
		return -11
	}

	return code
}

// Sema ...
type Sema struct {
	name  string
	value int
}

func doInsert(values []Sema) {
	if !batchInsert {
		for _, sema := range values {
			createSemaphore(sema.name, sema.value)
		}
		return
	}
	// start transaction
	tx, err := db.Begin()
	if err != nil {
		logrus.Errorf("err:%v\n", err)
	}
	defer tx.Rollback()

	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	sqlText := `
		INSERT INTO t_semaphore(name,value,value0,job)
		VALUES($1,$2,$2,$3)
		ON CONFLICT (name,job) DO UPDATE SET (value,value0) = ($2,$2)
	`

	batchSize := 100
	for i := 0; i < len(values); i += batchSize {
		stmt, err := tx.Prepare(sqlText)
		if err != nil {
			logrus.Errorf("err:%v\n", err)
		}
		defer stmt.Close()

		end := i + batchSize
		if end > len(values) {
			end = len(values)
		}

		for _, v := range values[i:end] {
			if _, err := stmt.Exec(v.name, v.value, jobID); err != nil {
				logrus.Errorf("err:%v\n", err)
			}
		}
		if err = tx.Commit(); err != nil {
			logrus.Errorf("err:%v\n", err)
		}

		fmt.Printf("[%d..%d], %d row(s) inserted.\n", i, end, end-i)

		// start next batch
		if tx, err = db.Begin(); err != nil {
			logrus.Errorf("err:%v\n", err)
		}
	}
}
