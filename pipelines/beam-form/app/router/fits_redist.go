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

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromFitsRedist(body string, headers map[string]string) error {
	defer func() {
		common.AddTimeStamp("leave-fromFitsRedist()")
	}()
	// task-bdy: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	idx := strings.LastIndex(body, "/ch")
	if idx == -1 {
		return errors.E("invalid task-body format", "task-body", body)
	}
	cubeID := body[:idx]
	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	semaName := fmt.Sprintf("fits-done:%s", cubeID)
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	semaVal, err := semaphore.AddValue(semaName, vtaskID, appID, -1)
	if err != nil {
		return errors.WrapE(err, 2, "semaphore-decrement",
			"sema-name", semaName, "app-id", appID, "vtask-id", vtaskID)
	}
	if semaVal > 0 {
		// 24ch not done.
		return nil
	}

	return errors.WrapE(toFitsMerge(body, headers), "toFitsMerge()",
		"task-body", body, "headers", headers)
}

func toFitsRedist(body string, fromHeaders map[string]string) error {
	// input task-body: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	obsID, p0, p1, _, _, _, err := strparse.ParseParts(body)
	if err != nil {
		return errors.WrapE(err, "invalid task-body format", "task-body", body)
	}

	semaName := fromHeaders["_vtask_size_sema"]
	ss := strings.Split(semaName, ":")
	groupIndex, _ := strconv.Atoi(ss[len(ss)-1])
	fmt.Printf("group-index=%d\n", groupIndex)
	ips := node.GetIPAddrListByGroupIndex(groupIndex)
	fromIP := fromHeaders["from_ip"]

	// 1. 读取共享变量表 pointing-data-root。
	varValues := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:%s/p%05d", obsID, p)
		if v, err := getPointingVariable(varName, appID); err != nil {
			logrus.Errorf("variable-get %s, err-info:%v\n", varName, err)
			varValues = append(varValues, "")
		} else {
			varValues = append(varValues, v)
		}
	}

	toIPs := []string{}
	// 2. 生成待分发IP列表。若不存在，按需创建共享变量表 pointing-data-root
	// 判断 varValues 这个字符串切片里，是否所有元素都是空字符串 ""
	if slices.IndexFunc(varValues, func(s string) bool { return s != "" }) == -1 {
		// varValues all empty string ""
		// 变量不存在，创建共享变量组
		// 从队列中读取可分配的节点
		prestoIPs, err := queue.PopN(p1 - p0 + 1)
		if err != nil {
			return errors.WrapE(err, 2, "queue pop error", "num", p1-p0+1)
		}
		fmt.Println("presto-ips:", prestoIPs)

		for p := p0; p <= p1; p++ {
			i := p - p0
			pointingDir := fmt.Sprintf("%s/p%05d", obsID, p)
			varName := "pointing-data-root:" + pointingDir
			var varValue, ip string
			if i < len(prestoIPs) {
				// 类型1，非组内IP地址
				ip = prestoIPs[i]
				varValue = prestoIPs[i]
			} else {
				// 类型2、类型3，组内地址
				// 自增长的index
				varValue, err = vPath.GetPath("stageing-24ch", pointingDir)
				if err != nil {
					return errors.WrapE(err, 9, "vPath.GetPath()",
						"category", "stageing-24ch", "key", pointingDir)
				}
				if ips[i] == fromIP {
					ip = "localhost"
				} else {
					ip = ips[i]
				}
			}
			setPointingVariable(varName, varValue, appID)
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
		"SINK_MODULE": "fits-redist",
	}
	common.AddTimeStamp("before-add-tasks")
	_, err = task.Add(body, hs, envVars)
	return errors.WrapE(err, "add-task",
		"task-body", body, "headers", hs, "envs", envVars)
}
