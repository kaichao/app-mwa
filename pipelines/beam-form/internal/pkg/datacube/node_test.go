package datacube

import (
	"fmt"
	"os"
	"testing"
)

func TestGetNodeNameByTimeChannel(t *testing.T) {
	nodeNames = []string{}
	loadNodeNamesMock(72)

	fmt.Println("node-names:", nodeNames)

	os.Setenv("DATACUBE_FILE", "../../../dataset.yaml")
	cube := GetDataCube("1255803168")
	fmt.Println("cube:", cube)
	for tm := cube.TimeBegin; tm < cube.TimeBegin+5*cube.TimeStep; tm += cube.TimeStep {
		for ch := 109; ch <= 132; ch++ {
			// fmt.Printf("t=%d,ch=%d,node:%s\n", tm, ch, cube.GetNodeNameByTimeChannel(tm, ch))
			fmt.Printf("%s ", cube.GetNodeNameByTimeChannel(tm, ch))
		}
		fmt.Println()
	}
}

// 从节点数量生成NodeName列表
func loadNodeNamesMock(n int) {
	if n < 24 {
		for i := 0; i < n; i++ {
			nodeNames = append(nodeNames, fmt.Sprintf("n%03d", i))
		}
	} else {
		for j := 0; j < n/24; j++ {
			for i := 0; i < 24; i++ {
				nodeNames = append(nodeNames, fmt.Sprintf("g%02dn%02d", j, i))
			}
		}
	}
}
