package queue_test

import (
	"beamform/internal/queue"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func init() {
	os.Setenv("PATH", "/usr/local/bin:$PATH")
}
func TestPush(t *testing.T) {
	err := queue.Push("192.168.0.1", 1.0)
	if err != nil {
		logrus.Println(err)
	}
	err = queue.Push("192.168.0.2", 2.0)
	if err != nil {
		logrus.Println(err)
	}
}

func TestPopN(t *testing.T) {
	ips, err := queue.PopN(3)

	if err != nil {
		logrus.Println(err)
	} else {
		fmt.Println(ips)
	}
}

func TestQuery(t *testing.T) {
	if err := queue.Query(); err != nil {
		logrus.Errorf("err:%v\n", err)
	}
}

func TestClear(t *testing.T) {
	if err := queue.Clear(); err != nil {
		logrus.Errorf("err:%v\n", err)
	}
}
