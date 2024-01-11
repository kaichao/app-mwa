package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

var (
	fromFuncs = map[string]func(string, map[string]string) int{
		"":         defaultFunc,
		"dir-list": fromDirList,
		// "dir-list":           fromDirListTest,
		"copy-unpack":        fromCopyUnpack,
		"cluster-copy-tar":   fromClusterCopyTar,
		"beam-maker":         fromBeamMaker,
		"fits-dist":          fromFitsDist,
		"fits-merger":        fromFitsMerger,
		"data-grouping-main": fromDataGroupingMain,
	}
)

func main() {
	logger.Infoln("00, Entering message-router")
	if len(os.Args) < 3 {
		logger.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
		os.Exit(1)
	}

	fmt.Println("arg0:", os.Args[0])
	fmt.Println("arg1:", os.Args[1])
	fmt.Println("arg2:", os.Args[2])

	logger.Infof("01, after number of arguments verification, message-body:%s,message-header:%s.\n",
		os.Args[1], os.Args[2])
	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &headers); err != nil {
		logger.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	logger.Infoln("02, after JSON format verification of headers")

	logger.Infoln("04, from-job not null")
	doMessageRoute := fromFuncs[headers["from_job"]]
	if doMessageRoute == nil {
		logger.Warnf("from_job not set in message-router, from_job=%s ,message=%s\n",
			headers["from_job"], os.Args[1])
		os.Exit(4)
	}

	logger.Infoln("05, message-router not null")
	exitCode := doMessageRoute(os.Args[1], headers)
	if exitCode != 0 {
		logger.Errorf("error found, error-code=%d\n", exitCode)
	}
	os.Exit(exitCode)
}

func defaultFunc(message string, headers map[string]string) int {
	fmt.Println("start-message:", os.Args[1])
	// 初始的启动消息（数据集ID）
	ss := strings.Split(os.Args[1], "~")
	if len(ss) != 3 {
		fmt.Fprintf(os.Stderr, "Invalid message format, msg-body:%s\n", os.Args[1])
		return 3
	}
	if dataset := parseDataSet(ss[2]); dataset == nil {
		fmt.Fprintf(os.Stderr, "Invalid dataset format, metadata:%s\n", ss[2])
		return 4
	} else {
		// metadata message
		initDataGrouping(dataset)

		createDatUsedSemaphores(dataset)

		createDatReadySemaphores(dataset)
		createFits1chReadySemaphores(dataset)
	}

	m := fmt.Sprintf("dir-list,%s~%s", ss[0], ss[1])
	scalebox.AppendToFile("/work/messages.txt", m)
	return 0
}
func fromCopyUnpack(message string, headers map[string]string) int {
	scalebox.AppendToFile("/work/messages.txt", "data-grouping-main,dat,"+message)
	return 0
}

func fromClusterCopyTar(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815_ch109.dat.zst.tar
	ss := regexp.MustCompile("ch([0-9]{3})").FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid message format, message=%s", message)
		return 91
	}
	fmt.Println("[INFO]input-message:", message)
	channel, _ := strconv.Atoi(ss[1])
	// ch := n - 109

	m := "/data/mwa/tar~" + message
	return sendChannelAwareMessage(m, "copy-unpack", channel)
}

func fromBeamMaker(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010795/00001/ch123.fits
	re := regexp.MustCompile("^([0-9]+/[0-9]+_[0-9]+)/[0-9]+/ch([0-9]{3}).fits$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromBeamMaker()\n", message)
	}
	sema := fmt.Sprintf("dat-used:%s/ch%s", ss[1], ss[2])
	n := countDown(sema)
	fmt.Printf("sema: %s,value:%d\n", sema, n)
	if n == 0 {
		removeDatFiles(sema)
	}

	if localMode {
		ss := strings.Split(message, "/")
		if len(ss) != 4 {
			fmt.Fprintf(os.Stderr, "invalid message format, message=%s \n", message)
		}
		nPointing, _ := strconv.Atoi(ss[2])
		fromIP := headers["from_ip"]
		fmt.Printf("n=%d,numNodesPerGroup=%d\n", nPointing, numNodesPerGroup)
		fmt.Printf("num of hosts=%d,index=%d\n", len(hosts), (nPointing-1)%numNodesPerGroup)
		toIP := hosts[(nPointing-1)%numNodesPerGroup]

		if fromIP != toIP {
			sinkJob := "fits-dist"
			format := "/dev/shm/scalebox/mydata/mwa/1ch~%s~root@%s/dev/shm/scalebox/mydata/mwa/1ch"
			m := fmt.Sprintf(format, message, toIP)
			cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, fromIP, m)
			fmt.Printf("cmdTxt:%s\n", cmdTxt)
			code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
			fmt.Printf("stdout for task-add:\n%s\n", stdout)
			fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
			return code
		}
	}
	sinkJob := "data-grouping-main"
	m := sinkJob + ",fits," + message
	scalebox.AppendToFile("/work/messages.txt", m)

	return 0
}

func removeDatFiles(sema string) {
	// 1257010784/1257010786_1257010795/109
	// dat-used:1257010784/1257010786_1257010815/ch114
	ss := regexp.MustCompile("[/_]").Split(sema, -1)
	ds := strings.Split(ss[0], ":")[1]
	beg, _ := strconv.Atoi(ss[1])
	end, _ := strconv.Atoi(ss[2])
	ch := ss[3]
	fmt.Println("sema:", sema)
	fmt.Printf("In removeDatFiles(),ds=%s,beg=%d,end=%d,ch=%s\n", ds, beg, end, ch)
	for i := beg; i <= end; i++ {
		fileName := fmt.Sprintf("mwa/dat/%s/%s_%d_%s.dat", ds, ds, i, ch)
		fmt.Printf(" file-name:%s\n", fileName)
		if localMode {
			cmdTxt := "ssh 10.11.16.79 rm -f /tmp/scalebox/mydata/" + fileName
			fmt.Println("cmd-text:", cmdTxt)
			code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
			fmt.Printf("stdout for rm-file:\n%s\n", stdout)
			fmt.Fprintf(os.Stderr, "stderr for rm-file:\n%s\n", stderr)
			if code != 0 {
				os.Exit(code)
			}
			cmdTxt = "ssh 10.11.16.80 rm -f /tmp/scalebox/mydata/" + fileName
			code, stdout, stderr = scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
			fmt.Printf("stdout for rm-file:\n%s\n", stdout)
			fmt.Fprintf(os.Stderr, "stderr for rm-file:\n%s\n", stderr)
			if code != 0 {
				os.Exit(code)
			}
		} else {
			cmdTxt := "rm -f /data/" + fileName
			fmt.Println("cmd-text:", cmdTxt)
			code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 10)
			fmt.Printf("stdout for rm-file:\n%s\n", stdout)
			fmt.Fprintf(os.Stderr, "stderr for rm-file:\n%s\n", stderr)
			if code != 0 {
				os.Exit(code)
			}
		}
	}

}
func fromFitsDist(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815/00001/ch129.fits
	sinkJob := "data-grouping-main"
	m := sinkJob + ",fits," + message
	scalebox.AppendToFile("/work/messages.txt", m)

	return 0
}

func fromFitsMerger(message string, headers map[string]string) int {
	return 0
}
