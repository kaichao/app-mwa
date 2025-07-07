package main

import (
	"regexp"

	"github.com/sirupsen/logrus"
)

/*
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
			url := os.Getenv("SOURCE_TAR_ROOT")
			if url == "" {
				url = sourcePicker.GetNext()
			}

			fmt.Printf("message=%s,source_url=%s\n", m, url)

			hs := common.SetJSONAttribute(parts[1], "source_url", url)
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
*/
func fromFitsPush(m string, headers map[string]string) int {
	// mwa/24ch/1257617424/p00021/t1257617426_1257617505.fits.zst
	re := regexp.MustCompile(`^mwa/24ch/([0-9]+/p[0-9]+)/t[0-9]+_[0-9]+`)
	ss := re.FindStringSubmatch(m)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", m)
		return 1
	}
	return toCrossAppPresto(ss[1])
}
