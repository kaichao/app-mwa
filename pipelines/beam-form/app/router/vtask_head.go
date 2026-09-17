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
	"fmt"
	"strconv"
	"strings"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/vtask"
)

func fromVtaskHead(body string, headers map[string]string) error {
	vtaskID, err := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	if err != nil {
		return errors.WrapE(err, "no valid vtask-id", "vtask-id", headers["_vtask_id"])
	}

	var pb, pe int
	n, err := fmt.Sscanf(strings.Split(body, "/")[1], "p%d_%d", &pb, &pe)
	if err != nil || n != 2 {
		return errors.WrapE(err, "parse cube-id", "cube-id", body)
	}

	// root task 由标准 wait-queue 模块的 run.sh（--direct）创建，
	// 不携带业务 header；在此补齐 _vtask_cube_name，
	// 随 _vtask_ 前缀 header 沿 sink 链自动传播到下游 core 模块
	headers["_vtask_cube_name"] = body

	// 分组流控信号量的操作，选择节点组
	err = toPullUnpack(body, headers)
	if err != nil {
		return errors.WrapE(err, 2, "to-pull-unpack",
			"body", body, "headers", headers)
	}

	// 恢复信号量，使得后续wait-queue可持续
	semaName := "vtask_size:wait-queue"
	if _, err := semaphore.AddValue(semaName, 1, appID); err != nil {
		return errors.WrapE(err, "restore-sema", "sema-name", semaName, "app-id", appID)
	}

	semaName = "cube-vtask-done:" + body
	semaValue := pe - pb + 1
	// err = semaphore.Create(semaName, semaValue, vtaskID, appID)
	err = vtask.CreateSemaphore(semaName, semaValue, vtaskID, appID)
	return errors.WrapE(err, 3, "semaphore.Create()",
		"sema-name", semaName, "sema-value", semaValue, "app-id", appID, "vtask-id", vtaskID)
}
