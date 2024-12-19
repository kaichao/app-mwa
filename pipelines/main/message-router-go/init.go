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

	// 计算节点的IP列表
	ips = []string{}
	// 计算节点的集群hostname
	hosts = []string{}
)

func init() {
	var err error

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
	fmt.Printf("host-num=%d\n", num)
	toHost := ips[num]
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, message)
	if len(headers) > 0 {
		h, err := json.Marshal(headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "headers:%v,JSON marshaling failed:%v\n", headers, err)
		} else {
			cmdTxt = fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s --headers '%s' %s", sinkJob, toHost, h, message)
		}
	}

	return misc.ExecCommandReturnExitCode(cmdTxt, 60)
}

func sendJobRefMessage(message string, headers map[string]string, sinkJob string) int {
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s %s", sinkJob, message)
	if len(headers) > 0 {
		h, err := json.Marshal(headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "headers:%v,JSON marshaling failed:%v\n", headers, err)
		} else {
			cmdTxt = fmt.Sprintf("scalebox task add --sink-job %s --headers '%s' %s", sinkJob, h, message)
		}
	}

	return misc.ExecCommandReturnExitCode(cmdTxt, 60)
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
	// fmt.Fprintln(os.Stderr, "sqlText:\n", sqlText)
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
	fmt.Println("ips:", ips)
}

// ExecWithRetries ...
func ExecWithRetries(cmd string, numRetries int, timeout int) (int, string, string) {
	delay := 10 * time.Second
	var (
		code           int
		stdout, stderr string
	)

	for i := 0; i < numRetries; i++ {
		code, stdout, stderr = misc.ExecCommandReturnAll(cmd, timeout)
		if code == 0 {
			return code, stdout, stderr
		}
		fmt.Printf("num-of-retries:%d,cmd=%s\n", i+1, cmd)
		time.Sleep(delay)
		delay *= 2
		timeout *= 2
	}
	return code, stdout, stderr
}

// AddTimeStamp ...
func AddTimeStamp(label string) {
	fileName := os.Getenv("WORK_DIR") + "/timestamps.txt"
	timeStamp := time.Now().Format("2006-01-02T15:04:05.000000Z07:00")
	// fmt.Printf("timestamp:%s\n", timeStamp)
	misc.AppendToFile(fileName, timeStamp+","+label)
}
