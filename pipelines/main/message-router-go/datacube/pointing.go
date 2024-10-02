package datacube

import (
	"fmt"
	"os"
)

// GetNumOfPointingBatch ...
func (cube *DataCube) GetNumOfPointingBatch() int {
	result := (cube.PointingEnd - cube.PointingBegin) /
		(cube.NumPerBatch * cube.PointingStep)
	remainder := (cube.PointingEnd - cube.PointingBegin) %
		(cube.NumPerBatch * cube.PointingStep)
	if remainder > 0 {
		result++
	}
	return result
}

// 获得全部指向的指向区间(not used)
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

// GetPointingRangesByBatchIndex ...
func (cube *DataCube) GetPointingRangesByBatchIndex(batchIndex int) []int {
	numBatch := (cube.PointingEnd - cube.PointingBegin + 1) / (cube.NumPerBatch * cube.PointingStep)
	if batchIndex < 0 || batchIndex > numBatch {
		fmt.Fprintf(os.Stderr, "batch-index:%d is out of range, it should be [0..%d]\n", batchIndex, numBatch)
		return []int{}
	}
	pb := cube.PointingBegin + batchIndex*cube.NumPerBatch*cube.PointingStep
	pe := pb + cube.NumPerBatch*cube.PointingStep - 1

	return cube.getPointingRangesByBatch(pb, pe)
}

func (cube *DataCube) getPointingRangesByBatch(batchBegin, batchEnd int) []int {
	var ret []int
	if batchEnd > cube.PointingEnd {
		batchEnd = cube.PointingEnd
	}
	for p0 := batchBegin; p0 <= batchEnd; p0 += cube.PointingStep {
		p1 := p0 + cube.PointingStep - 1
		if p1 > cube.PointingEnd {
			p1 = cube.PointingEnd
		}
		ret = append(ret, p0, p1)
	}

	return ret
}

// GetPointingBatchIndex ...
// 获取当前指向所在的批次索引
func (cube *DataCube) GetPointingBatchIndex(p int) int {
	if cube.PointingBegin > p || p > cube.PointingEnd {
		return -1
	}
	return (p - cube.PointingBegin) / (cube.PointingStep * cube.NumPerBatch)
}

// GetPointingBatchRange ...
// 获得当前指向所在的批次指向区间
func (cube *DataCube) GetPointingBatchRange(p int) (int, int) {
	index := cube.GetPointingBatchIndex(p)
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
