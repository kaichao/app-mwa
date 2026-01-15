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

	"github.com/kaichao/scalebox/pkg/semagroup"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromVtaskHead(body string, headers map[string]string) int {
	// 分组流控信号量的操作，选择节点组
	fmt.Printf("IN fromVtaskHead(), headers:%v\n", headers)

	code := toPullUnpack(body, headers)
	if code != 0 {
		return code
	}

	// 恢复信号量，使得后续wait-queue可持续
	semaName := "vtask_size:wait-queue"
	if _, err := semaphore.AddValue(semaName, appID, 1); err != nil {
		logrus.Errorf("Error in semaphore.AddValue, sema-name:%s,err:%v\n",
			semaName, err)
		return 1
	}
	return 0
}

func toVtaskHead(cubeName string) int {
	// 手工处理信号量组
	groupName := ":slot_vtask_size:vtask-head:"
	v, err := semagroup.Decrement(groupName, appID)
	if err != nil {
		logrus.Errorf("semagroup-decrement error, err:%v\n", err)
		return 1
	}
	// v == `":slot_vtask_size:vtask-head:1":3`
	var slotSeq int
	_, err = fmt.Sscanf(v, `":slot_vtask_size:vtask-head:%d"`, &slotSeq)
	if err != nil {
		logrus.Errorf("Invalid format from semagroup-decrement, err=%v\n", err)
		return 2
	}
	headers := map[string]string{
		"_vtask_cube_name": cubeName,
		"to_slot_index":    fmt.Sprintf("%d", slotSeq),
		"_slot_seq":        fmt.Sprintf("%d", slotSeq),
	}

	// if len(node.Nodes) >= 24 {
	// 	semaName := "counter:cube-vtask"
	// 	os.Setenv("SEMAPHORE_AUTO_CREATE", "yes")
	// 	v, err := semaphore.AddValue(semaName, appID, 1)
	// 	if err != nil {
	// 		logrus.Errorf("In toCubeVtask(), err-info:%v\n", err)
	// 	} else {
	// 		headers["_cube_index"] = v
	// 	}
	// 	os.Unsetenv("SEMAPHORE_AUTO_CREATE")
	// }
	envs := map[string]string{
		"SINK_MODULE": "vtask-head",
	}
	return task.AddWithMapHeaders(cubeName, headers, envs)
}
