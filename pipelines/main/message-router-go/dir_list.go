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
	sinkJob := "copy-unpack"
	if !strings.HasPrefix(message, "/") {
		// remote file, copy to global storage
		sinkJob = "cluster-copy-tar"
		m = message + "~/data/mwa/tar"
		scalebox.AppendToFile("/work/messages.txt", sinkJob+","+m)
		return 0
	}

	ss := regexp.MustCompile("ch([0-9]{3})").FindStringSubmatch(message)
	if len(ss) != 2 {
		fmt.Fprintf(os.Stderr, "channel num not include in message:%s \n", message)
		os.Exit(1)
	}
	channel, _ := strconv.Atoi(ss[1])
	return sendChannelAwareMessage(m, sinkJob, channel)
}

func filterDataset(message string) bool {
	// 	/raid0/scalebox/mydata/mwa/tar~1257010784/1257010786_1257010815_ch120.dat.zst.tar
	re := regexp.MustCompile(".+~([0-9]+)/([0-9]+)_([0-9]+)_ch.+")
	ss1 := re.FindStringSubmatch(message)
	ds1 := ss1[1]
	begin1 := ss1[2]
	end1 := ss1[3]

	ss2, err := scalebox.GetTextFileLines(datasetFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err open file:%s\n", datasetFile)
	}
	for _, s := range ss2 {
		ss := strings.Split(s, ",")
		ds2 := ss[0]
		begin2 := ss[1]
		end2 := ss[2]
		if ds1 == ds2 && begin1 <= end2 && begin2 <= end1 {
			// interleaved
			return true
		}
	}
	return false
}
