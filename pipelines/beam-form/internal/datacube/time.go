package datacube

import (
	"github.com/sirupsen/logrus"
)

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
	ts := cube.GetTimeRanges()
	for i := 0; i < len(ts); i += 2 {
		if ts[i] <= t && t <= ts[i+1] {
			return ts[i], ts[i+1]
		}
	}
	return -1, -1
}

// GetTimeRanges ...
func (cube *DataCube) GetTimeRanges() []int {
	return cube.GetTimeRangesWithinInterval(cube.TimeBegin, cube.TimeEnd)
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

	if cube.TimeTailMerge && len(ret) > 2 {
		n := ret[len(ret)-1] - ret[len(ret)-2]
		ret = ret[:len(ret)-2]
		ret[len(ret)-1] += n + 1
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
	ts := cube.GetTimeRanges()
	for i := 0; i < len(ts); i += 2 {
		if ts[i] <= t && t <= ts[i+1] {
			return i / 2
		}
	}
	return -1
}

// GetTimeUnitIndex ...
func (cube *DataCube) GetTimeUnitIndex(t int) int {
	ts := cube.GetTimeUnits()
	for i := 0; i < len(ts); i += 2 {
		if ts[i] <= t && t <= ts[i+1] {
			return i / 2
		}
	}
	return -1
}
