package datacube_test

import (
	"beamform/internal/datacube"
	"reflect"
	"testing"
)

func TestGetTimeRanges(t *testing.T) {
	ts := cube.GetTimeRanges()
	// num of ranges: 24
	if len(ts) != 48 {
		t.Errorf("len(datacube.getTimeRanges()) = %d, expected %d", len(ts), 48)
	}
}

func TestGetTimeUnitsWithinInterval(t *testing.T) {
	testCases := []struct {
		t0, t1   int
		expected []int
	}{
		{1257010785, 1257010825, []int{1257010786, 1257010825}},
		{1257010786, 1257010825, []int{1257010786, 1257010825}},
		{1257010786, 1257010865, []int{1257010786, 1257010825, 1257010826, 1257010865}},
		{1257015506, 1257015583, []int{1257015506, 1257015545, 1257015546, 1257015583}},
		{1257015546, 1257015583, []int{1257015546, 1257015583}},
		{1257015546, 1257015584, []int{1257015546, 1257015583}},
	}
	for _, tc := range testCases {
		ts := cube.GetTimeUnitsWithinInterval(tc.t0, tc.t1)
		if !reflect.DeepEqual(ts, tc.expected) {
			t.Errorf("datacube.getTimeRangesByInterval(%d,%d) = %v, expected %v",
				tc.t0, tc.t1, ts, tc.expected)
		}
	}
}

func TestGetTimeRangesWithinInterval(t *testing.T) {
	testCases := []struct {
		t0, t1   int
		expected []int
	}{
		{1257010785, 1257010985, []int{1257010786, 1257010985}},
		{1257010786, 1257010985, []int{1257010786, 1257010985}},
		{1257010786, 1257011185, []int{1257010786, 1257010985, 1257010986, 1257011185}},
		{1257015186, 1257015583, []int{1257015186, 1257015385, 1257015386, 1257015583}},
		{1257015386, 1257015583, []int{1257015386, 1257015583}},
		{1257015386, 1257015584, []int{1257015386, 1257015583}},
	}
	for _, tc := range testCases {
		ts := cube.GetTimeRangesWithinInterval(tc.t0, tc.t1)
		if !reflect.DeepEqual(ts, tc.expected) {
			t.Errorf("datacube.getTimeRangesByInterval(%d,%d) = %v, expected %v",
				tc.t0, tc.t1, ts, tc.expected)
		}
	}
}

func TestGetTimeRange(t *testing.T) {
	testCases := []struct {
		t, e0, e1 int
	}{
		{1257010785, -1, -1},
		{1257010786, 1257010786, 1257010985},
		{1257010787, 1257010786, 1257010985},
		{1257010985, 1257010786, 1257010985},
		{1257010986, 1257010986, 1257011185},
		{1257010987, 1257010986, 1257011185},
		{1257011185, 1257010986, 1257011185},
		{1257015386, 1257015386, 1257015583},
		{1257015583, 1257015386, 1257015583},
		{1257015584, -1, -1},
	}

	for _, tc := range testCases {
		t0, t1 := cube.GetTimeRange(tc.t)
		if t0 != tc.e0 || t1 != tc.e1 {
			t.Errorf("datacube.getTimeRange(%d) = [%d %d], expected [%d,%d]",
				tc.t, t0, t1, tc.e0, tc.e1)
		}
	}
}

func TestGetTimeRangeIndex(t *testing.T) {
	// 1257010786 ~ 1257015583
	testCases := []struct {
		t, expected int
	}{
		{1257010785, -1},
		{1257010786, 0},
		{1257010787, 0},
		{1257010985, 0},
		{1257010986, 1},
		{1257010987, 1},
		{1257011185, 1},
		{1257015384, 22},
		{1257015385, 22},
		{1257015386, 23},
		{1257015583, 23},
		{1257015584, -1},
	}
	for _, tc := range testCases {
		idx := cube.GetTimeRangeIndex(tc.t)
		if idx != tc.expected {
			t.Errorf("datacube.getTimeRangeIndex(%d) = %d, expected %d",
				tc.t, idx, tc.expected)
		}
	}
}

func TestTimeTailMerge(t *testing.T) {
	cube := datacube.NewDataCube("1265983624")

	ts := cube.GetTimeRanges()
	// num of ranges: 24
	if len(ts) != 48 {
		t.Errorf("len(datacube.getTimeRanges()) = %d, expected %d", len(ts), 48)
	}

	testCases1 := []struct {
		t, expected int
	}{
		{1265988225, 22},
		{1265988226, 23},
		{1265988227, 23},
		{1265988425, 23},
		{1265988426, 23},
		{1265988429, 23},
		{1265988430, -1},
	}

	for _, tc := range testCases1 {
		idx := cube.GetTimeRangeIndex(tc.t)
		if idx != tc.expected {
			t.Errorf("datacube.getTimeRangeIndex(%d) = %d, expected %d",
				tc.t, idx, tc.expected)
		}
	}

	testCases2 := []struct {
		t, e0, e1 int
	}{
		{1265988225, 1265988026, 1265988225},
		{1265988226, 1265988226, 1265988429},
		{1265988227, 1265988226, 1265988429},
		{1265988425, 1265988226, 1265988429},
		{1265988426, 1265988226, 1265988429},
		{1265988429, 1265988226, 1265988429},
		{1265988430, -1, -1},
	}
	for _, tc := range testCases2 {
		t0, t1 := cube.GetTimeRange(tc.t)
		if t0 != tc.e0 || t1 != tc.e1 {
			t.Errorf("datacube.getTimeRange(%d) = [%d %d], expected [%d,%d]",
				tc.t, t0, t1, tc.e0, tc.e1)
		}
	}
}
