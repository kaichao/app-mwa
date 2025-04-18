package queue

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/kaichao/gopkg/exec"
	"github.com/sirupsen/logrus"
)

const queueKey = "QUEUE_HOSTS"

func getRedisHostPort() (string, int) {
	redisHost := os.Getenv("REDIS_HOST")
	fmt.Printf("In package queue, get env-var REDIS_HOST:%s\n", os.Getenv("REDIS_HOST"))
	if redisHost == "" {
		// 用于单元测试
		redisHost = "10.0.6.100"
	}
	redisPort, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))
	if redisPort == 0 {
		redisPort = 6379
	}
	return redisHost, redisPort
}

// Push ...
func Push(item string, priority float32) error {
	redisHost, redisPort := getRedisHostPort()
	timestamp := time.Now().UnixMilli()
	cmd := fmt.Sprintf(`redis-cli -h %s -p %d ZADD %s %f %s:%d`,
		redisHost, redisPort, queueKey, priority, item, timestamp)
	code, _, stderr, err := exec.RunReturnAll(cmd, 5)
	if err != nil {
		return err
	}

	if code != 0 {
		errMsg := fmt.Sprintf("Error with exit-code:%d", code)
		logrus.Errorln("stderr:\n", stderr)
		return errors.New(errMsg)
	}
	return nil
}

// PopN ...
func PopN(num int) ([]string, error) {
	redisHost, redisPort := getRedisHostPort()
	cmd := fmt.Sprintf(`redis-cli -h %s -p %d ZPOPMIN %s %d`,
		redisHost, redisPort, queueKey, num)
	fmt.Printf("redis-cmd:%s\n", cmd)
	code, stdout, stderr, err := exec.RunReturnAll(cmd, 10)
	if err != nil {
		return []string{}, err
	}
	if code != 0 {
		errMsg := fmt.Sprintf("Error with exit-code:%d", code)
		logrus.Errorln("stderr:\n", stderr)
		return []string{}, errors.New(errMsg)
	}

	return getParsed(stdout), nil
}

// Query ...
func Query() error {
	redisHost, redisPort := getRedisHostPort()
	cmd := fmt.Sprintf(`redis-cli -h %s -p %d ZRANGE %s 0 -1`,
		redisHost, redisPort, queueKey)
	code, stdout, stderr, err := exec.RunReturnAll(cmd, 5)
	if err != nil {
		return err
	}

	if code != 0 {
		errMsg := fmt.Sprintf("Error with exit-code:%d", code)
		logrus.Errorln("stderr:\n", stderr)
		return errors.New(errMsg)
	}
	fmt.Println("result:", getParsed(stdout))
	return nil
}

// Clear ...
func Clear() error {
	redisHost, redisPort := getRedisHostPort()
	cmd := fmt.Sprintf(`redis-cli -h %s -p %d DEL %s`,
		redisHost, redisPort, queueKey)
	code, stdout, stderr, err := exec.RunReturnAll(cmd, 5)
	if err != nil {
		return err
	}
	fmt.Printf("code:%d,stdout:%s\n", code, stdout)
	if code != 0 {
		errMsg := fmt.Sprintf("Error with exit-code:%d", code)
		logrus.Errorln("stderr:\n", stderr)
		return errors.New(errMsg)
	}
	return nil
}

func getParsed(s string) []string {
	re := regexp.MustCompile(`(^|\n)(.+):`)
	matches := re.FindAllStringSubmatch(s, -1)
	var secondColumn []string
	for _, inner := range matches {
		secondColumn = append(secondColumn, inner[2])
	}
	return secondColumn
}
