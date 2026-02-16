/*
fromNull是消息路由的首个执行模块，按预加载策略，加载原始数据。
*/
package main

import (
	"beamform/app/router/iopath"
	"beamform/internal/datacube"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromNull(body string, headers map[string]string) int {
	// 按需创建信号量pointing-done，用于标识pointing的完成（？）
	// cube := datacube.NewDataCube(body)
	// cube0 := datacube.NewDataCube(cube.ObsID)
	// if true || reflect.DeepEqual(cube.GetTimeRanges(), cube0.GetTimeRanges()) {
	// 	// 输入数据集包含全时段
	// 	lines := []string{}
	// 	size := len(cube.GetTimeRanges()) / 2
	// 	for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
	// 		line := fmt.Sprintf(`"pointing-done:%s/p%05d":%d`, cube.ObsID, p, size)
	// 		lines = append(lines, line)
	// 	}
	// 	err := semaphore.CreateSemaphores(lines, 0, appID, 500)
	// 	if err != nil {
	// 		logrus.Errorf("create semaphore pointing-done, err-info:%v", err)
	// 		return 1
	// 	}
	// }

	// 生成每个指向数据的独立root目录路径
	cube := datacube.NewDataCube(body)
	for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
		pointingDir := fmt.Sprintf("%s/p%05d", cube.ObsID, p)
		varName := "pointing-data-root:" + pointingDir
		if v, err := getPointingVariable(varName, appID); err != nil || v == "" {
			varValue, err := iopath.GetStagingRoot(pointingDir)
			if err != nil {
				logger.LogTracedErrorDefault(err)
				return 9
			}
			setPointingVariable(varName, varValue, appID)
		}
	}

	if !strings.HasPrefix(body, "/") {
		return toTarLoad(body)
	}

	// 不需preload，产生所有cube到wait-queue
	fmtCubicName := "%s/p%05d_%05d/t%d_%d"
	trs := cube.GetTimeRanges()
	for i := 0; i < len(trs); i += 2 {
		cubeName := fmt.Sprintf(fmtCubicName, cube.ObsID,
			cube.PointingBegin, cube.PointingEnd, trs[i], trs[i+1])
		if ret := toWaitQueue(cubeName); ret != 0 {
			return ret
		}
	}
	return 0
}

// toCrossAppPresto()
func toCrossAppPresto(pointing string) int {
	varName := "pointing-data-root:" + pointing
	varValue, err := getPointingVariable(varName, appID)
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
		"SINK_MODULE": "message-router-presto",
		"MODULE_ID":   "",
		"APP_ID":      fmt.Sprintf("%d", prestoAppID),
	}
	fmt.Printf("In toCrossAppPresto(), env:APP_ID=%s, MODULE_ID=%s, SINK_MODULE=%s,GRPC_SERVER=%s\n",
		envVars["APP_ID"], envVars["MODULE_ID"], envVars["SINK_MODULE"], os.Getenv("GRPC_SERVER"))

	_, err = task.AddWithMapHeaders(pointing, headers, envVars)
	if err != nil {
		logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 1
	}
	return 0
}
