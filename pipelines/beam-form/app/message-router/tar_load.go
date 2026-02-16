/*
tar-load从外部存储预加载原始打包tar文件到HPC存储
# 模块介绍

## tar加载顺序
- 以time-range为顺序
- 直接用基于task-id为加载顺序，不使用独立sort_tag

# 消息路由中代码说明

## fromTarLoad
- 处理tar-load返回的消息

## toTarLoad
- 给tar-load发送消息的逻辑

*/

package main

import (
	"beamform/app/message-router/iopath"
	"beamform/internal/datacube"
	"fmt"
	"os"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
)

// body: 1267459410_1267459449_ch109.dat.tar.zst
// headers:
//   - _cube_name: 1257010784/p00001_00960/t1257012766_1257012965
func fromTarLoad(body string, headers map[string]string) int {
	cubeName := headers["_cube_name"]
	semaName := "tar-ready:" + cubeName
	n, err := semaphore.AddValue(semaName, 0, appID, -1)
	if err != nil {
		logger.LogTracedErrorDefault(err)
		// logrus.Errorf("semaphore-decrement, name=%s,err-info:%v\n", semaName, err)
		return 1
	}
	if n <= 0 {
		// 若支持分组级slot，则发给pull-unpack
		// 否则发给wait-queue
		return toWaitQueue(cubeName)
	}
	return 0
}

// toTarLoad
// - 产生tar-load的消息
// - 初始化信号量：tar-ready
// - 初始化共享变量：cube-stor-index, cube对应的主机编号(30..80)
//
// datasetID
// - 1257010784
// - 1257010784/p00001_00960
// - 1257010784/p_00960
// - 1257010784/p00001_
// - 1257010784/p00001_00960/t1257012766_1257012965
// - 1257010784/p00001_00960/t1257012766_
// - 1257010784/p00001_00960/t_1257012965
func toTarLoad(datasetID string) int {
	// 按顺序产生tar-load的任务
	cube := datacube.NewDataCube(datasetID)
	sourceURL := fmt.Sprintf("%s/mwa/tar/%s", iopath.GetOriginRoot(), cube.ObsID)
	fmtTarZst := `%d_%d_ch%d.dat.tar.zst`
	taskLines := []string{}
	// tar-ready信号量
	semaLines := []string{}

	trs := cube.GetTimeRanges()
	for i := 0; i < len(trs); i += 2 {
		tus := cube.GetTimeUnitsWithinInterval(trs[i], trs[i+1])
		cubeName := fmt.Sprintf("%s/p%05d_%05d/t%d_%d",
			cube.ObsID, cube.PointingBegin, cube.PointingEnd, trs[i], trs[i+1])
		semaName := "tar-ready:" + cubeName
		semaValue := len(tus) / 2 * cube.NumOfChannels
		semaLines = append(semaLines, fmt.Sprintf(`"%s":%d`, semaName, semaValue))
		for k := 0; k < len(tus); k += 2 {
			for j := 0; j < cube.NumOfChannels; j++ {
				ch := cube.ChannelBegin + j
				fileName := fmt.Sprintf(fmtTarZst, tus[k], tus[k+1], ch)
				root, err := iopath.GetPreloadRoot(cube.ObsID + "/" + fileName)
				if err != nil {
					// logrus.Errorf("error:%T,%v\n", err, err)
					logger.LogTracedErrorDefault(err)
					return 1
				}
				targetURL := fmt.Sprintf("%s/tar/%s", root, cube.ObsID)
				body := fmt.Sprintf(`%s,{"target_url":"%s","_cube_name":"%s"}`,
					fileName, targetURL, cubeName)
				taskLines = append(taskLines, body)
			}
		}
	}

	// 信号量重置，使得可以多次重新加载打包文件
	os.Setenv("CONFLICT_ACTION", "OVERWRITE")
	defer func() {
		os.Unsetenv("CONFLICT_ACTION")
	}()
	if err := semaphore.CreateSemaphores(semaLines, 0, appID, 100); err != nil {
		logger.LogTracedErrorDefault(err)
		return 2
	}

	headers := map[string]string{
		"source_url": sourceURL,
	}
	envs := map[string]string{
		"SINK_MODULE":     "tar-load",
		"CONFLICT_ACTION": "OVERWRITE",
	}

	if _, err := task.AddTasksWithMapHeaders(taskLines, headers, envs); err != nil {
		logger.LogTracedErrorDefault(err)
		return 3
	}
	return 0
}

// 文件大小：
// - 单个tar.zst文件：10GB
// - 解包后文件：12.5GB
var (
// 存储用的账号，以160秒计，存放2组
// 打包文件数 = 24 * 4 * 2 = 192
// 占用空间： (10+12.5) * 192 = 4320 GB
)
