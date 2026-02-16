package main

import (
	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromWaitQueue(body string, headers map[string]string) int {
	return toVtaskHead(body)
}

func toWaitQueue(cubeName string) int {
	// cube-name: 1257010784/p00001_00960/t1257012766_1257012965
	headers := map[string]string{}
	envs := map[string]string{
		"SINK_MODULE":     "wait-queue",
		"CONFLICT_ACTION": "OVERWRITE",
	}

	_, err := task.AddWithMapHeaders(cubeName, headers, envs)
	if err != nil {
		logger.LogTracedErrorDefault(err)
		return 1
	}

	return 0
}
