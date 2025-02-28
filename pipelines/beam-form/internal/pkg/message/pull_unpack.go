package message

import (
	"beamform/internal/pkg/datacube"
	"fmt"
	"regexp"
	"strconv"
)

// ProcessForPullUnpack ...
//
//	messages : array of message
//
// dat-ready's sema-pair list, '\n' as separator
func ProcessForPullUnpack(m string) ([]string, string) {
	re := regexp.MustCompile("^([0-9]+)((/p([0-9]+)_([0-9]+))(/t([0-9]+)_([0-9]+))?)?$")
	ss := re.FindStringSubmatch(m)
	dataset := ss[1]
	cube := datacube.GetDataCube(dataset)
	var (
		pBegin, pEnd int
		ts           []int
	)
	if ss[7] != "" {
		// 	1257010784/p00001_00960/t1257012766_1257012965
		t0, _ := strconv.Atoi(ss[7])
		t1, _ := strconv.Atoi(ss[8])
		ts = append(ts, t0, t1)
	} else {
		// 	1257010784/p00001_00960
		// 	1257010784
		ts = cube.GetTimeRanges()
	}
	if ss[4] != "" {
		// 	1257010784/p00001_00960/t1257012766_1257012965
		// 	1257010784/p00001_00960
		pBegin, _ = strconv.Atoi(ss[4])
		pEnd, _ = strconv.Atoi(ss[5])
	} else {
		// 	1257010784
		pBegin = cube.PointingBegin
		pEnd = cube.PointingEnd
	}

	messages := []string{}
	semas := ""
	prefix := fmt.Sprintf("%s/p%05d_%05d", dataset, pBegin, pEnd)
	for i := 0; i < cube.NumOfChannels; i++ {
		for j := 0; j < len(ts); j += 2 {
			hValue := fmt.Sprintf("%s/t%d_%d/ch%d", prefix, ts[j], ts[j+1], cube.ChannelBegin+i)
			header := fmt.Sprintf(`{"target_subdir":"%s"}`, hValue)
			tus := cube.GetTimeUnitsWithinInterval(ts[j], ts[j+1])
			for k := 0; k < len(tus); k += 2 {
				body := fmt.Sprintf("%s/%d_%d_ch%d.dat.tar.zst", dataset, tus[k], tus[k+1], cube.ChannelBegin+i)
				messages = append(messages, body+","+header)
			}

			semaName := "dat-ready:" + hValue
			semaVal := len(tus) / 2
			semaPair := fmt.Sprintf(`"%s":%d`, semaName, semaVal)
			semas += semaPair + "\n"
		}
	}
	return messages, semas
}
