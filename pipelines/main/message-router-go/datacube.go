package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

type dataCube struct {
	// prefix ':' type ':' sub-id
	DatasetID string

	// for type "H", x-coord
	// HorizontalWidth string

	// for type "V", y-coord
	// VerticalStart  string
	// VerticalHeight string

	ChannelBegin string
	ChannelSize  string

	TimeBegin  string
	TimeLength string
	// 30的倍数
	TimeStep string

	PointingBegin string
	PointingEnd   string
	// 通常为24
	PointingStep string
	NumPerBatch  string
}

// DataCube ...
type DataCube struct {
	// prefix ':' type ':' sub-id
	DatasetID string

	// for type "H", x-coord
	// HorizontalWidth int

	// for type "V", y-coord
	// VerticalStart  int
	// VerticalHeight int

	ChannelBegin int
	ChannelSize  int

	TimeBegin  int
	TimeLength int
	// 30的倍数
	TimeStep int

	PointingBegin int
	PointingEnd   int
	// 通常为24
	PointingStep int
	NumPerBatch  int
}

func getDataCube(datasetID string) *DataCube {
	cmdText := "scalebox dataset get-metadata " + datasetID
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Fprintf(os.Stderr, "stderr for dataset-get-metadata:\n%s\n", stderr)
	if code != 0 {
		fmt.Fprintf(os.Stderr, "[WARN] error for dataset-get-metadata dataset=%s in getDataCube()\n", datasetID)
		return nil
	}

	var ds dataCube
	if err := json.Unmarshal([]byte(stdout), &ds); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-datacube definition
		return nil
	}
	var datacube DataCube
	datacube.DatasetID = ds.DatasetID

	datacube.ChannelSize, _ = strconv.Atoi(ds.ChannelSize)

	datacube.TimeBegin, _ = strconv.Atoi(ds.TimeBegin)
	datacube.TimeLength, _ = strconv.Atoi(ds.TimeLength)
	datacube.TimeStep, _ = strconv.Atoi(ds.TimeStep)

	datacube.PointingBegin, _ = strconv.Atoi(ds.PointingBegin)
	datacube.PointingEnd, _ = strconv.Atoi(ds.PointingEnd)
	datacube.PointingStep, _ = strconv.Atoi(ds.PointingStep)
	datacube.NumPerBatch, _ = strconv.Atoi(ds.NumPerBatch)

	return &datacube
}

func (datacube *DataCube) getTimeRanges() []int {
	var ret []int

	for y := 0; y < datacube.TimeLength; y += datacube.TimeStep {
		y0 := datacube.TimeBegin + y
		y1 := y0 + datacube.TimeStep - 1
		if y1 > datacube.TimeBegin+datacube.TimeLength-1 {
			y1 = datacube.TimeBegin + datacube.TimeLength - 1
		}
		ret = append(ret, y0, y1)
	}
	return ret
}

func (datacube *DataCube) getTimeRange(t int) (int, int) {
	for y := 0; y < datacube.TimeLength; y += datacube.TimeStep {
		y0 := datacube.TimeBegin + y
		y1 := y0 + datacube.TimeStep - 1
		if y1 > datacube.TimeBegin+datacube.TimeLength-1 {
			y1 = datacube.TimeBegin + datacube.TimeLength - 1
		}
		if y0 <= t && t <= y1 {
			return y0, y1
		}
	}
	fmt.Fprintf(os.Stderr, "timestamp %d is out of range [%d..%d]\n",
		t, datacube.TimeBegin, datacube.TimeBegin+datacube.TimeLength-1)
	return -2, -1
}

func (datacube *DataCube) getPointingRanges() map[int]int {
	ret := make(map[int]int)
	for i := datacube.PointingBegin; i <= datacube.PointingEnd; i += datacube.PointingStep {
		j := i + datacube.PointingStep - 1
		if j > datacube.PointingEnd {
			j = datacube.PointingEnd
		}
		ret[i] = j
	}
	return ret
}
