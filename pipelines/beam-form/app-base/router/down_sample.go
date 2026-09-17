package main

import (
	"beamform/internal/strparse"
	"fmt"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/semaphore"
)

func fromDownSample(body string, headers map[string]string) error {
	// input body: 1257010784/p00001_00024/t1257012766_1257012965/ch109
	obsID, pBegin, pEnd, t0, t1, _, err := strparse.ParseParts(body)
	if err != nil {
		return errors.WrapE(err, 1, "parse task-body", "task-body", body)
	}

	cubeID := fmt.Sprintf("%s/p%05d_%05d/t%d_%d", obsID, pBegin, pEnd, t0, t1)
	// semaphore: fits-done:1257010784/p00001_00024/t1257010786_1257010985
	sema := "fits-done:" + cubeID
	semaVal, err := semaphore.AddValue(sema, -1, appID)
	if err != nil {
		return errors.WrapE(err, 2, "semaphore-decrement",
			"sema-name", sema, "app-id", appID)
	}
	if semaVal > 0 {
		// 24ch not done.
		return nil
	}

	return toFitsMerge(cubeID)
}
