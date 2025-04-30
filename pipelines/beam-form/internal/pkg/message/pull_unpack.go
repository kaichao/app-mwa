package message

import (
	"beamform/internal/pkg/datacube"
	"beamform/internal/pkg/node"
	"fmt"
	"os"

	"github.com/kaichao/scalebox/pkg/common"
)

// GetMessagesForPullUnpack ...
func GetMessagesForPullUnpack(m string, hostBound bool) []string {
	dataset, pBegin, pEnd, tBegin, tEnd, _, err := ParseParts(m)
	if err != nil {
		return []string{}
	}
	cube := datacube.GetDataCube(dataset)
	ts := cube.GetTimeRangesWithinInterval(tBegin, tEnd)
	messages := []string{}
	withPointingPath := os.Getenv("WITH_POINTING_PATH") == "yes"
	prefix := dataset
	if withPointingPath {
		prefix = fmt.Sprintf("%s/p%05d_%05d", dataset, pBegin, pEnd)
	}

	numGroups := len(node.NodeNames) / 24
	for j := 0; j < len(ts); j += 2 {
		for i := 0; i < cube.NumOfChannels; i++ {
			hValue := fmt.Sprintf("%s/t%d_%d/ch%d",
				prefix, ts[j], ts[j+1], cube.ChannelBegin+i)
			headers := common.SetJSONAttribute("{}", "target_subdir", hValue)
			if hostBound {
				headers = common.SetJSONAttribute(headers, "to_host",
					node.GetNodeNameByTimeChannel(cube, ts[j], i+cube.ChannelBegin))
			}
			if numGroups > 0 && j < 2*numGroups {
				// > 24节点，首次加载设置更高的带宽
				headers = common.SetJSONAttribute(headers, "bw_limit", os.Getenv("FIRST_BW_LIMIT"))
			}
			tus := cube.GetTimeUnitsWithinInterval(ts[j], ts[j+1])
			for k := 0; k < len(tus); k += 2 {
				body := fmt.Sprintf("%s/p%05d_%05d/%d_%d_ch%d.dat.tar.zst",
					dataset, pBegin, pEnd, tus[k], tus[k+1], cube.ChannelBegin+i)
				messages = append(messages, body+","+headers)
			}
		}
	}
	return messages
}
