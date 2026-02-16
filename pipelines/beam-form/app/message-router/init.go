package main

import (
	"beamform/internal/cache"
	"os"
	"strconv"

	"github.com/kaichao/scalebox/pkg/global"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

var (
	appID int
)

func init() {
	level, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetReportCaller(true)
}

func init() {
	os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))

	moduleID, _ := strconv.Atoi(os.Getenv("MODULE_ID"))
	appID = cache.GetAppIDByModuleID(moduleID)
}

func getPointingVariable(varName string, appID int) (string, error) {
	if os.Getenv("USE_GLOBAL_POINTING") == "yes" {
		return global.Get(varName)
	}
	return variable.GetValue(varName, 0, appID)
}

func setPointingVariable(varName string, varValue string, appID int) error {
	if os.Getenv("USE_GLOBAL_POINTING") == "yes" {
		return global.Set(varName, varValue)
	}
	return variable.Set(varName, varValue, 0, appID)
}
