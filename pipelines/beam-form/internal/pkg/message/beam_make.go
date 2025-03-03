package message

import (
	"beamform/internal/pkg/datacube"
	"fmt"
	"regexp"
	"strconv"
)

// ParseForBeamMake ...
// return:
//
// messages:
// semaphores:	dat-ready/dat-done/fits-done/pointing-done
//
//	pointing-done
func ParseForBeamMake(m string) ([]string, string) {
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
	ps := cube.GetPointingRangesByInterval(pBegin, pEnd)

	semaDatReady := ""
	semaDatDone := ""
	semaFitsDone := ""
	// fits-done:1257010784/p00001/t1257010786_1257010985
	nTimeRanges := len(ts) / 2

	messages := []string{}
	for k := 0; k < len(ps); k += 2 {
		for j := 0; j < len(ts); j += 2 {
			id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d`, dataset, ps[k], ps[k+1], ts[j], ts[j+1])
			for i := 0; i < cube.NumOfChannels; i++ {
				m := fmt.Sprintf("%s/p%05d_%05d/t%d_%d/ch%03d",
					dataset, ps[k], ps[k+1], ts[j], ts[j+1], cube.ChannelBegin+i)
				messages = append(messages, m)
				semaPair := fmt.Sprintf(`"dat-ready:%s/ch%d":%d`, id, cube.ChannelBegin+i, nTimeRanges)
				semaDatReady += semaPair + "\n"
			}

			semaPair := fmt.Sprintf(`"dat-done:%s":%d`, id, 24)
			semaDatDone += semaPair + "\n"
			// semaPair := fmt.Sprintf(`"fits-done:%s/p%05d_%05d/t%d_%d":%d`,
			// 	dataset, ps[k], ps[k+1], ts[j], ts[j+1], 24)
			semaPair = fmt.Sprintf(`"fits-done:%s":%d`, id, 24)
			semaFitsDone += semaPair + "\n"
		}
	}

	semaPointingDone := ""
	// pointing-done:1257010784/p00001
	for p := pBegin; p <= pEnd; p++ {
		if ss[7] == "" {
			semaPair := fmt.Sprintf(`"pointing-done:%s/p%05d":%d`,
				dataset, p, nTimeRanges)
			semaPointingDone += semaPair + "\n"
		}
	}
	semaphores := semaDatReady + semaDatDone + semaFitsDone + semaPointingDone
	return messages, semaphores
}
