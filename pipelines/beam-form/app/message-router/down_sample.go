package main

import (
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromDownSample(m string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromDownSample()")
	}()
	return toFitsRedist(m, headers)
}

func toDownSample(body string) int {
	envVars := map[string]string{
		"SINK_JOB": "down-sample",
	}
	return task.Add(body, "{}", envVars)
}
