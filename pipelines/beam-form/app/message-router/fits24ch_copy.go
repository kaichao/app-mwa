/*
fits24ch的数据从节点存储拷贝到HPC存储
*/
package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromFits24chCopy(body string, headers map[string]string) int {
	// body == mwa/24ch/1257010784/p00001/t1257010786_1257010965.fits.zst
	re := regexp.MustCompile(`^mwa/24ch/([0-9]+/p[0-9]+)/`)
	ss := re.FindStringSubmatch(body)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", body)
		return 1
	}

	// 信号量pointing-done / vtask-cube-done的减1操作
	semaPairs := map[string]int{}
	// semaphore: pointing-done:1257010784/p00001
	semaName0 := "pointing-done:" + ss[1]
	semaPairs[semaName0] = -1
	vtaskCubeName := headers["_vtask_cube_name"]
	semaName1 := "cube-vtask-done:" + vtaskCubeName
	semaPairs[semaName1] = -1
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	m, err := semaphore.AddMapValues(semaPairs, vtaskID, appID)
	if err != nil {
		logrus.Errorf("error while decrement semaphore,sema-pairs=%v, err:%v\n",
			semaPairs, err)
		return 1
	}

	semaVal := m[semaName1]
	if semaVal > 0 {
		// cube not done.
		return 0
	}

	return toVtaskTail(ss[1])
}

func toFits24chCopy(fileName, targetURL string) int {
	// headers := common.SetJSONAttribute("{}", "target_url", targetURL)
	headers := fmt.Sprintf(`{"target_url":"%s"}`, targetURL)
	envVars := map[string]string{
		"SINK_MODULE": "fits24ch-copy",
	}

	_, err := task.Add(fileName, headers, envVars)
	if err != nil {
		logrus.Errorf("task.AddWithMapHeaders(),err:%v\n", err)
		return 1
	}
	return 0
}
