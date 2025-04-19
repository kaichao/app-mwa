package main

import (
	"beamform/internal/pkg/message"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kaichao/gopkg/exec"
	"github.com/kaichao/scalebox/pkg/common"
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

	// host-bound==false
	messages := message.GetMessagesForPullUnpack(os.Args[1], false)
	for _, m := range messages {
		common.AppendToFile("my-messages.txt", m)
	}

	var headerOption string
	if v := os.Getenv("SOURCE_URL"); v != "" {
		headerOption = fmt.Sprintf("%s -h source_url=%s", headerOption, v)
	}
	if v := os.Getenv("TARGET_URL"); v != "" {
		headerOption = fmt.Sprintf("%s -h target_url=%s", headerOption, v)
	}
	cmd := fmt.Sprintf(`scalebox task add --sink-job=pull-unpack %s --task-file my-messages.txt`,
		headerOption)
	code, err := exec.RunReturnExitCode(cmd, 300)
	if err != nil {
		os.Exit(126)
	}

	os.Exit(code)
}
