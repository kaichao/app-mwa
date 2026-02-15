package iopath

import (
	"beamform/internal/aggpath"
	"beamform/internal/cache"
	"beamform/internal/picker"
	"fmt"
	"os"
	"strconv"

	"github.com/kaichao/gopkg/errors"
	"gopkg.in/yaml.v2"
)

func IsPreloadMode() bool {
	return os.Getenv("PRELOAD_MODE") != ""
}

func GetOriginRoot() string {
	if v := os.Getenv("ORIGIN_ROOT"); v != "" {
		return v
	}

	if config == nil {
		if err := loadIOPathConfig(); err != nil {
			println("Error loading YAML config:", err.Error())
			return ""
		}
		println("YAML config loaded successfully")
	}

	m := map[string]float64{}
	for _, p := range config.Origin.WeightedPaths {
		m[p.Path] = p.Weight
	}
	// Origin没有AGG_PATH
	delete(m, "AGG_PATH")
	return picker.NewWeightedPicker(m).GetNext()
}

// 波束合成的输入路径，包括打包文件(*.tar)、解包后文件（*。dat）
// func GetPreloadRoot0(index int) string {
// 	if v := os.Getenv("PRELOAD_ROOT"); v != "" {
// 		return v
// 	}

// 	if config == nil {
// 		if err := loadIOPathConfig(); err != nil {
// 			println("Error loading YAML config:", err.Error())
// 			return ""
// 		}
// 		println("YAML config loaded successfully")
// 	}
// 	return config.Preload.GetIndexedPath(index)
// }

func GetPreloadRoot(path string) (string, error) {
	if v := os.Getenv("PRELOAD_ROOT"); v != "" {
		return v, nil
	}

	// 给preload独立配置，避免加载路径错位
	os.Setenv("AGGPATH_CONF", "/preload.ap-conf")
	defer os.Unsetenv("AGGPATH_CONF")
	if config == nil {
		if err := loadIOPathConfig(); err != nil {
			return "", errors.WrapE(err, "loadIOPathConfig()")
		}
	}

	p, err := config.Preload.GetWeightedPath(path)
	return p, errors.WrapE(err, "GetWeightedPath()", "path", path)
}

// 24ch文件
// func GetStagingRoot0(index int) string {
// 	if v := os.Getenv("STAGING_ROOT"); v != "" {
// 		return v
// 	}

// 	if config == nil {
// 		if err := loadIOPathConfig(); err != nil {
// 			println("Error loading YAML config:", err.Error())
// 			return ""
// 		}
// 		println("YAML config loaded successfully")
// 	}
// 	return config.Staging.GetIndexedPath(index)
// }

func GetStagingRoot(path string) (string, error) {
	if v := os.Getenv("STAGING_ROOT"); v != "" {
		return v, nil
	}

	if config == nil {
		if err := loadIOPathConfig(); err != nil {
			return "", errors.WrapE(err, "loadIOPathConfig()")
		}
	}

	p, err := config.Staging.GetWeightedPath(path)
	return p, errors.WrapE(err, "GetWeightedPath()", "path", path)
}

// PathWeight 表示带权重的路径
type PathWeight struct {
	Path       string  `yaml:"path"`
	Weight     float64 `yaml:"weight"`
	CapacityGB int     `yaml:"capacity_gb,omitempty"`
}

// IndexRange 表示索引范围
// type IndexRange struct {
// 	StartIndex int `yaml:"start_index"`
// 	EndIndex   int `yaml:"end_index"`
// 	CapacityGB int `yaml:"capacity_gb"`
// }

// IOPathConfig 表示IO路径配置
type IOPathConfig struct {
	Name          string       `yaml:"name"`
	WeightedPaths []PathWeight `yaml:"weighted_paths"`
	// CombinedPath  []IndexRange `yaml:"combined_path"`

	indexes      []int
	weightPicker *picker.WeightedPicker
	currentIndex int
}

func (iopc *IOPathConfig) setup() {
	// for _, v := range iopc.CombinedPath {
	// 	for i := v.StartIndex; i <= v.EndIndex; i++ {
	// 		iopc.indexes = append(iopc.indexes, i)
	// 	}
	// }
	m := map[string]float64{}
	for _, p := range iopc.WeightedPaths {
		m[p.Path] = p.Weight
	}
	iopc.weightPicker = picker.NewWeightedPicker(m)
}

// func (iopc *IOPathConfig) GetIndexedPath(index int) string {
// 	path := iopc.weightPicker.GetNext()
// 	if path != "AGG_PATH" {
// 		return path
// 	}

// 	i := index % len(iopc.indexes)
// 	if index < 0 {
// 		iopc.currentIndex++
// 		i = iopc.currentIndex % len(iopc.indexes)
// 	}
// 	return fmt.Sprintf("/public/home/cstu0%03d/scalebox/mydata",
// 		iopc.indexes[i])
// }

func (iopc *IOPathConfig) GetWeightedPath(path string) (string, error) {
	p := iopc.weightPicker.GetNext()
	if p != "AGG_PATH" {
		return p, nil
	}

	fmt.Fprintf(os.Stderr, "name:%s,path:%s\n", iopc.Name, path)

	moduleID, _ := strconv.Atoi(os.Getenv("MODULE_ID"))
	appID := cache.GetAppIDByModuleID(moduleID)
	confFile := os.Getenv("AGGPATH_CONF")
	if confFile == "" {
		confFile = "/mwa.ap-conf"
	}
	ap, err := aggpath.New(appID, confFile)
	if err != nil {
		return "", errors.WrapE(err, "aggpath.New()")
	}
	p, err = ap.GetMemberPath(iopc.Name, path)
	if err != nil {
		return "", errors.WrapE(err, "aggpath.GetMemberPath()", "category", iopc.Name, "path", path)
	}

	return p, nil
}

// IOPath 表示整个YAML文件的结构
type IOPath struct {
	// 数据的全局输入
	Origin IOPathConfig `yaml:"origin"`
	// 计算的全局输入
	Preload IOPathConfig `yaml:"preload"`
	// 计算的本地输入

	// 计算的本地输出

	// 计算的全局输出
	Staging IOPathConfig `yaml:"staging"`
	// 数据的全局输出
	Final IOPathConfig `yaml:"final"`
}

var config *IOPath

// loadIOPathConfig 从YAML文件加载IO路径配置
func loadIOPathConfig() error {
	iopathFile := os.Getenv("IOPATH_FILE")
	if iopathFile == "" {
		iopathFile = "/io-path.yaml"
	}

	// 尝试在当前目录查找
	data, err := os.ReadFile(iopathFile)
	if err != nil {
		return errors.WrapE(err, "os.ReadFile()", "file-name", iopathFile)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return errors.WrapE(err, "yaml.Unmarshal()", "yaml-text", data)
	}

	config.Origin.setup()
	config.Preload.setup()
	config.Staging.setup()
	config.Final.setup()

	return nil
}
