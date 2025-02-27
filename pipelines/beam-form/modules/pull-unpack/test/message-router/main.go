package main

import (
	"beamform/internal/pkg/message"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 3 {
		logrus.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
		os.Exit(1)
	}

	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &headers); err != nil {
		logrus.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	if headers["from_job"] == "pull-unpack" {
		logrus.Printf("message from pull-unpack")
		os.Exit(0)
	}

	messages, _ := message.ProcessForPullUnpack(os.Args[1])
	for _, m := range messages {
		misc.AppendToFile("my_messages.txt", m)
	}
	var headerOption string
	if v := os.Getenv("SOURCE_URL"); v != "" {
		headerOption = fmt.Sprintf("%s -h source_url=%s", headerOption, v)
	}
	if v := os.Getenv("TARGET_URL"); v != "" {
		headerOption = fmt.Sprintf("%s -h target_url=%s", headerOption, v)
	}
	cmd := fmt.Sprintf(`scalebox task add --sink-job=pull-unpack %s --task-file my_messages.txt`,
		headerOption)
	code := misc.ExecCommandReturnExitCode(cmd, 300)
	os.Exit(code)
}
