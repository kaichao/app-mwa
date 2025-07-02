/*
  实现vtask流控及管理

  - 镜像名：scalebox/agent
  - 输入消息：
  - 输出消息：
  - headers：

  - task分发排序：
  - 流控参数：
  - 环境变量：

*/

package main

import (
	"github.com/kaichao/scalebox/pkg/task"
)

func fromCubeVtask(body string, headers map[string]string) int {
	return toPullUnpack(body)
}

func toCubeVtask(cubeID string) int {
	headers := map[string]string{}
	envs := map[string]string{
		"SINK_JOB": "cube-vtask",
	}
	return task.AddWithMapHeaders(cubeID, headers, envs)
}
