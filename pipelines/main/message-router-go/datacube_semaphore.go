package main

import "fmt"

func (cube *DataCube) createDatReadySemaphores() {
	// TARGET: beam-maker
	// 1257010784/1257010786_1257010815/112/00001_00024
	// 1257010784/1257010786_1257010815/112
	ts := cube.getTimeRanges()
	semaArr := []Sema{}
	for i := 0; i < len(ts); i += 2 {
		// all dat files in current range
		initValue := ts[i+1] - ts[i] + 1
		for ch := 109; ch <= 132; ch++ {
			sema := cube.getSemaDatReadyName(ts[i], ch)
			fmt.Printf("In createDatReadySemaphores(),sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}
func (cube *DataCube) createPointingBatchLeftSemaphores() {
	// pointing-batch-left:1257010784/t1257010786_1257010845/ch119
	initValue := cube.getNumOfPointingBatch()
	semaArr := []Sema{}

	ts := cube.getTimeRangesWithinInterval(cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
	for i := 0; i < len(ts); i += 2 {
		// all dat files in current range
		for ch := 109; ch <= 132; ch++ {
			sema := cube.getSemaPointingBatchLeftName(ts[i], ch)
			fmt.Printf("In createPointingBatchLeftSemaphores(), sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}
func (cube *DataCube) createFits24chReadySemaphores() {
	// TARGET: fits-merger
	// 1257010784/p00024/t1257010786_1257010815
	// 24-channel
	initValue := 24

	ts := cube.getTimeRanges()
	semaArr := []Sema{}

	for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
		for i := 0; i < len(ts); i += 2 {
			sema := cube.getSemaFits24chReadyName(p, ts[i])
			fmt.Printf("In createFits24chReadySemaphores(), sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}

func (cube *DataCube) createDatProcessedSemaphores() {
	// dat-processed:1257010784/p00001_00096/t1257010846_1257010905/ch111
	// first batch
	ts := cube.getTimeRanges()
	semaArr := []Sema{}

	fmt.Printf("cube:%v\n", cube)

	for i := 0; i < len(ts); i += 2 {
		for ch := 109; ch <= 132; ch++ {
			for pIndex := 0; pIndex < cube.getNumOfPointingBatch(); pIndex++ {
				p := cube.PointingBegin + pIndex*cube.PointingStep*cube.NumPerBatch
				p0, p1 := cube.getPointingBatchRange(p)
				sema := cube.getSemaDatProcessedName(p, ts[i], ch)
				fmt.Printf("In createDatProcessedSemaphores(), sema:%s,init-value:%d\n", sema, p1-p0+1)
				semaArr = append(semaArr, Sema{name: sema, value: p1 - p0 + 1})
			}
		}
	}
	doInsert(semaArr)
}

//	func (cube *DataCube) createLocalTarPullProgressCountSemaphores() {
//		arr := cube.getTimeUnits()
//		lenTimeUnits := len(arr) / 2
//		initValue := lenTimeUnits * cube.getNumOfPointingBatch() * 24 / len(ips)
//		fmt.Printf("LocalTarPullProgressCount, initValue=%d,lenTimeUnits=%d,numBatches=%d\n",
//			initValue, lenTimeUnits, cube.getNumOfPointingBatch())
//		semaArr := []Sema{}
//		for _, h := range ips {
//			sema := "progress-counter_local-tar-pull:" + h
//			semaArr = append(semaArr, Sema{name: sema, value: initValue})
//		}
//		doInsert(semaArr)
//	}

func (cube *DataCube) createPullUnpackProgressCountSemaphores() {
	arr := cube.getTimeUnits()
	lenTimeUnits := len(arr) / 2
	initValue := lenTimeUnits * cube.getNumOfPointingBatch() * 24 / len(ips)
	fmt.Printf("PullUnpackProgressCount, initValue=%d,lenTimeUnits=%d,numBatches=%d\n",
		initValue, lenTimeUnits, cube.getNumOfPointingBatch())
	semaArr := []Sema{}
	for _, h := range ips {
		sema := "progress-counter_pull-unpack:" + h
		semaArr = append(semaArr, Sema{name: sema, value: initValue})
	}
	doInsert(semaArr)
}

func (cube *DataCube) createBeamMakerProgressCountSemaphores() {
	arr := cube.getTimeRanges()
	lenTimeRanges := len(arr) / 2
	lenPointings := cube.PointingEnd - cube.PointingBegin + 1
	initValue := lenTimeRanges * lenPointings * 24 / len(ips)
	fmt.Printf("BeamMakerProgressCount, initValue=%d, lenTimeRanges=%d,lenPointings=%d\n",
		initValue, lenTimeRanges, lenPointings)
	semaArr := []Sema{}
	for _, h := range ips {
		sema := "progress-counter_beam-maker:" + h
		semaArr = append(semaArr, Sema{name: sema, value: initValue})
	}
	doInsert(semaArr)
}

func (cube *DataCube) getSemaPointingBatchIndex(t int, ch int) int {
	return doPointingBatchIndex(cube, t, ch, getSemaphore)
}

func (cube *DataCube) countDownSemaPointingBatchIndex(t int, ch int) int {
	return doPointingBatchIndex(cube, t, ch, countDown)
}

func doPointingBatchIndex(cube *DataCube, t int, ch int, op func(string) int) int {
	t0, t1 := cube.getTimeRange(t)
	sema := fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%d",
		cube.DatasetID, t0, t1, ch)
	n := op(sema)
	index := cube.getNumOfPointingBatch() - n
	fmt.Printf("In doGetPointingBatchIndex(), sema:%s, num-of-batch=%d, op(sema)=%d,index=%d \n",
		sema, cube.getNumOfPointingBatch(), n, index)
	return index
}

func (cube *DataCube) getSemaDatReadyName(t, ch int) string {
	tb, te := cube.getTimeRange(t)
	sema := fmt.Sprintf("dat-ready:%s/t%d_%d/ch%d", cube.DatasetID, tb, te, ch)
	return sema
}

func (cube *DataCube) getSemaPointingBatchLeftName(t, ch int) string {
	tb, te := cube.getTimeRange(t)
	sema := fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%d", cube.DatasetID, tb, te, ch)
	return sema
}
func (cube *DataCube) getSemaDatProcessedName(p, t, ch int) string {
	p0, p1 := cube.getPointingBatchRange(p)
	tb, te := cube.getTimeRange(t)

	sema := fmt.Sprintf("dat-processed:%s/p%05d_%05d/t%d_%d/ch%d",
		cube.DatasetID, p0, p1, tb, te, ch)

	return sema
}
func (cube *DataCube) getSemaFits24chReadyName(p, t int) string {
	tb, te := cube.getTimeRange(t)
	sema := fmt.Sprintf("fits-24ch-ready:%s/p%05d/t%d_%d",
		cube.DatasetID, p, tb, te)
	return sema
}
