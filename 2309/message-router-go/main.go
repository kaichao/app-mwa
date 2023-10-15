package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

var (
	funcs = map[string]func(string, map[string]string) int{
		"beam-maker":  fromBeamMaker,
		"fits-merger": fromFitsMerger,
	}
)

func main() {
	if len(os.Args) < 3 {
		logger.Fatalf("cmdline params: expected=2,actual=%d\n", len(os.Args))
		os.Exit(1)
	}

	params := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &params); err != nil {
		logger.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	if params["from_job"] == "" {
		logger.Warnf("from_job not set in task headers,message=%s\n", os.Args[1])
		os.Exit(0)
	}

	doMessageRoute := funcs[params["from_job"]]
	if doMessageRoute == nil {
		logger.Warnf("message-router not set,from_job=%s ,message=%s\n", params["from_job"], os.Args[1])
		os.Exit(3)
	}

	exitCode := doMessageRoute(os.Args[1], params)
	if exitCode != 0 {
		logger.Errorf("error found, error-code=%d\n", exitCode)
	}
	os.Exit(exitCode)
}

func fromBeamMaker(message string, params map[string]string) int {
	// 1257010784_1257010986_1257011185_132_001
	regex := regexp.MustCompile(`^([0-9]+)_([0-9]+_[0-9]+)_([0-9]+)_([0-9]+)$`)

	if strings.HasSuffix(message, ".fits") {
		// 非压缩fits文件
		scalebox.AppendToFile("/work/messages.txt", "fits2fil,"+dataRootMain+"/fits%"+message)
	} else if regex.MatchString(message) {
		// 压缩fits文件
		scalebox.AppendToFile("/work/messages.txt", "decompress,"+message)
	} else {
		logger.Errorf("File extension error in fromFitsPull(), text:'%s'\n", message)
		return 101
	}
	return 0
}

func fromFitsMerger(message string, params map[string]string) int {
	// appendToFile("/work/messages.txt", "fits2fil,"+dataRootMain+"/decompressed%"+message)
	return 0
}

// messager-router-prep --> messager-router-main
func fromMessageRouterPrep(message string, params map[string]string) int {
	// 时间戳写到cluster-main，用于画图
	// Dec+4352_12_05/20221202/Dec+4352_12_05_arcdrift-M01_0001.fil,2022-12-02-00:10:26
	re := regexp.MustCompile(`^(([^/]+)/.+),([0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{2}:[0-9]{2}:[0-9]{2})$`)
	matches := re.FindStringSubmatch(message)
	if matches != nil {
		filFile := matches[1]
		dataset := matches[2]
		ts := matches[3]
		tsFile := fmt.Sprintf("/local%s/fil/%s/timestamp.txt", dataRootMain, dataset)
		scalebox.AppendToFile(tsFile, filFile+" "+ts)
	}

	return 0
}
