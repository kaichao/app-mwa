package datacube

import (
	"fmt"
	"reflect"
	"testing"
)

func TestGetTimeIndex(t *testing.T) {
	// 1257010786 ~ 1257015583
	testCases := []struct {
		t, expected int
	}{
		{1257010785, -1},
		{1257010786, 0},
		{1257010787, 0},
		{1257010935, 0},
		{1257010936, 1},
		{1257010937, 1},
		{1257011085, 1},
		{1257015434, 30},
		{1257015435, 30},
		{1257015436, 31},
		{1257015583, 31},
		{1257015584, -1},
	}
	datacube := getMyDataCube()
	for _, tc := range testCases {
		result := datacube.getTimeRangeIndex(tc.t)
		if result != tc.expected {
			t.Errorf("datacube.getTimeIndex(%d) = %d, expected %d",
				tc.t, result, tc.expected)
		}
	}
}

func TestGetTimeRanges(t *testing.T) {
	datacube := getMyDataCube()
	ts := datacube.GetTimeRanges()
	fmt.Println(ts)
	// num of ranges: 32
	if len(ts) != 64 {
		t.Errorf("len(datacube.getTimeRanges()) = %d, expected %d", len(ts), 64)
	}
}

func TestGetTimeUnitsWithinInterval(t *testing.T) {
	testCases := []struct {
		t0, t1   int
		expected []int
	}{
		{1257010785, 1257010815, []int{1257010786, 1257010815}},
		{1257010786, 1257010815, []int{1257010786, 1257010815}},
		{1257010786, 1257010845, []int{1257010786, 1257010815, 1257010816, 1257010845}},
		{1257015526, 1257015583, []int{1257015526, 1257015555, 1257015556, 1257015583}},
		{1257015556, 1257015583, []int{1257015556, 1257015583}},
		{1257015556, 1257015584, []int{1257015556, 1257015583}},
	}
	datacube := getMyDataCube()
	for _, tc := range testCases {
		ts := datacube.GetTimeUnitsWithinInterval(tc.t0, tc.t1)
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
		{1257010785, 1257010935, []int{1257010786, 1257010935}},
		{1257010786, 1257010935, []int{1257010786, 1257010935}},
		{1257010786, 1257011085, []int{1257010786, 1257010935, 1257010936, 1257011085}},
		{1257015286, 1257015583, []int{1257015286, 1257015435, 1257015436, 1257015583}},
		{1257015436, 1257015583, []int{1257015436, 1257015583}},
		{1257015436, 1257015584, []int{1257015436, 1257015583}},
	}
	datacube := getMyDataCube()
	for _, tc := range testCases {
		ts := datacube.GetTimeRangesWithinInterval(tc.t0, tc.t1)
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
		{1257010786, 1257010786, 1257010935},
		{1257010787, 1257010786, 1257010935},
		{1257010935, 1257010786, 1257010935},
		{1257010936, 1257010936, 1257011085},
		{1257010937, 1257010936, 1257011085},
		{1257011085, 1257010936, 1257011085},
		{1257015436, 1257015436, 1257015583},
		{1257015583, 1257015436, 1257015583},
		{1257015584, -1, -1},
	}

	datacube := getMyDataCube()
	for _, tc := range testCases {
		t0, t1 := datacube.GetTimeRange(tc.t)
		if t0 != tc.e0 || t1 != tc.e1 {
			t.Errorf("datacube.getTimeRange(%d) = [%d %d], expected [%d,%d]",
				tc.t, t0, t1, tc.e0, tc.e1)
		}
	}
}
