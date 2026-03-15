package main

import (
	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromDownSample(body string, headers map[string]string) error {
	defer func() {
		common.AddTimeStamp("leave-fromDownSample()")
	}()

	err := toFitsRedist(body, headers)
	return errors.WrapE(err, "toFitsRedist()",
		"task-body", body, "headers", headers)
}

func toDownSample(body string, fromHeaders map[string]string) error {
	headers := map[string]string{
		"_sort_tag": fromHeaders["_sort_tag"],
		"sort_tag":  fromHeaders["_sort_tag"],
	}
	envVars := map[string]string{
		"SINK_MODULE": "down-sample",
	}
	_, err := task.AddWithMapHeaders(body, headers, envVars)
	return errors.WrapE(err, "add-task",
		"task-body", body, "headers", headers, "envs", envVars)
}
