package main

import (
	"os"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/module"
	"github.com/sirupsen/logrus"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	module.InitLogging()
	logrus.Infoln("00, Entering main-router")
	// 参数数量校验由 module.Run 完成，这里只为保留原有的 01 日志；
	// 条件与 module.Run 的校验一致，不足 3 个参数时不打印。
	if len(os.Args) >= 3 {
		logrus.Infof("01, after number of arguments verification, task-body:%s,task-header:%s.\n",
			os.Args[1], os.Args[2])
	}
	os.Exit(module.Run(fromFuncs, "from_module"))
}

// wrapHandler 补齐 module.Run 未提供的扩展点：headers 解析成功与路由命中
// 的 Info 日志、以及 before-mr / before-exit 两个时间戳采样。
// module.Run 在参数不足、headers 非法、handler 未命中时提前返回，不会走到这里，
// 与改造前这三条路径不打日志、不打时间戳的行为一致。
func wrapHandler(f module.HandlerFunc) module.HandlerFunc {
	return func(body string, headers map[string]string) error {
		logrus.Infoln("02, after JSON format verification of headers")
		logrus.Infoln("03, main-router not null")
		common.AddTimeStamp("before-mr")

		err := f(body, headers)

		common.AddTimeStamp("before-exit")
		return err
	}
}

var (
	fromFuncs = map[string]module.HandlerFunc{
		"": wrapHandler(defaultFunc),
		// "main-router": wrapHandler(fromMessageRouter),
		"down-sample": wrapHandler(fromDownSample),
		"fits-merge":  wrapHandler(fromFitsMerge),
	}
)
