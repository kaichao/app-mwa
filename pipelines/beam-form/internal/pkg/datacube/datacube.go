package datacube

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

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
	TimeEnd      int `yaml:"timeEnd"`
	// 单个打包文件的时长（40秒）
	TimeUnit int `yaml:"timeUnit"`
	// 单次beam-maker的时长，通常为40的倍数；120/160/200/240/320/400
	TimeStep int `yaml:"timeStep"`
	// 24节点为单元的分区数量

	PointingBegin int `yaml:"pointingBegin"`
	PointingEnd   int `yaml:"pointingEnd"`
	// 单次beam-maker处理的指向数，通常取24的倍数
	PointingStep int `yaml:"pointingStep"`
}

var (
	datacubeFile = "/dataset.yaml"

	// GetDataCube ...
	GetDataCube = getDataCubeFromFile
)

func getDataCubeFromFile(datasetID string) *DataCube {
	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		logrus.Errorln("Error:", err)
	}
	fmt.Println("Current Directory:", dir)

	config := map[string]map[string]DataCube{}

	if f := os.Getenv("DATACUBE_FILE"); f != "" {
		datacubeFile = f
	}
	fmt.Printf("datacube-file:%s\n", datacubeFile)
	yamlFile, err := os.ReadFile(datacubeFile)
	if err != nil {
		logrus.Fatalf("Read yaml file %s, err:%v", datacubeFile, err)
	}

	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		logrus.Errorf("Error parsing yaml file %s, err:%v", datacubeFile, err)
	}

	// fmt.Println("config:", config)

	re := regexp.MustCompile(`^([0-9]+)(/p([0-9]+)_([0-9]+))?$`)
	ss := re.FindStringSubmatch(datasetID)

	if len(ss) == 0 {
		logrus.Errorf("Invalid format, datasetID=%s\n", datasetID)
		return nil
	}
	datasetID = ss[1]
	p0, _ := strconv.Atoi(ss[3])
	p1, _ := strconv.Atoi(ss[4])

	cube := config["datasets"][datasetID]
	cube.DatasetID = datasetID
	if cube.NumOfSeconds == 0 {
		cube.NumOfSeconds = cube.TimeEnd - cube.TimeBegin + 1
	}

	if cube.NumOfChannels == 0 {
		cube.NumOfChannels = 24
	}
	if cube.TimeUnit == 0 {
		cube.TimeUnit = 40
	}
	if cube.TimeStep == 0 {
		cube.TimeStep = 200
	}
	if cube.PointingBegin == 0 {
		cube.PointingBegin = 1
	}
	if cube.PointingStep == 0 {
		cube.PointingStep = 24
	}
	// 设定定制的time_step，用于测试
	if v, _ := strconv.Atoi(os.Getenv("TIME_STEP")); v > 0 {
		cube.TimeStep = v
	}
	// 设定定制的pointing_begin
	if v, _ := strconv.Atoi(os.Getenv("POINTING_BEGIN")); v > 0 {
		cube.PointingBegin = v
	}
	// 设定定制的pointing_end
	if v, _ := strconv.Atoi(os.Getenv("POINTING_ENG")); v > 0 {
		cube.PointingEnd = v
	}

	if p0 > 0 {
		cube.PointingBegin = p0
		cube.PointingEnd = p1
	}

	return &cube
}

// GetHostIndex ...
func (cube *DataCube) GetHostIndex(t, index, numHosts int) int {
	// index := ch - cube.ChannelBegin
	if numHosts < 24 {
		return index % numHosts
	} else if numHosts == 24 {
		return index
	}
	rangeIndex := cube.getTimeRangeIndex(t)
	numSeg := numHosts / 24
	return (rangeIndex%numSeg)*24 + index
}

// ToCubeString ...
func (cube *DataCube) ToCubeString() string {
	return fmt.Sprintf(`
			cube: 
				t0=%d,t1=%d, tstep=%d
				p0=%d,p1=%d, pstep=%d \n
		`,
		cube.TimeBegin, cube.TimeEnd, cube.TimeStep,
		cube.PointingBegin, cube.PointingEnd, cube.PointingStep)
}
