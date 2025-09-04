package main

import (
	"beamform/internal/node"
	"beamform/internal/queue"
	"beamform/internal/strparse"
	"fmt"
	"net"
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
	idx := strings.LastIndex(m, "/ch")
	if idx == -1 {
		logrus.Errorf("invalid message format from fits-redist, message=%s\n", m)
		return 1
	}
	cubeID := m[:idx]
	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	semaName := fmt.Sprintf("fits-done:%s", cubeID)
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

func toFitsRedist(m string, fromHeaders map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	obsID, p0, p1, _, _, _, err := strparse.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}

	// cube := datacube.NewDataCube(obsID)
	cubeIndex, _ := strconv.Atoi(fromHeaders["_cube_index"])
	cubeIndex--
	fmt.Printf("cube-index=%d\n", cubeIndex)
	ips := node.GetIPAddrListByCubeIndex(cubeIndex)
	// ips := node.GetIPAddrListByTime(cube, t0)
	fromIP := fromHeaders["from_ip"]

	// 1. 读取共享变量表 pointing-data-root。
	varValues := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:%s/p%05d", obsID, p)
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
			varName := fmt.Sprintf("pointing-data-root:%s/p%05d", obsID, p)
			var varValue, ip string
			if i < len(prestoIPs) {
				// 类型1，非组内IP地址
				ip = prestoIPs[i]
				varValue = prestoIPs[i]
			} else {
				// 类型2、类型3，组内地址
				// varValue = os.Getenv("TARGET_24CH_ROOT")
				// if varValue == "" {
				// 	varValue = targetPicker.GetNext()
				// }
				varValue = getStagingRoot(p)
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
	hs := fmt.Sprintf(`{"target_hosts":"%s","sort_tag":"%s"}`,
		strings.Join(toIPs, ","), fromHeaders["_sort_tag"])

	envVars := map[string]string{
		"SINK_JOB": "fits-redist",
	}
	common.AddTimeStamp("before-send-messages")
	return task.Add(m, hs, envVars)
}
