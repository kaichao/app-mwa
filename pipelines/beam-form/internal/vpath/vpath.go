package vpath

import (
	"github.com/kaichao/gopkg/errors"
)

// VirtualPath 虚拟路径管理器
// 提供加权路径选择和聚合目录功能
type VirtualPath struct {
	name        string
	appID       int
	config      *Config            // 主配置（用于向后兼容）
	configs     map[string]*Config // category名称 -> 配置（新设计）
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
// 加载YAML文件中的所有配置，合并它们
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

	// 合并所有配置
	mergedConfig := &Config{
		Name:            "merged-config",
		WeightedPaths:   make([]WeightedPathConfig, 0),
		AggregatedPaths: make([]AggregatedPathConfig, 0),
	}

	// 合并所有WeightedPaths
	for configName, config := range allConfigs {
		// 设置配置名称（用于调试）
		if config.Name == "" {
			config.Name = configName
		}

		// 合并WeightedPaths
		mergedConfig.WeightedPaths = append(mergedConfig.WeightedPaths, config.WeightedPaths...)

		// 合并AggregatedPaths
		mergedConfig.AggregatedPaths = append(mergedConfig.AggregatedPaths, config.AggregatedPaths...)

		// 使用第一个配置的AggregatorType（如果有）
		if config.AggregatorType != "" && mergedConfig.AggregatorType == "" {
			mergedConfig.AggregatorType = config.AggregatorType
		}
	}

	// 调试信息
	// fmt.Printf("DEBUG NewVirtualPath: merged AggregatorType=%s, AggregatedPaths count=%d\n",
	//     mergedConfig.AggregatorType, len(mergedConfig.AggregatedPaths))
	// for i, ap := range mergedConfig.AggregatedPaths {
	//     fmt.Printf("  [%d] Name=%s, Members=%v\n", i, ap.Name, ap.Members)
	// }

	// 创建VirtualPath
	return newFromConfig(appID, mergedConfig)
}

// NewVirtualPathForCategory 创建指定category的VirtualPath实例
func NewVirtualPathForCategory(appID int, configFile, category string) (*VirtualPath, error) {
	// 加载指定category的配置
	config, err := loadConfigFromYAML(configFile, category)
	if err != nil {
		return nil, errors.WrapE(err, "load config from YAML")
	}

	// 创建VirtualPath
	return newFromConfig(appID, config)
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

	vp := &VirtualPath{
		name:        config.Name,
		appID:       appID,
		config:      config,
		selectors:   make(map[string]Selector),
		aggregators: make(map[string]Aggregator),
	}

	// 初始化选择器（按category分组）
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
	// 按category分组WeightedPaths
	categoryMap := make(map[string][]WeightedPathConfig)

	for _, wp := range vp.config.WeightedPaths {
		// 使用wp.Category作为分组key
		// 注意：wp.Category在config.go中已经正确设置
		category := wp.Category
		if category == "" {
			// 如果Category为空，使用默认
			category = "default"
		}
		categoryMap[category] = append(categoryMap[category], wp)
	}

	// 为每个category创建选择器
	for category, configs := range categoryMap {
		selector := NewWeightedSelector(category, configs)
		vp.selectors[category] = selector
	}

	return nil
}

// initAggregators 初始化聚合器
func (vp *VirtualPath) initAggregators() error {
	// 收集所有需要聚合器的category
	categories := make(map[string]bool)
	for _, wp := range vp.config.WeightedPaths {
		if wp.Type == "aggregated" && wp.Category != "" {
			categories[wp.Category] = true
		}
	}

	// 为每个category创建聚合器
	for category := range categories {
		if _, exists := vp.aggregators[category]; exists {
			continue // 已经存在（可能从AggregatedPaths创建的）
		}

		var aggregator Aggregator
		var err error

		// 检查是否有对应的AggregatedPathConfig
		var apConfig *AggregatedPathConfig
		for i, ap := range vp.config.AggregatedPaths {
			if ap.Name == category {
				apConfig = &vp.config.AggregatedPaths[i]
				break
			}
		}

		// 确定聚合器类型
		aggregatorType := vp.config.AggregatorType
		if aggregatorType == "" {
			// 默认使用scalebox
			aggregatorType = "scalebox"
		}

		// 如果有AggregatedPathConfig，但aggregatorType不是memory，则使用memory（向后兼容）
		// 注意：这主要是为了测试，生产环境应该使用scalebox
		if apConfig != nil && aggregatorType != "memory" {
			// 如果有AggregatedPathConfig但未指定memory类型，使用memory（测试场景）
			aggregatorType = "memory"
		}

		// 调试信息
		// fmt.Printf("DEBUG initAggregators: category=%s, hasAggregatedPathConfig=%v, aggregatorType=%s, AggregatedPaths count=%d\n",
		//     category, apConfig != nil, aggregatorType, len(vp.config.AggregatedPaths))

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

	// 向后兼容：仍然为AggregatedPaths创建聚合器（如果还没创建）
	for _, apConfig := range vp.config.AggregatedPaths {
		if _, exists := vp.aggregators[apConfig.Name]; !exists {
			aggregator := NewMemoryAggregator(apConfig)
			vp.aggregators[apConfig.Name] = aggregator
		}
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
	for _, wp := range vp.config.WeightedPaths {
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
	// 查找category对应的配置，确定是否是aggregated类型
	for _, wp := range vp.config.WeightedPaths {
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

// getConfig 获取配置（用于测试，包内可见）
func (vp *VirtualPath) getConfig() *Config {
	return vp.config
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
