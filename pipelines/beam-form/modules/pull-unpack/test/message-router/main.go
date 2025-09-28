package main

import (
	"beamform/internal/datacube"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func main() {
	if len(os.Args) < 3 {
		logrus.Errorf("usage: %s <headers> <message>\nparameters expect=2,actual=%d\n",
			os.Args[0], len(os.Args)-1)
		os.Exit(1)
	}

	headers := make(map[string]string)
	if err := json.Unmarshal([]byte(os.Args[2]), &headers); err != nil {
		logrus.Fatalf("err:%v\n", err)
		os.Exit(2)
	}

	if headers["from_module"] == "pull-unpack" {
		logrus.Printf("message from pull-unpack")
		os.Exit(0)
	}

	code := toPullUnpack(os.Args[1], headers)

	os.Exit(code)
}

func toPullUnpack(body string, fromHeaders map[string]string) int {
	cube := datacube.NewDataCube(body)
	fmt.Println(cube.ToCubeString())
	trs := cube.GetTimeRanges()
	fmt.Printf("len(trs)=%d,trs=%v\n", len(trs), trs)

	prefix := fmt.Sprintf("%s/p%05d_%05d", cube.ObsID, cube.PointingBegin, cube.PointingEnd)

	messages := []string{}

	for j := 0; j < cube.NumOfChannels; j++ {
		ch := cube.ChannelBegin + j
		for i := 0; i < len(trs); i += 2 {

			targetSubDir := fmt.Sprintf("%s/t%d_%d/ch%d", cube.ObsID, trs[i], trs[i+1], ch)
			headers := fmt.Sprintf(`{"target_subdir":"%s"}`, targetSubDir)

			sourceURL := os.Getenv("SOURCE_TAR_ROOT")
			headers = common.SetJSONAttribute(headers, "source_url", sourceURL)
			tus := cube.GetTimeUnitsWithinInterval(trs[i], trs[i+1])

			for k := 0; k < len(tus); k += 2 {
				m := fmt.Sprintf("%s/%d_%d_ch%d.dat.tar.zst", prefix, tus[k], tus[k+1], ch)
				messages = append(messages, m+","+headers)
			}
		}
	}

	envs := map[string]string{
		"SINK_MODULE": "pull-unpack",
	}

	return task.AddTasks(messages, "{}", envs)
}
