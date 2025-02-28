package main

import (
	"encoding/json"
	"os"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Infoln("00, Entering message-router")
	if len(os.Args) < 3 {
		logrus.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
		os.Exit(1)
	}

	logrus.Infof("01, after number of arguments verification, message-body:%s,message-header:%s.\n",
		os.Args[1], os.Args[2])
	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &headers); err != nil {
		logrus.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	logrus.Infoln("02, after JSON format verification of headers")

	doMessageRoute := fromFuncs[headers["from_job"]]
	if doMessageRoute == nil {
		logrus.Warnf("from_job not set/not existed in message-router, from_job=%s ,message=%s\n",
			headers["from_job"], os.Args[1])
		os.Exit(4)
	}

	misc.AddTimeStamp("before-mr")
	logrus.Infoln("03, message-router not null")
	exitCode := doMessageRoute(os.Args[1], headers)
	if exitCode != 0 {
		logrus.Errorf("error found, error-code=%d\n", exitCode)
	}
	misc.AddTimeStamp("before-exit")
	os.Exit(exitCode)
}

var (
	fromFuncs = map[string]func(string, map[string]string) int{
		"":               defaultFunc,
		"message-router": fromMessageRouter,
		"down-sample":    fromDownSample,
		"fits-merge":     fromFitsMerge,
	}
)
