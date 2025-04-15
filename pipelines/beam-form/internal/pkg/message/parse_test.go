package message_test

import (
	"beamform/internal/pkg/message"
	"os"
	"testing"
)

func TestParseParts(t *testing.T) {
	os.Setenv("DATACUBE_FILE", "../../../dataset.yaml")
	testCases := []struct {
		m       string
		dataset string
		p0, p1  int
		t0, t1  int
		ch      int
	}{
		{"1257010784", "1257010784", 1, 12985, 1257010786, 1257015583, -1},
		{"1257010784/p00001_00960", "1257010784", 1, 960, 1257010786, 1257015583, -1},
		{"1257010784/t1257012766_1257012965", "1257010784", 1, 12985, 1257012766, 1257012965, -1},
		{"1257010784/p00001_00960/t1257012766_1257012965", "1257010784", 1, 960, 1257012766, 1257012965, -1},
		{"1257010784/p00001_00960/t1257012766_1257012965/ch109", "1257010784", 1, 960, 1257012766, 1257012965, 109},
	}
	for _, tc := range testCases {
		dataset, p0, p1, t0, t1, ch, err := message.ParseParts(tc.m)

		if err != nil || dataset != tc.dataset ||
			p0 != tc.p0 || p1 != tc.p1 ||
			t0 != tc.t0 || t1 != tc.t1 ||
			ch != tc.ch {
			t.Errorf("message.ParseParts(%s) ,[dataset,p0,p1,t0,t1]=[%s,%d,%d,%d,%d] , expected [%s,%d,%d,%d,%d]",
				tc.m, dataset, p0, p1, t0, t1, tc.dataset, tc.p0, tc.p1, tc.t0, tc.t1)
		}
	}
}
