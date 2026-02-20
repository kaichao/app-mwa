package main

import (
	"beamform/internal/datacube"
	"beamform/internal/node"
	"beamform/internal/strparse"
	"fmt"
	"os"
	"strconv"

	"github.com/kaichao/gopkg/exec"
	"github.com/kaichao/gopkg/logger"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromBeamMake(body string, headers map[string]string) int {
	// body: 1257617424/p00049_00072/t1257617426_1257617505/ch111
	// sema: dat-done:1257010784/p00001_00960/t1257010786_1257010985/ch109
	obsID, _, _, t0, t1, ch, err := strparse.ParseParts(body)
	if err != nil {
		logrus.Errorf("Parse task-body, body=%s,err=%v\n", body, err)
		return 1
	}

	// 用obsID，但可能有边界对齐问题？
	semaName := fmt.Sprintf("dat-done:%s/ch%d", headers["_vtask_cube_name"], ch)
	vtaskID, _ := strconv.ParseInt(headers["_vtask_id"], 10, 64)
	// 信号量操作
	v, err := semaphore.AddValue(semaName, vtaskID, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, err-info:%v\n", err)
		return 3
	}
	// 若信号量为0，则删除dat文件目录（？）
	if v <= 0 {
		ipAddr := headers["from_ip"]
		sshPort, _ := strconv.Atoi(os.Getenv("SSH_PORT"))
		if sshPort == 0 {
			sshPort = 22
		}
		sshUser := os.Getenv("SSH_USER")
		if sshUser == "" {
			sshUser = "root"
		}
		config := exec.SSHConfig{
			User:       sshUser,
			Host:       ipAddr,
			Port:       sshPort,
			Background: true,
		}
		subDatDir := fmt.Sprintf(`%s/t%d_%d/ch%d`, obsID, t0, t1, ch)
		cmd := fmt.Sprintf(`rm -rf %s/mydata/mwa/dat/%s`,
			os.Getenv("LOCAL_TMPDIR"), subDatDir)
		_, stdout, stderr, err := exec.RunSSHCommand(config, cmd, 30)
		if err != nil {
			logrus.Warnf("exec-cmd:%s\nstdout:\n%s\nstderr:\n%s\nerr-info:\n%v\n",
				cmd, stdout, stderr, err)
		}

		if globalDatDir := headers["_global_dat_dir"]; globalDatDir != "" {
			cmd := fmt.Sprintf(`rm -rf %s/dat/%s`,
				globalDatDir, subDatDir)
			_, stdout, stderr, err := exec.RunSSHCommand(config, cmd, 30)
			if err != nil {
				logrus.Warnf("exec-cmd:%s\nstdout:\n%s\nstderr:\n%s\nerr-info:\n%v\n",
					cmd, stdout, stderr, err)
			}

			// sub-path: 1302282040/t1302282041_1302282200/ch126
			if err := vPath.ReleasePath("global-dat", subDatDir); err != nil {
				logger.LogTracedErrorDefault(err)
			}
		}
	}

	return toDownSample(body, headers)
}

func toBeamMake(cubeID string, ch int, fromHeaders map[string]string) int {
	cube := datacube.NewDataCube(cubeID)
	ps := cube.GetPointingRangesByInterval(cube.PointingBegin, cube.PointingEnd)
	tasks := []string{}
	for k := 0; k < len(ps); k += 2 {
		body := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/ch%d`,
			cube.ObsID, ps[k], ps[k+1], cube.TimeBegin, cube.TimeEnd, ch)
		// 加上排序标签
		var sortTag string
		if os.Getenv("RUN_MODE") == "full_parallel" {
			sortTag = fmt.Sprintf("p%05d:t%d", ps[k], cube.TimeBegin)
		} else if len(node.Nodes) >= 24 {
			sortTag = fmt.Sprintf("t%d:p%05d", cube.TimeBegin, ps[k])
		} else {
			// 多个channel用单个节点计算
			sortTag = fmt.Sprintf("t%d:ch%d:p%05d", cube.TimeBegin, ch, ps[k])
		}
		headers := fmt.Sprintf(`{"sort_tag":"%s","_sort_tag":"%s"}`,
			sortTag, sortTag)

		globalDatDir := fromHeaders["_global_dat_dir"]
		if globalDatDir != "" {
			headers, _ = common.SetJSONAttribute(headers,
				"_global_dat_dir", globalDatDir)
		}

		line := body + "," + headers
		tasks = append(tasks, line)
	}

	common.AddTimeStamp("before-send-messages")
	envVars := map[string]string{
		"SINK_MODULE":     "beam-make",
		"TIMEOUT_SECONDS": "600",
	}
	headers := `{}`
	if _, err := task.AddTasks(tasks, headers, envVars); err != nil {
		logrus.Errorf("err:%v\n", err)
		return 1
	}
	return 0
}
