package message

import (
	"beamform/internal/pkg/datacube"
	"errors"
	"regexp"
	"strconv"
)

// ParseParts ...
//
// 参数：
//   - m: 消息体。其格式
//     1257010784
//     1257010784/p00001_00960
//     1257010784/t1257012766_1257012965
//     1257010784/p00001_00960/t1257012766_1257012965
//
// 返回值：
//   - dataset, p0, p1, t0, t1
//
// 环境变量：
//
//	-
func ParseParts(m string) (string, int, int, int, int, error) {
	re := regexp.MustCompile(`^([0-9]+)(/p([0-9]+)_([0-9]+))?(/t([0-9]+)_([0-9]+))?$`)
	ss := re.FindStringSubmatch(m)
	if len(ss) == 0 {
		return "", 0, 0, 0, 0, errors.New("invalid message format")
	}
	dataset := ss[1]
	cube := datacube.GetDataCube(dataset)

	var (
		p0, p1, t0, t1 int
	)
	if ss[3] == "" {
		p0 = cube.PointingBegin
		p1 = cube.PointingEnd
	} else {
		p0, _ = strconv.Atoi(ss[3])
		p1, _ = strconv.Atoi(ss[4])
	}
	if ss[6] == "" {
		t0 = cube.TimeBegin
		t1 = cube.TimeEnd
	} else {
		t0, _ = strconv.Atoi(ss[6])
		t1, _ = strconv.Atoi(ss[7])
	}
	return dataset, p0, p1, t0, t1, nil
}
