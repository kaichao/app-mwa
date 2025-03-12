package main

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/message"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

func defaultFunc(msg string, headers map[string]string) int {
	defer func() {
		misc.AddTimeStamp("leave-defaultFunc()")
	}()
	misc.AddTimeStamp("enter-defaultFunc()")

	// messages, sema := message.ParseForPullUnpack(msg)
	messages := message.GetMessagesForPullUnpack(msg)
	sema := message.GetSemaphores(msg)
	misc.AppendToFile("my-sema.txt", sema)
	cmd := "scalebox semaphore create --sema-file my-sema.txt"
	if code := misc.ExecCommandReturnExitCode(cmd, 600); code > 0 {
		return code
	}
	fmt.Printf("num-of-messages:%d,num-of-sema:%d\n", len(messages), len(sema))
	for _, m := range messages {
		misc.AppendToFile("my-tasks.txt", m)
	}

	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	cmd = "scalebox task add --sink-job=pull-unpack --task-file my-tasks.txt"
	return misc.ExecCommandReturnExitCode(cmd, 600)
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
	cmd := fmt.Sprintf(`scalebox semaphore decrement %s`, sema)
	s := misc.ExecCommandReturnStdout(cmd, 600)
	semaVal, err := strconv.Atoi(s)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s\n", sema)
		return 2
	}
	if semaVal == 0 {
		ps := cube.GetPointingRangesByInterval(p0, p1)
		for k := 0; k < len(ps); k += 2 {
			body := fmt.Sprintf("%s/p%05d_%05d/t%d_%d/%s",
				obsID, ps[k], ps[k+1], t0, t1, ch)
			misc.AppendToFile("my-tasks.txt", body)
		}
		cmd = `scalebox task add --sink-job=beam-make --task-file my-tasks.txt`
		return misc.ExecCommandReturnExitCode(cmd, 10)
	}
	return 0
}

func fromMessageRouter(message string, headers map[string]string) int {
	return 0
}
func fromBeamMake(message string, headers map[string]string) int {
	// 信号量操作
	// 若信号量为0，则删除dat文件目录（？）
	return 0
}

func fromDownSample(message string, headers map[string]string) int {
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
	cmd := fmt.Sprintf("scalebox semaphore decrement fits-done:%s", ss[1])
	misc.AppendToFile("custom-out.txt", cmd)
	fmt.Printf("cmd=%s\n", cmd)
	s := misc.ExecCommandReturnStdout(cmd, 5)
	fmt.Printf("run-cmd,stdout=%s\n", s)
	if s == "-32768" {
		// error while decrement semaphore
		return 1
	}
	if s != "0" {
		// 24ch not done.
		return 0
	}
	// output message: 1257010784/p00023/t1257010786_1257010965
	taskFile := "my-tasks.txt"
	fmt.Println("000,task-file:", taskFile)
	for p := pBegin; p <= pEnd; p++ {
		m := fmt.Sprintf("%s/p%05d/%s", ds, p, t)
		fmt.Println("001,message:", m)
		misc.AppendToFile(taskFile, m)
	}
	cmd = "scalebox task add --sink-job=fits-merge --task-file=my-tasks.txt"
	code := misc.ExecCommandReturnExitCode(cmd, 120)
	return code
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
	cmd := fmt.Sprintf("scalebox semaphore decrement pointing-done:%s", ss[1])
	s := misc.ExecCommandReturnStdout(cmd, 5)
	if s == "-32768" {
		// error while decrement semaphore
		return 1
	}
	if s != "0" {
		// pointing not done.
		return 0
	}

	return 0
}
