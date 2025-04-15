package datacube

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// GetTimeChannelIndex ...
// 用于pull-unpack、redist
func (cube *DataCube) GetTimeChannelIndex(t int, ch int, numHosts int) int {
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
