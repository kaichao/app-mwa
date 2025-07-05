package main

import (
	"beamform/internal/datacube"
	"fmt"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/sirupsen/logrus"
)

func defaultFunc(msg string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-defaultFunc()")
	}()
	common.AddTimeStamp("enter-defaultFunc()")
	// input message:
	// 	1257010784
	// 	1257010784/p00001_00960
	// 	1257010784/p00001_00960/t1257012766_1257012965

	cube := datacube.NewDataCube(msg)

	semaFitsDone := ""
	// fits-done:1257010784/p00001/t1257010786_1257010985
	pRanges := cube.GetPointingRanges()
	tRanges := cube.GetTimeRanges()
	for k := 0; k < len(pRanges); k += 2 {
		for j := 0; j < len(tRanges); j += 2 {
			id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d`,
				cube.ObsID, pRanges[k], pRanges[k+1], tRanges[j], tRanges[j+1])
			semaPair := fmt.Sprintf(`"fits-done:%s":%d`, id, 24)
			semaFitsDone += semaPair + "\n"
		}
	}
	common.AppendToFile("my-semas.txt", semaFitsDone)

	semaPointingDone := ""
	// pointing-done:1257010784/p00001
	nTimeRanges := len(tRanges) / 2
	for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
		semaPair := fmt.Sprintf(`"pointing-done:%s/p%05d":%d`,
			cube.ObsID, p, nTimeRanges)
		semaPointingDone += semaPair + "\n"
	}
	common.AppendToFile("my-semas.txt", semaPointingDone)

	if err := semaphore.CreateFileSemaphores("my-semas.txt", appID, 100); err != nil {
		logrus.Errorf("semaphore-create,err-info:%v\n", err)
		return 1
	}
	common.AddTimeStamp("after-semaphores")

	return toBeamMake(msg)
}

// func fromMessageRouter(message string, headers map[string]string) int {
// 	return 0
// }
