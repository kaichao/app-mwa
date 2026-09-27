package main

import (
	"beamform/internal/vpath"
	"os"
	"strconv"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/global"
	"github.com/kaichao/scalebox/pkg/module"
	"github.com/kaichao/scalebox/pkg/variable"
)

var (
	appID int
	vPath *vpath.Engine
)

// initApp 初始化应用级依赖：APP_ID 与 /vpath.yaml。
// 必须由 main() 在 module.InitLogging() 之后、module.Run() 之前显式调用——
// 放在 init() 里会早于日志初始化，module.LogEntry 尚为 nil。
func initApp() {
	appID, _ = strconv.Atoi(os.Getenv("APP_ID"))

	var err error
	vPath, err = vpath.Load("/vpath.yaml", appID)
	if err != nil {
		logger.LogError(err, module.LogEntry)
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
