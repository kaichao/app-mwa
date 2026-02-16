package main

import (
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromDownSample(body string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromDownSample()")
	}()
	return toFitsRedist(body, headers)
}

func toDownSample(body string, fromHeaders map[string]string) int {
	headers := map[string]string{
		"_sort_tag": fromHeaders["_sort_tag"],
		"sort_tag":  fromHeaders["_sort_tag"],
	}
	envVars := map[string]string{
		"SINK_MODULE": "down-sample",
	}
	_, err := task.AddWithMapHeaders(body, headers, envVars)
	if err != nil {
		logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 1
	}
	return 0
}
