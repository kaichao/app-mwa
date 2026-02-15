package main

import (
	"strconv"

	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromVtaskTail(body string, headers map[string]string) int {
	// 恢复应用级编程信号量
	// vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	// 待确认 vtask-id = 0 ?
	vtaskID := int64(0)
	semaName := ":" + headers["_vtask_size_sema"]
	if _, err := semaphore.AddValue(semaName, vtaskID, appID, 1); err != nil {
		logger.LogTracedErrorDefault(err)
		logrus.Errorf("In fromVtaskTail(), sema-name=%s,err=%v\n", semaName, err)
		return 1
	}
	return 0
}

// pointingID:
func toVtaskTail(pointingID string, fromHeaders map[string]string) int {
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
	m, err := semaphore.AddMapValues(semaPairs, vtaskID, appID)
	if err != nil {
		logrus.Errorf("error while decrement semaphore,sema-pairs=%v, err:%v\n",
			semaPairs, err)
		return 1
	}

	if semaVal := m[semaName1]; semaVal > 0 {
		// cube not done.
		return 0
	}

	headers := map[string]string{}
	envs := map[string]string{
		"SINK_MODULE": "vtask-tail",
	}

	if _, err := task.AddWithMapHeaders(pointingID, headers, envs); err != nil {
		logger.LogTracedErrorDefault(err)
		// logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 1
	}
	return 0
}
