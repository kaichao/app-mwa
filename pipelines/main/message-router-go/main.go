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
		"unpack":           fromUnpack,
		"cluster-copy-tar": fromClusterCopyTar,
		"beam-maker":       fromBeamMaker,
		"down-sampler":     fromDownSampler,
		"fits-dist":        fromFitsDist,
		"fits-merger":      fromFitsMerger,
	}
)

func main() {
	logger.Infoln("00, Entering message-router")
	if len(os.Args) < 3 {
		logger.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
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

	doMessageRoute := fromFuncs[headers["from_job"]]
	if doMessageRoute == nil {
		logger.Warnf("from_job not set/not existed in message-router, from_job=%s ,message=%s\n",
			headers["from_job"], os.Args[1])
		os.Exit(4)
	}

	logger.Infoln("03, message-router not null")
	exitCode := doMessageRoute(os.Args[1], headers)
	if exitCode != 0 {
		logger.Errorf("error found, error-code=%d\n", exitCode)
	}
	os.Exit(exitCode)
}

func defaultFunc(message string, headers map[string]string) int {
	// 初始的启动消息（数据集ID）
	// /raid0/scalebox/mydata/mwa/tar~1257010784
	ss := strings.Split(message, "~")
	if len(ss) != 2 {
		fmt.Fprintf(os.Stderr, "Invalid message format, msg-body:%s\n", message)
		return 3
	}
	dataset := getDataSet(ss[1])
	if dataset == nil {
		fmt.Fprintf(os.Stderr, "Invalid dataset format, metadata:%s\n", ss[2])
		return 4
	}

	createDatUsedSemaphores(dataset)
	createDatReadySemaphores(dataset)
	createFits24chReadySemaphores(dataset)

	m := fmt.Sprintf("dir-list,%s~%s", ss[0], ss[1])
	scalebox.AppendToFile("/work/messages.txt", m)
	return 0
}

func fromUnpack(message string, headers map[string]string) int {
	// 	1257010784/1257010784_1257010790_ch120.dat
	re := regexp.MustCompile("^([0-9]+)_([0-9]+)_ch([0-9]{3}).dat$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromCopyUnpack()\n", message)
		return 11
	}

	// 1257010784_1257010790_ch112.dat
	dataset := getDataSet(ss[1])
	if dataset == nil {
		fmt.Fprintf(os.Stderr, "[WARN] unknown dataset:%s in fromCopyUnpack()\n", ss[1])
		return 12
	}

	t, _ := strconv.Atoi(ss[2])
	t0, t1 := dataset.getTimeRange(t)
	sema := fmt.Sprintf("dat-ready:%s/%d_%d/%s", ss[1], t0, t1, ss[3])

	if n := countDown(sema); n == 0 {
		channel, _ := strconv.Atoi(ss[3])
		for b, e := range getPointingRanges() {
			m := fmt.Sprintf("%s/%d_%d/%s/%05d_%05d", ss[1], t0, t1, ss[3], b, e)
			ret := sendNodeAwareMessage(m, make(map[string]string), "beam-maker", channel-109)
			if ret != 0 {
				return ret
			}
		}
	}

	return 0
}

func fromClusterCopyTar(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815_ch109.dat.zst.tar
	ss := regexp.MustCompile("([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]{3})").FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid message format, message=%s", message)
		return 21
	}
	dataset := getDataSet(ss[1])
	ts, _ := strconv.Atoi(ss[2])
	b, e := dataset.getTimeRange(ts)
	channel, _ := strconv.Atoi(ss[3])

	m := fmt.Sprintf("/data/mwa/tar~%s~%d_%d", message, b, e)

	return sendNodeAwareMessage(m, make(map[string]string), "unpack", channel-109)
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
		removeLocalDatFiles(sema)
	}

	ch, _ := strconv.Atoi(ss[2])
	return sendNodeAwareMessage(message, make(map[string]string), "down-sampler", ch-109)
}

func fromDownSampler(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010795/00001/ch123.fits.zst
	if !localMode {
		return toFitsMerger(message, headers)
	}

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
		format := "/dev/shm/scalebox/mydata/mwa/1chx~%s~root@%s/dev/shm/scalebox/mydata/mwa/1chx"
		m := fmt.Sprintf(format, message, toIP)
		cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, fromIP, m)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 20)
		fmt.Printf("stdout for task-add:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
		return code
	}
	return toFitsMerger(message, headers)
}

func removeLocalDatFiles(sema string) {
	// 1257010784/1257010786_1257010795/109
	// dat-used:1257010784/1257010786_1257010815/ch114
	ss := regexp.MustCompile("[/_]").Split(sema, -1)
	ds := strings.Split(ss[0], ":")[1]
	beg, _ := strconv.Atoi(ss[1])
	end, _ := strconv.Atoi(ss[2])
	ch := ss[3]
	fmt.Println("sema:", sema)
	fmt.Printf("In removeDatFiles(),ds=%s,beg=%d,end=%d,ch=%s\n", ds, beg, end, ch)

	if localMode {
		dir := fmt.Sprintf("/tmp/scalebox/mydata/mwa/dat/%s/%s/%d_%d/", ds, ch, beg, end)
		num, _ := strconv.Atoi(ch[2:])
		i := (num - 109) % numNodesPerGroup
		fmt.Printf("\n")
		cmdTxt := fmt.Sprintf("ssh %s rm -rf %s", hosts[i], dir)
		fmt.Println("cmd-text:", cmdTxt)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
		fmt.Printf("stdout for rm-dat-files:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for rm-dat-files:\n%s\n", stderr)
		if code != 0 {
			os.Exit(code)
		}
	} else {
		dir := fmt.Sprintf("/data/mwa/dat/%s/%s/%d_%d/", ds, ch, beg, end)
		cmdTxt := fmt.Sprintf("rm -rf %s", dir)
		fmt.Println("cmd-text:", cmdTxt)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
		fmt.Printf("stdout for rm-dat-files:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for rm-dat-files:\n%s\n", stderr)
		if code != 0 {
			os.Exit(code)
		}
	}
}

func fromFitsDist(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815/00005/ch124.fits.zst
	return toFitsMerger(message, headers)
}

func toFitsMerger(message string, headers map[string]string) int {
	// input-message:
	// 		1257010784/1257010786_1257010815/00001/ch129.fits.zst
	re := regexp.MustCompile("^([0-9]+/[0-9]+_[0-9]+/([0-9]{5}))/ch[0-9]{3}.fits.zst$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in toFitsMerger()\n", message)
		return 1
	}
	// semaphore:
	// 		fits-24ch-ready:1257010784/1257010786_1257010815/00029
	sema := fmt.Sprintf("fits-24ch-ready:%s", ss[1])

	if n := countDown(sema); n == 0 {
		// 1257010784/1257010786_1257010815/00022
		pointing, _ := strconv.Atoi(ss[2])
		return sendNodeAwareMessage(ss[1], make(map[string]string), "fits-merger", pointing-1)
	}

	return 0
}

func fromFitsMerger(message string, headers map[string]string) int {
	// 1257010784/00022/1257010786_1257010815
	ss := strings.Split(message, "/")
	pointing, _ := strconv.Atoi(ss[1])
	fmt.Printf("pointing:%d\n", pointing)
	return sendNodeAwareMessage(message, make(map[string]string), "presto-search", pointing-1)
}
