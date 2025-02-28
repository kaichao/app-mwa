package main

import (
	"beamform/internal/pkg/message"
	"fmt"
	"os"
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
	messages, semaFitsDone, semaPointingDone := message.ProcessForBeamMake(msg)
	misc.AppendToFile("/work/custom-out.txt",
		fmt.Sprintf("n_messages:%d,nsemaFitsDone:%d,nsemaPointingDone:%d\n",
			len(messages), len(semaFitsDone), len(semaPointingDone)))

	misc.AddTimeStamp("after-ProcessForBeamMake()")

	misc.AppendToFile("my-sema-fits-done.txt", semaFitsDone)
	cmd := `scalebox semaphore create --sema-file my-sema-fits-done.txt`
	if code := misc.ExecCommandReturnExitCode(cmd, 600); code != 0 {
		return code
	}
	// batchSize := 120
	// for i := 0; i < len(semaFitsDone); i += batchSize {
	// 	end := i + batchSize
	// 	if end > len(semaFitsDone) {
	// 		end = len(semaFitsDone)
	// 	}
	// 	cmd := fmt.Sprintf(fmtSemaCreate, strings.Join(semaFitsDone[i:end], ","))
	// 	if code := misc.ExecCommandReturnExitCode(cmd, 600); code != 0 {
	// 		return code
	// 	}
	// }

	misc.AddTimeStamp("after-semaFitsDone")
	misc.AppendToFile("my-sema-pointing-done.txt", semaPointingDone)
	cmd = `scalebox semaphore create --sema-file my-sema-pointing-done.txt`
	if code := misc.ExecCommandReturnExitCode(cmd, 600); code != 0 {
		return code
	}

	misc.AddTimeStamp("after-semaPointingDone")

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
	misc.AppendToFile(os.Getenv("WORK_DIR")+"/custom-out.txt", cmd)
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
