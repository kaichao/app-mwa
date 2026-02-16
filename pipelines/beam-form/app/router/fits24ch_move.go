/*
fits24ch的数据从节点存储拷贝到HPC存储
*/
package main

import (
	"fmt"
	"regexp"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromFits24chMove(body string, headers map[string]string) int {
	// body == mwa/24ch/1257010784/p00001/t1257010786_1257010965.fits.zst
	re := regexp.MustCompile(`^mwa/24ch/([0-9]+/p[0-9]+)/`)
	ss := re.FindStringSubmatch(body)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", body)
		return 1
	}

	return toVtaskTail(ss[1], headers)
}

func toFits24chMove(fileName, targetURL string) int {
	headers := fmt.Sprintf(`{"target_url":"%s"}`, targetURL)
	envVars := map[string]string{
		"SINK_MODULE": "fits24ch-move",
	}

	if _, err := task.Add(fileName, headers, envVars); err != nil {
		logger.LogTracedErrorDefault(err)
		// logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 1
	}
	return 0
}
