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
func GetPreloadRoot(ch int) string {
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

	m := map[string]float64{}
	for _, p := range config.Preload.WeightedPaths {
		m[p.Path] = p.Weight
	}

	path := picker.NewWeightedPicker(m).GetNext()
	if path != "COMBINED_PATH" {
		return path
	}

	i := ch % len(config.Preload.indexes)
	path = fmt.Sprintf("/public/home/cstu00%02d/scalebox/mydata",
		config.Preload.indexes[i])

	return path
}

// 24ch文件
func GetStagingRoot(pt int) string {
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

	if wpStaging == nil {
		m := map[string]float64{}
		for _, wp := range config.Staging.WeightedPaths {
			m[wp.Path] = wp.Weight
		}
		wpStaging = picker.NewWeightedPicker(m)
	}

	path := wpStaging.GetNext()
	if path != "COMBINED_PATH" {
		return path
	}

	// i := pt % len(config.Staging.indexes)
	i := index % len(config.Staging.indexes)
	index++
	path = fmt.Sprintf("/public/home/cstu00%02d/scalebox/mydata",
		config.Staging.indexes[i])

	return path
}

var index int = 0

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
	indexes       []int
	weightPicker  *picker.WeightedPicker
}

// IOPath 表示整个YAML文件的结构
type IOPath struct {
	Origin  IOPathConfig `yaml:"origin"`
	Preload IOPathConfig `yaml:"preload"`
	Staging IOPathConfig `yaml:"staging"`
	Final   IOPathConfig `yaml:"final"`
}

func (ipc *IOPathConfig) GetRoot(n int) {

}

var (
	config    *IOPath
	wpStaging *picker.WeightedPicker
)

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

	for _, v := range config.Origin.CombinedPath {
		for i := v.StartIndex; i <= v.EndIndex; i++ {
			config.Origin.indexes = append(config.Origin.indexes, i)
		}
	}
	for _, v := range config.Preload.CombinedPath {
		for i := v.StartIndex; i <= v.EndIndex; i++ {
			config.Preload.indexes = append(config.Preload.indexes, i)
		}
	}
	for _, v := range config.Staging.CombinedPath {
		for i := v.StartIndex; i <= v.EndIndex; i++ {
			config.Staging.indexes = append(config.Staging.indexes, i)
		}
	}
	for _, v := range config.Final.CombinedPath {
		for i := v.StartIndex; i <= v.EndIndex; i++ {
			config.Final.indexes = append(config.Final.indexes, i)
		}
	}

	// 打印combined path信息，包括capacity_gb
	// println("Preload weighted paths:", len(config.Preload.WeightedPaths))
	for i, path := range config.Preload.CombinedPath {
		println("  ", i, ": start_index=", path.StartIndex, ", end_index=", path.EndIndex, ", capacity_gb=", path.CapacityGB)
	}

	// println("Staging weighted paths:", len(config.Staging.WeightedPaths))
	for i, path := range config.Staging.CombinedPath {
		println("  ", i, ": start_index=", path.StartIndex, ", end_index=", path.EndIndex, ", capacity_gb=", path.CapacityGB)
	}

	return nil
}
