package main

import (
	"beamform/internal/message"
	"fmt"
	"regexp"
	"strconv"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func defaultFunc(msg string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-defaultFunc()")
	}()
	common.AddTimeStamp("enter-defaultFunc()")
	// input message:
	// 	1257010784
	// 	1257010784/p00001_00960
	// 	1257010784/p00001_00960/t1257012766_1257012965
	// messages, semas := message.ParseForBeamMake(msg)
	common.AppendToFile("my-semas.txt", message.GetSemaphores(msg))
	if err := semaphore.CreateFileSemaphores("", appID, 100); err != nil {
		logrus.Errorf("semaphore-create,err-info:%v\n", err)
		return 1
	}
	common.AddTimeStamp("after-semaphores")

	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	messages := message.GetMessagesForBeamMake(msg)
	common.AppendToFile("custom-out.txt",
		fmt.Sprintf("n_messages:%d\n", len(messages)))

	envVars := map[string]string{
		"SINK_JOB":        "beam-make",
		"TIMEOUT_SECONDS": "1800",
	}
	return task.AddTasks(messages, "", envVars)
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
	sema := "fits-done:" + ss[1]
	v, err := semaphore.AddValue(sema, appID, -1)
	// semaValue, err := semaphore.Decrement(sema)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
	semaValue, _ := strconv.Atoi(v)
	if semaValue > 0 {
		// 24ch not done.
		return 0
	}

	// output message: 1257010784/p00023/t1257010786_1257010965
	messages := []string{}
	for p := pBegin; p <= pEnd; p++ {
		m := fmt.Sprintf("%s/p%05d/%s", ds, p, t)
		messages = append(messages, m)
	}

	envVars := map[string]string{
		"SINK_JOB": "fits-merge",
	}
	return task.AddTasks(messages, "", envVars)
}

func fromFitsMerge(message string, headers map[string]string) int {
	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", message)
		return 1
	}

	// semaphore: pointing-done:1257010784/p00001
	sema := "pointing-done:" + ss[1]
	v, err := semaphore.AddValue(sema, appID, -1)
	// semaValue, err := semaphore.Decrement(sema)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
	semaValue, _ := strconv.Atoi(v)
	if semaValue > 0 {
		// 24ch not done.
		return 0
	}

	return 0
}
