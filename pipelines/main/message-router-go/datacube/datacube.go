package datacube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type dataCube struct {
	DatasetID string

	ChannelBegin  string
	NumOfChannels string

	TimeBegin    string
	NumOfSeconds string
	// 40
	TimeUnit string
	// 40的倍数
	TimeStep string

	PointingBegin string
	PointingEnd   string
	// 通常取24
	PointingStep string
	NumPerBatch  string
}

// DataCube ...
//
//	Time Dimension: TimeUnit, TimeRange
//
//	Pointing Demension: PointingRange, PointingBatch
type DataCube struct {
	DatasetID string `yaml:"datasetID"`

	ChannelBegin  int `yaml:"channelBegin"`
	NumOfChannels int `yaml:"numOfChannels"`

	TimeBegin    int `yaml:"timeBegin"`
	NumOfSeconds int `yaml:"numOfSeconds"`
	TimeEnd      int `yaml:"timeEnd"`
	// 单个打包文件的时长（40秒）
	TimeUnit int `yaml:"timeUnit"`
	// 单次beam-maker的时长，通常为40的倍数
	TimeStep int `yaml:"timeStep"`

	PointingBegin int `yaml:"pointingBegin"`
	PointingEnd   int `yaml:"pointingEnd"`
	// 单次beam-maker处理的指向数，通常取24的倍数
	PointingStep int `yaml:"pointingStep"`
	// 单批次beam-maker的执行次数，batchIndex从0起
	NumPerBatch int `yaml:"numPerBatch"`
}

func getDataCubeFromDB(datasetID string) *DataCube {
	cmdText := "scalebox dataset get-metadata " + datasetID
	code, stdout, stderr := misc.ExecShellCommandWithExitCode(cmdText, 10)
	fmt.Fprintf(os.Stderr, "stderr for dataset-get-metadata:\n%s\n", stderr)
	if code != 0 {
		fmt.Fprintf(os.Stderr, "[WARN] error for dataset-get-metadata dataset=%s in getDataCube()\n", datasetID)
		return nil
	}

	var (
		dc       dataCube
		datacube DataCube
	)
	if err := json.Unmarshal([]byte(stdout), &dc); err != nil {
		// skip non-json format error
		if !strings.HasPrefix(err.Error(), "invalid character") {
			fmt.Printf("error parsing, err-info:%v\n", err)
		}
		// non-datacube definition
		return nil
	}

	datacube.DatasetID = dc.DatasetID

	datacube.ChannelBegin, _ = strconv.Atoi(dc.ChannelBegin)
	datacube.NumOfChannels, _ = strconv.Atoi(dc.NumOfChannels)

	datacube.TimeBegin, _ = strconv.Atoi(dc.TimeBegin)
	datacube.NumOfSeconds, _ = strconv.Atoi(dc.NumOfSeconds)
	datacube.TimeUnit, _ = strconv.Atoi(dc.TimeUnit)
	datacube.TimeStep, _ = strconv.Atoi(dc.TimeStep)

	datacube.PointingBegin, _ = strconv.Atoi(dc.PointingBegin)
	datacube.PointingEnd, _ = strconv.Atoi(dc.PointingEnd)
	datacube.PointingStep, _ = strconv.Atoi(dc.PointingStep)
	datacube.NumPerBatch, _ = strconv.Atoi(dc.NumPerBatch)

	return &datacube
}

var (
	datacubeFile = "/dataset.yaml"
	// datacubeFile = "/dataset-perf-test.yaml"
	// datacubeFile = "/dataset-base.yaml"

	// GetDataCube ...
	GetDataCube = getDataCubeFromFile
)

func getDataCubeFromFile(datasetID string) *DataCube {
	config := map[string]map[string]map[string]DataCube{}

	yamlFile, err := ioutil.ReadFile(datacubeFile)
	if err != nil {
		logrus.Fatalf("Read yaml file %s, err:%v", datacubeFile, err)
	}

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		logrus.Errorf("Error parsing yaml file %s, err:%v", datacubeFile, err)
	}

	fmt.Println("config:", config)
	cube := config["datasets"][datasetID]["metadata"]
	if cube.NumOfSeconds == 0 {
		cube.NumOfSeconds = cube.TimeEnd - cube.TimeBegin + 1
	}
	fmt.Println(cube)
	return &cube
}
