package main

import (
	"beamform/internal/vpath"
	"os"
	"strconv"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/global"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

var (
	appID int
	vPath *vpath.VirtualPath

	logEntry *logrus.Entry
)

func init() {
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
	formatter := &logrus.TextFormatter{
		DisableQuote: true,
	}
	logrus.SetFormatter(formatter)

	// 配置logger
	log := logrus.New()
	log.SetLevel(level)
	if level >= logrus.DebugLevel {
		// debug / trace
		log.SetFormatter(formatter)
	} else {
		log.SetFormatter(&logrus.JSONFormatter{})
	}
	logEntry = logrus.NewEntry(log)
}

func init() {
	os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))

	appID, _ = strconv.Atoi(os.Getenv("APP_ID"))

	var err error
	vPath, err = vpath.NewVirtualPath(appID, "/vpath.yaml")
	if err != nil {
		logger.LogError(err, logEntry)
	}
}

func getPointingVariable(varName string, appID int) (string, error) {
	if os.Getenv("USE_GLOBAL_POINTING") == "yes" {
		return global.Get(varName)
	}
	return variable.GetValue(varName, appID)
}

func setPointingVariable(varName string, varValue string, appID int) error {
	if os.Getenv("USE_GLOBAL_POINTING") == "yes" {
		return global.Set(varName, varValue)
	}
	return variable.Set(varName, varValue, appID)
}
