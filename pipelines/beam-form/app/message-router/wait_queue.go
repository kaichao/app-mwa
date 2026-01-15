package main

import (
	"fmt"
	"strings"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromWaitQueue(body string, headers map[string]string) int {
	return toVtaskHead(body)
}

func toWaitQueue(cubeName string) int {
	// cube-name: 1257010784/p00001_00960/t1257012766_1257012965
	headers := map[string]string{}
	envs := map[string]string{
		"SINK_MODULE": "wait-queue",
	}
	code := task.AddWithMapHeaders(cubeName, headers, envs)
	if code != 0 {
		return code
	}

	var pb, pe int
	n, err := fmt.Sscanf(strings.Split(cubeName, "/")[1], "p%d_%d", &pb, &pe)
	if err != nil || n != 2 {
		logrus.Errorf("error parsing cubeID=%s, err-info:%v\n", cubeName, err)
		return 2
	}
	semaName := "cube-vtask-done:" + cubeName
	semaValue := pe - pb + 1
	if err = semaphore.Create(semaName, semaValue, appID); err != nil {
		logrus.Errorf("Semaphore-create error, sema-name:%s,sema-value:%d,app-id:%d,err-info:%v\n",
			semaName, semaValue, appID, err)
		return 3
	}
	return 0
}
