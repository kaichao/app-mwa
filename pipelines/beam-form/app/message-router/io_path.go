package main

import (
	"beamform/internal/picker"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func isPreloadMode() bool {
	return os.Getenv("PRELOAD_MODE") != ""
}
func getOriginRoot() string {
	if v := os.Getenv("ORIGIN_ROOT"); v != "" {
		return v
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
func getPreloadRoot(ch int) string {
	if v := os.Getenv("PRELOAD_ROOT"); v != "" {
		return v
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
func getStagingRoot(pt int) string {
	if v := os.Getenv("STAGING_ROOT"); v != "" {
		return v
	}

	m := map[string]float64{}
	for _, wp := range config.Staging.WeightedPaths {
		m[wp.Path] = wp.Weight
	}

	path := picker.NewWeightedPicker(m).GetNext()
	if path != "COMBINED_PATH" {
		return path
	}

	i := pt % len(config.Staging.indexes)
	path = fmt.Sprintf("/public/home/cstu00%02d/scalebox/mydata",
		config.Staging.indexes[i])

	return path
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
	indexes       []int
}

// IOPath 表示整个YAML文件的结构
type IOPath struct {
	Origin  IOPathConfig `yaml:"origin"`
	Preload IOPathConfig `yaml:"preload"`
	Staging IOPathConfig `yaml:"staging"`
	Final   IOPathConfig `yaml:"final"`
}

var (
	config *IOPath
)

// loadIOPathConfig 从YAML文件加载IO路径配置
func loadIOPathConfig() (*IOPath, error) {
	iopathFile := os.Getenv("IOPATH_FILE")
	if iopathFile == "" {
		iopathFile = "/io-path.yaml"
	}
	// 尝试在当前目录查找
	data, err := os.ReadFile(iopathFile)
	if err != nil {
		return nil, err
	}

	var config IOPath
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
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

	return &config, nil
}

// 测试YAML读取功能
func init() {
	var err error
	config, err = loadIOPathConfig()
	if err != nil {
		println("Error loading YAML config:", err.Error())
		return
	}

	println("YAML config loaded successfully")

	// 打印combined path信息，包括capacity_gb
	println("Preload weighted paths:", len(config.Preload.WeightedPaths))
	for i, path := range config.Preload.CombinedPath {
		println("  ", i, ": start_index=", path.StartIndex, ", end_index=", path.EndIndex, ", capacity_gb=", path.CapacityGB)
	}

	println("Staging weighted paths:", len(config.Staging.WeightedPaths))
	for i, path := range config.Staging.CombinedPath {
		println("  ", i, ": start_index=", path.StartIndex, ", end_index=", path.EndIndex, ", capacity_gb=", path.CapacityGB)
	}
}
