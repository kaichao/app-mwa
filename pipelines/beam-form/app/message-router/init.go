package main

import (
	"beamform/internal/cache"
	"os"
	"strconv"

	"github.com/kaichao/scalebox/pkg/global"
	"github.com/kaichao/scalebox/pkg/variable"
)

var (
	appID int
)

func init() {
	os.Setenv("REDIS_HOST", os.Getenv("GRPC_SERVER"))

	moduleID, _ := strconv.Atoi(os.Getenv("MODULE_ID"))
	appID = cache.GetAppIDByModuleID(moduleID)
}

func getPointingVariable(varName string, appID int) (string, error) {
	if os.Getenv("USE_GLOBAL_POINTING") == "yes" {
		return global.Get(varName)
	}
	return variable.Get(varName, appID)
}

func setPointingVariable(varName string, varValue string, appID int) error {
	if os.Getenv("USE_GLOBAL_POINTING") == "yes" {
		return global.Set(varName, varValue)
	}
	return variable.Set(varName, varValue, appID)
}

func addTasks(sinkModule string, bodies []string) error {
	return nil
}
