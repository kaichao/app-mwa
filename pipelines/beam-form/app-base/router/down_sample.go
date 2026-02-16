package main

import (
	"beamform/internal/strparse"
	"fmt"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/sirupsen/logrus"
)

func fromDownSample(body string, headers map[string]string) int {
	// input body: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	obsID, pBegin, pEnd, t0, t1, _, err := strparse.ParseParts(body)
	if err != nil {
		logrus.Errorf("message parsing, err-info:%v", err)
		return 1
	}

	cubeID := fmt.Sprintf("%s/p%05d_%05d/t%d_%d", obsID, pBegin, pEnd, t0, t1)
	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	sema := "fits-done:" + cubeID
	semaVal, err := semaphore.AddValue(sema, 0, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
	if semaVal > 0 {
		// 24ch not done.
		return 0
	}

	return toFitsMerge(cubeID)
}
