/*
  fits24ch的数据从HPC存储拷贝到外部存储
  运行在I/O节点上
*/

package main

import (
	"fmt"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromFits24chUnload(body string, headers map[string]string) int {
	// 仅纪录
	return 0
}

func toFits24chUnload(fileName, targetURL string) int {
	headers := fmt.Sprintf(`{"target_url":"%s"}`, targetURL)
	envVars := map[string]string{
		"SINK_MODULE": "fits24ch-unload",
	}
	if _, err := task.Add(fileName, headers, envVars); err != nil {
		logger.LogTracedErrorDefault(err)
		return 1
	}
	return 0
}
