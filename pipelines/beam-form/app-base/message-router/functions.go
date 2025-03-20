package main

import (
	"beamform/internal/pkg/message"
	"beamform/internal/pkg/semaphore"
	"beamform/internal/pkg/task"
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
	// messages, semas := message.ParseForBeamMake(msg)
	semas := message.GetSemaphores(msg)
	if err := semaphore.Create(semas); err != nil {
		logrus.Errorf("semaphore-create,errInfo:%v\n", err)
		logrus.Errorf("semaphore:\n%s\n", semas)
		return 1
	}
	misc.AddTimeStamp("after-semaphores")

	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	messages := message.GetMessagesForBeamMake(msg)
	misc.AppendToFile("custom-out.txt",
		fmt.Sprintf("n_messages:%d,num-of-semas:%d\n", len(messages), len(semas)))
	return task.AddTasks("beam-make", messages, "", 1800)
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
	semaValue, err := semaphore.Decrement(sema)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
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
	return task.AddTasks("fits-merge", messages, "", 120)
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
	semaValue, err := semaphore.Decrement(sema)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
	if semaValue > 0 {
		// 24ch not done.
		return 0
	}

	return 0
}
