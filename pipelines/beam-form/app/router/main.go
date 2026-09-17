package main

import (
	"os"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/module"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// 启动期日志必须先于 initApp()：initApp 出错时要用 module.LogEntry 上报。
	// module.Run 内部会再配置一次，两次调用幂等。
	module.InitLogging()
	initApp()
	common.AddTimeStamp("before-mr")
	os.Exit(module.Run(fromFuncs, "from_module"))
}

var (
	fromFuncs = map[string]module.HandlerFunc{
		"":            fromNull,
		"tar-load":    fromTarLoad,
		"wait-queue":  fromWaitQueue,
		"vtask-head":  fromVtaskHead,
		"pull-unpack": fromPullUnpack,
		"beam-make":   fromBeamMake,
		"down-sample": fromDownSample,
		"fits-redist": fromFitsRedist,
		// 可以在波束合成节点、或脉冲星搜索节点上运行
		"fits-merge": fromFitsMerge,
		// 在波束合成节点做fits-merge，将24ch移动到计算节点上
		"fits24ch-move":   fromFits24chMove,
		"vtask-tail":      fromVtaskTail,
		"fits24ch-unload": fromFits24chUnload,
	}
)
