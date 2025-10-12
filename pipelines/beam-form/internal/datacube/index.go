package datacube

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetTimeChannelIndexEx ...
// 用于to beam-make、redist
// 不再使用 ？
func (cube *DataCube) GetTimeChannelIndexEx(t int, ch int, numHosts int) int {
	if (numHosts < 2) || (numHosts%24 != 0 && 24%numHosts != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", numHosts)
		return -1
	}
	if numHosts <= 24 {
		// 24的约数
		return (ch - cube.ChannelBegin) % numHosts
	}

	// 24的倍数，也可支持24的约数？
	indexTime := cube.GetTimeRangeIndex(t)
	indexCH := ch - cube.ChannelBegin

	// 获取上一级调用者的信息
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		logrus.Errorln("无法获取调用者信息")
		return -1
	}
	funcName := runtime.FuncForPC(pc).Name()

	// 判断调用者为GetNodeNameByTimeChannel
	if strings.HasSuffix(funcName, ".GetNodeNameByTimeChannel") {
		if os.Getenv("INTERLEAVED_DAT") == "yes" {
			numGroups := numHosts / 24
			// 不同channel数据的计算量不同，交叉分布有助于计算需求在各节点上均衡分布
			if (indexTime/numGroups)%2 == 1 {
				indexCH = 23 - indexCH
			}
		}
	}
	return (indexTime*cube.NumOfChannels + indexCH) % numHosts
}

// GetTimePointingIndex ...
// 用于脉冲星搜索
func (cube *DataCube) GetTimePointingIndex(t int, p int, numHosts int) int {
	if (numHosts < 2) || (numHosts%24 != 0 && 24%numHosts != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", numHosts)
		return -1
	}

	fmt.Printf("p=%d, beg=%d\n", p, cube.PointingBegin)
	if numHosts <= 24 {
		// 24的约数
		return (p - cube.PointingBegin) % numHosts
	}
	n := cube.GetTimeRangeIndex(t) % (numHosts / 24)
	return n*24 + (p-cube.PointingBegin)%24
}

// GetTimeChannelIndex ...
func (cube *DataCube) GetTimeChannelIndex(t int, ch int) int {
	indexTime := cube.GetTimeUnitIndex(t)
	indexCH := ch - cube.ChannelBegin

	return indexTime*cube.NumOfChannels + indexCH
}
