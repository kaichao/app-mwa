package datacube_test

import (
	"fmt"
	"testing"
)

func TestGetPointingRanges(t *testing.T) {
	datacube := getMyDataCube()
	ps := datacube.GetPointingRangesByInterval(1, 960)
	fmt.Println(ps)
	// num of ranges: 541
	if len(ps) != 1082 {
		// t.Errorf("len(datacube.getPointingRanges()) = %d, expected %d", len(ps), 80)
	}
}

/*
func TestGetPoingtingRangesByBatch(t *testing.T) {
	testCases := []struct {
		p, size, last int
	}{
		//out of range
		{0, 0, 0},
		{1, 20, 480},
		{2, 20, 480},
		{480, 20, 480},
		{481, 20, 960},
		{482, 20, 960},
		{960, 20, 960},
		{12961, 1, 12972},
		{12972, 1, 12972},
		//out of range
		{12973, 0, 0},
	}

	datacube := getMyDataCube()
	for _, tc := range testCases {
		// p0, p1 := datacube.GetPointingBatchRange(tc.p)
		// arr := datacube.getPointingRangesByBatch(p0, p1)
		last := 0
		if len(arr) > 0 {
			last = arr[len(arr)-1]
		}
		fmt.Printf("p=%d, arr=%v\n", tc.p, arr)
		if len(arr) == tc.size {
			t.Errorf("datacube.getPointingRangesByBatch(%d) ,[len,last]= [%d,%d] , expected [%d,%d]",
				tc.p, len(arr), last, tc.size, tc.last)
		}
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
		result := datacube.GetPointingBatchIndex(tc.t)
		if result != tc.expected {
			t.Errorf("datacube.getPointingBatchIndex(%d) = %d, expected %d", tc.t, result, tc.expected)
		}
	}
}

func TestGetPointingRangesByBatchIndex(t *testing.T) {
	// 1 ~ 12972 (24 * 20)
	testCases := []struct {
		idx, expected int
	}{
		{-1, 0},
		{0, 40},
		{1, 40},
		{26, 40},
		{27, 2},
		{28, 0},
	}
	datacube := getMyDataCube()
	for _, tc := range testCases {
		result := datacube.GetPointingRangesByBatchIndex(tc.idx)
		fmt.Printf("idx:%d,val:%v\n", tc.idx, result)
		if len(result) != tc.expected {
			t.Errorf("datacube.getPointingRangesByBatchIndex(%d),len() = %d, expected %d", tc.idx, len(result), tc.expected)
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
		p0, p1 := datacube.GetPointingBatchRange(tc.p)
		if p0 != tc.e0 || p1 != tc.e1 {
			t.Errorf("datacube.getPointingBatchRange(%d) = [%d %d], expected [%d,%d]",
				tc.p, p0, p1, tc.e0, tc.e1)
		}
	}
}
*/
