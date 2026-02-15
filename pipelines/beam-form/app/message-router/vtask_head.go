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

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/semagroup"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromVtaskHead(body string, headers map[string]string) int {
	vtaskID, err := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	if err != nil {
		logrus.Errorf("_vtask_id=%s, no valid vtask-id in headers, err:%v\n", headers["_vtask_id"], err)
		return 1
	}

	var pb, pe int
	n, err := fmt.Sscanf(strings.Split(body, "/")[1], "p%d_%d", &pb, &pe)
	if err != nil || n != 2 {
		logrus.Errorf("error parsing cubeID=%s, err-info:%v\n", body, err)
		return 1
	}

	// 分组流控信号量的操作，选择节点组
	code := toPullUnpack(body, headers)
	if code != 0 {
		return code
	}

	// 恢复信号量，使得后续wait-queue可持续
	semaName := "vtask_size:wait-queue"
	if _, err := semaphore.AddValue(semaName, 0, appID, 1); err != nil {
		logger.LogTracedErrorDefault(err)
		// logrus.Errorf("Error in semaphore.AddValue, sema-name:%s,err:%v\n",semaName, err)
		return 2
	}

	semaName = "cube-vtask-done:" + body
	semaValue := pe - pb + 1
	if err = semaphore.Create(semaName, semaValue, vtaskID, appID); err != nil {
		logger.LogTracedErrorDefault(err)
		// logrus.Errorf("Semaphore-create error, sema-name:%s,sema-value:%d,app-id:%d,err-info:%v\n", semaName, semaValue, appID, err)
		return 3
	}

	return 0
}

func toVtaskHead(cubeName string) int {
	// 手工处理信号量组
	groupName := ":slot_vtask_size:vtask-head:"
	semaName, _, err := semagroup.Decrement(groupName, appID)
	fmt.Printf("In toVtaskHead(),sema-name:%s,seq#%s#\n", semaName, semaName[len(groupName):])
	slotSeq, _ := strconv.Atoi(semaName[len(groupName):])
	if err != nil {
		logger.LogTracedErrorDefault(err)
		return 1
	}
	headers := map[string]string{
		"_vtask_cube_name": cubeName,
		"to_slot_index":    fmt.Sprintf("%d", slotSeq),
		"_slot_seq":        fmt.Sprintf("%d", slotSeq),
	}

	envs := map[string]string{
		"SINK_MODULE": "vtask-head",
	}

	if _, err = task.AddWithMapHeaders(cubeName, headers, envs); err != nil {
		logger.LogTracedErrorDefault(err)
		// logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 2
	}

	return 0
}
