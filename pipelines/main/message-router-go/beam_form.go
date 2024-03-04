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
	re := regexp.MustCompile("^([0-9]+)/p[0-9]+/(t[0-9]+_[0-9]+)/(ch([0-9]{3})).fits$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[WARN]message:%s not valid format in fromBeamMaker()\n", message)
		return 3
	}
	sema := fmt.Sprintf("dat-used:%s/%s/%s", ss[1], ss[2], ss[3])
	n := countDown(sema)
	fmt.Printf("sema: %s,value:%d\n", sema, n)
	if n == 0 {
		removeLocalDatFiles(sema)
	}

	ch, _ := strconv.Atoi(ss[4])
	return sendNodeAwareMessage(message, make(map[string]string), "down-sampler", ch-109)
}

func fromDownSampler(message string, headers map[string]string) int {
	// 1257010784/p00001/t1257010786_1257010795/ch123.fits.zst
	if !localMode {
		return toFitsMerger(message, headers)
	}

	re := regexp.MustCompile("^[0-9]+/p([0-9]+)/t[0-9]+_[0-9]+/ch[0-9]+.fits.zst$")
	ss := re.FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "invalid message format, message=%s \n", message)
		return 3
	}
	nPointing, _ := strconv.Atoi(ss[1])
	fromIP := headers["from_ip"]
	fmt.Printf("n=%d,numNodesPerGroup=%d\n", nPointing, numNodesPerGroup)
	fmt.Printf("num of hosts=%d,index=%d\n", len(hosts), (nPointing-1)%numNodesPerGroup)
	toIP := hosts[(nPointing-1)%numNodesPerGroup]

	if fromIP != toIP {
		sinkJob := "fits-redist"
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
	return sendNodeAwareMessage(message, make(map[string]string), "presto-search", pointing-1)
}
