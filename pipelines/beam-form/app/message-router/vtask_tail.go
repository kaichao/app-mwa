package main

import (
	"strconv"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromVtaskTail(body string, headers map[string]string) int {
	// 恢复应用级编程信号量
	semaName := ":" + headers["_vtask_size_sema"]
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	if _, err := semaphore.AddValue(semaName, vtaskID, appID, 1); err != nil {
		logrus.Errorf("In fromVtaskTail(), sema-name=%s,err=%v\n", semaName, err)
		return 1
	}
	return 0
}

func toVtaskTail(cubeID string) int {
	headers := map[string]string{}
	envs := map[string]string{
		"SINK_MODULE": "vtask-tail",
	}

	_, err := task.AddWithMapHeaders(cubeID, headers, envs)
	if err != nil {
		logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 1
	}
	return 0
}
