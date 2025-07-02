/*
fromNull是消息路由的首个执行模块，按PRELOAD_MODE指定的预加载策略，加载原始数据。
*/
package main

import (
	"beamform/internal/datacube"
	"fmt"
	"os"
	"reflect"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
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
