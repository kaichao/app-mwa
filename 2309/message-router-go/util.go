package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func appendToFile(fileName string, line string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open file %s error,err-info:%v\n", fileName, err)
		fmt.Fprintln(os.Stderr, os.Args)
		os.Exit(3)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	writer.WriteString(line + "\n")
	writer.Flush()
}

func execCommand(myCmd string) string {
	cmd := exec.Command("bash", "-c", myCmd)
	output, err := cmd.Output()
	logger.Infof("IN execCmd(), cmd=%s,stdout=%s\n", myCmd, string(output))
	if err != nil {
		logger.Errorf("ERROR in execCmd(): cmd=%s,err=%v\n", myCmd, err)
		return ""
	}
	// 删除尾部的\n
	return strings.Replace(string(output), "\n", "", -1)
}
