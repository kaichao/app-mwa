package main

import (
	"strconv"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/vtask"
)

func fromVtaskTail(body string, headers map[string]string) error {
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	semaName := ":" + headers["_vtask_size_sema"]
	err := vtask.UnbindResource(appID, vtaskID, semaName)

	return errors.WrapE(err, "vtask-unbind-resource",
		"sema-name", semaName, "app-id", appID, "vtask-id", vtaskID)
}

// pointingID:
func toVtaskTail(pointingID string, fromHeaders map[string]string) error {
	vtaskID, _ := strconv.ParseInt(fromHeaders["_vtask_id"], 10, 64)
	// 信号量pointing-done / vtask-cube-done的减1操作
	semaPairs := map[string]int{}
	// semaphore: pointing-done:1257010784/p00001
	// TODO: 稍等再加入
	// semaName0 := "pointing-done:" + pointingID
	// semaPairs[semaName0] = -1
	semaName1 := "cube-vtask-done:" + fromHeaders["_vtask_cube_name"]
	semaPairs[semaName1] = -1
	// 执行减一操作
	m, err := vtask.AddSemaphoreMapValues(semaPairs, vtaskID, appID)
	if err != nil {
		return errors.WrapE(err, "decrement semaphore",
			"app-id", appID, "vtask-id", vtaskID, "sema-pairs", semaPairs)
	}

	if semaVal := m[semaName1]; semaVal > 0 {
		// cube not done.
		return nil
	}

	headers := map[string]string{}
	envs := map[string]string{
		"SINK_MODULE": "vtask-tail",
	}

	body := fromHeaders["_vtask_cube_name"]
	_, err = task.AddWithMapHeaders(body, headers, envs)
	return errors.WrapE(err, "add-task",
		"task-body", body, "task-headers", headers, "envs", envs)
}
