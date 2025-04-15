package message

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/node"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/kaichao/gopkg/common"
)

// ParseForPullUnpack ...
//
//	messages : array of message
//
// dat-ready's sema-pair list, '\n' as separator
// deprecated.
func ParseForPullUnpack(m string) ([]string, string) {
	re := regexp.MustCompile("^([0-9]+)((/p([0-9]+)_([0-9]+))(/t([0-9]+)_([0-9]+))?)?$")
	ss := re.FindStringSubmatch(m)
	dataset := ss[1]
	cube := datacube.GetDataCube(dataset)
	var (
		pBegin, pEnd int
		tBegin, tEnd int
	)
	if ss[7] != "" {
		// 	1257010784/p00001_00960/t1257012766_1257012965
		tBegin, _ = strconv.Atoi(ss[7])
		tEnd, _ = strconv.Atoi(ss[8])
	} else {
		// 	1257010784/p00001_00960
		// 	1257010784
		tBegin = cube.TimeBegin
		tEnd = cube.TimeEnd
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

	ts := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
	messages := []string{}
	semas := ""
	prefix := fmt.Sprintf("%s/p%05d_%05d", dataset, pBegin, pEnd)
	for i := 0; i < cube.NumOfChannels; i++ {
		for j := 0; j < len(ts); j += 2 {
			hValue := fmt.Sprintf("%s/t%d_%d/ch%d",
				prefix, ts[j], ts[j+1], cube.ChannelBegin+i)
			header := fmt.Sprintf(`{"target_subdir":"%s"}`, hValue)
			tus := cube.GetTimeUnitsWithinInterval(ts[j], ts[j+1])
			for k := 0; k < len(tus); k += 2 {
				body := fmt.Sprintf("%s/p%05d_%05d/%d_%d_ch%d.dat.tar.zst",
					dataset, pBegin, pEnd, tus[k], tus[k+1], cube.ChannelBegin+i)
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

// GetMessagesForPullUnpack ...
func GetMessagesForPullUnpack(m string) []string {
	dataset, pBegin, pEnd, tBegin, tEnd, _, err := ParseParts(m)
	if err != nil {
		return []string{}
	}
	cube := datacube.GetDataCube(dataset)
	ts := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
	messages := []string{}
	withPointingPath := os.Getenv("WITH_POINTING_PATH") == "yes"
	prefix := dataset
	if withPointingPath {
		prefix = fmt.Sprintf("%s/p%05d_%05d", dataset, pBegin, pEnd)
	}
	for j := 0; j < len(ts); j += 2 {
		for i := 0; i < cube.NumOfChannels; i++ {
			hValue := fmt.Sprintf("%s/t%d_%d/ch%d",
				prefix, ts[j], ts[j+1], cube.ChannelBegin+i)
			headers := common.SetJSONAttribute("{}", "target_subdir", hValue)
			headers = common.SetJSONAttribute(headers, "to_host",
				node.GetNodeNameByTimeChannel(cube, ts[j], i+cube.ChannelBegin))
			// header := fmt.Sprintf(`{"target_subdir":"%s"}`, hValue)
			tus := cube.GetTimeUnitsWithinInterval(ts[j], ts[j+1])
			for k := 0; k < len(tus); k += 2 {
				body := fmt.Sprintf("%s/p%05d_%05d/%d_%d_ch%d.dat.tar.zst",
					dataset, pBegin, pEnd, tus[k], tus[k+1], cube.ChannelBegin+i)
				messages = append(messages, body+","+headers)
			}
		}
	}
	return messages
}
