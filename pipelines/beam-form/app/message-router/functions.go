package main

import (
	"beamform/internal/pkg/cache"
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/message"
	"beamform/internal/pkg/node"
	"beamform/internal/pkg/queue"
	"beamform/internal/pkg/semaphore"
	"fmt"
	"math/rand/v2"
	"net"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kaichao/gopkg/common"
	"github.com/kaichao/gopkg/exec"
	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

func defaultFunc(msg string, headers map[string]string) int {
	defer func() {
		misc.AddTimeStamp("leave-defaultFunc()")
	}()
	misc.AddTimeStamp("enter-defaultFunc()")

	cmd := "scalebox variable get datasets"
	val, err := exec.RunReturnStdout(cmd, 5)
	if err != nil {
		return 125
	}
	if val == "" {
		val = msg
	} else {
		val += "," + msg
	}
	cmd = "scalebox variable set datasets " + msg
	code, err := exec.RunReturnExitCode(cmd, 5)
	if err != nil {
		return 125
	}
	if code != 0 {
		return code
	}

	// host-bound
	messages := message.GetMessagesForPullUnpack(msg, true)
	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// 1266932744/p00001_00960/1266933866_1266933905_ch112.dat.tar.zst
	if code := task.AddTasks("pull-unpack", messages, "", 600); code > 0 {
		return code
	}
	sema := message.GetSemaphores(msg)
	if err := semaphore.Create(sema); err != nil {
		return 1
	}
	fmt.Printf("num-of-messages:%d,num-of-sema:%d\n", len(messages), len(sema))
	return 0
}

func fromPullUnpack(msg string, headers map[string]string) int {
	// input message: 1257617424/p00001_00096/1257617426_1257617465_ch112.dat.tar.zst
	// - target_dir:1257617424/t1257617426_1257617505/ch111
	// semaphore: dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109
	re := regexp.MustCompile(`^(([0-9]+)/p([0-9]+)_([0-9]+))/([0-9]+)_[0-9]+_(ch[0-9]+).dat.tar.zst$`)
	ss := re.FindStringSubmatch(msg)
	if len(ss) == 0 {
		logrus.Errorf("Invalid Message Format, body=%s\n", msg)
		return 1
	}
	prefix := ss[1]
	obsID := ss[2]
	p0, _ := strconv.Atoi(ss[3])
	p1, _ := strconv.Atoi(ss[4])
	t, _ := strconv.Atoi(ss[5])
	ch := ss[6]
	cube := datacube.GetDataCube(obsID)
	t0, t1 := cube.GetTimeRange(t)

	sema := fmt.Sprintf(`dat-ready:%s/t%d_%d/%s`, prefix, t0, t1, ch)
	semaVal, err := semaphore.Decrement(sema)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s\n", sema)
		return 2
	}
	if semaVal > 0 {
		return 0
	}
	ps := cube.GetPointingRangesByInterval(p0, p1)
	messages := []string{}
	for k := 0; k < len(ps); k += 2 {
		body := fmt.Sprintf("%s/p%05d_%05d/t%d_%d/%s",
			obsID, ps[k], ps[k+1], t0, t1, ch)
		messages = append(messages, body)
	}
	return task.AddTasks("beam-make", messages, "", 600)
}

func fromMessageRouter(message string, headers map[string]string) int {
	return 0
}
func fromBeamMake(m string, headers map[string]string) int {
	// message: 1257617424/p00049_00072/t1257617426_1257617505/ch111
	// sema: dat-done:1257010784/p00001_00960/t1257010786_1257010985/ch109
	obsID, p0, _, t0, t1, ch, err := message.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}
	suffix := fmt.Sprintf("t%d_%d/ch%d", t0, t1, ch)
	pStr := fmt.Sprintf("%05d", p0)

	// 考虑到环境变量等因素影响，找到准确的指向范围
	var ps0, ps1 string
	cmd := "scalebox variable get datasets"
	val, err := exec.RunReturnStdout(cmd, 5)
	if err != nil {
		return 125
	}
	re := regexp.MustCompile(`^[0-9]+/p([0-9]+)_([0-9]+)`)
	for _, ds := range strings.Split(val, ",") {
		ss := re.FindStringSubmatch(ds)
		if len(ss) == 0 {
			logrus.Errorf("Invalid Format of message, dataset=%s\n", ds)
			return 1
		}
		if ss[1] <= pStr && pStr <= ss[2] {
			ps0 = ss[1]
			ps1 = ss[2]
			break
		}
	}
	if ps0 == "" {
		logrus.Errorln("dataset not found in variable datasets")
		return 2
	}
	// 用obsID，但可能有边界对齐问题？
	semaName := fmt.Sprintf("dat-done:%s/p%s_%s/%s", obsID, ps0, ps1, suffix)
	// 信号量操作
	v, err := semaphore.Decrement(semaName)
	if err != nil {
		logrus.Errorf("semaphore-decrement, err-info:%v\n", err)
		return 3
	}
	// 若信号量为0，则删除dat文件目录（？）
	if v == 0 {
		ipAddr := headers["from_ip"]
		sshPort := os.Getenv("SSH_PORT")
		if sshPort == "" {
			sshPort = "22"
		}
		sshUser := os.Getenv("SSH_USER")
		if sshUser == "" {
			sshUser = "root"
		}
		cmd := fmt.Sprintf(`ssh -p %s %s@%s rm -rf /tmp/scalebox/mydata/mwa/dat/%s/%s`,
			sshPort, sshUser, ipAddr, obsID, suffix)
		fmt.Printf("cmd:%s\n", cmd)
		code, err := exec.RunReturnExitCode(cmd, 60)
		if err != nil {
			return 125
		}
		if code != 0 {
			return code
		}
	}
	return task.Add("down-sample", m, "")
}

func fromDownSample(m string, headers map[string]string) int {
	// 获取24个指向的对应的IP地址，
	// 0.
	// 1. 从队列中读取24个消息，分配给给相关指向；（类型1）
	// 2. 若有部分指向未有对应消息，则分发给计算组内IP地址（类型2/类型3）
	// 3. 写共享变量pointing-data-root
	// 4. 完成target_hosts的数据采集，向fits-redist发送task对应消息

	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// 产生hosts列表
	dataset, p0, p1, t0, _, _, err := message.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}
	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	appID := cache.GetAppIDByJobID(jobID)

	cube := datacube.GetDataCube(dataset)

	ips := node.GetIPAddrListByTime(cube, t0)

	fromIP := headers["from_ip"]

	// 读取共享变量表。
	varValues := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:p%05d", p)
		if v, err := variable.Get(varName, appID); err != nil {
			logrus.Errorf("variable-get %s, err-info:%v\n", varName, err)
			varValues = append(varValues, "")
		} else {
			varValues = append(varValues, v)
		}
	}

	toIPs := []string{}
	// varValues all empty string ""
	if slices.IndexFunc(varValues, func(s string) bool { return s != "" }) == -1 {
		fmt.Println("AAA")
		// create variables
		list, err := queue.PopN(p1 - p0 + 1)
		if err != nil {
			fmt.Printf("Queue pop error, err-info:%v\n", err)
			logrus.Errorf("Queue pop error, err-info:%v\n", err)
			return 2
		}

		for p := p0; p <= p1; p++ {
			i := p - p0
			varName := fmt.Sprintf("pointing-data-root:p%05d", p)
			var varValue, ip string
			if i < len(list) {
				// 类型1
				ip = list[i]
				varValue = list[i]
			} else if ips[i] == fromIP {
				// 类型2、类型3
				ip = "localhost"
				varValue = "/raid0/scalebox/mydata/mwa"
			} else {
				// 类型2、类型3
				ip = ips[i]
				varValue = "/raid0/scalebox/mydata/mwa"
			}
			variable.Set(varName, varValue, appID)
			toIPs = append(toIPs, ip)
		}
	} else {
		// 从共享变量表中读取
		// 若是远端IP（类型1）
		for p := p0; p <= p1; p++ {
			i := p - p0
			ip := net.ParseIP(varValues[i])
			if ip != nil && ip.To4() != nil {
				// ipv4 addr
				toIPs = append(toIPs, varValues[i])
			} else if ips[i] == fromIP {
				toIPs = append(toIPs, "localhost")
			} else {
				toIPs = append(toIPs, ips[i])
			}
		}
	}

	hs := fmt.Sprintf(`{"target_hosts":"%s"}`, strings.Join(toIPs, ","))
	code := task.Add("fits-redist", m, hs)
	return code
}

// WeightedItem ...
type WeightedItem struct {
	Value  string
	Weight float64
}

// WeightedRandomChoice ...
func WeightedRandomChoice(items []WeightedItem) string {
	var total float64
	var maxItem WeightedItem

	for _, item := range items {
		total += item.Weight
		if item.Weight > maxItem.Weight {
			maxItem = item
		}
	}

	r := rand.Float64() * total
	for _, item := range items {
		if r < item.Weight {
			return item.Value
		}
		r -= item.Weight
	}

	// fallback: 返回权值最大项
	return maxItem.Value
}

func fromFitsRedist(m string, headers map[string]string) int {
	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	appID := cache.GetAppIDByJobID(jobID)

	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	ds, p0, p1, t0, t1, _, err := message.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}

	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	semaName := fmt.Sprintf("fits-done:%s/p%05d_%05d/t%d_%d",
		ds, p0, p1, t0, t1)
	semaVal, err := semaphore.Decrement(semaName)
	if err != nil {
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	cube := datacube.GetDataCube(ds)
	// output message: 1257010784/p00023/t1257010786_1257010965
	messages := []string{}
	for p := p0; p <= p1; p++ {
		varName := fmt.Sprintf("pointing-data-root:p%05d", p)
		varValue, err := variable.Get(varName, appID)
		if err != nil {
			logrus.Errorf("variable-get, err-info:%v\n", err)
			return 11
		}

		headers := ""
		toHost := node.GetNodeNameByPointingTime(cube, p, t0)
		if ip := net.ParseIP(varValue); ip != nil && ip.To4() != nil {
			// IPv4地址（类型1）
			// 设置"to_ip"头
			headers = common.SetJSONAttribute(headers, "to_ip", varValue)
		} else if strings.Contains(varValue, "@") {
			// 远端存储（类型3）
			headers = common.SetJSONAttribute(headers, "to_host", toHost)
			// 24ch存放在/dev/shm
		} else {
			// 共享存储（类型2）
			headers = common.SetJSONAttribute(headers, "to_host", toHost)
			// 24ch存放在共享存储
		}
		m := fmt.Sprintf(`%s/p%05d/t%d_%d,%s`, ds, p, t0, t1, headers)
		fmt.Printf("var-value:%s,to-host:%s,headers=%s,m=%s\n",
			varValue, toHost, headers, m)
		messages = append(messages, m)
	}
	return task.AddTasks("fits-merge", messages, "", 600)
}

func fromFitsMerge(message string, headers map[string]string) int {
	jobID, _ := strconv.Atoi(os.Getenv("JOB_ID"))
	appID := cache.GetAppIDByJobID(jobID)

	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p([0-9]+))(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", message)
		return 1
	}

	varName := fmt.Sprintf("pointing-data-root:p%s", ss[2])
	varValue, err := variable.Get(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}
	if strings.Contains(varValue, "@") {
		// 共享变量pointing-data-root，若为类型3，给fits-push发消息，推送到远端ssh存储
		m := ""
		return task.AddTasks("fits-merge", []string{m}, "", 600)
	}

	// semaphore: pointing-ready:1257010784/p00001
	sema := "pointing-done:" + ss[1]
	semaVal, err := semaphore.Decrement(sema)
	if err != nil {
		// error while decrement semaphore
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	// 给presto-search流水线发消息

	return 0
}

func fromFitsPush(message string, headers map[string]string) int {
	// 信号量pointing-done的操作

	// 给presto-search流水线发消息

	return 0
}
