package main

import (
	"fmt"
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
		result := datacube.getTimeIndex(tc.t)
		if result != tc.expected {
			t.Errorf("datacube.getTimeIndex(%d) = %d, expected %d",
				tc.t, result, tc.expected)
		}
	}
}

func TestGetTimeRanges(t *testing.T) {
	datacube := getMyDataCube()
	ts := datacube.getTimeRanges()
	fmt.Println(ts)
	// num of ranges: 32
	if len(ts) != 64 {
		t.Errorf("len(datacube.getTimeRanges()) = %d, expected %d", len(ts), 64)
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
		t0, t1 := datacube.getTimeRange(tc.t)
		if t0 != tc.e0 || t1 != tc.e1 {
			t.Errorf("datacube.getTimeRange(%d) = [%d %d], expected [%d,%d]",
				tc.t, t0, t1, tc.e0, tc.e1)
		}
	}
}

func TestGetPoingtingRanges(t *testing.T) {
	datacube := getMyDataCube()
	ts := datacube.getPointingRanges()
	fmt.Println(ts)
	// num of ranges: 541
	if len(ts) != 1082 {
		t.Errorf("len(datacube.getPointingRanges()) = %d, expected %d", len(ts), 64)
	}
}

func TestGetPointingBatchIndex(t *testing.T) {
	// 1 ~ 12972 (24 * 20)
	testCases := []struct {
		t, expected int
	}{
		{0, -1},
		{1, 0},
		{2, 0},
		{480, 0},
		{481, 1},
		{960, 1},
		{961, 2},
		{12960, 26},
		{12961, 27},
		{12971, 27},
		{12972, 27},
		{12973, -1},
	}
	datacube := getMyDataCube()
	for _, tc := range testCases {
		result := datacube.getPointingBatchIndex(tc.t)
		if result != tc.expected {
			t.Errorf("datacube.getPointingBatchIndex(%d) = %d, expected %d", tc.t, result, tc.expected)
		}
	}
}

func TestGetPointingBatchRange(t *testing.T) {
	testCases := []struct {
		p, e0, e1 int
	}{
		{0, -1, -1},
		{1, 1, 480},
		{2, 1, 480},
		{480, 1, 480},
		{481, 481, 960},
		{482, 481, 960},
		{960, 481, 960},
		{12961, 12961, 12972},
		{12972, 12961, 12972},
		{12973, -1, -1},
	}

	datacube := getMyDataCube()
	for _, tc := range testCases {
		p0, p1 := datacube.getPointingBatchRange(tc.p)
		if p0 != tc.e0 || p1 != tc.e1 {
			t.Errorf("datacube.getPointingBatchRange(%d) = [%d %d], expected [%d,%d]",
				tc.p, p0, p1, tc.e0, tc.e1)
		}
	}
}

func TestGetPoingtingBatchRanges(t *testing.T) {
	datacube := getMyDataCube()
	ts := datacube.getPointingBatchRanges()
	fmt.Println(ts)
	// num of ranges: 28, 12972/480 = 27.025
	if len(ts) != 56 {
		t.Errorf("len(datacube.getPointingRanges()) = %d, expected %d", len(ts), 64)
	}
}

func getMyDataCube() *DataCube {
	datacube := &DataCube{
		DatasetID:     "1257010784",
		ChannelBegin:  109,
		NumOfChannels: 24,

		TimeBegin:    1257010786,
		NumOfSeconds: 4798,
		TimeStep:     150,

		PointingBegin: 1,
		PointingEnd:   12972,
		PointingStep:  24,
		NumPerBatch:   20,
	}

	return datacube
}
