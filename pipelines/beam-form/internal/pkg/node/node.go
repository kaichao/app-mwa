package node

import (
	"beamform/internal/pkg/datacube"
)

// GetNodeNameByTimeChannel ...
func GetNodeNameByTimeChannel(cube *datacube.DataCube, t int, ch int) string {
	index := cube.GetTimeChannelIndex(t, ch, len(NodeNames))
	return NodeNames[index]
}

// GetIPAddrListByTime ...
func GetIPAddrListByTime(cube *datacube.DataCube, t int) []string {
	ips := []string{}
	for ch := 109; ch < 133; ch++ {
		index := cube.GetTimeChannelIndex(t, ch, len(NodeNames))
		ips = append(ips, nodeIPs[index])
	}
	return ips
}

// GetNodeNameByPointingTime ...
func GetNodeNameByPointingTime(cube *datacube.DataCube, p int, t int) string {
	index := cube.GetTimePointingIndex(t, p, len(NodeNames))
	return NodeNames[index]
}
