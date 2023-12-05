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
		"dir-list":          fromDirList,
		"data-grouping-dat": fromDataGroupingDat,
		"beam-maker":        fromBeamMaker,
		"fits-merger":       fromFitsMerger,
	}

	currentDirections = "00001_00003"
)

func main() {
	logger.Infoln("00, Entering message-router")
	if len(os.Args) < 3 {
		logger.Fatalf("cmdline params: expected=2,actual=%d\n", len(os.Args))
		os.Exit(1)
	}

	logger.Infof("01, after number of arguments verification, message-body:%s,message-header:%s.\n",
		os.Args[1], os.Args[2])
	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &headers); err != nil {
		logger.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	logger.Infoln("02, after JSON format verification of headers")
	if headers["from_job"] == "" {
		// 初始的启动消息（数据集ID）
		scalebox.AppendToFile("/work/messages.txt", "dir-list,"+os.Args[1])
		os.Exit(0)
	}

	logger.Infoln("04, from-job not null")
	doMessageRoute := funcs[headers["from_job"]]
	if doMessageRoute == nil {
		logger.Warnf("from_job not set in message-router, from_job=%s ,message=%s\n", headers["from_job"], os.Args[1])
		os.Exit(3)
	}

	logger.Infoln("05, message-router not null")
	exitCode := doMessageRoute(os.Args[1], headers)
	if exitCode != 0 {
		logger.Errorf("error found, error-code=%d\n", exitCode)
	}
	os.Exit(exitCode)
}

func fromDirList(message string, params map[string]string) int {
	var m string
	if dataset := parseDataSet(message); dataset != nil {
		initDataGrouping(dataset)
	} else {
		// 文件项
		m = "data-grouping-dat,dat," + message
	}
	if m != "" {
		scalebox.AppendToFile("/work/messages.txt", m)
		logger.Infof("11, message emitted by dir-list :%s.\n", m)
	}
	return 0
}

func fromDataGroupingDat(message string, params map[string]string) int {
	//  input: 1257010784/1257010784_1257010790_ch132.dat,...,1257010784/1257010784_1257010799_ch132.dat
	//	output: 1257010784/1257010986_1257011185/132/00001_00003
	ms := strings.Split(message, ",")
	first := ms[0]
	last := ms[len(ms)-1]
	re := regexp.MustCompile("^([0-9]+)/[0-9]+_([0-9]+)_ch([0-9]{3}).dat$")
	ss := re.FindStringSubmatch(first)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Not valid message format, message:%s\n", message)
		return 99
	}
	ds := ss[1]
	start := ss[2]
	ch := ss[3]

	ss = re.FindStringSubmatch(last)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Not valid message format, message:%s\n", message)
		return 99
	}
	end := ss[2]
	m := fmt.Sprintf("%s/%s_%s/%s/%s", ds, start, end, ch, currentDirections)
	scalebox.AppendToFile("/work/messages.txt", "beam-maker,"+m)

	return 0
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
