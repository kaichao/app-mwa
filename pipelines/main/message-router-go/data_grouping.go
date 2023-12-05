package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

// DataSet ...
type DataSet struct {
	// prefix ':' type ':' sub-id
	DatasetID string

	// for type "H", x-coord
	HorizontalWidth int

	// for type "V", y-coord
	VerticalStart  int
	VerticalHeight int
}

func parseDataSet(t string) *DataSet {
	var ds DataSet
	if err := json.Unmarshal([]byte(t), &ds); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-dataset definition
		return nil
	}
	return &ds
}

func initDataGrouping(dataset *DataSet) {
	fmtDatDataSet := ` {
		"datasetID":"dat:%s",
		"keyGroupRegex":"^([0-9]+)/[0-9]+_([0-9]+)_ch([0-9]{3}).dat$",
		"keyGroupIndex":"1,3,2",
		"sinkJob":"message-router-main",

		"groupType":"V",
		"verticalStart": %d,
		"verticalHeight": %d,
		"groupSize": %d,
		"interleaved":true
	}
	`
	numPerGroup := 10
	// 删除空格
	format := regexp.MustCompile("\\s+").ReplaceAllString(fmtDatDataSet, "")
	s := fmt.Sprintf(format, dataset.DatasetID, dataset.VerticalStart, dataset.VerticalHeight, numPerGroup)

	scalebox.AppendToFile("/work/messages.txt", "data-grouping-dat,"+s)
}
