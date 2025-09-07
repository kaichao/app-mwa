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
	"beamform/internal/datacube"
	"fmt"
	"os"
	"strconv"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

// body: 1267459410_1267459449_ch109.dat.tar.zst
// headers:
//   - _cube_id: 1257010784/p00001_00960/t1257012766_1257012965
func fromTarLoad(body string, headers map[string]string) int {
	if os.Getenv("PRELOAD_MODE") == "preload-only" {
		return 0
	}
	cubeID := headers["_cube_id"]
	semaName := "tar-ready:" + cubeID
	v, err := semaphore.AddValue(semaName, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, name=%s,err-info:%v\n", semaName, err)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		logrus.Errorf("semaphore-decrement, atoi error, name=%s,err-info:%v\n", semaName, err)
	}
	if n <= 0 {
		// 若支持分组级slot，则发给pull-unpack
		// 否则发给cube-vtask
		return toCubeVtask(cubeID)
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
	// 按顺序产生file-copy消息
	cube := datacube.NewDataCube(datasetID)
	sourceURL := fmt.Sprintf("%s/mwa/tar/%s", getOriginRoot(), cube.ObsID)
	fmtTarZst := `%d_%d_ch%d.dat.tar.zst`
	bodies := []string{}
	semas := []*semaphore.Sema{}
	// vars := []string{}

	trs := cube.GetTimeRanges()
	for i := 0; i < len(trs); i += 2 {
		tus := cube.GetTimeUnitsWithinInterval(trs[i], trs[i+1])
		cubeID := fmt.Sprintf("%s/p%05d_%05d/t%d_%d",
			cube.ObsID, cube.PointingBegin, cube.PointingEnd, trs[i], trs[i+1])
		semaName := "tar-ready:" + cubeID
		semaValue := len(tus) / 2 * cube.NumOfChannels
		semas = append(semas, &semaphore.Sema{Name: semaName, Value: semaValue})

		for k := 0; k < len(tus); k += 2 {
			for j := 0; j < cube.NumOfChannels; j++ {
				ch := cube.ChannelBegin + j
				// 	storIndex++
				// 	if storIndex > storEnd {
				// 		storIndex = storBegin
				// 	}
				// vars = append(vars, fmt.Sprintf(`cube-stor-index:%s/ch%03d,%d`, cubeID, ch, storIndex))
				// cubeURL := fmt.Sprintf("cstu00%d@60.245.128.14:65010/public/home/cstu00%d/mydata/mwa/tar",
				// 	storIndex, storIndex)
				targetURL := fmt.Sprintf("%s/mwa/tar/%s",
					// targetURL := fmt.Sprintf("cstu0030@60.245.128.14:65010%s/mwa/tar/%s",
					getPreloadRoot(ch-cube.ChannelBegin), cube.ObsID)
				fileName := fmt.Sprintf(fmtTarZst, tus[k], tus[k+1], ch)
				body := fmt.Sprintf(`%s,{"target_url":"%s","_cube_id":"%s"}`,
					fileName, targetURL, cubeID)
				bodies = append(bodies, body)
			}
		}
	}
	headers := map[string]string{
		"source_url": sourceURL,
	}
	envs := map[string]string{
		"SINK_JOB": "tar-load",
	}
	// for _, line := range vars {
	// 	ss := strings.Split(line, ",")
	// 	err := variable.Set(ss[0], ss[1], appID)
	// 	if err != nil {
	// 		logrus.Errorf("create variable, name=%s,value=%s,err-info:%v\n",
	// 			ss[0], ss[1], err)
	// 		return 2
	// 	}
	// }
	if err := semaphore.CreateSemaphores(semas, appID, 100); err != nil {
		logrus.Errorf("Create semaphore, err-info:%v\n", err)
		return 1
	}

	return task.AddTasksWithMapHeaders(bodies, headers, envs)
}

// 文件大小：
// - 单个tar.zst文件：10GB
// - 解包后文件：12.5GB
var (
// 存储用的账号，以160秒计，存放2组
// 打包文件数 = 24 * 4 * 2 = 192
// 占用空间： (10+12.5) * 192 = 4320 GB
)
