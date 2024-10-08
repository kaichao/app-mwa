package datacube

import (
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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
	// 单次beam-maker的时长，通常为40的倍数；120/160/200/240/320/400
	TimeStep int `yaml:"timeStep"`
	// 24节点为单元的分区数量
	NumPerSeg int `yaml:"numPerSeg"`

	PointingBegin int `yaml:"pointingBegin"`
	PointingEnd   int `yaml:"pointingEnd"`
	// 单次beam-maker处理的指向数，通常取24的倍数
	PointingStep int `yaml:"pointingStep"`
	// 单批次beam-maker的执行次数，batchIndex从0起
	NumPerBatch int `yaml:"numPerBatch"`
}

var (
	datacubeFile = "/dataset.yaml"
	// datacubeFile = "/dataset-perf-test.yaml"
	// datacubeFile = "/dataset-base.yaml"

	// GetDataCube ...
	GetDataCube = getDataCubeFromFile
)

func getDataCubeFromFile(datasetID string) *DataCube {
	config := map[string]map[string]DataCube{}

	yamlFile, err := ioutil.ReadFile(datacubeFile)
	if err != nil {
		logrus.Fatalf("Read yaml file %s, err:%v", datacubeFile, err)
	}

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		logrus.Errorf("Error parsing yaml file %s, err:%v", datacubeFile, err)
	}

	fmt.Println("config:", config)
	cube := config["datasets"][datasetID]
	if cube.NumOfSeconds == 0 {
		cube.NumOfSeconds = cube.TimeEnd - cube.TimeBegin + 1
	}
	if cube.NumPerSeg == 0 {
		cube.NumPerSeg = 1
	}

	return &cube
}

// GetNumWithBlockID ...
func (cube *DataCube) GetNumWithBlockID(t int, num int) int {
	if cube.NumPerSeg == 1 {
		return num
	}

	index := cube.getTimeRangeIndex(t)
	return (index%cube.NumPerSeg)*24 + num
}
