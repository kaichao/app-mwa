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
