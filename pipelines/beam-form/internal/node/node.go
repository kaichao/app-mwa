package node

import (
	"beamform/internal/datacube"
	"os"
)

// GetIPAddrListByGroupIndex ...
// - 用于toRedist中本组分发
func GetIPAddrListByGroupIndex(groupIndex int) []string {
	var ret []string
	if len(Nodes) <= 24 {
		for i := 0; i < 24; i++ {
			ret = append(ret, Nodes[i%len(Nodes)].IPAddr)
		}
	} else {
		// >= 48 nodes
		numGroup := len(Nodes) / 24
		// groupIndex : [0..numGroup-1]
		start := (groupIndex % numGroup) * 24
		for i := 0; i < 24; i++ {
			ret = append(ret, Nodes[i+start].IPAddr)
		}
	}
	return ret
}

// GetNodeNameByIndexChannel ...
// - 用于pull-unpack
func GetNodeNameByIndexChannel(cube *datacube.DataCube, indexGroup int, ch int) string {
	ch -= cube.ChannelBegin
	numGroups := len(Nodes) / 24
	var index int
	if len(Nodes) < 24 {
		index = ch % len(Nodes)
	} else {
		if os.Getenv("INTERLEAVED_DAT") == "yes" && indexGroup%2 == 1 {
			ch = 23 - ch
		}
		index = (indexGroup%numGroups)*24 + ch
	}

	// fmt.Fprintf(os.Stderr, "len(nodes)=%d,index-group=%d,ch=%d\n", len(Nodes), indexGroup, ch)
	return Nodes[index].Name
}

// GetNodeNameByPointingTime ...
// - 用于fits-merge
func GetNodeNameByPointingTime(cube *datacube.DataCube, p int, t int) string {
	index := cube.GetTimePointingIndex(t, p, len(Nodes))
	return Nodes[index].Name
}
