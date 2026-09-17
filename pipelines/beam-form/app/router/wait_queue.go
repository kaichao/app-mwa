package main

import (
	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromWaitQueue(body string, headers map[string]string) error {
	// 标准 wait-queue 模块的 run.sh 已完成 vtask bind + add-subtask --direct，
	// 且不写 sink-tasks.txt，task 在 wait-queue slot 终结，正常无回流。
	// 若因自定义 run.sh 等原因回流，空操作终结，避免重复 bind 造成额度双扣。
	return nil
}

func toWaitQueue(cubeName string) error {
	// cube-name: 1257010784/p00001_00960/t1257012766_1257012965
	headers := map[string]string{}
	envs := map[string]string{
		"SINK_MODULE":     "wait-queue",
		"CONFLICT_ACTION": "OVERWRITE",
	}

	_, err := task.AddWithMapHeaders(cubeName, headers, envs)
	return errors.WrapE(err, "add-tasks", "task-body", cubeName)
}
