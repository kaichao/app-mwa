package main

import (
	"beamform/internal/message"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/kaichao/gopkg/exec"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromBeamMake(m string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromBeamMake()")
	}()
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
		logrus.Errorf("m=%s,datasets=%s, dataset not found in variable datasets\n", m, val)

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

	common.AddTimeStamp("before-send-messages")
	envVars := map[string]string{
		"SINK_JOB": "down-sample",
	}
	return task.Add(m, "{}", envVars)
}
