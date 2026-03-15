/*
  fits24ch的数据从HPC存储拷贝到外部存储
  运行在I/O节点上
*/

package main

import (
	"fmt"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromFits24chUnload(body string, headers map[string]string) error {
	// 仅纪录，不处理。
	return nil
}

func toFits24chUnload(fileName, targetURL string) error {
	headers := fmt.Sprintf(`{"target_url":"%s"}`, targetURL)
	envVars := map[string]string{
		"SINK_MODULE": "fits24ch-unload",
	}
	_, err := task.Add(fileName, headers, envVars)
	return errors.WrapE(err, "add-task",
		"task-body", fileName, "headers", headers, "envs", envVars)
}
