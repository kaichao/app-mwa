package node

import (
	"beamform/internal/datacube"
	"fmt"
	"os"
)

// GetNodeNameByIndexChannel ...
// - 用于toBeamMake
func GetNodeNameByIndexChannel(cube *datacube.DataCube, indexGroup int, ch int) string {
	var index int
	if len(Nodes) < 24 {
		index = (ch - cube.ChannelBegin) % len(Nodes)
	} else {
		numGroups := len(Nodes) / 24
		index = (indexGroup%numGroups)*24 + ch - cube.ChannelBegin
	}

	fmt.Fprintf(os.Stderr, "len(nodes)=%d,index-group=%d,ch=%d\n", len(Nodes), indexGroup, ch)
	return Nodes[index].Name
}

// GetNodeNameByTimeChannel ...
// - 用于toBeamMake
func GetNodeNameByTimeChannel(cube *datacube.DataCube, t int, ch int) string {
	index := cube.GetTimeChannelIndex(t, ch, len(Nodes))
	return Nodes[index].Name
}

// GetIPAddrListByTime ...
// - 用于toRedist
func GetIPAddrListByTime(cube *datacube.DataCube, t int) []string {
	ips := []string{}
	for ch := 109; ch < 133; ch++ {
		index := cube.GetTimeChannelIndex(t, ch, len(Nodes))
		ips = append(ips, Nodes[index].IPAddr)
	}
	return ips
}

// GetNodeNameByPointingTime ...
func GetNodeNameByPointingTime(cube *datacube.DataCube, p int, t int) string {
	index := cube.GetTimePointingIndex(t, p, len(Nodes))
	return Nodes[index].Name
}
