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
		// 输入数据集包含全时段，创建信号量pointing-done
		fileName := "my-sema.txt"
		size := len(cube.GetTimeRanges()) / 2
		for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
			common.AppendToFile(fileName, fmt.Sprintf(`"pointing-done:%s/p%05d":%d`, cube.ObsID, p, size))
		}
		err := semaphore.CreateFileSemaphores(fileName, appID, 500)
		if err != nil {
			logrus.Errorf("create semaphore pointing-done, err-info:%v", err)
			return 1
		}
	}

	if isPreloadMode() {
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

// toCrossAppPresto()
func toCrossAppPresto(pointing string) int {
	varName := "pointing-data-root:" + pointing
	varValue, err := variable.Get(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get,name:%s, err-info:%v\n", varName, err)
		return 11
	}

	prestoAppID, err := strconv.Atoi(os.Getenv("PRESTO_APP_ID"))
	if err != nil {
		logrus.Errorln("No valid PRESTO_APP_ID")
		return 12
	}
	// IPv4地址（类型1）， 设置"to_ip"头
	headers := map[string]string{
		"source_url": varValue,
	}
	// 给presto-search流水线发消息
	envVars := map[string]string{
		"SINK_JOB": "message-router-presto",
		"JOB_ID":   "",
		"APP_ID":   fmt.Sprintf("%d", prestoAppID),
	}
	fmt.Printf("In toCrossAppPresto(), env:APP_ID=%s, JOB_ID=%s, SINK_JOB=%s,GRPC_SERVER=%s\n",
		envVars["APP_ID"], envVars["JOB_ID"], envVars["SINK_JOB"], os.Getenv("GRPC_SERVER"))
	return task.AddWithMapHeaders(pointing, headers, envVars)
}
