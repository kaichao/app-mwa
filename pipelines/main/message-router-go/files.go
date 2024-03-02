package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

func fromDirList(message string, headers map[string]string) int {
	fmt.Println("message:", message)
	// 	/raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch120.dat.zst.tar
	// /data/mwa/tar~1257010784/1257010786_1257010815_ch129.dat.zst.tar
	m := message
	if !strings.Contains(message, "~") {
		m = "/data/mwa/tar~" + message
	}
	if !filterDataset(m) {
		// filtered
		return 0
	}

	if !strings.HasPrefix(message, "/") {
		// remote file, copy to global storage
		sinkJob := "cluster-tar-pull"
		m = message + "~/data/mwa/tar"
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+m)
		return 0
	}

	// /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch111.dat.zst.tar
	ss := strings.Split(message, "~")
	if len(ss) != 2 {
		fmt.Fprintf(os.Stderr, "invalide message format, message:%s\n", message)
	}
	return toLocalTarPull(ss[1], headers)
}

func fromClusterTarPull(message string, headers map[string]string) int {
	return toLocalTarPull(message, headers)
}

func toLocalTarPull(message string, headers map[string]string) int {
	// message: 1257010784/1257010786_1257010815_ch109.dat.zst.tar

	fmt.Printf("to-local-pull,message:%s\n", message)

	// input-message:
	// 		1257010784/1257010786_1257010815/00001/ch129.fits.zst
	ss := regexp.MustCompile("([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]{3})").FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid message format, message=%s", message)
		return 21
	}
	dataset := getDataSet(ss[1])
	ts, _ := strconv.Atoi(ss[2])
	// b, e := dataset.getTimeRange(ts)
	channel, _ := strconv.Atoi(ss[3])

	// m = fmt.Sprintf("%s~%d_%d", m, b, e)
	cmdTxt := `scalebox cluster get-parameter rsync_info`
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
	fmt.Printf("stdout for get-cluster-parameter rsync_info:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for get-cluster-parameter rsync_info:\n%s\n", stderr)
	if code != 0 {
		return code
	}
	ss = strings.Split(strings.TrimSpace(stdout), "#")
	sss := strings.Split(ss[0], ":")
	if len(ss) != 4 || len(sss) != 2 {
		fmt.Fprintf(os.Stderr, "Invalid return text from get-cluster-parameter rsync_info:\n%s\n", stdout)
		return 1
	}

	prefix := fmt.Sprintf("%s%s/mwa/tar~", ss[3], sss[1])
	fmt.Println("prefix:", prefix)
	// prefix := "root@10.200.1.100/raid0/scalebox/mydata/mwa/tar~"
	suffix := "~/dev/shm/scalebox/mydata/mwa/tar"
	m := prefix + message + suffix

	h := make(map[string]string)
	h["sorted_tag"] = fmt.Sprintf("%06d", dataset.getSortedNumber(ts, channel, tStep))

	return sendNodeAwareMessage(m, h, "local-tar-pull", channel-109)
}

func fromLocalTarPull(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815_ch109.dat.zst.tar
	re := regexp.MustCompile(`^([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+)`)
	matches := re.FindStringSubmatch(message)

	fmt.Printf("message:%s, matches:%v\n", message, matches)

	if len(matches) < 4 {
		fmt.Fprintf(os.Stderr, "Invalid message format, message:%s\n", message)
		return 1
	}
	ch, _ := strconv.Atoi(matches[3])
	dataset := getDataSet(matches[1])
	ts, _ := strconv.Atoi(matches[2])
	b, e := dataset.getTimeRange(ts)
	m := fmt.Sprintf("%s~%d_%d", message, b, e)

	fmt.Printf("message:%s, matches:%v,channel:%d\n", m, matches, ch)

	return sendNodeAwareMessage(m, make(map[string]string), "unpack", ch-109)
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

// unpack的处理顺序编号
func (dataset *DataSet) getSortedNumber(t int, channel int, groupSize int) int {
	x := channel - 109
	y := (t - dataset.VerticalStart) / groupSize
	return y*24 + x
}

func filterDataset(message string) bool {
	// 	/raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch120.dat.zst.tar
	re := regexp.MustCompile(".+~([0-9]+)/([0-9]+)_([0-9]+)_ch.+")
	ss := re.FindStringSubmatch(message)
	datasetID := ss[1]
	begin1, _ := strconv.Atoi(ss[2])
	end1, _ := strconv.Atoi(ss[3])

	dataset := getDataSet(datasetID)
	begin2 := dataset.VerticalStart
	end2 := dataset.VerticalStart + dataset.VerticalHeight - 1
	// interleaved
	return begin1 <= end2 && begin2 <= end1
}

func removeLocalDatFiles(sema string) int {
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
		cmdTxt := fmt.Sprintf("ssh %s rm -rf %s", hosts[i], dir)
		fmt.Println("cmd-text:", cmdTxt)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
		fmt.Printf("stdout for rm-dat-files:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for rm-dat-files:\n%s\n", stderr)
		if code != 0 {
			return code
		}
	} else {
		dir := fmt.Sprintf("/data/mwa/dat/%s/%s/%d_%d/", ds, ch, beg, end)
		cmdTxt := fmt.Sprintf("rm -rf %s", dir)
		fmt.Println("cmd-text:", cmdTxt)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
		fmt.Printf("stdout for rm-dat-files:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for rm-dat-files:\n%s\n", stderr)
		if code != 0 {
			return code
		}
	}
	return 0
}
