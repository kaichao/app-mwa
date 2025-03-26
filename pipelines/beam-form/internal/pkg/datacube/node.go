package datacube

import (
	"strings"

	"github.com/sirupsen/logrus"
)

// GetNodeNameByTimeChannel ...
func (cube *DataCube) GetNodeNameByTimeChannel(t int, ch int) string {
	// fmt.Printf("ch=%d\n", ch)
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
	// 24的倍数，也可支持24的约数？
	indexTime := cube.getTimeRangeIndex(t)
	indexCH := ch - cube.ChannelBegin
	index := (indexTime*cube.NumOfChannels + indexCH) % n
	return nodeNames[index]
}

// GetNodeNameListByTime ...
func (cube *DataCube) GetNodeNameListByTime(t int) string {
	n := len(nodeNames)
	if (n < 2) || (n%24 != 0 && 24%n != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", n)
		return ""
	}
	if n <= 24 {
		// 24的约数
		var hosts []string
		// indexCh := (ch - cube.ChannelBegin) % n
		for i := 0; i < 24/n; i++ {
			hosts = append(hosts, nodeIPs...)
		}
		return strings.Join(hosts, ",")
	}

	return ""
}

// GetNodeNameByPointing ...
func (cube *DataCube) GetNodeNameByPointing(p int) string {
	n := len(nodeNames)
	if (n < 2) || (n%24 != 0 && 24%n != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", n)
		return ""
	}
	if n <= 24 {
		// 24的约数
		index := (p - cube.PointingBegin) % n
		// n := (ch - cube.ChannelBegin) % n
		// fmt.Printf("pointing:%d,ch:%d,index-p:%d,index-ch:%d\n", p, ch, index, n)
		// if n == index {
		// 	return "localhost"
		// }
		return nodeNames[index]
	}
	index := (p - cube.PointingBegin) % n
	return nodeNames[index]
}
