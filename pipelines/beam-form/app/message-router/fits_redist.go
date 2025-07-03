package main

import (
	"beamform/internal/datacube"
	"beamform/internal/node"
	"beamform/internal/queue"
	"beamform/internal/strparse"
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

func fromFitsRedist(m string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromFitsRedist()")
	}()
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	ds, p0, p1, t0, t1, _, err := strparse.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}

	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	semaName := fmt.Sprintf("fits-done:%s/p%05d_%05d/t%d_%d",
		ds, p0, p1, t0, t1)
	v, err := semaphore.AddValue(semaName, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema:%s, err:%v\n", semaName, err)
		return 1
	}
	semaVal, _ := strconv.Atoi(v)
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	return toFitsMerge(m)
}

func toFitsRedist(m string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	dataset, p0, p1, t0, _, _, err := strparse.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}

	cube := datacube.NewDataCube(dataset)
	ips := node.GetIPAddrListByTime(cube, t0)
	fromIP := headers["from_ip"]

	// 1. 读取共享变量表 pointing-data-root。
	varValues := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:%s/p%05d", dataset, p)
		if v, err := variable.Get(varName, appID); err != nil {
			logrus.Errorf("variable-get %s, err-info:%v\n", varName, err)
			varValues = append(varValues, "")
		} else {
			varValues = append(varValues, v)
		}
	}

	toIPs := []string{}
	// 2. 生成待分发IP列表。若不存在，按需创建共享变量表 pointing-data-root
	if slices.IndexFunc(varValues, func(s string) bool { return s != "" }) == -1 {
		// varValues all empty string ""
		// 变量不存在，创建共享变量组
		// 从队列中读取可分配的节点
		prestoIPs, err := queue.PopN(p1 - p0 + 1)
		if err != nil {
			fmt.Printf("Queue pop error, err-info:%v\n", err)
			logrus.Errorf("Queue pop error, err-info:%v\n", err)
			return 2
		}
		fmt.Println("presto-ips:", prestoIPs)

		for p := p0; p <= p1; p++ {
			i := p - p0
			varName := fmt.Sprintf("pointing-data-root:%s/p%05d", dataset, p)
			var varValue, ip string
			if i < len(prestoIPs) {
				// 类型1，非组内IP地址
				ip = prestoIPs[i]
				varValue = prestoIPs[i]
			} else {
				// 类型2、类型3，组内地址
				varValue = os.Getenv("TARGET_24CH_ROOT")
				if varValue == "" {
					varValue = targetPicker.GetNext()
				}
				if ips[i] == fromIP {
					ip = "localhost"
				} else {
					ip = ips[i]
				}
			}
			variable.Set(varName, varValue, appID)
			toIPs = append(toIPs, ip)
		}
	} else {
		// 从已有共享变量表中读取
		for p := p0; p <= p1; p++ {
			i := p - p0
			ip := net.ParseIP(varValues[i])
			if ip != nil && ip.To4() != nil {
				// 若是远端IP（类型1），ipv4 addr
				toIPs = append(toIPs, varValues[i])
			} else if ips[i] == fromIP {
				toIPs = append(toIPs, "localhost")
			} else {
				toIPs = append(toIPs, ips[i])
			}
		}
	}

	// 3. 完成target_hosts的数据采集，向fits-redist发送task对应消息
	hs := fmt.Sprintf(`{"target_hosts":"%s"}`, strings.Join(toIPs, ","))

	envVars := map[string]string{
		"SINK_JOB": "fits-redist",
	}
	common.AddTimeStamp("before-send-messages")
	return task.Add(m, hs, envVars)
}
