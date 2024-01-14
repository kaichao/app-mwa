package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
)

type dataSet struct {
	// prefix ':' type ':' sub-id
	DatasetID string

	// for type "H", x-coord
	HorizontalWidth string

	// for type "V", y-coord
	VerticalStart  string
	VerticalHeight string
}

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

/*
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
*/
func getDataSet(datasetID string) *DataSet {
	cmdText := "scalebox dataset get-metadata " + datasetID
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Fprintf(os.Stderr, "stderr for dataset-get-metadata:\n%s\n", stderr)
	if code != 0 {
		fmt.Fprintf(os.Stderr, "[WARN] error for dataset-get-metadata dataset=%s in getDataSet()\n", datasetID)
		return nil
	}

	var ds dataSet
	if err := json.Unmarshal([]byte(stdout), &ds); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-dataset definition
		return nil
	}
	var dataset DataSet
	dataset.DatasetID = ds.DatasetID
	dataset.HorizontalWidth, _ = strconv.Atoi(ds.HorizontalWidth)
	dataset.VerticalStart, _ = strconv.Atoi(ds.VerticalStart)
	dataset.VerticalHeight, _ = strconv.Atoi(ds.VerticalHeight)

	return &dataset
}

func (dataset *DataSet) getTimeRanges() []int {
	var ret []int

	for y := 0; y < dataset.VerticalHeight; y += tStep {
		y0 := dataset.VerticalStart + y
		y1 := y0 + tStep - 1
		if y1 > dataset.VerticalStart+dataset.VerticalHeight-1 {
			y1 = dataset.VerticalStart + dataset.VerticalHeight - 1
		}
		ret = append(ret, y0, y1)
	}
	return ret
}

func (dataset *DataSet) getTimeRange(t int) (int, int) {
	for y := 0; y < dataset.VerticalHeight; y += tStep {
		y0 := dataset.VerticalStart + y
		y1 := y0 + tStep - 1
		if y1 > dataset.VerticalStart+dataset.VerticalHeight-1 {
			y1 = dataset.VerticalStart + dataset.VerticalHeight - 1
		}
		if y0 <= t && t <= y1 {
			return y0, y1
		}
	}
	fmt.Fprintf(os.Stderr, "timestamp %d is out of range [%d..%d]\n",
		t, dataset.VerticalStart, dataset.VerticalStart+dataset.VerticalHeight-1)
	return -2, -1
}
