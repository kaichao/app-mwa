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

func toDownSample(body string, fromHeaders map[string]string) int {
	// headers := fmt.Sprintf(`{"_cube_index":"%s"}`, fromHeaders["_cube_index"])
	headers := map[string]string{
		"_cube_index": fromHeaders["_cube_index"],
		"_sort_tag":   fromHeaders["_sort_tag"],
		"sort_tag":    fromHeaders["_sort_tag"],
	}
	envVars := map[string]string{
		"SINK_JOB": "down-sample",
	}
	return task.AddWithMapHeaders(body, headers, envVars)
}
