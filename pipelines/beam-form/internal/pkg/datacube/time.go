package datacube

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// not used
func (cube *DataCube) getTimeRangeIndex(t int) int {
	t -= cube.TimeBegin

	if 0 > t || t >= cube.NumOfSeconds {
		fmt.Fprintf(os.Stderr, "[WARN]timestamp %d is out of range [%d..%d]\n",
			t, cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
		return -1
	}
	return t / cube.TimeStep
}

// not used
func (cube *DataCube) getTimeUnit(t int) (int, int) {
	t -= cube.TimeBegin
	if 0 > t || t >= cube.NumOfSeconds {
		logrus.Warnf("getTimeUnit(), timestamp %d is out of range [%d..%d]\n",
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

// GetTimeRange ...
func (cube *DataCube) GetTimeRange(t int) (int, int) {
	fmt.Println("cube:", cube)
	t -= cube.TimeBegin
	if 0 > t || t >= cube.NumOfSeconds {
		logrus.Warnf("getTimeRange(),timestamp %d is out of range [%d..%d]\n",
			t+cube.TimeBegin, cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
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

// GetTimeRanges ...
func (cube *DataCube) GetTimeRanges() []int {
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

// GetTimeRangesWithinInterval ...
func (cube *DataCube) GetTimeRangesWithinInterval(lower, upper int) []int {
	var ret []int
	lower -= cube.TimeBegin
	upper -= cube.TimeBegin
	if lower < 0 {
		lower = 0
	}
	for t := lower; t < upper; t += cube.TimeStep {
		t0 := t / cube.TimeStep * cube.TimeStep
		t1 := t0 + cube.TimeStep - 1
		if t1 > cube.NumOfSeconds-1 {
			t1 = cube.NumOfSeconds - 1
		}
		ret = append(ret, cube.TimeBegin+t0, cube.TimeBegin+t1)
	}
	return ret
}

// GetTimeUnits ...
func (cube *DataCube) GetTimeUnits() []int {
	var ret []int
	for t := 0; t < cube.NumOfSeconds; t += cube.TimeUnit {
		t0 := cube.TimeBegin + t
		t1 := t0 + cube.TimeUnit - 1
		if t1 > cube.TimeBegin+cube.NumOfSeconds-1 {
			t1 = cube.TimeBegin + cube.NumOfSeconds - 1
		}
		ret = append(ret, t0, t1)
	}
	return ret
}

// GetTimeUnitsWithinInterval ...
func (cube *DataCube) GetTimeUnitsWithinInterval(lower, upper int) []int {
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

// GetTimeRangeIndex ...
func (cube *DataCube) GetTimeRangeIndex(t int) int {
	index := (t - cube.TimeBegin) / cube.TimeStep
	fmt.Printf("begin=%d,step=%d,t=%d,index=%d\n",
		cube.TimeBegin, cube.TimeStep, t, index)
	return index
}
