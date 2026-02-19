package vpath

import (
	"github.com/kaichao/gopkg/errors"
)

// VirtualPath 虚拟路径管理器
// 提供加权路径选择和聚合目录功能
type VirtualPath struct {
	name        string
	appID       int
	configs     map[string]*Config // category名称 -> 配置
	selectors   map[string]Selector
	aggregators map[string]Aggregator
}

// Config 虚拟路径配置
type Config struct {
	// Name 配置名称
	Name string `yaml:"-"` // 不从YAML读取，由代码设置

	// WeightedPaths 加权路径配置
	WeightedPaths []WeightedPathConfig `yaml:"weighted_paths"`

	// AggregatedPaths 聚合目录配置
	// 注意：在新的设计中，聚合目录由信号量确定，不通过配置文件
	// 保留此字段用于向后兼容和测试
	AggregatedPaths []AggregatedPathConfig `yaml:"-"`

	// AggregatorType 聚合器类型
	// "memory": 使用MemoryAggregator（测试用）
	// "scalebox": 使用ScaleboxAggregator（生产用）
	// 如果未指定，默认根据AggregatedPaths是否为空决定
	AggregatorType string `yaml:"aggregator_type,omitempty"`
}

// WeightedPathConfig 加权路径配置
type WeightedPathConfig struct {
	// Path 路径或路径标识符
	// 可以是具体路径，也可以是特殊标识符如"AGG_PATH"
	Path string `yaml:"path"`

	// Weight 权重值，大于等于0
	// 权重为0的路径永远不会被选择，但允许配置
	// 所有路径的权重之和必须大于0
	Weight float64 `yaml:"weight"`

	// Type 路径类型
	// "static": 静态路径，直接返回Path
	// "aggregated": 聚合路径，需要从聚合目录分配
	// 注意：根据path是否等于"AGG_PATH"自动推断
	Type string `yaml:"-"` // 不从YAML读取，由代码推断

	// Category 聚合目录分类（仅对AGG_PATH有效）
	// 对应YAML中的category字段
	Category string `yaml:"category,omitempty"`

	// CapacityGB 容量需求（GB）
	// 对于static路径：表示路径容量
	// 对于aggregated路径：表示每次分配需要的容量（对应YAML中的need_gb）
	CapacityGB int `yaml:"capacity_gb,omitempty"`

	// NeedGB 仅对AGG_PATH有效，从YAML的need_gb字段读取
	// 临时字段，用于解析
	NeedGB int `yaml:"need_gb,omitempty"`
}

// AggregatedPathConfig 聚合目录配置
type AggregatedPathConfig struct {
	// Name 聚合目录名称
	Name string `yaml:"name"`

	// CapacityGB 总容量（GB）
	CapacityGB int `yaml:"capacity_gb"`

	// Members 成员路径列表
	Members []string `yaml:"members"`
}

// NewVirtualPath 创建VirtualPath实例（主入口）
// appID: 应用ID，用于区分不同应用
// configFile: 配置文件路径（YAML格式）
// 加载YAML文件中的所有配置，不合并它们，保留原始配置结构
func NewVirtualPath(appID int, configFile string) (*VirtualPath, error) {
	// 加载所有配置
	allConfigs, err := loadAllConfigsFromYAML(configFile)
	if err != nil {
		return nil, errors.WrapE(err, "load all configs from YAML")
	}

	// 检查配置数量
	if len(allConfigs) == 0 {
		return nil, errors.E("no configuration found in YAML file")
	}

	// 创建VirtualPath，不合并配置
	vp := &VirtualPath{
		name:        "multi-config",
		appID:       appID,
		configs:     allConfigs,
		selectors:   make(map[string]Selector),
		aggregators: make(map[string]Aggregator),
	}

	// 初始化选择器（按配置名称分组）
	if err := vp.initSelectors(); err != nil {
		return nil, errors.WrapE(err, "init selectors")
	}

	// 初始化聚合器
	if err := vp.initAggregators(); err != nil {
		return nil, errors.WrapE(err, "init aggregators")
	}

	return vp, nil
}

// NewVirtualPathFromConfig 从Config创建VirtualPath实例（用于测试）
func NewVirtualPathFromConfig(appID int, config *Config) (*VirtualPath, error) {
	return newFromConfig(appID, config)
}

// newFromConfig 从Config创建VirtualPath实例（内部使用）
func newFromConfig(appID int, config *Config) (*VirtualPath, error) {
	if config == nil {
		return nil, errors.E("config is nil")
	}

	if config.Name == "" {
		return nil, errors.E("config.Name is empty")
	}

	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, errors.WrapE(err, "validate config")
	}

	// 创建configs map，包含单个配置
	configs := make(map[string]*Config)
	configs[config.Name] = config

	vp := &VirtualPath{
		name:        config.Name,
		appID:       appID,
		configs:     configs,
		selectors:   make(map[string]Selector),
		aggregators: make(map[string]Aggregator),
	}

	// 初始化选择器（按配置名称分组）
	if err := vp.initSelectors(); err != nil {
		return nil, errors.WrapE(err, "init selectors")
	}

	// 初始化聚合器
	if err := vp.initAggregators(); err != nil {
		return nil, errors.WrapE(err, "init aggregators")
	}

	return vp, nil
}

// initSelectors 初始化选择器
func (vp *VirtualPath) initSelectors() error {
	// 使用configs（按配置名称分组）
	for configName, config := range vp.configs {
		// 为每个配置创建选择器，key是配置名称
		selector := NewWeightedSelector(configName, config.WeightedPaths)
		vp.selectors[configName] = selector
	}

	return nil
}

// initAggregators 初始化聚合器
func (vp *VirtualPath) initAggregators() error {
	// 收集所有需要聚合器的category
	categories := make(map[string]bool)

	// 收集所有配置中的aggregated路径
	for _, config := range vp.configs {
		for _, wp := range config.WeightedPaths {
			if wp.Type == "aggregated" && wp.Category != "" {
				categories[wp.Category] = true
			}
		}
	}

	// 为每个category创建聚合器
	for category := range categories {
		if _, exists := vp.aggregators[category]; exists {
			continue // 已经存在
		}

		var aggregator Aggregator
		var err error

		// 检查是否有对应的AggregatedPathConfig
		var apConfig *AggregatedPathConfig
		// 从所有configs中查找
		for _, config := range vp.configs {
			for i, ap := range config.AggregatedPaths {
				if ap.Name == category {
					apConfig = &config.AggregatedPaths[i]
					break
				}
			}
			if apConfig != nil {
				break
			}
		}

		// 确定聚合器类型
		aggregatorType := ""
		// 从第一个配置中获取AggregatorType
		for _, config := range vp.configs {
			if config.AggregatorType != "" {
				aggregatorType = config.AggregatorType
				break
			}
		}
		if aggregatorType == "" {
			// 默认使用scalebox
			aggregatorType = "scalebox"
		}

		// 如果有AggregatedPathConfig，但aggregatorType不是memory，则使用memory（向后兼容）
		if apConfig != nil && aggregatorType != "memory" {
			aggregatorType = "memory"
		}

		switch aggregatorType {
		case "memory":
			if apConfig == nil {
				// 如果没有配置，创建一个默认的（仅用于测试）
				apConfig = &AggregatedPathConfig{
					Name:       category,
					CapacityGB: 1000, // 默认容量
					Members:    []string{"/default/path1", "/default/path2"},
				}
			}
			aggregator = NewMemoryAggregator(*apConfig)

		case "scalebox":
			aggregator, err = NewScaleboxAggregator(category, vp.appID)
			if err != nil {
				return errors.WrapE(err, "create ScaleboxAggregator", "category", category)
			}

		default:
			return errors.E("unknown aggregator type", "type", aggregatorType)
		}

		vp.aggregators[category] = aggregator
	}

	return nil
}

// GetPath 获取路径
// category: 路径分类，对应WeightedPaths中的配置
// key: 唯一标识符，用于聚合路径的分配
func (vp *VirtualPath) GetPath(category, key string) (string, error) {
	selector, ok := vp.selectors[category]
	if !ok {
		return "", errors.E("category not found", "category", category)
	}

	// 选择路径
	selectedPath := selector.Select()

	// 查找对应的配置
	var wpConfig *WeightedPathConfig
	// 使用configs，需要找到category对应的配置
	config, ok := vp.configs[category]
	if !ok {
		return "", errors.E("config not found for category", "category", category)
	}
	for _, wp := range config.WeightedPaths {
		if wp.Path == selectedPath {
			wpConfig = &wp
			break
		}
	}

	if wpConfig == nil {
		return "", errors.E("selected path config not found", "path", selectedPath)
	}

	// 根据类型处理
	switch wpConfig.Type {
	case "static":
		return wpConfig.Path, nil
	case "aggregated":
		// 从聚合器分配路径
		aggregator, ok := vp.aggregators[wpConfig.Category]
		if !ok {
			return "", errors.E("aggregator not found", "category", wpConfig.Category)
		}
		return aggregator.Allocate(key, wpConfig.CapacityGB)
	default:
		return "", errors.E("unknown path type", "type", wpConfig.Type)
	}
}

// ReleasePath 释放路径
// category: 路径分类
// key: 唯一标识符
func (vp *VirtualPath) ReleasePath(category, key string) error {
	// 使用configs，需要找到category对应的配置
	config, ok := vp.configs[category]
	if !ok {
		return errors.E("config not found for category", "category", category)
	}

	// 查找aggregated类型的路径
	for _, wp := range config.WeightedPaths {
		if wp.Type == "aggregated" {
			// 释放聚合路径
			aggregator, ok := vp.aggregators[wp.Category]
			if !ok {
				return errors.E("aggregator not found", "category", wp.Category)
			}
			return aggregator.Release(key, wp.CapacityGB)
		}
	}

	// 如果不是aggregated类型，无需释放
	return nil
}

// validateConfig 验证配置
func validateConfig(config *Config) error {
	// 过滤掉权重<=0的项
	filteredPaths := make([]WeightedPathConfig, 0, len(config.WeightedPaths))
	totalWeight := 0.0

	for _, wp := range config.WeightedPaths {
		if wp.Path == "" {
			return errors.E("WeightedPaths[%d].Path is empty", len(filteredPaths))
		}
		if wp.Weight < 0 {
			return errors.E("WeightedPaths[%d].Weight must be >= 0", len(filteredPaths))
		}

		// 跳过权重<=0的项
		if wp.Weight <= 0 {
			continue
		}

		// 推断类型（如果未设置）
		if wp.Type == "" {
			if wp.Path == "AGG_PATH" {
				wp.Type = "aggregated"
			} else {
				wp.Type = "static"
			}
		}

		if wp.Type != "static" && wp.Type != "aggregated" {
			return errors.E("WeightedPaths[%d].Type must be 'static' or 'aggregated'", len(filteredPaths))
		}

		if wp.Type == "aggregated" {
			if wp.Category == "" {
				return errors.E("WeightedPaths[%d].Category is empty for aggregated type", len(filteredPaths))
			}
			if wp.CapacityGB <= 0 {
				return errors.E("WeightedPaths[%d].CapacityGB must be > 0 for aggregated type", len(filteredPaths))
			}
		}

		totalWeight += wp.Weight
		filteredPaths = append(filteredPaths, wp)
	}

	// 更新配置中的WeightedPaths
	config.WeightedPaths = filteredPaths

	// 检查总权重是否大于0
	if totalWeight <= 0 {
		return errors.E("total weight of all WeightedPaths must be > 0")
	}

	// 注意：在新的设计中，AggregatedPaths不由配置文件确定
	// 由ScaleboxAggregator通过信号量管理
	// 所以这里不验证AggregatedPaths

	return nil
}
