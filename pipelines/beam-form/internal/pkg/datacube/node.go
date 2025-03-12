package datacube

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// GetNodeNameByTimeChannel ...
func (cube *DataCube) GetNodeNameByTimeChannel(t int, ch int) string {
	fmt.Printf("ch=%d\n", ch)
	n := len(nodeNames)
	if (n < 2) || (n%24 != 0 && 24%n != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", n)
		return ""
	}

	if n < 24 {
		// 24的约数
		index := (ch - cube.ChannelBegin) % n
		return nodeNames[index]
	}
	// 24的倍数
	indexTime := cube.getTimeRangeIndex(t)
	indexCH := ch - cube.ChannelBegin
	index := (indexTime*cube.NumOfChannels + indexCH) % n
	return nodeNames[index]
}

// GetNodeNameByPointing ...
func (cube *DataCube) GetNodeNameByPointing(t int, ch int) string {
	return ""
}
