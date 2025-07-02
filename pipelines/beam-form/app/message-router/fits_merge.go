package main

import (
	"beamform/internal/datacube"
	"beamform/internal/node"
	"beamform/internal/strparse"
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

func fromFitsMerge(m string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromFitsMerge()")
	}()
	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(m)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", m)
		return 1
	}

	varName := fmt.Sprintf("pointing-data-root:%s", ss[1])
	varValue, err := variable.Get(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}
	if strings.Contains(varValue, "@") {
		// 共享变量pointing-data-root，若为类型3，给fits-push发消息，推送到远端ssh存储
		msg := fmt.Sprintf("mwa/24ch/%s.fits.zst", m)
		headers := common.SetJSONAttribute("{}", "target_url", varValue)
		// headers = common.SetJSONAttribute("{}", "target_jump_servers", "root@10.200.1.100")

		envVars := map[string]string{
			"SINK_JOB": "fits-push",
		}
		return task.Add(msg, headers, envVars)
	}

	common.AddTimeStamp("before-send-messages")
	return doCrossAppTaskAdd(ss[1])
}

func toFitsMerge(m string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	ds, p0, p1, t0, t1, _, err := strparse.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}

	cube := datacube.GetDataCube(ds)
	// output message: 1257010784/p00023/t1257010786_1257010965
	messages := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:%s/p%05d", ds, p)
		varValue, err := variable.Get(varName, appID)
		if err != nil {
			logrus.Errorf("variable-get, var-name:%s, err-info:%v\n", varName, err)
			return 11
		}

		headers := ""
		// BUG: 节点数量少，补充数据时，pointing计数不对齐，toHost不准确
		toHost := node.GetNodeNameByPointingTime(cube, p, t0)
		if ip := net.ParseIP(varValue); ip != nil && ip.To4() != nil {
			// IPv4地址（类型1）， 设置"to_ip"头
			headers = common.SetJSONAttribute(headers, "to_ip", varValue)
			headers = common.SetJSONAttribute(headers,
				"output_root", "/tmp/scalebox/mydata")
		} else if strings.Contains(varValue, "@") {
			// 远端存储（类型3）
			headers = common.SetJSONAttribute(headers, "to_host", toHost)
			headers = common.SetJSONAttribute(headers,
				"output_root", "/dev/shm/scalebox/mydata")
			// 24ch存放在/dev/shm
		} else {
			// 共享存储（类型2）
			headers = common.SetJSONAttribute(headers, "to_host", toHost)
			// 24ch存放在共享存储
			headers = common.SetJSONAttribute(headers,
				"output_root", varValue)
		}
		m := fmt.Sprintf(`%s/p%05d/t%d_%d,%s`, ds, p, t0, t1, headers)
		fmt.Printf("var-value:%s,to-host:%s,headers=%s,m=%s\n",
			varValue, toHost, headers, m)
		messages = append(messages, m)
	}

	common.AddTimeStamp("before-send-messages")
	envVars := map[string]string{
		"SINK_JOB":        "fits-merge",
		"TIMEOUT_SECONDS": "600",
	}
	return task.AddTasks(messages, "{}", envVars)
}
