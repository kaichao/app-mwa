package main

import (
	"beamform/internal/datacube"
	"beamform/internal/strparse"
	"fmt"
	"os"

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
	obsID, _, _, t0, t1, ch, err := strparse.ParseParts(m)
	if err != nil {
		logrus.Errorf("Parse message, body=%s,err=%v\n", m, err)
		return 1
	}
	suffix := fmt.Sprintf("t%d_%d/ch%d", t0, t1, ch)

	// 用obsID，但可能有边界对齐问题？
	// semaName := fmt.Sprintf("dat-done:%s/p%05d_%05d/t%d_%d/ch%d", obsID, ps0, ps1, suffix)
	cubeID := headers["_cube_id"]
	semaName := fmt.Sprintf("dat-done:%s/ch%d", cubeID, ch)
	// 信号量操作
	v, err := semaphore.AddValue(semaName, appID, -1)
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
		code, err := exec.RunReturnExitCode(cmd, 60)
		if err != nil {
			logrus.Errorf("Remove dat dir, cmd=%s,err-info=%v\n", cmd, err)
			return 125
		}
		if code != 0 {
			return code
		}
	}

	common.AddTimeStamp("before-send-messages")
	return toDownSample(m, headers)
}

func toBeamMake(cubeID string, ch int, fromHeaders map[string]string) int {
	cube := datacube.NewDataCube(cubeID)
	ps := cube.GetPointingRangesByInterval(cube.PointingBegin, cube.PointingEnd)
	messages := []string{}
	for k := 0; k < len(ps); k += 2 {
		body := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/ch%d`,
			cube.ObsID, ps[k], ps[k+1], cube.TimeBegin, cube.TimeEnd, ch)
		// 加上排序标签
		/*
			if os.Getenv("POINTING_FIRST") == "yes" {
				body = fmt.Sprintf(`%s,{"sort_tag":"p%05d:t%d"}`,
					body, ps[k], t0)
			}
		*/
		messages = append(messages, body)
	}
	fmt.Printf("num-of-messages in fromPullUnpack():%d\n", len(messages))
	common.AddTimeStamp("before-send-messages")
	envVars := map[string]string{
		"SINK_JOB":        "beam-make",
		"TIMEOUT_SECONDS": "600",
	}
	headers := fmt.Sprintf(`{"_cube_id":"%s","_cube_index":"%s"}`,
		cubeID, fromHeaders["_cube_index"])
	return task.AddTasks(messages, headers, envVars)
}
