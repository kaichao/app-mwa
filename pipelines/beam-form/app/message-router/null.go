/*
fromNull是消息路由的首个执行模块，按PRELOAD_MODE指定的预加载策略，加载原始数据。
*/
package main

import (
	"beamform/internal/datacube"
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

func fromNull(body string, headers map[string]string) int {
	cube := datacube.NewDataCube(body)
	cube0 := datacube.NewDataCube(cube.ObsID)
	if reflect.DeepEqual(cube.GetTimeRanges(), cube0.GetTimeRanges()) {
		// 输入数据集包含全时段，创建信号量pointing-ready
		fileName := "my-sema.txt"
		size := len(cube.GetTimeRanges()) / 2
		for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
			common.AppendToFile(fileName, fmt.Sprintf(`"pointing-ready:%s/p%05d",%d`, cube.ObsID, p, size)+"\n")
		}
		err := semaphore.CreateFileSemaphores(fileName, appID, 500)
		if err != nil {
			logrus.Errorf("create semaphore pointing-ready, err-info:%v", err)
			return 1
		}
	}

	if os.Getenv("PRELOAD_MODE") != "none" {
		return toTarLoad(body)
	}

	// 产生所有cube
	fmtCubicID := "%s/p%05d_%05d/t%d_%d"
	trs := cube.GetTimeRanges()
	for i := 0; i < len(trs); i += 2 {
		cubeID := fmt.Sprintf(fmtCubicID, cube.ObsID,
			cube.PointingBegin, cube.PointingEnd, trs[i], trs[i+1])
		if ret := toCubeVtask(cubeID); ret != 0 {
			return ret
		}
	}
	return 0
}

// toCrossApp()
func doCrossAppTaskAdd(pointing string) int {
	// 信号量pointing-done的操作
	// semaphore: pointing-done:1257010784/p00001
	sema := "pointing-done:" + pointing
	v, err := semaphore.AddValue(sema, appID, -1)
	if err != nil {
		logrus.Errorf("error while decrement semaphore,sema=%s, err:%v\n",
			sema, err)
		return 1
	}
	semaVal, _ := strconv.Atoi(v)
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	varName := "pointing-data-root:" + pointing
	varValue, err := variable.Get(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}

	prestoAppID, err := strconv.Atoi(os.Getenv("PRESTO_APP_ID"))
	if err != nil {
		logrus.Errorln("no valid PRESTO_APP_ID")
		return 12
	}
	// IPv4地址（类型1）， 设置"to_ip"头
	headers := common.SetJSONAttribute("{}", "source_url", varValue)
	// 给presto-search流水线发消息
	envVars := map[string]string{
		"SINK_JOB": "message-router-presto",
		"JOB_ID":   "",
		"APP_ID":   fmt.Sprintf("%d", prestoAppID),
	}
	fmt.Printf("In doCrossAppTaskAdd(), env:APP_ID=%s, JOB_ID=%s, SINK_JOB=%s,GRPC_SERVER=%s\n",
		envVars["APP_ID"], envVars["JOB_ID"], envVars["SINK_JOB"], os.Getenv("GRPC_SERVER"))
	return task.Add(pointing, headers, envVars)
}
