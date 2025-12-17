package iopath

import (
	"beamform/internal/picker"
	"fmt"
	"os"

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
	// Origin没有COMBINED_PATH
	delete(m, "COMBINED_PATH")
	return picker.NewWeightedPicker(m).GetNext()
}

// 波束合成的输入路径，包括打包文件(*.tar)、解包后文件（*。dat）
func GetPreloadRoot(index int) string {
	if v := os.Getenv("PRELOAD_ROOT"); v != "" {
		return v
	}

	if config == nil {
		if err := loadIOPathConfig(); err != nil {
			println("Error loading YAML config:", err.Error())
			return ""
		}
		println("YAML config loaded successfully")
	}
	return config.Preload.GetIndexedPath(index)
}

// 24ch文件
func GetStagingRoot(index int) string {
	if v := os.Getenv("STAGING_ROOT"); v != "" {
		return v
	}

	if config == nil {
		if err := loadIOPathConfig(); err != nil {
			println("Error loading YAML config:", err.Error())
			return ""
		}
		println("YAML config loaded successfully")
	}
	return config.Staging.GetIndexedPath(index)
}

// PathWeight 表示带权重的路径
type PathWeight struct {
	Path       string  `yaml:"path"`
	Weight     float64 `yaml:"weight"`
	CapacityGB int     `yaml:"capacity_gb,omitempty"`
}

// IndexRange 表示索引范围
type IndexRange struct {
	StartIndex int `yaml:"start_index"`
	EndIndex   int `yaml:"end_index"`
	CapacityGB int `yaml:"capacity_gb"`
}

// IOPathConfig 表示IO路径配置
type IOPathConfig struct {
	WeightedPaths []PathWeight `yaml:"weighted_paths"`
	CombinedPath  []IndexRange `yaml:"combined_path"`

	indexes      []int
	weightPicker *picker.WeightedPicker
	currentIndex int
}

func (iopc *IOPathConfig) setup() {
	for _, v := range iopc.CombinedPath {
		for i := v.StartIndex; i <= v.EndIndex; i++ {
			iopc.indexes = append(iopc.indexes, i)
		}
	}
	m := map[string]float64{}
	for _, p := range iopc.WeightedPaths {
		m[p.Path] = p.Weight
	}
	iopc.weightPicker = picker.NewWeightedPicker(m)
}

func (iopc *IOPathConfig) GetIndexedPath(index int) string {
	path := iopc.weightPicker.GetNext()
	if path != "COMBINED_PATH" {
		return path
	}

	i := index % len(iopc.indexes)
	if index < 0 {
		iopc.currentIndex++
		i = iopc.currentIndex % len(iopc.indexes)
	}
	return fmt.Sprintf("/public/home/cstu0%03d/scalebox/mydata",
		iopc.indexes[i])
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

	fmt.Println("io-path file:", iopathFile)
	// 尝试在当前目录查找
	data, err := os.ReadFile(iopathFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	config.Origin.setup()
	config.Preload.setup()
	config.Staging.setup()
	config.Final.setup()

	return nil
}
