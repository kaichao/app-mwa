package main

import (
	"beamform/internal/datacube"
	"fmt"
	"os"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
)

func toBeamMake(msg string) int {
	// output message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	cube := datacube.NewDataCube(msg)

	tRanges := cube.GetTimeRanges()
	pRanges := cube.GetPointingRanges()

	messages := []string{}
	pointingRange := fmt.Sprintf("p%05d_%05d", cube.PointingBegin, cube.PointingEnd)
	headers := common.SetJSONAttribute("{}", "pointing_range", pointingRange)
	fmt.Println("headers:", headers)
	withPointingPath := os.Getenv("WITH_POINTING_PATH") == "yes"
	for k := 0; k < len(pRanges); k += 2 {
		for j := 0; j < len(tRanges); j += 2 {
			for i := 0; i < cube.NumOfChannels; i++ {
				m := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/ch%03d`,
					cube.ObsID, pRanges[k], pRanges[k+1],
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
	common.AppendToFile("custom-out.txt",
		fmt.Sprintf("n_messages:%d\n", len(messages)))

	envVars := map[string]string{
		"SINK_MODULE":     "beam-make",
		"TIMEOUT_SECONDS": "1800",
	}
	return task.AddTasks(messages, "", envVars)
}
