package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"mr/datacube"

	"github.com/kaichao/scalebox/pkg/misc"
)

func fromDirList(message string, headers map[string]string) int {
	// 	1257010784/1257010786_1257010815_ch120.dat.tar.zst
	if !filterDataCube(message) {
		// filtered
		return 0
	}
	hs := map[string]string{
		"source_url": "/cluster_data_root/mwa/tar",
	}
	// message: 1257010784/1257010786_1257010815_ch111.dat.tar.zst
	return toPullUnpack(message, hs)
}

/*
	func fromDirList(message string, headers map[string]string) int {
		// 	1257010784/1257010786_1257010815_ch120.dat.tar.zst
		if !filterDataCube(message) {
			// filtered
			return 0
		}

		AddTimeStamp("")
		m := fmt.Sprintf("%s~b%02d", message, getBatchIndex(message))
		if os.Getenv("ENABLE_CLUSTER_DIST") == "yes" {
			// 以共享存储中转
			hs := map[string]string{
				"source_url": headers["source_url"],
				"target_url": os.Getenv("CLUSTER_DATA_ROOT") + "/mwa/tar",
			}
			return sendJobRefMessage(m, hs, "cluster-dist")
		}

		hs := map[string]string{
			"source_url": headers["source_url"],
		}
		return toPullUnpack(m, hs)
	}
*/
func fromClusterDist(message string, headers map[string]string) int {
	hs := map[string]string{
		"source_url": "/cluster_data_root/mwa/tar",
	}
	// message: 1257010784/1257010786_1257010815_ch111.dat.tar.zst
	return toPullUnpack(message, hs)
}

func toPullUnpack(message string, headers map[string]string) int {
	// CASE 1: FROM: remote cluster(with jump-servers)
	// 	message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst
	// 	message: 1257010784/1257015316_1257015345_ch122.dat.tar.zst

	// CASE 2: FROM: beam-maker
	// message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst

	// CASE 3: FROM: dir-list && local
	// source_url: astro@10.100.1.30:10022:/data1/mydata/mwa/tar
	// message: 1257010784/1257010786_1257010815_ch111.dat.tar.zst

	// CASE 4: FROM: cluster-dist
	// source_url: /work1/cstu0036/mydata/mwa/tar
	// message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst

	// only packed file
	// m := strings.Split(message, "~")[0]

	// input-message:
	// 		1257010784/1257010786_1257010815_ch109.dat.tar.zst~b00
	ss := regexp.MustCompile("([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]{3})").
		FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid message format, message=%s", message)
		return 21
	}
	cube := datacube.GetDataCube(ss[1])
	t0, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])

	tb, te := cube.GetTimeRange(t0)
	targetDir := fmt.Sprintf("/tmp/scalebox/mydata/mwa/dat/%s/ch%d/%d_%d",
		ss[1], ch, tb, te)

	// hs := map[string]string{}
	headers["target_url"] = targetDir
	if os.Getenv("JUMP_SERVERS") != "" {
		headers["jump_servers"] = os.Getenv("JUMP_SERVERS")
	}
	// if headers["from_job"] == "beam-maker" {
	// CASE 2: FROM: beam-maker
	// message: 1257010784/1257010786_1257010815_ch109.dat.tar.zst

	// prefix := strings.Split(os.Getenv("DATASET_URI"), "~")[0]
	// // Replace the first occurrence of '/' with ':/'
	// modified := strings.Replace(prefix, "/", ":/", 1)
	// h["source_url"] = modified

	// 通过headers中的sorted_tag，设定显式排序
	headers["sorted_tag"] = getSortedTagForDataPull(cube, t0, ch)

	// add batch-index to message body.
	// batchIndex := getSemaPointingBatchIndex(cube, t0, ch)
	// m = fmt.Sprintf("%s~b%02d", m, batchIndex)

	num := cube.GetHostIndex(t0, ch-cube.ChannelBegin, len(ips))

	fmt.Printf("In toPullUnpack(), ch=%d, num=%d, t0=%d\n", ch, num, t0)
	return sendNodeAwareMessage(message, headers, "pull-unpack", num)
}

// 1257010784/1257010786_1257010815_ch109.dat.tar.zst
func getBatchIndex(packFile string) int {
	ss := regexp.MustCompile("([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]{3})").FindStringSubmatch(packFile)
	if ss == nil {
		return -1
	}
	cube := datacube.GetDataCube(ss[1])
	t0, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])
	return getSemaPointingBatchIndex(cube, t0, ch)
}

func fromPullUnpack(message string, headers map[string]string) int {
	// 	1257010784/1257010784_1257010790_ch120.dat~b01
	re := regexp.MustCompile("^([0-9]+)_([0-9]+)_ch([0-9]{3}).dat~(.+)$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromCopyUnpack()\n", message)
		return 11
	}

	// 1257010784_1257010790_ch112.dat
	cube := datacube.GetDataCube(ss[1])
	if cube == nil {
		fmt.Fprintf(os.Stderr, "[WARN] unknown datacube:%s in fromCopyUnpack()\n", ss[1])
		return 12
	}

	t, _ := strconv.Atoi(ss[2])
	ch, _ := strconv.Atoi(ss[3])
	tb, te := cube.GetTimeRange(t)

	index := cube.GetHostIndex(t, ch-cube.ChannelBegin, len(ips))
	// index := (ch - 109) % len(hosts)
	sema := fmt.Sprintf("progress-counter_pull-unpack:%s", hosts[index])
	countDown(sema)

	sema = getSemaDatReadyName(cube, t, ch)
	AddTimeStamp("before-sema-dec")
	// 信号量dat-ready减1
	if n := countDown(sema); n > 0 {
		misc.AppendToFile(os.Getenv("WORK_DIR")+"/extra-attributes.txt", fmt.Sprintf("sema:%d", n))
		// 该group未全部就绪
		return 0
	} else if n < 0 {
		// 出错
		return 1
	}

	// 单批次完成，信号量dat-ready触发
	batchIndex := getSemaPointingBatchIndex(cube, t, ch)
	batchIndex = 0
	arr := cube.GetPointingRangesByBatchIndex(batchIndex)

	fmt.Printf("001:message=%s,ds=%s,t=%d, ch=%v\n", message, ss[1], t, ch)
	fmt.Printf("In fromUnpack(), batch-index=%d,p-ranges:%v\n", batchIndex, arr)
	fileName := os.Getenv("WORK_DIR") + "/task-body.txt"
	for i := 0; i < len(arr); i += 2 {
		p0 := arr[i]
		p1 := arr[i+1]

		m := fmt.Sprintf("%s/%d_%d/%s/%05d_%05d", ss[1], tb, te, ss[3], p0, p1)
		// 通过headers中的sorted_tag，设定显式排序
		line := fmt.Sprintf(`%s,{"sorted_tag":"%s"}`, m, getSortedTagForBeamForm(cube, tb, p0, ch))
		misc.AppendToFile(fileName, line)
		// h := map[string]string{"sorted_tag": getSortedTagForBeamForm(cube, tb, p0, ch)}
		// ret := sendJobRefMessage(m, h, "beam-maker")
		// if ret != 0 {
		// 	return ret
		// }
	}
	cmdTxt := "scalebox task add --sink-job beam-maker"
	code := misc.ExecCommandReturnExitCode(cmdTxt, 120)
	return code
}

func filterDataCube(message string) bool {
	// 	1257010784/1257010786_1257010815_ch120.dat.tar.zst
	re := regexp.MustCompile("([0-9]+)/([0-9]+)_([0-9]+)_ch.+")
	ss := re.FindStringSubmatch(message)
	datasetID := ss[1]
	begin1, _ := strconv.Atoi(ss[2])
	end1, _ := strconv.Atoi(ss[3])

	datacube := datacube.GetDataCube(datasetID)
	begin2 := datacube.TimeBegin
	end2 := datacube.TimeBegin + datacube.NumOfSeconds - 1
	// interleaved
	return begin1 <= end2 && begin2 <= end1
}

func removeLocalDatFiles(sema string) int {
	// 1257010784/1257010786_1257010795/109
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
	dir := fmt.Sprintf("/tmp/scalebox/mydata/mwa/dat/%s/%s/%d_%d/", ds, ch, beg, end)
	num, _ := strconv.Atoi(ch[2:])
	cube := datacube.GetDataCube(ss[1])
	num = cube.GetHostIndex(beg, num-cube.ChannelBegin, len(ips))
	fmt.Printf("In removeLocalDatFiles(), ch=%s, num=%d, t0=%d\n", ch, num, beg)

	defaultUser := os.Getenv("DEFAULT_USER")
	sshPort := 50022
	cmdTxt = fmt.Sprintf("ssh -p %d %s@%s rm -rf %s", sshPort, defaultUser, ips[num], dir)
	fmt.Println("cmd-text:", cmdTxt)
	code, stdout, stderr := ExecWithRetries(cmdTxt, 5, 30)
	fmt.Printf("stdout for rm-dat-files:\n%s\n", stdout)
	fmt.Fprintf(os.Stderr, "stderr for rm-dat-files:\n%s\n", stderr)

	return code
}
