package main

import (
	"beamform/internal/strparse"
	"fmt"
	"strconv"

	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/sirupsen/logrus"
)

func fromDownSample(message string, headers map[string]string) int {
	// input message: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	// re := regexp.MustCompile(`^(([0-9]+)/p([0-9]+)_([0-9]+)/(t[0-9]+_[0-9]+))(/ch[0-9]+)$`)
	// ss := re.FindStringSubmatch(message)
	// if ss == nil {
	// 	logrus.Errorf("Invalid format, message:%s\n", message)
	// 	return 1
	// }
	// fmt.Println("message-parts:", ss)
	// ds := ss[2]
	// pBegin, _ := strconv.Atoi(ss[3])
	// pEnd, _ := strconv.Atoi(ss[4])
	// t := ss[5]

	obsID, pBegin, pEnd, t0, t1, _, err := strparse.ParseParts(message)
	if err != nil {
		logrus.Errorf("message parsing, err-info:%v", err)
		return 1
	}

	cubeID := fmt.Sprintf("%s/p%05d_%05d/t%d_%d", obsID, pBegin, pEnd, t0, t1)
	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	sema := "fits-done:" + cubeID
	v, err := semaphore.AddValue(sema, 0, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s,err-info=%v\n", sema, err)
		return 2
	}
	semaValue, _ := strconv.Atoi(v)
	if semaValue > 0 {
		// 24ch not done.
		return 0
	}

	return toFitsMerge(cubeID)
}
