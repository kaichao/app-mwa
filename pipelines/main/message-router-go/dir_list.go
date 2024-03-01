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
		sinkJob := "cluster-copy"
		m = message + "~/data/mwa/tar"
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+m)
		return 0
	}

	// /raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch111.dat.zst.tar
	ss := strings.Split(message, "~")
	if len(ss) != 2 {
		fmt.Fprintf(os.Stderr, "invalide message format, message:%s\n", message)
	}
	return toLocalPull(ss[1], headers)
}

func fromClusterCopy(message string, headers map[string]string) int {
	return toLocalPull(message, headers)
}

func toLocalPull(message string, headers map[string]string) int {
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

	prefix := "root@10.200.1.100/raid0/scalebox/mydata/mwa/tar~"
	suffix := "~/dev/shm/scalebox/mydata/mwa/tar"
	m := prefix + message + suffix

	h := make(map[string]string)
	h["sorted_tag"] = fmt.Sprintf("%06d", dataset.getSortedNumber(ts, channel, tStep))

	return sendNodeAwareMessage(m, h, "local-copy", channel-109)
}

func fromLocalCopy(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010815_ch109.dat.zst.tar
	re := regexp.MustCompile(`^([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]+)`)
	matches := re.FindStringSubmatch(message)

	fmt.Printf("message:%s, matches:%v\n", message, matches)

	if len(matches) < 4 {
		fmt.Fprintf(os.Stderr, "invalid message format, message:%s\n", message)
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
