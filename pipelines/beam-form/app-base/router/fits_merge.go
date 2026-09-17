package main

import (
	"beamform/internal/strparse"
	"fmt"
	"regexp"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
)

func fromFitsMerge(body string, headers map[string]string) error {
	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(body)
	if ss == nil {
		return errors.E("invalid task-body format", "task-body", body)
	}

	// semaphore: pointing-done:1257010784/p00001
	sema := "pointing-done:" + ss[1]
	semaValue, err := semaphore.AddValue(sema, -1, appID)
	if err != nil {
		return errors.WrapE(err, 2, "semaphore-decrement",
			"sema-name", sema, "app-id", appID)
	}
	if semaValue > 0 {
		// 24ch not done.
		return nil
	}

	return nil
}

func toFitsMerge(cubeID string) error {
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
		return errors.WrapE(err, 1, "add-tasks",
			"task-lines", tasks, "envs", envVars)
	}
	return nil
}
