package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"
	scalebox "github.com/kaichao/scalebox/golang/misc"
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

	db *sql.DB
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
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}

func initHosts() {
	sqlText := `
		SELECT hostname,ip_addr
		FROM t_host
		WHERE cluster=$1 AND hostname LIKE 'n-%'
		ORDER BY 1
		LIMIT $2
	`

	clustName := os.Getenv("CLUSTER")
	numOfNodes, _ := strconv.Atoi(os.Getenv("NUM_OF_NODES"))
	fmt.Printf("num-of-nodes:%d in cluster %s\n", numOfNodes, clustName)
	rows, err := getDB().Query(sqlText, clustName, numOfNodes)
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

func getDB() *sql.DB {
	if db == nil {
		dbHost := os.Getenv("PGHOST")
		if dbHost == "" {
			dbHost = scalebox.GetLocalIP()
		}
		dbPort := os.Getenv("PGPORT")
		if dbPort == "" {
			dbPort = "5432"
		}
		databaseURL := fmt.Sprintf("postgres://scalebox:changeme@%s:%s/scalebox", dbHost, dbPort)
		// set database connection
		var err error
		if db, err = sql.Open("pgx", databaseURL); err != nil {
			log.Fatal("Unable to connect to database:", err)
		}
		db.SetConnMaxLifetime(500)
		db.SetMaxIdleConns(50)
		db.SetMaxOpenConns(20)
		// db.Stats()
	}
	return db
}
