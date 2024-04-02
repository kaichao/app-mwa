package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

func fromBeamMaker(message string, headers map[string]string) int {
	// 1257010784/p00009/t1257010786_1257010845/ch111.fits
	re := regexp.MustCompile("^([0-9]+)/p([0-9]+)/t([0-9]+)_([0-9]+)/(ch([0-9]{3})).fits$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromBeamMaker()\n", message)
		return 3
	}
	cube := getDataCube(ss[1])

	p, _ := strconv.Atoi(ss[2])
	tb, _ := strconv.Atoi(ss[3])
	te, _ := strconv.Atoi(ss[4])
	ch, _ := strconv.Atoi(ss[6])

	index := (ch - 109) % len(hosts)
	sema := "beam-maker-progress-count:" + hosts[index]
	countDown(sema)

	// p0, p1 := cube.getPointingBatchRange(p)
	// sema := fmt.Sprintf("dat-processed:%s/p%05d_%05d/t%s_%s/%s", ss[1], p0, p1, ss[3], ss[4], ss[5])
	sema = cube.getSemaDatProcessedName(p, tb, ch)
	n := countDown(sema)
	fmt.Printf("In fromBeamMaker(),sema: %s,value:%d\n", sema, n)
	if n != 0 {
		// 该batch中还未处理完
		return sendNodeAwareMessage(message, make(map[string]string), "down-sampler", ch-109)
	}

	removeLocalDatFiles(sema)

	// 数据删除，修改信号量值
	batchIndex := cube.countDownSemaPointingBatchIndex(tb, ch)
	fmt.Printf("In fromBeamMaker(),batch-index=%d\n", batchIndex)
	// index := cube.getPointingBatchIndex(p0)
	if batchIndex < 0 || batchIndex >= cube.getNumOfPointingBatch() {
		// 数据已经全部处理完成，没有新的Batch
		fmt.Printf("In fromBeamMaker(),batch-index=%d,no-new data \n", batchIndex)
		return sendNodeAwareMessage(message, make(map[string]string), "down-sampler", ch-109)
	}

	// reset semaphore dat-ready(以TimeRange为单位)
	sema = cube.getSemaDatReadyName(tb, ch)
	fmt.Printf("In fromBeamMaker(), sema:%s,init-value:%d\n", sema, te-tb+1)
	createSemaphore(sema, te-tb+1)

	//	reset local-tar-pull消息（以TimeUnit为单位）
	sortedTag := cube.getSortedTag(tb, ch)

	fmt.Printf("In fromBeamMaker(),tb=%d,ch=%d,sortedTag:%s\n", tb, ch, sortedTag)

	tarr := cube.getTimeUnitsWithinInterval(tb, te)
	for i := 0; i < len(tarr); i += 2 {
		t0 := tarr[i]
		t1 := tarr[i+1]
		fmtMessage := "%s/%d_%d_ch%d.dat.tar.zst"
		m := fmt.Sprintf(fmtMessage, ss[1], t0, t1, ch)
		toLocalTarPull(m, headers)
	}
	return sendNodeAwareMessage(message, make(map[string]string), "down-sampler", ch-109)
}

func fromDownSampler(message string, headers map[string]string) int {
	// 1257010784/p00001/t1257010786_1257010795/ch123.fits.zst
	// if !localMode {
	// 	return toFitsMerger(message, headers)
	// }

	re := regexp.MustCompile("^[0-9]+/p([0-9]+)/t[0-9]+_[0-9]+/ch[0-9]+.fits.zst$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "invalid message format, message=%s \n", message)
		return 3
	}
	nPointing, _ := strconv.Atoi(ss[1])
	fromIP := headers["from_ip"]
	fmt.Printf("n=%d,numNodesPerGroup=%d\n", nPointing, len(ips))
	fmt.Printf("num of hosts=%d,index=%d\n", len(ips), (nPointing-1)%len(ips))
	toIP := ips[(nPointing-1)%len(ips)]

	if fromIP != toIP {
		sinkJob := "fits-redist"
		// format := "/dev/shm/scalebox/mydata/mwa/1chx~%s~root@%s/dev/shm/scalebox/mydata/mwa/1chx"
		// m := fmt.Sprintf(format, message, toIP)
		// cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, fromIP, m)
		format := "root@%s/dev/shm/scalebox/mydata/mwa/1chx~%s~/dev/shm/scalebox/mydata/mwa/1chx"
		m := fmt.Sprintf(format, fromIP, message)
		cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toIP, m)
		code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdTxt, 20)
		fmt.Printf("stdout for task-add:\n%s\n", stdout)
		fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
		return code
	}
	return toFitsMerger(message, headers)
}

func fromFitsRedist(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815/00005/ch124.fits.zst
	return toFitsMerger(message, headers)
}

func toFitsMerger(message string, headers map[string]string) int {
	// input-message:
	// 		1257010784/p00001/t1257010786_1257010815/ch129.fits.zst
	re := regexp.MustCompile("^([0-9]+/p([0-9]{5})/t[0-9]+_[0-9]+)/ch[0-9]{3}.fits.zst$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in toFitsMerger()\n", message)
		return 1
	}
	// semaphore:
	// 		fits-24ch-ready:1257010784/p00029/t1257010786_1257010815
	sema := fmt.Sprintf("fits-24ch-ready:%s", ss[1])
	if n := countDown(sema); n == 0 {
		// 1257010784/1257010786_1257010815/00022
		pointing, _ := strconv.Atoi(ss[2])
		return sendNodeAwareMessage(ss[1], make(map[string]string), "fits-merger", pointing-1)
	}

	return 0
}

func fromFitsMerger(message string, headers map[string]string) int {
	// 1257010784/p00022/t1257010786_1257010815
	re := regexp.MustCompile(`p([0-9]+)/`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "invalid message format in fromFitsMerger(), message=%s \n", message)
		return 3
	}
	pointing, _ := strconv.Atoi(ss[1])
	fmt.Printf("pointing:%d\n", pointing)
	m := fmt.Sprintf(`/dev/shm/scalebox/mydata/mwa/24ch~%s.fits.zst~scalebox@159.226.237.136/raid0/scalebox/mydata/mwa/24ch`, message)
	return sendNodeAwareMessage(m, make(map[string]string), "fits-24ch-push", pointing-1)
}
