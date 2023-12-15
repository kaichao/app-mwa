package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

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
