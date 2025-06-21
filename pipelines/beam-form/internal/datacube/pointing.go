package datacube

import "fmt"

// 获得全部指向的指向区间(not used)
func (cube *DataCube) getPointingRanges() []int {
	return cube.GetPointingRangesByInterval(cube.PointingBegin, cube.PointingEnd)
}

// GetPointingRangesByInterval ...
func (cube *DataCube) GetPointingRangesByInterval(pBegin, pEnd int) []int {
	var ret []int
	for p0 := pBegin; p0 <= pEnd; p0 += cube.PointingStep {
		p1 := p0 + cube.PointingStep - 1
		if p1 > cube.PointingEnd {
			p1 = cube.PointingEnd
		}
		if p1 > pEnd {
			p1 = pEnd
		}
		ret = append(ret, p0, p1)
	}

	fmt.Printf("pBegin:%d,pEnd:%d,ret:%v\n", pBegin, pEnd, ret)
	return ret
}
