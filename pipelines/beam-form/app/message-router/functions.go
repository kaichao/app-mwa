package main

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/message"
	"beamform/internal/pkg/node"
	"beamform/internal/pkg/semaphore"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kaichao/gopkg/exec"
	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/kaichao/scalebox/pkg/task"
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
	// 1. 从队列中读取24个消息，分配给给相关指向；（类型1）
	// 2. 若有部分指向未有对应消息，则分发给计算组内IP地址（类型2/类型3）
	// 3. 写共享变量pointing-data-root
	// 4. 完成target_hosts的数据采集，向fits-redist发送task对应消息

	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// 产生hosts列表
	dataset, _, _, t0, _, _, err := message.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}
	cube := datacube.GetDataCube(dataset)
	nodes := node.GetNodeNameListByTime(cube, t0)
	// local-ip-addr -> "localhost"
	fromIP := headers["from_ip"]
	ips := []string{}
	for _, s := range nodes {
		if s == fromIP {
			ips = append(ips, "localhost")
		} else {
			ips = append(ips, s)
		}
	}

	hs := fmt.Sprintf(`{"target_hosts":"%s"}`, strings.Join(ips, ","))
	code := task.Add("fits-redist", m, hs)
	return code
}

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
	semaVal, err := semaphore.Decrement(semaName)
	if err != nil {
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	// ds := fmt.Sprintf("%s/p%05d_%05d",)
	cube := datacube.GetDataCube(ds)
	// output message: 1257010784/p00023/t1257010786_1257010965
	messages := []string{}
	for p := p0; p <= p1; p++ {
		toHost := node.GetNodeNameByPointingTime(cube, p, t0)
		m := fmt.Sprintf(`%s/p%05d/t%d_%d,{"to_host":"%s"}`, ds, p, t0, t1, toHost)
		messages = append(messages, m)
	}
	return task.AddTasks("fits-merge", messages, "", 600)
}

func fromFitsMerge(message string, headers map[string]string) int {
	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", message)
		return 1
	}

	// 查共享变量pointing-data-root，若为类型3，给fits-push发消息，推送到远端ssh存储

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

	return 0
}

func fromFitsPush(message string, headers map[string]string) int {
	// 信号量pointing-done的操作，给presto-search流水线发消息

	return 0
}
