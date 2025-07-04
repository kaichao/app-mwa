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
	"beamform/internal/node"
	"fmt"
	"os"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromCubeVtask(body string, headers map[string]string) int {

	// 分组流控信号量的操作，选择节点组

	fmt.Printf("IN fromCubeVtask(), headers:%v\n", headers)

	return toPullUnpack(body, headers)
}

func toCubeVtask(cubeID string) int {
	headers := map[string]string{}
	if len(node.Nodes) >= 24 {
		semaName := "counter:cube-vtask"
		os.Setenv("SEMAPHORE_AUTO_CREATE", "yes")
		v, err := semaphore.AddValue(semaName, appID, 1)
		if err != nil {
			logrus.Errorf("In toCubeVtask(), err-info:%v\n", err)
		} else {
			headers["_cube_index"] = v
		}
		os.Unsetenv("SEMAPHORE_AUTO_CREATE")
	}
	envs := map[string]string{
		"SINK_JOB": "cube-vtask",
	}
	return task.AddWithMapHeaders(cubeID, headers, envs)
}
