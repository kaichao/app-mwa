package node

import (
	"beamform/internal/datacube"
	"os"
)

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

// GetNodeNameByPointingTime ...
// func GetNodeNameByPointingTime(cube *datacube.DataCube, p int, t int) string {
// 	index := cube.GetTimePointingIndex(t, p, len(Nodes))
// 	return Nodes[index].Name
// }

// GetNodeNameByGroupIndexPointing ...
// - 用于fits-merge任务生成。
func GetNodeNameByGroupIndexPointing(groupIndex, p int) string {
	groupSize := 24
	// TODO: 如果节点数小于24，会报错。
	// 仅考虑单组小于24节点的情况。
	// 如果多组，总节点数依然小于24，比如，2组8个节点共16个节点的情况，还需再考虑
	if len(Nodes) < 24 {
		groupSize = len(Nodes)
	}
	index := groupIndex*groupSize + (p-1)%groupSize
	return Nodes[index].Name
}
