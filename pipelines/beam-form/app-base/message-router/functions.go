package main

import (
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
	// input message:
	// 	1257010784
	// 	1257010784/p00001_00960
	// 	1257010784/p00001_00960/t1257012766_1257012965
	messages, semas := message.ParseForBeamMake(msg)
	misc.AppendToFile("custom-out.txt",
		fmt.Sprintf("n_messages:%d,num-of-semas:%d\n", len(messages), len(semas)))

	misc.AddTimeStamp("after-ProcessForBeamMake()")

	misc.AppendToFile("my-semas.txt", semas)
	cmd := `scalebox semaphore create --sema-file my-semas.txt`
	if code := misc.ExecCommandReturnExitCode(cmd, 600); code != 0 {
		return code
	}

	misc.AddTimeStamp("after-semaphores")

	for _, m := range messages {
		misc.AppendToFile("my-tasks.txt", m)
	}

	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	cmd = "scalebox task add --sink-job=beam-make --task-file my-tasks.txt"
	return misc.ExecCommandReturnExitCode(cmd, 3600)
}

func fromMessageRouter(message string, headers map[string]string) int {
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
	if s == "-32768" {
		// error while decrement semaphore
		return 2
	}
	if s != "0" {
		// 24ch not done.
		return 0
	}
	// output message: 1257010784/p00023/t1257010786_1257010965
	taskFile := "my-tasks.txt"
	for p := pBegin; p <= pEnd; p++ {
		m := fmt.Sprintf("%s/p%05d/%s", ds, p, t)
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
		return 2
	}
	if s != "0" {
		// pointing not done.
		return 0
	}

	return 0
}
