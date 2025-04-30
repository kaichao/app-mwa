package main

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/message"
	"beamform/internal/pkg/node"
	"beamform/internal/pkg/queue"
	"fmt"
	"net"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/kaichao/gopkg/exec"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

func defaultFunc(msg string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-defaultFunc()")
	}()
	common.AddTimeStamp("enter-defaultFunc()")

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
	messages := []string{}
	for _, m := range message.GetMessagesForPullUnpack(msg, true) {
		parts := strings.SplitN(m, ",", 2)
		hs := common.SetJSONAttribute(parts[1], "source_url", sourcePicker.GetNext())
		// 交叉分布、首组限速
		messages = append(messages, fmt.Sprintf(`%s,%s`, parts[0], hs))
	}
	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// 1266932744/p00001_00960/1266933866_1266933905_ch112.dat.tar.zst

	envVars := map[string]string{
		"SINK_JOB":        "pull-unpack",
		"TIMEOUT_SECONDS": "600",
	}
	if code := task.AddTasks(messages, "{}", envVars); code > 0 {
		return code
	}

	common.AppendToFile("my-sema.txt", message.GetSemaphores(msg))
	if err := semaphore.CreateFileSemaphores("my-sema.txt", appID, 100); err != nil {
		logrus.Errorf("Create semaphores, err-info:%v\n", err)
		return 1
	}
	// if err := semaphore.Create(sema); err != nil {
	// 	return 1
	// }
	fmt.Printf("num-of-messages:%d\n", len(messages))
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
	v, err := semaphore.AddValue(sema, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s\n", sema)
		return 2
	}
	semaVal, _ := strconv.Atoi(v)
	if semaVal > 0 {
		return 0
	}
	ps := cube.GetPointingRangesByInterval(p0, p1)
	messages := []string{}
	for k := 0; k < len(ps); k += 2 {
		body := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/%s`,
			obsID, ps[k], ps[k+1], t0, t1, ch)
		// 加上排序标签
		if os.Getenv("POINTING_FIRST") == "yes" {
			body = fmt.Sprintf(`%s,{"sorted_tag":"p%05d:t%d"}`,
				body, ps[k], t0)
		}
		messages = append(messages, body)
	}
	fmt.Println("messages in fromPullUnpack():", messages)
	envVars := map[string]string{
		"SINK_JOB":        "beam-make",
		"TIMEOUT_SECONDS": "600",
	}
	return task.AddTasks(messages, "{}", envVars)
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
	v, err := semaphore.AddValue(semaName, appID, -1)
	// v, err := semaphore.Decrement(semaName)
	if err != nil {
		logrus.Errorf("semaphore-decrement, err-info:%v\n", err)
		return 3
	}
	// 若信号量为0，则删除dat文件目录（？）
	if v == "0" {
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

	envVars := map[string]string{
		"SINK_JOB": "down-sample",
	}
	return task.Add(m, "{}", envVars)
}

func fromDownSample(m string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	dataset, p0, p1, t0, _, _, err := message.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}

	cube := datacube.GetDataCube(dataset)
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
				varValue = targetPicker.GetNext()
				// varValue = weightedTargetSimple()
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
	code := task.Add(m, hs, envVars)
	return code
}

/*
func weightedTarget() string {
	jsonFile := fmt.Sprintf("/%s-target.json", os.Getenv("CLUSTER"))
	data, _ := os.ReadFile(jsonFile)
	m := map[string]float64{}
	json.Unmarshal(data, &m)

	var total, maxWeight float64
	var maxKey string
	for _, w := range m {
		total += w
	}

	r := rand.Float64() * total
	for k, w := range m {
		if w > maxWeight {
			maxWeight = w
			maxKey = k
		}
		if r < w {
			return k
		}
		r -= w
	}

	// fallback: 返回权重最大项
	return maxKey
}
*/
// 包级变量
/*
var (
	theoryPercent map[string]float64
	historyCounts map[string]int
	totalCount    int
)

func init() {
	jsonFile := fmt.Sprintf("/%s-target.json", os.Getenv("CLUSTER"))
	data, _ := os.ReadFile(jsonFile)
	weights := map[string]float64{}
	json.Unmarshal(data, &weights)

	theoryPercent = make(map[string]float64)
	historyCounts = make(map[string]int)
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	for key, w := range weights {
		theoryPercent[key] = w / totalWeight
		historyCounts[key] = 0
	}
}

func weightedTargetSimple() string {
	var firstKey string
	for key := range theoryPercent {
		firstKey = key
		if totalCount == 0 || float64(historyCounts[key])/float64(totalCount) < theoryPercent[key] {
			historyCounts[key]++
			totalCount++
			return key
		}
	}
	historyCounts[firstKey]++
	totalCount++
	return firstKey
}
*/

func fromFitsRedist(m string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	ds, p0, p1, t0, t1, _, err := message.ParseParts(m)
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

	envVars := map[string]string{
		"SINK_JOB":        "fits-merge",
		"TIMEOUT_SECONDS": "600",
	}
	return task.AddTasks(messages, "{}", envVars)
}

func fromFitsMerge(m string, headers map[string]string) int {
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

	return doCrossAppTaskAdd(ss[1])
}

func fromFitsPush(m string, headers map[string]string) int {
	// mwa/24ch/1257617424/p00021/t1257617426_1257617505.fits.zst
	re := regexp.MustCompile(`^mwa/24ch/([0-9]+/p[0-9]+)/t[0-9]+_[0-9]+`)
	ss := re.FindStringSubmatch(m)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", m)
		return 1
	}
	return doCrossAppTaskAdd(ss[1])
}

func doCrossAppTaskAdd(pointing string) int {
	// 信号量pointing-done的操作
	// semaphore: pointing-done:1257010784/p00001
	sema := "pointing-done:" + pointing
	v, err := semaphore.AddValue(sema, appID, -1)
	if err != nil {
		logrus.Errorf("error while decrement semaphore,sema=%s, err:%v\n",
			sema, err)
		return 1
	}
	semaVal, _ := strconv.Atoi(v)
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	varName := "pointing-data-root:" + pointing
	varValue, err := variable.Get(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}

	prestoAppID, err := strconv.Atoi(os.Getenv("PRESTO_APP_ID"))
	if err != nil {
		logrus.Errorln("no valid PRESTO_APP_ID")
		return 12
	}
	// IPv4地址（类型1）， 设置"to_ip"头
	headers := common.SetJSONAttribute("{}", "source_url", varValue)
	// 给presto-search流水线发消息
	envVars := map[string]string{
		"SINK_JOB": "message-router-presto",
		"JOB_ID":   "",
		"APP_ID":   fmt.Sprintf("%d", prestoAppID),
	}
	fmt.Printf("In doCrossAppTaskAdd(), env:APP_ID=%s, JOB_ID=%s, SINK_JOB=%s,GRPC_SERVER=%s\n",
		envVars["APP_ID"], envVars["JOB_ID"], envVars["SINK_JOB"], os.Getenv("GRPC_SERVER"))
	return task.Add(pointing, headers, envVars)
}
