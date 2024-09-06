package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger

	ips = []string{}
	// ips = []string{"10.11.16.79", "10.11.16.76", "10.11.16.75"}
	// ips            = []string{"10.11.16.79", "10.11.16.80", "10.11.16.76", "10.11.16.75"}
	// hosts = []string{"n0.dcu", "n1.dcu", "n2.dcu", "n3.dcu"}
	hosts = []string{}

	workDir string

	// db *sql.DB
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

	initHosts()
}

func sendNodeAwareMessage(message string, headers map[string]string, sinkJob string, num int) int {
	toHost := ips[num%len(ips)]
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, message)
	if len(headers) > 0 {
		h, err := json.Marshal(headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "headers:%v,JSON marshaling failed:%v\n", headers, err)
		} else {
			cmdTxt = fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s --headers '%s' %s", sinkJob, toHost, h, message)
		}
	}

	fmt.Printf("cmd-text for task-add:%s\n", cmdTxt)
	code, stdout, stderr := misc.ExecShellCommandWithExitCode(cmdTxt, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func sendJobRefMessage(message string, headers map[string]string, sinkJob string) int {
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s  %s", sinkJob, message)
	if len(headers) > 0 {
		h, err := json.Marshal(headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "headers:%v,JSON marshaling failed:%v\n", headers, err)
		} else {
			cmdTxt = fmt.Sprintf("scalebox task add --sink-job %s --headers '%s' %s", sinkJob, h, message)
		}
	}

	fmt.Printf("cmd-text for task-add:%s\n", cmdTxt)
	code, stdout, stderr := misc.ExecShellCommandWithExitCode(cmdTxt, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func initHosts() {
	// 计算节点以c-开始
	sqlFmt := `
		SELECT hostname,ip_addr
		FROM t_host
		WHERE cluster=$1 AND hostname LIKE '%v-%%'
		ORDER BY 1
		LIMIT $2
	`
	prefix := strings.Split(os.Getenv("NODES"), "-")[0]
	sqlText := fmt.Sprintf(sqlFmt, prefix)
	fmt.Fprintln(os.Stderr, "sqlText:\n", sqlText)
	clustName := os.Getenv("CLUSTER")
	numOfNodes, _ := strconv.Atoi(os.Getenv("NUM_OF_NODES"))
	fmt.Printf("num-of-nodes:%d in cluster %s\n", numOfNodes, clustName)
	rows, err := misc.GetDB().Query(sqlText, clustName, numOfNodes)
	defer rows.Close()
	if err != nil {
		logrus.Errorf("query t_host error: %v\n", err)
	}

	var hostname, ipAddr string
	for rows.Next() {
		err := rows.Scan(&hostname, &ipAddr)
		if err == nil {
			hosts = append(hosts, hostname)
			ips = append(ips, ipAddr)
		} else {
			logrus.Errorln(err)
		}
	}
}

// ExecWithRetries ...
func ExecWithRetries(cmd string, numRetries int) (int, string, string) {
	delay := 30 * time.Second
	var (
		code           int
		stdout, stderr string
	)

	for i := 0; i < numRetries; i++ {
		code, stdout, stderr = misc.ExecShellCommandWithExitCode(cmd, -1)
		if code == 0 {
			return code, stdout, stderr
		}
		fmt.Printf("num-of-retries:%d,cmd=%s\n", i+1, cmd)
		time.Sleep(delay)
		delay *= 2
	}
	return code, stdout, stderr
}
