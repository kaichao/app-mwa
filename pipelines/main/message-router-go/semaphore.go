package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/exec"
	"github.com/kaichao/scalebox/pkg/postgres"
	"github.com/sirupsen/logrus"
)

func createSemaphore(semaName string, defaultValue int) int {
	cmdText := fmt.Sprintf("scalebox semaphore create %s %d", semaName, defaultValue)
	code, _ := exec.ExecCommandReturnExitCode(cmdText, 15)
	return code
}

func countDown(semaName string) int {
	cmdText := fmt.Sprintf("scalebox semaphore decrement %s", semaName)
	stdout, _ := exec.ExecCommandReturnStdout(cmdText, 15)
	if stdout == "" {
		return math.MinInt
	}
	if val, err := strconv.Atoi(strings.TrimSpace(stdout)); err != nil {
		fmt.Fprintf(os.Stderr, "err to convert semaphore value:%s\n%v\n", stdout, err)
	} else {
		return val
	}

	return math.MinInt
}

func getSemaphore(semaName string) int {
	cmdText := fmt.Sprintf("scalebox semaphore get %s", semaName)
	stdout, _ := exec.ExecCommandReturnStdout(cmdText, 15)
	if stdout == "" {
		return math.MinInt
	}
	if val, err := strconv.Atoi(strings.TrimSpace(stdout)); err == nil {
		fmt.Fprintf(os.Stderr, "err to convert semaphore value:\n%v\n", err)
	} else {
		return val
	}

	return math.MinInt
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
	tx, err := postgres.GetDB().Begin()
	if err != nil {
		logrus.Errorf("err:%v\n", err)
	}
	defer tx.Rollback()

	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	sqlText := `
		INSERT INTO t_semaphore(name,value,value0,app)
		SELECT $1,$2,$2,app FROM t_job WHERE id=$3
		ON CONFLICT (name,app) DO UPDATE SET (value,value0) = ($2,$2)
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
		if tx, err = postgres.GetDB().Begin(); err != nil {
			logrus.Errorf("err:%v\n", err)
		}
	}
}

var (
	// used for semaphore batch-insert
	batchInsert bool
)

func init() {
	batchInsert = os.Getenv("BATCH_INSERT") == "yes"
}
