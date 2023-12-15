package main

import (
	"os"
	"strconv"

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

	hosts               = []string{"10.11.16.79", "10.11.16.80"}
	numNodesPerGroup    int
	numPointingsPerCalc int

	localMode bool

	datasetFile = "/work/.scalebox/dataset-v.txt"
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

	if numNodesPerGroup, err = strconv.Atoi(os.Getenv("NUM_NODES_PER_GROUP")); err != nil {
		numNodesPerGroup = 24
	}
	if numPointingsPerCalc, err = strconv.Atoi(os.Getenv("NUM_POINTINGS_PER_CALC")); err != nil {
		numPointingsPerCalc = 24
	}

	localMode = os.Getenv("LOCAL_MODE") == "yes"
}
