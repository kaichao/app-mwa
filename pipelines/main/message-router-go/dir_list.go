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
	// sinkJob := "unpack"
	if !strings.HasPrefix(message, "/") {
		// remote file, copy to global storage
		sinkJob := "cluster-copy"
		m = message + "~/data/mwa/tar"
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+m)
		return 0
	}

	ss := regexp.MustCompile("([0-9]+)/([0-9]+)_[0-9]+_ch([0-9]{3})").FindStringSubmatch(message)
	if ss == nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Invalid message format, message=%s", message)
		return 21
	}
	dataset := getDataSet(ss[1])
	ts, _ := strconv.Atoi(ss[2])
	b, e := dataset.getTimeRange(ts)
	channel, _ := strconv.Atoi(ss[3])

	m = fmt.Sprintf("%s~%d_%d", m, b, e)

	h := make(map[string]string)
	h["sorted_tag"] = fmt.Sprintf("%06d", dataset.getSortedNumber(ts, channel, tStep))
	return sendNodeAwareMessage(m, h, "unpack", channel-109)
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
