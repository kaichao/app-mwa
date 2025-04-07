package datacube

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// GetNodeNameByTimeChannel ...
func (cube *DataCube) GetNodeNameByTimeChannel(t int, ch int) string {
	// fmt.Printf("ch=%d\n", ch)
	n := len(NodeNames)
	if (n < 2) || (n%24 != 0 && 24%n != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", n)
		return ""
	}

	if n < 24 {
		// 24的约数
		index := (ch - cube.ChannelBegin) % n
		return NodeNames[index]
	}
	// 24的倍数，也可支持24的约数？
	indexTime := cube.getTimeRangeIndex(t)
	indexCH := ch - cube.ChannelBegin
	index := (indexTime*cube.NumOfChannels + indexCH) % n
	return NodeNames[index]
}

// GetNodeNameListByTime ...
func (cube *DataCube) GetNodeNameListByTime(t int) string {
	n := len(NodeNames)
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
	index := cube.getTimeRangeIndex(t)
	i := index % (n / 24)
	hosts := nodeIPs[i*24 : (i+1)*24]
	return strings.Join(hosts, ",")
}

// GetNodeNameByPointingTime ...
func (cube *DataCube) GetNodeNameByPointingTime(p int, t int) string {
	n := len(NodeNames)
	if (n < 2) || (n%24 != 0 && 24%n != 0) {
		logrus.Warnf("The number of nodes is %d, should be a multiple or a divisor of 24\n", n)
		return ""
	}
	if n <= 24 {
		// 24的约数
		index := (p - cube.PointingBegin) % n
		return NodeNames[index]
	}
	tIndex := cube.getTimeRangeIndex(t)
	i := tIndex % (n / 24)
	index := i*24 + (p-cube.PointingBegin)%24
	fmt.Printf("tIndex=%d,t=%d; p=%d, begin=%d\n",
		tIndex, t, p, cube.PointingBegin)
	return NodeNames[index]
}
