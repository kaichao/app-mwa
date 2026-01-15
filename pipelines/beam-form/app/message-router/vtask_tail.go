package main

import (
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromVtaskTail(body string, headers map[string]string) int {
	// 恢复应用级编程信号量
	semaName := ":" + headers["_vtask_size_sema"]
	if _, err := semaphore.AddValue(semaName, appID, 1); err != nil {
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
	return task.AddWithMapHeaders(cubeID, headers, envs)
}
