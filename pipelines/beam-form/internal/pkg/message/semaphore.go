package message

import (
	"beamform/internal/pkg/datacube"
	"fmt"
)

// GetSemaphores 获取消息对应的信号量列表
//
// 参数：
//   - m: 消息体。其格式
//     1257010784
//     1257010784/p00001_00960
//     1257010784/t1257012766_1257012965
//     1257010784/p00001_00960/t1257012766_1257012965
//
// 返回值：
//   - string: '\n'分隔的信号量列表。信号量格式："sema_name":sema_value
//
// 环境变量：
//   - []string: 计算出的商
func GetSemaphores(m string) string {
	dataset, pBegin, pEnd, tBegin, tEnd, _, err := ParseParts(m)
	if err != nil {
		return ""
	}
	cube := datacube.GetDataCube(dataset)

	tRanges := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
	pRanges := cube.GetPointingRangesByInterval(pBegin, pEnd)

	semaDatReady := ""
	semaDatDone := ""
	nPRanges := len(pRanges) / 2
	for j := 0; j < len(tRanges); j += 2 {
		tUnits := cube.GetTimeUnitsWithinInterval(tRanges[j], tRanges[j+1])
		nTimeUnits := len(tUnits) / 2
		for i := 0; i < cube.NumOfChannels; i++ {
			id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/ch%d`,
				dataset, pBegin, pEnd, tRanges[j], tRanges[j+1], cube.ChannelBegin+i)
			semaPair := fmt.Sprintf(`"dat-ready:%s":%d`, id, nTimeUnits)
			semaDatReady += semaPair + "\n"

			semaPair = fmt.Sprintf(`"dat-done:%s":%d`, id, nPRanges)
			semaDatDone += semaPair + "\n"
		}
	}

	semaFitsDone := ""
	// fits-done:1257010784/p00001/t1257010786_1257010985
	for k := 0; k < len(pRanges); k += 2 {
		for j := 0; j < len(tRanges); j += 2 {
			id := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d`, dataset, pRanges[k], pRanges[k+1], tRanges[j], tRanges[j+1])
			semaPair := fmt.Sprintf(`"fits-done:%s":%d`, id, 24)
			semaFitsDone += semaPair + "\n"
		}
	}

	semaPointingDone := ""
	// pointing-done:1257010784/p00001
	nTimeRanges := len(tRanges) / 2
	for p := pBegin; p <= pEnd; p++ {
		semaPair := fmt.Sprintf(`"pointing-done:%s/p%05d":%d`,
			dataset, p, nTimeRanges)
		semaPointingDone += semaPair + "\n"
	}
	semaphores := semaDatReady + semaDatDone + semaFitsDone + semaPointingDone

	return semaphores
}
