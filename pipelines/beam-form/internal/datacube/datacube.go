package datacube

import (
	"beamform/internal/strparse"
	"fmt"
	"os"
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
	ObsID     string

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

	TimeTailMerge bool `yaml:"timeTailMerge"`
}

var (
	datacubeFile = "/dataset.yaml"
)

// NewDataCube ...
func NewDataCube(datasetID string) *DataCube {
	obsID, p0, p1, t0, t1, _, err := strparse.ParseParts(datasetID)
	if err != nil {
		logrus.Errorf("New dataset, err-info:%v\n", err)
		return nil
	}

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
		logrus.Fatalf("Error parsing yaml file %s, err:%v", datacubeFile, err)
	}

	cube := config["datasets"][obsID]

	cube.ObsID = obsID
	cube.DatasetID = datasetID

	if p0 > 0 {
		cube.PointingBegin = p0
	}

	if p1 > 0 {
		cube.PointingEnd = p1
	}

	if t0 > 0 {
		cube.TimeBegin = t0
	}
	if t1 > 0 {
		cube.TimeEnd = t1
	}

	// set default value
	if cube.NumOfChannels == 0 {
		cube.NumOfChannels = 24
	}
	if cube.TimeUnit == 0 {
		cube.TimeUnit = 40
	}
	if cube.TimeStep == 0 {
		cube.TimeStep = 200
	}
	if cube.TimeEnd == 0 {
		cube.TimeEnd = cube.TimeBegin + cube.NumOfSeconds - 1
	}
	if cube.NumOfSeconds == 0 {
		cube.NumOfSeconds = cube.TimeEnd - cube.TimeBegin + 1
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
	// 设定定制的time_begin，用于测试
	if v, _ := strconv.Atoi(os.Getenv("TIME_BEGIN")); v > 0 {
		cube.TimeBegin = v
	}
	// 设定定制的time_end，用于测试
	if v, _ := strconv.Atoi(os.Getenv("TIME_END")); v > 0 {
		cube.TimeEnd = v
	}
	// 设定定制的pointing_begin
	if v, _ := strconv.Atoi(os.Getenv("POINTING_BEGIN")); v > 0 {
		cube.PointingBegin = v
	}
	// 设定定制的pointing_end
	if v, _ := strconv.Atoi(os.Getenv("POINTING_ENG")); v > 0 {
		cube.PointingEnd = v
	}

	return &cube
}

// // GetDataCubeFromFile ...
// func GetDataCubeFromFile(obsID string) *DataCube {
// 	// 获取当前工作目录
// 	dir, err := os.Getwd()
// 	if err != nil {
// 		logrus.Errorln("Error:", err)
// 	}
// 	fmt.Println("Current Directory:", dir)

// 	config := map[string]map[string]DataCube{}

// 	if f := os.Getenv("DATACUBE_FILE"); f != "" {
// 		datacubeFile = f
// 	}
// 	fmt.Printf("datacube-file:%s\n", datacubeFile)
// 	yamlFile, err := os.ReadFile(datacubeFile)
// 	if err != nil {
// 		logrus.Fatalf("Read yaml file %s, err:%v", datacubeFile, err)
// 	}

// 	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
// 		logrus.Fatalf("Error parsing yaml file %s, err:%v", datacubeFile, err)
// 	}

// 	re := regexp.MustCompile(`^([0-9]+)(/p([0-9]+)_([0-9]+))?$`)
// 	ss := re.FindStringSubmatch(obsID)

// 	if len(ss) == 0 {
// 		logrus.Errorf("Invalid format, obsID=%s\n", obsID)
// 		return nil
// 	}
// 	obsID = ss[1]
// 	p0, _ := strconv.Atoi(ss[3])
// 	p1, _ := strconv.Atoi(ss[4])

// 	cube := config["datasets"][obsID]
// 	cube.ObsID = obsID

// 	if cube.NumOfChannels == 0 {
// 		cube.NumOfChannels = 24
// 	}
// 	if cube.TimeUnit == 0 {
// 		cube.TimeUnit = 40
// 	}
// 	if cube.TimeStep == 0 {
// 		cube.TimeStep = 200
// 	}
// 	if cube.TimeEnd == 0 {
// 		cube.TimeEnd = cube.TimeBegin + cube.NumOfSeconds - 1
// 	}
// 	if cube.PointingBegin == 0 {
// 		cube.PointingBegin = 1
// 	}
// 	if cube.PointingStep == 0 {
// 		cube.PointingStep = 24
// 	}
// 	// 设定定制的time_step，用于测试
// 	if v, _ := strconv.Atoi(os.Getenv("TIME_STEP")); v > 0 {
// 		cube.TimeStep = v
// 	}
// 	// 设定定制的time_begin，用于测试
// 	if v, _ := strconv.Atoi(os.Getenv("TIME_BEGIN")); v > 0 {
// 		cube.TimeBegin = v
// 	}
// 	// 设定定制的time_end，用于测试
// 	if v, _ := strconv.Atoi(os.Getenv("TIME_END")); v > 0 {
// 		cube.TimeEnd = v
// 	}
// 	// 设定定制的pointing_begin
// 	if v, _ := strconv.Atoi(os.Getenv("POINTING_BEGIN")); v > 0 {
// 		cube.PointingBegin = v
// 	}
// 	// 设定定制的pointing_end
// 	if v, _ := strconv.Atoi(os.Getenv("POINTING_ENG")); v > 0 {
// 		cube.PointingEnd = v
// 	}

// 	if cube.NumOfSeconds == 0 {
// 		cube.NumOfSeconds = cube.TimeEnd - cube.TimeBegin + 1
// 	}

// 	if p0 > 0 {
// 		cube.PointingBegin = p0
// 		cube.PointingEnd = p1
// 	}

// 	return &cube
// }

// ToCubeString ...
func (cube *DataCube) ToCubeString() string {
	return fmt.Sprintf(`
		cube:
			obs-id=%s
			dataset-id=%s
			ch0=%d,ch1=%d,
			t0=%d,t1=%d, num-of-seconds=%d, tstep=%d, tunit=%d,
			p0=%d,p1=%d, pstep=%d
		`,
		cube.ObsID,
		cube.DatasetID,
		cube.ChannelBegin, cube.ChannelBegin+cube.NumOfChannels-1,
		cube.TimeBegin, cube.TimeEnd, cube.NumOfSeconds, cube.TimeStep, cube.TimeUnit,
		cube.PointingBegin, cube.PointingEnd, cube.PointingStep)
}
