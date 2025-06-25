package datacube_test

import (
	"beamform/internal/datacube"
	"os"
	"testing"
)

func TestGetTimeChannelIndex(t *testing.T) {
	os.Setenv("DATACUBE_FILE", "../../../dataset.yaml")
	cube := datacube.GetDataCube("1257010784")

	testCases := []struct {
		t     int
		ch    int
		len   int
		index int
	}{
		{1257010786, 109, 3, 0},
		{1257010786, 112, 3, 0},
		{1257010986, 112, 3, 0},
		{1257010786, 109, 24, 0},
		{1257010986, 109, 24, 0},
		{1257010786, 109, 48, 0},
		{1257010986, 109, 48, 24},
	}
	for _, tc := range testCases {
		index := cube.GetTimeChannelIndex(tc.t, tc.ch, tc.len)
		if index != tc.index {
			t.Errorf("cube.GetTimeChannelIndex(%d,%d,%d) ,result:%d , expected %d",
				tc.t, tc.ch, tc.len, index, tc.index)
		}
	}
}

func TestGetTimePointingIndex(t *testing.T) {
	os.Setenv("DATACUBE_FILE", "../../../dataset.yaml")
	cube := datacube.GetDataCube("1257010784")

	testCases := []struct {
		t     int
		p     int
		len   int
		index int
	}{
		{1257010786, 1, 3, 0},
		{1257010786, 4, 3, 0},
		{1257010986, 1, 3, 0},
		{1257010786, 1, 24, 0},
		{1257010986, 1, 24, 0},
		{1257010786, 1, 48, 0},
		{1257010986, 1, 48, 24},
	}
	for _, tc := range testCases {
		index := cube.GetTimePointingIndex(tc.t, tc.p, tc.len)
		if index != tc.index {
			t.Errorf("cube.GetTimePointingIndex(%d,%d,%d) ,result:%d , expected %d",
				tc.t, tc.p, tc.len, index, tc.index)
		}
	}

}
