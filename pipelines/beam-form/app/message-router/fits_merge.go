package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

func fromFitsMerge(m string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromFitsMerge()")
	}()
	// 1257010784/p00001/t1257010786_1257010965
	re := regexp.MustCompile(`^([0-9]+/p[0-9]+)(/t[0-9]+_[0-9]+)$`)
	ss := re.FindStringSubmatch(m)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", m)
		return 1
	}

	varName := fmt.Sprintf("pointing-data-root:%s", ss[1])
	varValue, err := variable.Get(varName, appID)
	if err != nil {
		logrus.Errorf("variable-get, err-info:%v\n", err)
		return 11
	}
	if strings.Contains(varValue, "@") {
		// 共享变量pointing-data-root，若为类型3，给fits-push发消息，推送到远端ssh存储
		msg := fmt.Sprintf("mwa/24ch/%s.fits.zst", m)
		headers := common.SetJSONAttribute("{}", "target_url", varValue)
		// headers = common.SetJSONAttribute("{}", "target_jump_servers", "root@10.200.1.100")

		envVars := map[string]string{
			"SINK_JOB": "fits-push",
		}
		return task.Add(msg, headers, envVars)
	}

	common.AddTimeStamp("before-send-messages")
	return doCrossAppTaskAdd(ss[1])
}
