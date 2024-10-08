package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"mr/datacube"
)

func fromBeamMaker(message string, headers map[string]string) int {
	defer func() {
		AddTimeStamp("before-leave-fromBeamMaker()")
	}()
	// 1257010784/p00009/t1257010786_1257010845/ch111.fits
	re := regexp.MustCompile("^([0-9]+)/p([0-9]+)/t([0-9]+)_([0-9]+)/(ch([0-9]{3})).fits$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromBeamMaker()\n", message)
		return 3
	}
	cube := datacube.GetDataCube(ss[1])

	p, _ := strconv.Atoi(ss[2])
	tb, _ := strconv.Atoi(ss[3])
	te, _ := strconv.Atoi(ss[4])
	ch, _ := strconv.Atoi(ss[6])

	AddTimeStamp("before-sema-progress-counter")
	index := (ch - 109) % len(hosts)
	sema := fmt.Sprintf("progress-counter_beam-maker_s%02d:%s", index/24, hosts[index])

	countDown(sema)

	AddTimeStamp("after-sema-progress-counter")
	sema = getSemaDatProcessedName(cube, p, tb, ch)
	n := countDown(sema)
	fmt.Printf("In fromBeamMaker(),sema: %s,value:%d\n", sema, n)
	if n != 0 {
		AddTimeStamp("before-sendJobRefMessage()")
		// 该batch中还未处理完
		return sendJobRefMessage(message, make(map[string]string), "down-sampler")
	}

	AddTimeStamp("before-removeLocalDatFiles()")
	removeLocalDatFiles(sema)
	AddTimeStamp("after-removeLocalDatFiles()")

	// 数据删除，修改信号量值
	batchIndex := countDownSemaPointingBatchIndex(cube, tb, ch)
	fmt.Printf("In fromBeamMaker(),batch-index=%d\n", batchIndex)
	// index := cube.getPointingBatchIndex(p0)
	if batchIndex < 0 || batchIndex >= cube.GetNumOfPointingBatch() {
		// 数据已经全部处理完成，没有新的Batch
		fmt.Printf("In fromBeamMaker(),batch-index=%d,no-new data \n", batchIndex)
		return sendJobRefMessage(message, make(map[string]string), "down-sampler")
	}

	AddTimeStamp("before-reset-sema-dat-ready")
	// reset semaphore dat-ready(以TimeRange为单位)
	sema = getSemaDatReadyName(cube, tb, ch)
	fmt.Printf("In fromBeamMaker(), sema:%s,init-value:%d\n", sema, te-tb+1)
	createSemaphore(sema, te-tb+1)
	AddTimeStamp("after-reset-sema-dat-ready")

	//	reset local-tar-pull消息（以TimeUnit为单位）
	sortedTag := getSortedTagForDataPull(cube, tb, ch)

	fmt.Printf("In fromBeamMaker(),tb=%d,ch=%d,sortedTag:%s\n", tb, ch, sortedTag)

	tarr := cube.GetTimeUnitsWithinInterval(tb, te)
	for i := 0; i < len(tarr); i += 2 {
		t0 := tarr[i]
		t1 := tarr[i+1]
		fmtMessage := "%s/%d_%d_ch%d.dat.tar.zst"
		m := fmt.Sprintf(fmtMessage, ss[1], t0, t1, ch)
		// toLocalTarPull(m, headers)
		toPullUnpack(m, headers)
	}
	AddTimeStamp("before-sendJobRefMessage()")

	return sendJobRefMessage(message, make(map[string]string), "down-sampler")
}

func fromDownSampler(message string, headers map[string]string) int {
	defer func() {
		AddTimeStamp("before-leave-fromDownSampler()")
	}()
	// 1257010784/p00001/t1257010786_1257010795/ch123.fits.zst
	re := regexp.MustCompile("^([0-9]+)/p([0-9]+)/t([0-9]+)_[0-9]+/ch[0-9]+.fits.zst$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "invalid message format, message=%s \n", message)
		return 3
	}
	nPointing, _ := strconv.Atoi(ss[2])
	t, _ := strconv.Atoi(ss[3])
	fromIP := headers["from_ip"]
	fmt.Printf("n=%d,numNodesPerGroup=%d\n", nPointing, len(ips))
	fmt.Printf("num of hosts=%d,index=%d\n", len(ips), (nPointing-1)%len(ips))

	num := (nPointing - 1) % len(ips)
	if len(ips) > 24 {
		// 24的倍数，multi-block
		cube := datacube.GetDataCube(ss[1])
		num = cube.GetNumWithBlockID(t, (nPointing-1)%24)
	}
	toIP := ips[num]

	fmt.Printf("In fromDownSampler(), nPointing=%d, num=%d, t=%d\n", nPointing, num, t)

	AddTimeStamp("before-fits-redist")
	if fromIP != toIP {
		sinkJob := "fits-redist"
		// format := "root@%s/dev/shm/scalebox/mydata/mwa/1chx~%s~/dev/shm/scalebox/mydata/mwa/1chx"
		// m := fmt.Sprintf(format, fromIP, message)
		// cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --to-ip %s %s", sinkJob, toIP, m)
		prefix := os.Getenv("DEFAULT_USER") + "@" + fromIP
		if os.Getenv("FITS_REDIST_MODE") == "RSYNC" {
			prefix = fmt.Sprintf("rsync://root@%s:50873", fromIP)
		} else {
			prefix = fmt.Sprintf("cstu0036@%s:50022", fromIP)
		}
		sourceURL := prefix + "/dev/shm/scalebox/mydata/mwa/1chx"
		hs := map[string]string{"source_url": sourceURL}

		AddTimeStamp("before-sendNodeAwareMessage()")
		return sendNodeAwareMessage(message, hs, sinkJob, num)

		// cmdTxt := fmt.Sprintf("scalebox task add --sink-job %s --header source_url=%s --to-ip %s %s",
		// 	sinkJob, sourceURL, toIP, message)
		// code, stdout, stderr := misc.ExecShellCommandWithExitCode(cmdTxt, 20)
		// fmt.Printf("stdout for task-add:\n%s\n", stdout)
		// fmt.Fprintf(os.Stderr, "stderr for task-add:\n%s\n", stderr)
		// return code
	}

	AddTimeStamp("before-toFitsMerger()")
	return toFitsMerger(message, headers)
}

func fromFitsRedist(message string, headers map[string]string) int {
	// message: 1257010784/1257010786_1257010815/00005/ch124.fits.zst
	return toFitsMerger(message, headers)
}

func toFitsMerger(message string, headers map[string]string) int {
	// message:
	// 		1257010784/p00001/t1257010786_1257010815/ch129.fits.zst
	re := regexp.MustCompile("^(([0-9]+)/p([0-9]{5})/t([0-9]+)_[0-9]+)/ch[0-9]{3}.fits.zst$")
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
		nPointing, _ := strconv.Atoi(ss[3])
		num := (nPointing - 1) % len(ips)
		if len(ips) > 24 {
			// 24倍数，multi-block
			cube := datacube.GetDataCube(ss[2])
			t, _ := strconv.Atoi(ss[4])
			num = cube.GetNumWithBlockID(t, (nPointing-1)%24)

			fmt.Printf("In toFitsMerger(), nPointing=%d, num=%d, t=%d\n", nPointing, num, t)
		}

		AddTimeStamp("before-sendNodeAwareMessage()")
		return sendNodeAwareMessage(ss[1], make(map[string]string), "fits-merger", num)
	}

	return 0
}

func fromFitsMerger(message string, headers map[string]string) int {
	defer func() {
		AddTimeStamp("before-leave-fromFitsMerger()")
	}()
	// message: 1257010784/p00022/t1257010786_1257010815
	re := regexp.MustCompile(`p([0-9]+)/`)
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "invalid message format in fromFitsMerger(), message=%s \n", message)
		return 3
	}
	pointing, _ := strconv.Atoi(ss[1])
	fmt.Printf("pointing:%d\n", pointing)
	m := fmt.Sprintf(`%s.fits.zst`, message)

	AddTimeStamp("before-sendJobRefMessage()")
	return sendJobRefMessage(m, make(map[string]string), "fits-24ch-push")
}
