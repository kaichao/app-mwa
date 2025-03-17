package message

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/json"
	"fmt"
	"os"
)

// ParseForBeamMake for shared storage
// return:
//
// messages:
// semaphores:	dat-ready/dat-done/fits-done/pointing-done
//
//	pointing-done
//
// deprecated.
// func ParseForBeamMake(m string) ([]string, string) {
// 	re := regexp.MustCompile("^([0-9]+)((/p([0-9]+)_([0-9]+))(/t([0-9]+)_([0-9]+))?)?$")
// 	ss := re.FindStringSubmatch(m)
// 	dataset := ss[1]
// 	cube := datacube.GetDataCube(dataset)
// 	var (
// 		pBegin, pEnd int
// 		tBegin, tEnd int
// 	)
// 	if ss[7] != "" {
// 		// 	1257010784/p00001_00960/t1257012766_1257012965
// 		tBegin, _ = strconv.Atoi(ss[7])
// 		tEnd, _ = strconv.Atoi(ss[8])
// 	} else {
// 		// 	1257010784/p00001_00960
// 		// 	1257010784
// 		tBegin = cube.TimeBegin
// 		tEnd = cube.TimeEnd
// 	}
// 	if ss[4] != "" {
// 		// 	1257010784/p00001_00960/t1257012766_1257012965
// 		// 	1257010784/p00001_00960
// 		pBegin, _ = strconv.Atoi(ss[4])
// 		pEnd, _ = strconv.Atoi(ss[5])
// 	} else {
// 		// 	1257010784
// 		pBegin = cube.PointingBegin
// 		pEnd = cube.PointingEnd
// 	}
// 	tRanges := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
// 	pRanges := cube.GetPointingRangesByInterval(pBegin, pEnd)

// 	semaDatReady := ""
// 	for j := 0; j < len(tRanges); j += 2 {
// 		tUnits := cube.GetTimeUnitsWithinInterval(tRanges[j], tRanges[j+1])
// 		nTimeUnits := len(tUnits) / 2
// 		for i := 0; i < cube.NumOfChannels; i++ {
// 			semaPair := fmt.Sprintf(`"dat-ready:%s/p%05d_%05d/t%d_%d/ch%d":%d`,
// 				dataset, pBegin, pEnd, tRanges[j], tRanges[j+1], cube.ChannelBegin+i, nTimeUnits)
// 			semaDatReady += semaPair + "\n"
// 		}
// 	}

// 	semaDatDone := ""
// 	semaFitsDone := ""
// 	// fits-done:1257010784/p00001/t1257010786_1257010985
// 	messages := []string{}
// 	nPointingRanges := len(pRanges) / 2
// 	for k := 0; k < len(pRanges); k += 2 {
// 		for j := 0; j < len(tRanges); j += 2 {
// 			for i := 0; i < cube.NumOfChannels; i++ {
// 				m := fmt.Sprintf("%s/p%05d_%05d/t%d_%d/ch%03d",
// 					dataset, pRanges[k], pRanges[k+1], tRanges[j], tRanges[j+1], cube.ChannelBegin+i)
// 				messages = append(messages, m)
// 			}

// 			id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d`, dataset, pRanges[k], pRanges[k+1], tRanges[j], tRanges[j+1])
// 			semaPair := fmt.Sprintf(`"dat-done:%s":%d`, id, nPointingRanges)
// 			semaDatDone += semaPair + "\n"
// 			// semaPair := fmt.Sprintf(`"fits-done:%s/p%05d_%05d/t%d_%d":%d`,
// 			// 	dataset, ps[k], ps[k+1], ts[j], ts[j+1], 24)
// 			semaPair = fmt.Sprintf(`"fits-done:%s":%d`, id, 24)
// 			semaFitsDone += semaPair + "\n"
// 		}
// 	}

// 	semaPointingDone := ""
// 	// pointing-done:1257010784/p00001
// 	nTimeRanges := len(tRanges) / 2
// 	for p := pBegin; p <= pEnd; p++ {
// 		if ss[7] == "" {
// 			semaPair := fmt.Sprintf(`"pointing-done:%s/p%05d":%d`,
// 				dataset, p, nTimeRanges)
// 			semaPointingDone += semaPair + "\n"
// 		}
// 	}
// 	semaphores := semaDatReady + semaDatDone + semaFitsDone + semaPointingDone
// 	return messages, semaphores
// }

// GetMessagesForBeamMake ...
func GetMessagesForBeamMake(m string) []string {
	dataset, pBegin, pEnd, tBegin, tEnd, err := ParseParts(m)
	if err != nil {
		return []string{}
	}
	cube := datacube.GetDataCube(dataset)

	tRanges := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
	pRanges := cube.GetPointingRangesByInterval(pBegin, pEnd)

	messages := []string{}
	pointingRange := fmt.Sprintf("p%05d_%05d", pBegin, pEnd)
	headers := json.SetAttribute("{}", "pointing_range", pointingRange)
	fmt.Println("headers:", headers)
	withPointingPath := os.Getenv("WITH_POINTING_PATH") == "yes"
	for k := 0; k < len(pRanges); k += 2 {
		for j := 0; j < len(tRanges); j += 2 {
			for i := 0; i < cube.NumOfChannels; i++ {
				m := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/ch%03d`,
					dataset, pRanges[k], pRanges[k+1],
					tRanges[j], tRanges[j+1],
					cube.ChannelBegin+i)
				if withPointingPath {
					m += "," + headers
				}
				// fmt.Println("m=", m)
				messages = append(messages, m)
			}
		}
	}
	return messages
}
