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
	DatasetID string

	ChannelBegin  string
	NumOfChannels string

	TimeBegin    string
	NumOfSeconds string
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
	DatasetID string

	ChannelBegin  int
	NumOfChannels int

	TimeBegin    int
	NumOfSeconds int
	// 单次beam-maker的时长，通常为30的倍数
	TimeStep int

	PointingBegin int
	PointingEnd   int
	// 单次beam-maker处理的指向数，通常取24的倍数
	PointingStep int
	// 单批次beam-maker的执行次数
	NumPerBatch int
}

func getDataCube(datasetID string) *DataCube {
	cmdText := "scalebox dataset get-metadata " + datasetID
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Fprintf(os.Stderr, "stderr for dataset-get-metadata:\n%s\n", stderr)
	if code != 0 {
		fmt.Fprintf(os.Stderr, "[WARN] error for dataset-get-metadata dataset=%s in getDataCube()\n", datasetID)
		return nil
	}

	var (
		dc       dataCube
		datacube DataCube
	)
	if err := json.Unmarshal([]byte(stdout), &dc); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-datacube definition
		return nil
	}

	datacube.DatasetID = dc.DatasetID

	datacube.ChannelBegin, _ = strconv.Atoi(dc.ChannelBegin)
	datacube.NumOfChannels, _ = strconv.Atoi(dc.NumOfChannels)

	datacube.TimeBegin, _ = strconv.Atoi(dc.TimeBegin)
	datacube.NumOfSeconds, _ = strconv.Atoi(dc.NumOfSeconds)
	datacube.TimeStep, _ = strconv.Atoi(dc.TimeStep)

	datacube.PointingBegin, _ = strconv.Atoi(dc.PointingBegin)
	datacube.PointingEnd, _ = strconv.Atoi(dc.PointingEnd)
	datacube.PointingStep, _ = strconv.Atoi(dc.PointingStep)
	datacube.NumPerBatch, _ = strconv.Atoi(dc.NumPerBatch)

	return &datacube
}

func (datacube *DataCube) getTimeIndex(t int) int {
	t -= datacube.TimeBegin
	if 0 > t || t >= datacube.NumOfSeconds {
		fmt.Fprintf(os.Stderr, "[WARN]timestamp %d is out of range [%d..%d]\n",
			t, datacube.TimeBegin, datacube.TimeBegin+datacube.NumOfSeconds-1)
		return -1
	}
	return t / datacube.TimeStep
}

func (datacube *DataCube) getTimeRange(t int) (int, int) {
	// for y := 0; y < datacube.NumOfSeconds; y += datacube.TimeStep {
	// 	y0 := datacube.TimeBegin + y
	// 	y1 := y0 + datacube.TimeStep - 1
	// 	if y1 > datacube.TimeBegin+datacube.NumOfSeconds-1 {
	// 		y1 = datacube.TimeBegin + datacube.NumOfSeconds - 1
	// 	}
	// 	if y0 <= t && t <= y1 {
	// 		return y0, y1
	// 	}
	// }
	// fmt.Fprintf(os.Stderr, "[WARN]timestamp %d is out of range [%d..%d]\n",
	// 	t, datacube.TimeBegin, datacube.TimeBegin+datacube.NumOfSeconds-1)
	// return -1, -1

	t -= datacube.TimeBegin
	if 0 > t || t >= datacube.NumOfSeconds {
		fmt.Fprintf(os.Stderr, "[WARN]timestamp %d is out of range [%d..%d]\n",
			t, datacube.TimeBegin, datacube.TimeBegin+datacube.NumOfSeconds-1)
		return -1, -1
	}
	index := t / datacube.TimeStep
	t0 := datacube.TimeBegin + index*datacube.TimeStep
	t1 := t0 + datacube.TimeStep - 1
	if t1 > datacube.TimeBegin+datacube.NumOfSeconds-1 {
		t1 = datacube.TimeBegin + datacube.NumOfSeconds - 1
	}
	return t0, t1

}

func (datacube *DataCube) getTimeRanges() []int {
	var ret []int
	for t := 0; t < datacube.NumOfSeconds; t += datacube.TimeStep {
		t0 := datacube.TimeBegin + t
		t1 := t0 + datacube.TimeStep - 1
		if t1 > datacube.TimeBegin+datacube.NumOfSeconds-1 {
			t1 = datacube.TimeBegin + datacube.NumOfSeconds - 1
		}
		ret = append(ret, t0, t1)
	}
	return ret
}

func (datacube *DataCube) getPointingRanges() []int {
	var ret []int
	for p0 := datacube.PointingBegin; p0 <= datacube.PointingEnd; p0 += datacube.PointingStep {
		p1 := p0 + datacube.PointingStep - 1
		if p1 > datacube.PointingEnd {
			p1 = datacube.PointingEnd
		}
		ret = append(ret, p0, p1)
	}

	return ret
}

func (datacube *DataCube) getPointingBatchIndex(p int) int {
	if datacube.PointingBegin > p || p > datacube.PointingEnd {
		return -1
	}
	return (p - datacube.PointingBegin) / (datacube.PointingStep * datacube.NumPerBatch)
}

func (datacube *DataCube) getPointingBatchRange(p int) (int, int) {
	index := datacube.getPointingBatchIndex(p)
	if index == -1 {
		return -1, -1
	}
	p0 := datacube.PointingBegin + index*datacube.PointingStep*datacube.NumPerBatch
	p1 := p0 + datacube.PointingStep*datacube.NumPerBatch - 1
	if p1 > datacube.PointingEnd {
		p1 = datacube.PointingEnd
	}
	return p0, p1
}

func (datacube *DataCube) getPointingBatchRanges() []int {
	var ret []int
	for p0 := datacube.PointingBegin; p0 <= datacube.PointingEnd; p0 += datacube.PointingStep * datacube.NumPerBatch {
		p1 := p0 + datacube.PointingStep*datacube.NumPerBatch - 1
		if p1 > datacube.PointingEnd {
			p1 = datacube.PointingEnd
		}
		ret = append(ret, p0, p1)
	}

	return ret
}

// 三维datacube中，block的顺序号，用于local-tar-pull/cluster-tar-pull运行过程中的的排序
func (datacube *DataCube) getBlockOrder(t int, channel int) int {
	ch := channel - datacube.ChannelBegin
	tm := (t - datacube.TimeBegin) / datacube.TimeStep
	fmt.Printf("datacube.channelBegin:%d\n", datacube.ChannelBegin)
	fmt.Printf("datacube:%v\n", datacube)
	fmt.Println("ch=", ch)
	fmt.Println("tm=", tm)

	// 2位时间编码 + 2位通道编码
	return tm*100 + ch
}
