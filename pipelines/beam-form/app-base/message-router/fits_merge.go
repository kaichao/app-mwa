package main

import (
	"beamform/internal/strparse"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromFitsMerge(message string, headers map[string]string) int {
	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", message)
		return 1
	}

	// semaphore: pointing-done:1257010784/p00001
	sema := "pointing-done:" + ss[1]
	v, err := semaphore.AddValue(sema, 0, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
	semaValue, _ := strconv.Atoi(v)
	if semaValue > 0 {
		// 24ch not done.
		return 0
	}

	return 0
}

func toFitsMerge(cubeID string) int {
	obsID, pBegin, pEnd, t0, t1, _, _ := strparse.ParseParts(cubeID)

	// output task: 1257010784/p00023/t1257010786_1257010965
	tasks := []string{}
	for p := pBegin; p <= pEnd; p++ {
		m := fmt.Sprintf("%s/p%05d/t%d_%d", obsID, p, t0, t1)
		tasks = append(tasks, m)
	}

	envVars := map[string]string{
		"SINK_MODULE": "fits-merge",
	}
	if _, err := task.AddTasks(tasks, "", envVars); err != nil {
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	return 0
}
