package main

import (
	"beamform/internal/datacube"
	"beamform/internal/node"
	"beamform/internal/strparse"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/kaichao/scalebox/pkg/common"
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
		logrus.Errorf("Invalid format, task-body:%s\n", body)
		return 1
	}
	pointingID := ss[1]
	varName := fmt.Sprintf("pointing-data-root:%s", pointingID)
	varValue, err := getPointingVariable(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}

	fileName := fmt.Sprintf("mwa/24ch/%s.fits.zst", body)
	if os.Getenv("RUN_MODE") == "full-parallel" {
		// 全并行，需通过fits24ch-copy拷贝到脉冲星搜索的计算节点（全并行，fits合并在后续计算节点上？）
		return toFits24chCopy(fileName, varValue)
	}
	if strings.Contains(varValue, "@") {
		// 共享变量pointing-data-root，若为类型3，给fits24-unload发消息，推送到远端ssh存储
		if code := toFits24chUnload(fileName, varValue); code > 0 {
			return code
		}
	}

	return toVtaskTail(pointingID, headers)

	// return toCrossAppPresto(ss[1])
}

func toFitsMerge(body string) int {
	// task-body: 1257010784/p00001_00024/t1257012766_1257012965/ch109
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
				"output_root", "/public/home/cstu0100/scalebox/mydata/mwa")
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
