package main

import (
	"beamform/internal/datacube"
	"beamform/internal/node"
	"beamform/internal/strparse"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromFitsMerge(body string, headers map[string]string) int {
	// body == 1257010784/p00001/t1257010786_1257010965
	defer func() {
		common.AddTimeStamp("leave-fromFitsMerge()")
	}()
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(body)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", body)
		return 1
	}

	varName := fmt.Sprintf("pointing-data-root:%s", ss[1])
	varValue, err := getPointingVariable(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}
	if strings.Contains(varValue, "@") {
		// 共享变量pointing-data-root，若为类型3，给fits24-copy发消息，推送到远端ssh存储
		fileName := fmt.Sprintf("mwa/24ch/%s.fits.zst", body)
		return toFits24chCopy(fileName, varValue)
	}

	// 信号量pointing-done / vtask-cube-done的减1操作
	// semaphore: pointing-done:1257010784/p00001
	semaPointingDone := "pointing-done:" + ss[1]
	semaVal, err := semaphore.AddValue(semaPointingDone, 0, appID, -1)
	fmt.Printf("ret-val of sema=pointing-done:%d\n", semaVal)
	if err != nil {
		logrus.Errorf("decrement sema-pointing-done, err:%v\n", err)
		return 1
	}

	vtaskCubeName := headers["_vtask_cube_name"]
	semaCubeVtaskDone := "cube-vtask-done:" + vtaskCubeName
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	semaVal, err = semaphore.AddValue(semaCubeVtaskDone, vtaskID, appID, -1)
	fmt.Printf("ret-val of cube-vtask-done:%d\n", semaVal)
	if err != nil {
		logrus.Errorf("decrement sema-cube-vtask-done, err:%v\n", err)
		return 2
	}

	if semaVal > 0 {
		// cube not done.
		return 0
	}

	return toVtaskTail(ss[1])

	// common.AddTimeStamp("before-add-tasks")
	// return toCrossAppPresto(ss[1])
}

func toFitsMerge(body string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	ds, p0, p1, t0, t1, _, err := strparse.ParseParts(body)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", body, err)
		return 1
	}

	cube := datacube.NewDataCube(ds)
	// sink task: 1257010784/p00023/t1257010786_1257010965
	tasks := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:%s/p%05d", ds, p)
		varValue, err := getPointingVariable(varName, appID)
		if err != nil {
			logrus.Errorf("variable-get, var-name:%s, err-info:%v\n", varName, err)
			return 11
		}

		headers := ""
		// BUG: 节点数量少，补充数据时，pointing计数不对齐，toHost不准确
		toHost := node.GetNodeNameByPointingTime(cube, p, t0)
		if ip := net.ParseIP(varValue); ip != nil && ip.To4() != nil {
			// IPv4地址（类型1）， 设置"to_ip"头
			headers, _ = common.SetJSONAttribute(headers, "to_ip", varValue)
			headers, _ = common.SetJSONAttribute(headers,
				"output_root", os.Getenv("LOCAL_TMPDIR")+"/mydata")
		} else if strings.Contains(varValue, "@") {
			// 远端存储（类型3，远端ssh存储）
			headers, _ = common.SetJSONAttribute(headers, "to_host", toHost)
			// 24ch存放在/dev/shm
			// headers = common.SetJSONAttribute(headers,
			// 	"output_root", os.Getenv("LOCAL_SHMDIR")+"/mydata")
			// 24ch存放在共享存储
			headers, _ = common.SetJSONAttribute(headers,
				"output_root", "/public/home/cstu0100/scalebox/mydata")
		} else {
			// 共享存储（类型2）
			headers, _ = common.SetJSONAttribute(headers, "to_host", toHost)
			// 24ch存放在共享存储
			headers, _ = common.SetJSONAttribute(headers,
				"output_root", varValue)
		}
		m := fmt.Sprintf(`%s/p%05d/t%d_%d,%s`, ds, p, t0, t1, headers)
		fmt.Printf("var-value:%s,to-host:%s,headers=%s,m=%s\n",
			varValue, toHost, headers, m)
		tasks = append(tasks, m)
	}

	common.AddTimeStamp("before-add-tasks")
	envVars := map[string]string{
		"SINK_MODULE":     "fits-merge",
		"TIMEOUT_SECONDS": "600",
	}
	_, err = task.AddTasks(tasks, "{}", envVars)
	if err != nil {
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	return 0
}
