package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	scalebox "github.com/kaichao/scalebox/golang/misc"
	"gopkg.in/yaml.v2"
)

type dataCube struct {
	DatasetID string

	ChannelBegin  string
	NumOfChannels string

	TimeBegin    string
	NumOfSeconds string
	// 30
	TimeUnit string
	// 30的倍数
	TimeStep string

	PointingBegin string
	PointingEnd   string
	// 通常为24
	PointingStep string
	NumPerBatch  string
}

// DataCube ...
//
//	Time Dimension: TimeUnit, TimeRange
//
//	Pointing Demension: PointingRange, PointingBatch
type DataCube struct {
	DatasetID string

	ChannelBegin  int `yaml:"channelBegin"`
	NumOfChannels int `yaml:"numOfChannels"`

	TimeBegin    int `yaml:"timeBegin"`
	NumOfSeconds int `yaml:"numOfSeconds"`
	// 单个打包文件的时长（30秒）
	TimeUnit int `yaml:"timeUnit"`
	// 单次beam-maker的时长，通常为30的倍数
	TimeStep int `yaml:"timeStep"`

	PointingBegin int `yaml:"pointingBegin"`
	PointingEnd   int `yaml:"pointingEnd"`
	// 单次beam-maker处理的指向数，通常取24的倍数
	PointingStep int `yaml:"pointingStep"`
	// 单批次beam-maker的执行次数，batchIndex从0起
	NumPerBatch int `yaml:"numPerBatch"`
}

func getDataCube(datasetID string) *DataCube {
	cmdText := "scalebox dataset get-metadata " + datasetID
	code, stdout, stderr := scalebox.ExecShellCommandWithExitCode(cmdText, 10)
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

// 三维datacube中，给定顺序号，用于local-tar-pull/cluster-tar-pull运行过程中的的排序

func (cube *DataCube) getSortedTag(time int, ch int) string {
	batchIndex := cube.getSemaPointingBatchIndex(time, ch)
	// p := cube.getPointingBatchIndex(pointing)
	ch -= cube.ChannelBegin
	tm := (time - cube.TimeBegin) / cube.TimeStep
	fmt.Printf("datacube.channelBegin:%d\n", cube.ChannelBegin)
	fmt.Printf("datacube:%v\n", cube)
	fmt.Println("ch=", ch)
	fmt.Println("tm=", tm)

	// 2位指向批次码(pointing-batch) + 2位时间编码（time-range） + 2位通道编码（00~23）
	return fmt.Sprintf("%02d%02d%02d", batchIndex, tm, ch)
}

var (
	datacubeFile = "../dataset-base.yaml"
)

func getDataCubeFromFile(datasetID string) *DataCube {
	config := map[string]map[string]map[string]DataCube{}

	yamlFile, err := ioutil.ReadFile(datacubeFile)
	if err != nil {
		log.Fatalf("读取YAML文件出错：%v", err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("解析YAML文件出错：%v", err)
	}

	cube := config["datasets"][datasetID]["metadata"]
	fmt.Println(cube)
	return &cube
}
