package message

import (
	"beamform/internal/datacube"
	"fmt"
	"os"

	"github.com/kaichao/scalebox/pkg/common"
)

// GetMessagesForBeamMake ...
func GetMessagesForBeamMake(m string) []string {
	dataset, pBegin, pEnd, tBegin, tEnd, _, err := ParseParts(m)
	if err != nil {
		return []string{}
	}
	cube := datacube.GetDataCube(dataset)

	tRanges := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
	pRanges := cube.GetPointingRangesByInterval(pBegin, pEnd)

	messages := []string{}
	pointingRange := fmt.Sprintf("p%05d_%05d", pBegin, pEnd)
	headers := common.SetJSONAttribute("{}", "pointing_range", pointingRange)
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
