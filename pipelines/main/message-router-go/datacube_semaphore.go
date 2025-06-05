package main

import (
	"fmt"

	"mr/datacube"
)

func createDatReadySemaphores(cube *datacube.DataCube) {
	// TARGET: beam-maker
	// 1257010784/1257010786_1257010815/112/00001_00024
	// 1257010784/1257010786_1257010815/112
	ts := cube.GetTimeRanges()
	semaArr := []Sema{}
	for i := 0; i < len(ts); i += 2 {
		// all dat files in current range
		initValue := ts[i+1] - ts[i] + 1
		for ch := 109; ch <= 132; ch++ {
			sema := getSemaDatReadyName(cube, ts[i], ch)
			fmt.Printf("In createDatReadySemaphores(),sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}
func createPointingBatchLeftSemaphores(cube *datacube.DataCube) {
	// pointing-batch-left:1257010784/t1257010786_1257010845/ch119
	initValue := cube.GetNumOfPointingBatch()
	semaArr := []Sema{}

	ts := cube.GetTimeRangesWithinInterval(cube.TimeBegin, cube.TimeBegin+cube.NumOfSeconds-1)
	for i := 0; i < len(ts); i += 2 {
		// all dat files in current range
		for ch := 109; ch <= 132; ch++ {
			sema := getSemaPointingBatchLeftName(cube, ts[i], ch)
			fmt.Printf("In createPointingBatchLeftSemaphores(), sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}
func createFits24chReadySemaphores(cube *datacube.DataCube) {
	// TARGET: fits-merger
	// 1257010784/p00024/t1257010786_1257010815
	// 24-channel
	initValue := 24

	ts := cube.GetTimeRanges()
	semaArr := []Sema{}

	for p := cube.PointingBegin; p <= cube.PointingEnd; p++ {
		for i := 0; i < len(ts); i += 2 {
			sema := getSemaFits24chReadyName(cube, p, ts[i])
			fmt.Printf("In createFits24chReadySemaphores(), sema:%s,init-value:%d\n", sema, initValue)
			semaArr = append(semaArr, Sema{name: sema, value: initValue})
		}
	}
	doInsert(semaArr)
}

func createDatProcessedSemaphores(cube *datacube.DataCube) {
	// dat-processed:1257010784/p00001_00096/t1257010846_1257010905/ch111
	// first batch
	ts := cube.GetTimeRanges()
	semaArr := []Sema{}

	fmt.Printf("cube:%v\n", cube)

	for i := 0; i < len(ts); i += 2 {
		for ch := 109; ch <= 132; ch++ {
			for pIndex := 0; pIndex < cube.GetNumOfPointingBatch(); pIndex++ {
				p := cube.PointingBegin + pIndex*cube.PointingStep*cube.NumPerBatch
				p0, p1 := cube.GetPointingBatchRange(p)
				sema := getSemaDatProcessedName(cube, p, ts[i], ch)
				fmt.Printf("In createDatProcessedSemaphores(), sema:%s,init-value:%d\n", sema, p1-p0+1)
				semaArr = append(semaArr, Sema{name: sema, value: p1 - p0 + 1})
			}
		}
	}
	doInsert(semaArr)
}

func createPullUnpackProgressCountSemaphores(cube *datacube.DataCube) {
	arr := cube.GetTimeUnits()
	lenTimeUnits := len(arr) / 2

	initValue := lenTimeUnits * cube.GetNumOfPointingBatch() * 24 / len(ips) * cube.TimeUnit
	fmt.Printf("PullUnpackProgressCount, initValue=%d,lenTimeUnits=%d,numBatches=%d\n",
		initValue, lenTimeUnits, cube.GetNumOfPointingBatch())
	semaArr := []Sema{}
	for _, h := range hosts {
		// multi-seg support
		sema := fmt.Sprintf("task_progress:pull-unpack:%s", h)
		semaArr = append(semaArr, Sema{name: sema, value: initValue})
	}
	doInsert(semaArr)
}

func createBeamMakerProgressCountSemaphores(cube *datacube.DataCube) {
	arr := cube.GetTimeRanges()
	lenTimeRanges := len(arr) / 2
	lenPointings := cube.PointingEnd - cube.PointingBegin + 1
	initValue := lenTimeRanges * lenPointings * 24 / len(ips)
	fmt.Printf("BeamMakerProgressCount, initValue=%d, lenTimeRanges=%d,lenPointings=%d\n",
		initValue, lenTimeRanges, lenPointings)
	semaArr := []Sema{}
	for _, h := range hosts {
		// multi-seg support
		sema := fmt.Sprintf("task_progress:beam-maker:%s", h)
		semaArr = append(semaArr, Sema{name: sema, value: initValue})
	}
	doInsert(semaArr)
}

func getSemaPointingBatchIndex(cube *datacube.DataCube, t int, ch int) int {
	return doPointingBatchIndex(cube, t, ch, getSemaphore)
}

func countDownSemaPointingBatchIndex(cube *datacube.DataCube, t int, ch int) int {
	return doPointingBatchIndex(cube, t, ch, countDown)
}

func doPointingBatchIndex(cube *datacube.DataCube, t int, ch int, op func(string) int) int {
	t0, t1 := cube.GetTimeRange(t)
	sema := fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%d",
		cube.DatasetID, t0, t1, ch)
	n := op(sema)
	index := cube.GetNumOfPointingBatch() - n
	fmt.Printf("In doGetPointingBatchIndex(), sema:%s, num-of-batch=%d, op(sema)=%d,index=%d \n",
		sema, cube.GetNumOfPointingBatch(), n, index)
	return index
}

func getSemaDatReadyName(cube *datacube.DataCube, t, ch int) string {
	tb, te := cube.GetTimeRange(t)
	sema := fmt.Sprintf("dat-ready:%s/t%d_%d/ch%d", cube.DatasetID, tb, te, ch)
	return sema
}

func getSemaPointingBatchLeftName(cube *datacube.DataCube, t, ch int) string {
	tb, te := cube.GetTimeRange(t)
	sema := fmt.Sprintf("pointing-batch-left:%s/t%d_%d/ch%d", cube.DatasetID, tb, te, ch)
	return sema
}
func getSemaDatProcessedName(cube *datacube.DataCube, p, t, ch int) string {
	p0, p1 := cube.GetPointingBatchRange(p)
	tb, te := cube.GetTimeRange(t)

	sema := fmt.Sprintf("dat-processed:%s/p%05d_%05d/t%d_%d/ch%d",
		cube.DatasetID, p0, p1, tb, te, ch)

	return sema
}
func getSemaFits24chReadyName(cube *datacube.DataCube, p, t int) string {
	tb, te := cube.GetTimeRange(t)
	sema := fmt.Sprintf("fits-24ch-ready:%s/p%05d/t%d_%d",
		cube.DatasetID, p, tb, te)
	return sema
}

// 三维datacube中，给定顺序号，用于pull-unpack/cluster-dist运行过程中的的排序
func getSortTagForDataPull(cube *datacube.DataCube, time int, ch int) string {
	batchIndex := getSemaPointingBatchIndex(cube, time, ch)
	// p := cube.getPointingBatchIndex(pointing)
	ch -= cube.ChannelBegin
	tm := (time - cube.TimeBegin) / cube.TimeStep
	fmt.Printf("datacube.channelBegin:%d\n", cube.ChannelBegin)
	fmt.Printf("datacube:%v\n", cube)
	fmt.Println("ch=", ch)
	fmt.Println("tm=", tm)

	// 2位指向批次码(pointing-batch) + 2位通道编码（00~23）+ 2位时间编码（time-range）
	return fmt.Sprintf("b%02d-t%02d-ch%02d", batchIndex, tm, ch)
}

// 三维datacube中，给定顺序号，用于beam-form运行过程中的的排序
func getSortTagForBeamForm(cube *datacube.DataCube, time int, p int, ch int) string {
	batchIndex := getSemaPointingBatchIndex(cube, time, ch)
	ch -= cube.ChannelBegin
	tm := (time - cube.TimeBegin) / cube.TimeUnit
	fmt.Printf("datacube.channelBegin:%d\n", cube.ChannelBegin)
	fmt.Printf("datacube:%v\n", cube)
	fmt.Println("ch=", ch)
	fmt.Println("tm=", tm)

	// 2位指向批次码(pointing-batch) + 3位时间编码（time-range） 5位指向码 + 2位通道编码（00~23）
	return fmt.Sprintf("b%02d-t%03d-p%05d-ch%02d", batchIndex, tm, p, ch)
}
