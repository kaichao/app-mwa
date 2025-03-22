package main

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/message"
	"beamform/internal/pkg/semaphore"
	"beamform/internal/pkg/task"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

func defaultFunc(msg string, headers map[string]string) int {
	defer func() {
		misc.AddTimeStamp("leave-defaultFunc()")
	}()
	misc.AddTimeStamp("enter-defaultFunc()")

	cmd := "scalebox variable get datasets"
	val := misc.ExecCommandReturnStdout(cmd, 5)
	if val == "" {
		val = msg
	} else {
		val += "," + msg
	}
	cmd = "scalebox variable set datasets " + msg
	if code := misc.ExecCommandReturnExitCode(cmd, 5); code != 0 {
		return code
	}

	messages := message.GetMessagesForPullUnpack(msg)
	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
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
func fromBeamMake(message string, headers map[string]string) int {
	// message: 1257617424/p00049_00072/t1257617426_1257617505/ch111
	// sema: dat-done:1257010784/p00001_00960/t1257010786_1257010985/ch109
	re := regexp.MustCompile(`^(([0-9]+)/p([0-9]+)_[0-9]+)/(t[0-9]+_[0-9]+/ch[0-9]+)$`)
	ss := re.FindStringSubmatch(message)
	if len(ss) == 0 {
		logrus.Errorf("Invalid Message Format, body=%s\n", message)
		return 1
	}
	obsID := ss[2]
	// datasetID := ss[1]
	suffix := ss[4]

	p := ss[3]
	var p0, p1 string
	cmd := "scalebox variable get datasets"
	val := misc.ExecCommandReturnStdout(cmd, 5)
	re = regexp.MustCompile(`^[0-9]+/p([0-9]+)_([0-9]+)`)
	for _, ds := range strings.Split(val, ",") {
		ss := re.FindStringSubmatch(ds)
		if len(ss) == 0 {
			logrus.Errorf("Invalid Format of message, dataset=%s\n", ds)
			return 1
		}
		if ss[1] <= p && p <= ss[2] {
			p0 = ss[1]
			p1 = ss[2]
			break
		}
	}
	if p0 == "" {
		logrus.Errorln("dataset not found in variable datasets")
		return 2
	}
	// 用obsID，但可能有边界对齐问题？
	semaName := fmt.Sprintf("dat-done:%s/p%s_%s/%s", obsID, p0, p1, suffix)
	// 信号量操作
	v, err := semaphore.Decrement(semaName)
	if err != nil {
		logrus.Errorf("semaphore-decrement, err-info:%v\n", err)
		return 2
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
		if code := misc.ExecCommandReturnExitCode(cmd, 60); code != 0 {
			return code
		}
	}
	return task.Add("down-sample", message, "")
}

func fromDownSample(m string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// 产生hosts列表
	// dataset, p0, p1, t0, t1, err := message.ParseParts(m)
	dataset, _, _, t0, _, _ := message.ParseParts(m)
	cube := datacube.GetDataCube(dataset)
	nodes := cube.GetNodeNameListByTime(t0)
	hs := fmt.Sprintf(`{"target_hosts":"%s"}`, nodes)
	fmt.Printf("in fromDownSample(),hosts=%s\n", hs)
	code := task.Add("fits-redist", m, hs)
	fmt.Printf("Exit-code:%d\n", code)
	return code
}

func fromFitsRedist(message string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	re := regexp.MustCompile(`^(([0-9]+)/p([0-9]+)_([0-9]+)/(t[0-9]+_[0-9]+))(/ch[0-9]+)$`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", message)
		return 1
	}
	fmt.Println("message-parts:", ss)
	ds := ss[2]
	pBegin, _ := strconv.Atoi(ss[3])
	pEnd, _ := strconv.Atoi(ss[4])
	t := ss[5]

	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	sema := "fits-done:" + ss[1]
	semaVal, err := semaphore.Decrement(sema)
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
	for p := pBegin; p <= pEnd; p++ {
		toHost := cube.GetNodeNameByPointing(p)
		m := fmt.Sprintf(`%s/p%05d/%s,{"to_host":"%s"}`, ds, p, t, toHost)
		fmt.Println(m)
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
