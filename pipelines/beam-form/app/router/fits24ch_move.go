/*
fits24ch的数据从节点存储拷贝到HPC存储
*/
package main

import (
	"fmt"
	"regexp"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromFits24chMove(body string, headers map[string]string) error {
	// body == mwa/24ch/1257010784/p00001/t1257010786_1257010965.fits.zst
	re := regexp.MustCompile(`^mwa/24ch/([0-9]+/p[0-9]+)/`)
	ss := re.FindStringSubmatch(body)
	if ss == nil {
		return errors.E("invalid task-body format", "task-body", body)
	}

	err := toVtaskTail(ss[1], headers)
	return errors.WrapE(err, "toVtaskTail()",
		"pointing-id", ss[1], "headers", headers)
}

func toFits24chMove(fileName, targetURL string) error {
	headers := fmt.Sprintf(`{"target_url":"%s"}`, targetURL)
	envVars := map[string]string{
		"SINK_MODULE": "fits24ch-move",
	}
	_, err := task.Add(fileName, headers, envVars)
	return errors.WrapE(err, "add-task",
		"task-body", fileName, "headers", headers, "envs", envVars)
}
