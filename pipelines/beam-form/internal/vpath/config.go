package vpath

import (
	"os"
	"path/filepath"

	"github.com/kaichao/gopkg/errors"
	"gopkg.in/yaml.v3"
)

// yamlConfig 从YAML文件解析的配置结构
type yamlConfig map[string]categoryConfig

// categoryConfig 每个category的配置
type categoryConfig struct {
	WeightedPaths   []WeightedPathConfig   `yaml:"weighted_paths"`
	AggregatedPaths []AggregatedPathConfig `yaml:"aggregated_paths,omitempty"`
	AggregatorType  string                 `yaml:"aggregator_type,omitempty"`
}

// LoadConfigFromYAML 从YAML文件加载配置
// filename: YAML配置文件路径
// categoryName: 要加载的category名称（对应YAML中的顶级key）
func loadConfigFromYAML(filename, categoryName string) (*Config, error) {
	// 读取YAML文件
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.WrapE(err, "read YAML file")
	}

	// 解析YAML
	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, errors.WrapE(err, "parse YAML")
	}

	// 获取指定category的配置
	cc, ok := yc[categoryName]
	if !ok {
		return nil, errors.E("category not found in YAML", "category", categoryName)
	}

	// 转换为Config
	config := &Config{
		Name:            categoryName,
		WeightedPaths:   make([]WeightedPathConfig, 0, len(cc.WeightedPaths)),
		AggregatedPaths: cc.AggregatedPaths,
		AggregatorType:  cc.AggregatorType,
	}

	// 处理每个weighted path
	for _, wp := range cc.WeightedPaths {
		// 推断类型
		if wp.Path == "AGG_PATH" {
			wp.Type = "aggregated"
			// 对于AGG_PATH，need_gb覆盖capacity_gb
			if wp.NeedGB > 0 {
				wp.CapacityGB = wp.NeedGB
			}
			// 对于AGG_PATH，如果pool为空，使用配置名称作为pool
			if wp.Pool == "" {
				wp.Pool = categoryName
			}
		} else {
			wp.Type = "static"
			// 对于静态路径，不需要设置pool
		}

		config.WeightedPaths = append(config.WeightedPaths, wp)
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, errors.WrapE(err, "validate config")
	}

	return config, nil
}

// LoadAllConfigsFromYAML 从YAML文件加载所有category的配置
func loadAllConfigsFromYAML(filename string) (map[string]*Config, error) {
	// 读取YAML文件
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.WrapE(err, "read YAML file")
	}

	// 解析YAML
	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, errors.WrapE(err, "parse YAML")
	}

	// 转换为Config map
	configs := make(map[string]*Config)
	for categoryName, cc := range yc {
		config := &Config{
			Name:            categoryName,
			WeightedPaths:   make([]WeightedPathConfig, 0, len(cc.WeightedPaths)),
			AggregatedPaths: cc.AggregatedPaths,
			AggregatorType:  cc.AggregatorType,
		}

		// 处理每个weighted path
		for _, wp := range cc.WeightedPaths {
			// 推断类型
			if wp.Path == "AGG_PATH" {
				wp.Type = "aggregated"
				// 对于AGG_PATH，need_gb覆盖capacity_gb
				if wp.NeedGB > 0 {
					wp.CapacityGB = wp.NeedGB
				}
				// 对于AGG_PATH，如果pool为空，使用配置名称作为pool
				if wp.Pool == "" {
					wp.Pool = categoryName
				}
			} else {
				wp.Type = "static"
				// 对于静态路径，不需要设置pool
			}

			config.WeightedPaths = append(config.WeightedPaths, wp)
		}

		// 验证配置
		if err := validateConfig(config); err != nil {
			return nil, errors.WrapE(err, "validate config", "category", categoryName)
		}

		configs[categoryName] = config
	}

	return configs, nil
}

// saveConfigToYAML 将配置保存到YAML文件（包内可见）
func saveConfigToYAML(filename string, configs map[string]*Config) error {
	yc := make(yamlConfig)

	for name, config := range configs {
		cc := categoryConfig{
			WeightedPaths: make([]WeightedPathConfig, len(config.WeightedPaths)),
		}

		// 复制WeightedPaths，但清除内部字段
		for i, wp := range config.WeightedPaths {
			cc.WeightedPaths[i] = WeightedPathConfig{
				Path:       wp.Path,
				Weight:     wp.Weight,
				Pool:       wp.Pool,
				CapacityGB: wp.CapacityGB,
				NeedGB:     wp.NeedGB,
			}
		}

		yc[name] = cc
	}

	// 转换为YAML
	data, err := yaml.Marshal(yc)
	if err != nil {
		return errors.WrapE(err, "marshal YAML")
	}

	// 写入文件
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return errors.WrapE(err, "write YAML file")
	}

	return nil
}

// getConfigFileName 获取配置文件名（辅助函数，包内可见）
func getConfigFileName(baseDir, name string) string {
	return filepath.Join(baseDir, name+".yaml")
}
