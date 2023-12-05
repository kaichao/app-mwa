package main

import (
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

var (
	logger *logrus.Logger

	dataRootMain    string
	dataRootQiu     string
	rsyncPrefixMain string
	rsyncPrefixQiu  string

	// 每n个fits文件合并为1个fil文件
	numPerGroup int
	// 每次观测最大序列号
	maxSequence int
)

func init() {
	var (
		err error
	)
	logger = logrus.New()
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.WarnLevel
	}
	logger.SetLevel(level)
	logger.SetReportCaller(true)
}
