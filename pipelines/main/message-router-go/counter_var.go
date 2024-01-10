package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

var (
	counters = make(map[string]int)
	workDir  string

	db *sql.DB
)

func init() {
	// var err error

	workDir = os.Getenv("WORD_DIR")
	if workDir == "" {
		workDir = "/work"
	}
	// // set database connection
	// if db, err = sql.Open("sqlite3", workDir+"/.scalebox/sqlite.db"); err != nil {
	// 	logrus.Fatalln("Unable to open sqlite3 database:", err)
	// }
	// sqlText := `
	// 	CREATE TABLE IF NOT EXISTS t_counter (
	// 		id INTEGER PRIMARY KEY autoincrement,
	// 		uri_name TEXT,
	// 		value INT
	// 	);
	// 	CREATE UNIQUE INDEX IF NOT EXISTS i_count_0 ON t_counter(name,uri_name);
	// `

	// if _, err = db.Exec(sqlText); err != nil {
	// 	logrus.Errorln(err)
	// 	os.Exit(1)
	// }
}

func initCounters(dataset *DataSet) {
	begin, err := strconv.Atoi(os.Getenv("POINTING_BEGIN"))
	if err != nil || begin == 0 {
		begin = 1
	}
	end, err := strconv.Atoi(os.Getenv("POINTING_END"))
	if err != nil || end == 0 {
		end = 144
	}
	initValue := end - begin + 1

	arr := getRange(dataset)
	for ch := 109; ch <= 132; ch++ {
		for i := 0; i < len(arr); i += 2 {
			uri := fmt.Sprintf("remove-dat-file:%s/%d_%d/ch%d", dataset.DatasetID, arr[i], arr[i+1], ch)
			fmt.Printf("uri:%s,init-value:%d\n", uri, initValue)
			cmdText := fmt.Sprintf("scalebox latch create %s %d", uri, initValue)
			scalebox.ExecShellCommand(cmdText)
		}
	}
}

func getRange(dataset *DataSet) []int {
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
func addCounter(counterName string, defaultValue int) {
	counters[counterName] = defaultValue
}

func countDown(counterName string) int {
	cmdText := fmt.Sprintf("scalebox latch countdown %s", counterName)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func countDownN(counterName string, n int) int {
	m, ok := counters[counterName]
	if !ok {
		return -1
	}
	m -= n
	counters[counterName] = m
	return m
}
