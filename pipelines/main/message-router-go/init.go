package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	scalebox "github.com/kaichao/scalebox/golang/misc"
	"github.com/sirupsen/logrus"
)

var (
	db *sql.DB

	logger *logrus.Logger

	hosts = []string{"10.11.16.79", "10.11.16.76", "10.11.16.75"}
	// hosts            = []string{"10.11.16.79", "10.11.16.80", "10.11.16.76", "10.11.16.75"}
	// numNodesPerGroup int

	localMode bool

	workDir string
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

	localMode = os.Getenv("LOCAL_MODE") == "yes"

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
	if db, err = sql.Open("pgx", databaseURL); err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	db.SetConnMaxLifetime(500)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(20)
	db.Stats()
}

func sendNodeAwareMessage(message string, headers map[string]string, sinkJob string, num int) int {
	if !localMode {
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+message)
		return 0
	}

	toHost := hosts[num%len(hosts)]
	// cmdTxt := fmt.Sprintf("scalebox task add --upsert --sink-job %s --to-ip %s %s", sinkJob, toHost, message)
	cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toHost, message)
	if len(headers) > 0 {
		h, err := json.Marshal(headers)
		if err != nil {
			fmt.Fprintf(os.Stderr, "headers:%v,JSON marshaling failed:%v\n", headers, err)
		} else {
			// cmdTxt = fmt.Sprintf("scalebox task add --upsert --sink-job %s --to-ip %s --headers '%s' %s", sinkJob, toHost, h, message)
			cmdTxt = fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s --headers '%s' %s", sinkJob, toHost, h, message)
		}
	}

	fmt.Printf("cmd-text:%s\n", cmdTxt)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
	fmt.Printf("stdout for task-add:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
	return code
}
