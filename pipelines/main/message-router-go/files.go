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
	// 	/raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch120.dat.tar.zst
	// /data/mwa/tar~1257010784/1257010786_1257010815_ch129.dat.tar.zst
	m := message
	if !strings.Contains(message, "~") {
		m = "/data/mwa/tar~" + message
	}
	if !filterDataCube(m) {
		// filtered
		return 0
	}

	if os.Getenv("JUMP_SERVERS") == "" && !strings.HasPrefix(message, "/") {
		// no jump servers && remote file, copy to global storage
		sinkJob := "cluster-tar-pull"
		m = message + "~/data/mwa/tar"
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+m)
		return 0
	}

	// remote cluster(with jump-servers)
	// 	message: <user>@<ip-addr>/raid0/tmp/mwa/tar1257010784~1257010784/1257010786_1257010815_ch109.dat.tar.zst
	// local cluster
	// 	message: /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch111.dat.tar.zst
	return toLocalTarPull(message, headers)
}

func fromClusterTarPull(message string, headers map[string]string) int {
	// message: 1257010784/1257010786_1257010815_ch111.dat.tar.zst
	return toLocalTarPull(message, headers)
}

func toLocalTarPull(message string, headers map[string]string) int {
	// FROM: remote cluster(with jump-servers)
	// 	message: <user>@<ip-addr>/raid0/tmp/mwa/tar1257010784~1257010784/1257010786_1257010815_ch109.dat.tar.zst
	// 	message: <user>@<ip-addr>/raid0/tmp/mwa/new-tar1257010784~1257010784/1257015316_1257015345_ch122.dat.tar.zst

	// FROM: beam-maker
	// message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst

	// FROM: dir-list && local
	// message: /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch111.dat.tar.zst

	// FROM: cluster-tar-pull
	// message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst

	ss := strings.Split(message, "~")
	// only packed file
	m := ss[len(ss)-1]

	fmt.Printf("to-local-pull,message:%s\n m:%s\n", message, m)

	// input-message:
	// 		1257010784/1257010786_1257010815_ch109.dat.tar.zst
	ss = regexp.MustCompile("([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]{3})").FindStringSubmatch(m)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid message format, message=%s", message)
		return 21
	}
	cube := getDataCube(ss[1])
	t0, _ := strconv.Atoi(ss[2])
	// b, e := datacube.getTimeRange(ts)
	ch, _ := strconv.Atoi(ss[3])

	batchIndex := cube.getSemaPointingBatchIndex(t0, ch)
	// 通过headers中的sorted_tag，设定显式排序
	h := map[string]string{"sorted_tag": cube.getSortedTag(t0, ch)}
	// h := make(map[string]string)
	// h["sorted_tag"] = cube.getSortedTag(batchIndex, ts, channel)

	suffix := "~/dev/shm/scalebox/mydata/mwa/tar"
	prefix := ""
	if os.Getenv("JUMP_SERVERS") != "" {
		// remote && jump-servers
		// 	message: <user>@<ip-addr>/raid0/tmp/mwa/tar1257010784~1257010784/1257010786_1257010815_ch109.dat.tar.zst
		m = message + suffix
	} else {
		prefix = getLocalRsyncPrefix()
		if strings.HasPrefix(message, "/") {
			// local
			// message: /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch111.dat.tar.zst
			ss := strings.Split(message, "~")
			m = ss[len(ss)-1]
			m = prefix + m + suffix
		} else {
			// from cluster-tar-pull
			// message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst
			m = prefix + message + suffix
		}
	}

	// add batch-index to message body.
	m = fmt.Sprintf("%s~%02d", m, batchIndex)
	return sendNodeAwareMessage(m, h, "local-tar-pull", ch-109)
}

func getLocalRsyncPrefix() string {
	cmdTxt := `scalebox cluster get-parameter rsync_info`
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
	fmt.Printf("stdout for get-cluster-parameter rsync_info:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for get-cluster-parameter rsync_info:\n%s\n", stderr)
	if code != 0 {
		return ""
	}
	ss := strings.Split(strings.TrimSpace(stdout), "#")
	sss := strings.Split(ss[0], ":")
	if len(ss) != 4 || len(sss) != 2 {
		fmt.Fprintf(os.Stderr, "Invalid return text from get-cluster-parameter rsync_info:\n%s\n", stdout)
		return ""
	}

	return fmt.Sprintf("%s%s/mwa/tar~", ss[3], sss[1])
}

func fromLocalTarPull(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815_ch109.dat.tar.zst
	re := regexp.MustCompile(`^([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+)`)
	matches := re.FindStringSubmatch(message)
	fmt.Printf("message:%s, matches:%v\n", message, matches)

	if len(matches) < 4 {
		fmt.Fprintf(os.Stderr, "Invalid message format, message:%s\n", message)
		return 1
	}
	ch, _ := strconv.Atoi(matches[3])
	cube := getDataCube(matches[1])
	t, _ := strconv.Atoi(matches[2])

	batchIndex := cube.getSemaPointingBatchIndex(t, ch)
	// // 通过headers中的sorted_tag，设定显式排序
	// h := map[string]string{"sorted_tag": cube.getSortedTag(batchIndex, t, ch)}
	h := map[string]string{}

	tb, te := cube.getTimeRange(t)
	m := fmt.Sprintf("%s~%d_%d~%02d", message, tb, te, batchIndex)
	fmt.Printf("In fromLocalTarPull(),batchIndex=%d, ch=%d, message:%s\n", batchIndex, ch, m)

	// add batch-index to message body.
	m = fmt.Sprintf("%s~%02d", m, batchIndex)
	return sendNodeAwareMessage(m, h, "unpack", ch-109)
}

func fromUnpack(message string, headers map[string]string) int {
	// 	1257010784/1257010784_1257010790_ch120.dat~01
	re := regexp.MustCompile("^([0-9]+)_([0-9]+)_ch([0-9]{3}).dat~(.+)$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromCopyUnpack()\n", message)
		return 11
	}

	// 1257010784_1257010790_ch112.dat
	cube := getDataCube(ss[1])
	if cube == nil {
		fmt.Fprintf(os.Stderr, "[WARN] unknown datacube:%s in fromCopyUnpack()\n", ss[1])
		return 12
	}

	t, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])
	tb, te := cube.getTimeRange(t)

	// sema := fmt.Sprintf("dat-ready:%s/t%d_%d/ch%s", ss[1], tb, te, ss[3])

	sema := cube.getSemaDatReadyName(t, ch)
	// 信号量dat-ready减1
	if n := countDown(sema); n != 0 {
		return 0
	}

	// 信号量dat-ready触发
	// sema = fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%s", ss[1], tb, te, ss[3])
	// n := countDown(sema)
	// batchIndex := cube.getNumOfPointingBatch() - n
	batchIndex := cube.getSemaPointingBatchIndex(t, ch)
	arr := cube.getPointingRangesByBatchIndex(batchIndex)
	// arr := cube.getPointingRangesByBatch(cube.getPointingBatchRange(p))

	fmt.Printf("In fromUnpack(), batch-index=%d,p-ranges:%v\n", batchIndex, arr)
	for i := 0; i < len(arr); i += 2 {
		p0 := arr[i]
		p1 := arr[i+1]

		m := fmt.Sprintf("%s/%d_%d/%s/%05d_%05d", ss[1], tb, te, ss[3], p0, p1)
		// batchIndex := getPointingBatchIndex(cube, tb, channel)
		// 通过headers中的sorted_tag，设定显式排序
		h := map[string]string{"sorted_tag": cube.getSortedTag(tb, ch)}
		ret := sendNodeAwareMessage(m, h, "beam-maker", ch-109)
		if ret != 0 {
			return ret
		}
	}

	return 0
}

func filterDataCube(message string) bool {
	// 	/raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch120.dat.tar.zst
	re := regexp.MustCompile(".+~([0-9]+)/([0-9]+)_([0-9]+)_ch.+")
	ss := re.FindStringSubmatch(message)
	datasetID := ss[1]
	begin1, _ := strconv.Atoi(ss[2])
	end1, _ := strconv.Atoi(ss[3])

	datacube := getDataCube(datasetID)
	begin2 := datacube.TimeBegin
	end2 := datacube.TimeBegin + datacube.NumOfSeconds - 1
	// interleaved
	return begin1 <= end2 && begin2 <= end1
}

func removeLocalDatFiles(sema string) int {
	// 1257010784/1257010786_1257010795/109
	// dat-processed:1257010784/t1257010786_1257010815/ch114
	// dat-processed:1257010784/p00001_000096/t1257010786_1257010815/ch114
	re := regexp.MustCompile("dat-processed:([0-9]+)/p.+/t([0-9]+)_([0-9]+)/(ch[0-9]+)")
	ss := re.FindStringSubmatch(sema)
	ds := ss[1]
	beg, _ := strconv.Atoi(ss[2])
	end, _ := strconv.Atoi(ss[3])
	ch := ss[4]
	fmt.Println("In removeLocalDatFiles(), sema:", sema)
	fmt.Printf("In removeDatFiles(),ds=%s,beg=%d,end=%d,ch=%s\n", ds, beg, end, ch)

	var cmdTxt string
	if localMode {
		dir := fmt.Sprintf("/tmp/scalebox/mydata/mwa/dat/%s/%s/%d_%d/", ds, ch, beg, end)
		num, _ := strconv.Atoi(ch[2:])
		i := (num - 109) % len(hosts)
		cmdTxt = fmt.Sprintf("ssh %s rm -rf %s", hosts[i], dir)
	} else {
		dir := fmt.Sprintf("/data/mwa/dat/%s/%s/%d_%d/", ds, ch, beg, end)
		cmdTxt = fmt.Sprintf("rm -rf %s", dir)
	}
	fmt.Println("cmd-text:", cmdTxt)
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 600)
	fmt.Printf("stdout for rm-dat-files:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for rm-dat-files:\n%s\n", stderr)

	return code
}
