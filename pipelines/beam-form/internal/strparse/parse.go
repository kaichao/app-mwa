package strparse

import (
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
//     1257010784/p00001_00960/t1257012766_1257012965/ch109
//
// 返回值：
//   - obsid, p0, p1, t0, t1, ch, err
//
// 环境变量：
//
//	-
func ParseParts(m string) (string, int, int, int, int, int, error) {
	re := regexp.MustCompile(`^([0-9]+)(/p([0-9]+)?_([0-9]+)?)?(/t([0-9]+)?_([0-9]+)?)?(/ch([0-9]+))?$`)
	ss := re.FindStringSubmatch(m)
	if len(ss) == 0 {
		return "", 0, 0, 0, 0, 0, errors.New("invalid message format")
	}
	dataset := ss[1]
	// cube := datacube.GetDataCube(dataset)

	var (
		ch int
	)
	p0, _ := strconv.Atoi(ss[3])
	p1, _ := strconv.Atoi(ss[4])
	t0, _ := strconv.Atoi(ss[6])
	t1, _ := strconv.Atoi(ss[7])
	if ss[9] == "" {
		ch = -1
	} else {
		ch, _ = strconv.Atoi(ss[9])
	}
	return dataset, p0, p1, t0, t1, ch, nil
}
