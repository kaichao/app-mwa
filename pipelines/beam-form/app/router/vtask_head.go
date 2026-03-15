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
	"github.com/kaichao/scalebox/pkg/semagroup"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
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

	// 分组流控信号量的操作，选择节点组
	err = toPullUnpack(body, headers)
	if err != nil {
		return errors.WrapE(err, 2, "to-pull-unpack",
			"body", body, "headers", headers)
	}

	// 恢复信号量，使得后续wait-queue可持续
	semaName := "vtask_size:wait-queue"
	if _, err := semaphore.AddValue(semaName, 0, appID, 1); err != nil {
		return errors.WrapE(err, "restore-sema", "sema-name", semaName, "app-id", appID)
	}

	semaName = "cube-vtask-done:" + body
	semaValue := pe - pb + 1
	err = semaphore.Create(semaName, semaValue, vtaskID, appID)
	return errors.WrapE(err, 3, "semaphore.Create()",
		"sema-name", semaName, "sema-value", semaValue, "app-id", appID, "vtask-id", vtaskID)
}

func toVtaskHead(cubeName string) error {
	// 手工处理信号量组
	groupName := ":slot_vtask_size:vtask-head:"
	semaName, _, err := semagroup.Decrement(groupName, appID)
	fmt.Printf("In toVtaskHead(),sema-name:%s,seq#%s#\n", semaName, semaName[len(groupName):])
	slotSeq, _ := strconv.Atoi(semaName[len(groupName):])
	if err != nil {
		return errors.WrapE(err, "strconv.Atoi()",
			"value", semaName[len(groupName):], "sema-name", semaName)
	}

	headers := map[string]string{
		"_vtask_cube_name": cubeName,
		"to_slot_index":    fmt.Sprintf("%d", slotSeq),
		"_slot_seq":        fmt.Sprintf("%d", slotSeq),
	}
	envs := map[string]string{
		"SINK_MODULE": "vtask-head",
	}
	_, err = task.AddWithMapHeaders(cubeName, headers, envs)
	return errors.WrapE(err, 2, "add-task",
		"body", cubeName, "headers", headers, "envs", envs)
}
