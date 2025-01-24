package main

import (
	"beamform/internal/pkg/message"
	"fmt"
	"os"
	"strings"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 3 {
		logrus.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
		os.Exit(1)
	}

	messages := message.ProcessForPullUnpack(os.Args[1])
	fmt.Println("messages:")
	fmt.Println(messages)
	for _, m := range messages {
		ss := strings.Split(m, ",")
		cmd := fmt.Sprintf(`scalebox task add --sink-job=pull-unpack -h target_subdir=%s %s`,
			ss[1], ss[0])
		if code := misc.ExecCommandReturnExitCode(cmd, 10); code != 0 {
			os.Exit(code)
		}
	}
}
