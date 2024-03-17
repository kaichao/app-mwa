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
	// 30
	TimeUnit string
	// 30的倍数
	TimeStep string

	PointingBegin string
	PointingEnd   string
	// 通常为24
	PointingStep string
	NumPerBatch  string
}

// DataCube ...
//
//	Time Dimension: TimeUnit, TimeRange
//
//	Pointing Demension: PointingRange, PointingBatch
type DataCube struct {
	DatasetID string

	ChannelBegin  int
	NumOfChannels int

	TimeBegin    int
	NumOfSeconds int
	// 单个打包文件的时长（30秒）
	TimeUnit int
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
	datacube.TimeUnit, _ = strconv.Atoi(dc.TimeUnit)
	datacube.TimeStep, _ = strconv.Atoi(dc.TimeStep)

	datacube.PointingBegin, _ = strconv.Atoi(dc.PointingBegin)
	datacube.PointingEnd, _ = strconv.Atoi(dc.PointingEnd)
	datacube.PointingStep, _ = strconv.Atoi(dc.PointingStep)
	datacube.NumPerBatch, _ = strconv.Atoi(dc.NumPerBatch)

	return &datacube
}

func (cube *DataCube) getTimeIndex(t int) int {
	t -= cube.TimeBegin
	if 0 > t || t >= cube.NumOfSeconds {
		fmt.Fprintf(os.Stderr, "[WARN]timestamp %d is out of range [%d..%d]\n",
			t, cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
		return -1
	}
	return t / cube.TimeStep
}

func (cube *DataCube) getTimeUnit(t int) (int, int) {
	t -= cube.TimeBegin
	if 0 > t || t >= cube.NumOfSeconds {
		fmt.Fprintf(os.Stderr, "[WARN]getTimeUnit(), timestamp %d is out of range [%d..%d]\n",
			t, cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
		return -1, -1
	}
	index := t / cube.TimeUnit
	t0 := cube.TimeBegin + index*cube.TimeUnit
	t1 := t0 + cube.TimeUnit - 1
	if t1 > cube.TimeBegin+cube.NumOfSeconds-1 {
		t1 = cube.TimeBegin + cube.NumOfSeconds - 1
	}
	return t0, t1
}

func (cube *DataCube) getTimeRange(t int) (int, int) {
	t -= cube.TimeBegin
	if 0 > t || t >= cube.NumOfSeconds {
		fmt.Fprintf(os.Stderr, "[WARN]getTimeRange(),timestamp %d is out of range [%d..%d]\n",
			t, cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
		return -1, -1
	}
	index := t / cube.TimeStep
	t0 := cube.TimeBegin + index*cube.TimeStep
	t1 := t0 + cube.TimeStep - 1
	if t1 > cube.TimeBegin+cube.NumOfSeconds-1 {
		t1 = cube.TimeBegin + cube.NumOfSeconds - 1
	}
	return t0, t1
}

func (cube *DataCube) getTimeRanges() []int {
	var ret []int
	for t := 0; t < cube.NumOfSeconds; t += cube.TimeStep {
		t0 := cube.TimeBegin + t
		t1 := t0 + cube.TimeStep - 1
		if t1 > cube.TimeBegin+cube.NumOfSeconds-1 {
			t1 = cube.TimeBegin + cube.NumOfSeconds - 1
		}
		ret = append(ret, t0, t1)
	}
	return ret
}

func (cube *DataCube) getTimeUnitsByInterval(lower, upper int) []int {
	var ret []int
	lower -= cube.TimeBegin
	upper -= cube.TimeBegin
	if lower < 0 {
		lower = 0
	}
	for t := lower; t < upper; t += cube.TimeUnit {
		t0 := t / cube.TimeUnit * cube.TimeUnit
		t1 := t0 + cube.TimeUnit - 1
		if t1 > cube.NumOfSeconds-1 {
			t1 = cube.NumOfSeconds - 1
		}
		ret = append(ret, cube.TimeBegin+t0, cube.TimeBegin+t1)
	}
	return ret
}

func (cube *DataCube) getNumOfPointingBatch() int {
	result := (cube.PointingEnd - cube.PointingBegin) /
		(cube.NumPerBatch * cube.PointingStep)
	remainder := (cube.PointingEnd - cube.PointingBegin) %
		(cube.NumPerBatch * cube.PointingStep)
	if remainder > 0 {
		result++
	}
	return result
}

// 获得全部指向的指向区间
func (cube *DataCube) getPointingRanges() []int {
	var ret []int
	for p0 := cube.PointingBegin; p0 <= cube.PointingEnd; p0 += cube.PointingStep {
		p1 := p0 + cube.PointingStep - 1
		if p1 > cube.PointingEnd {
			p1 = cube.PointingEnd
		}
		ret = append(ret, p0, p1)
	}

	return ret
}

func (cube *DataCube) getPointingRangesByBatchIndex(batchIndex int) []int {
	return cube.getPointingRangesByBatch(cube.getPointingBatchRange(batchIndex))
}

func (cube *DataCube) getPointingRangesByBatch(batchBegin, batchEnd int) []int {
	var ret []int
	for p0 := batchBegin; p0 <= batchEnd; p0 += cube.PointingStep {
		p1 := p0 + cube.PointingStep - 1
		if p1 > cube.PointingEnd {
			p1 = cube.PointingEnd
		}
		ret = append(ret, p0, p1)
	}

	return ret
}

// 获取当前指向所在的批次索引
func (cube *DataCube) getPointingBatchIndex(p int) int {
	if cube.PointingBegin > p || p > cube.PointingEnd {
		return -1
	}
	return (p - cube.PointingBegin) / (cube.PointingStep * cube.NumPerBatch)
}

// 获得当前指向所在的批次指向区间
func (cube *DataCube) getPointingBatchRange(p int) (int, int) {
	index := cube.getPointingBatchIndex(p)
	if index == -1 {
		return -1, -1
	}
	p0 := cube.PointingBegin + index*cube.PointingStep*cube.NumPerBatch
	p1 := p0 + cube.PointingStep*cube.NumPerBatch - 1
	if p1 > cube.PointingEnd {
		p1 = cube.PointingEnd
	}
	return p0, p1
}

func (cube *DataCube) getPointingBatchRanges() []int {
	var ret []int
	for p0 := cube.PointingBegin; p0 <= cube.PointingEnd; p0 += cube.PointingStep * cube.NumPerBatch {
		p1 := p0 + cube.PointingStep*cube.NumPerBatch - 1
		if p1 > cube.PointingEnd {
			p1 = cube.PointingEnd
		}
		ret = append(ret, p0, p1)
	}

	return ret
}

// 三维datacube中，给定顺序号，用于local-tar-pull/cluster-tar-pull运行过程中的的排序
func (cube *DataCube) getSortedTag(pointing int, time int, channel int) string {
	p := cube.getPointingBatchIndex(pointing)
	ch := channel - cube.ChannelBegin
	tm := (time - cube.TimeBegin) / cube.TimeStep
	fmt.Printf("datacube.channelBegin:%d\n", cube.ChannelBegin)
	fmt.Printf("datacube:%v\n", cube)
	fmt.Println("ch=", ch)
	fmt.Println("tm=", tm)

	// 2位指向码 + 2位时间编码 + 2位通道编码

	return fmt.Sprintf("%02d%02d%02d", p, tm, ch)
}
