package vpath

// Selector 路径选择器接口
type Selector interface {
	// Select 选择一条路径
	Select() string

	// Name 返回选择器名称
	Name() string
}

// WeightedSelector 加权选择器
type WeightedSelector struct {
	name    string
	paths   []string
	weights []float64
	counts  []int
	total   int
}

// NewWeightedSelector 创建加权选择器
func NewWeightedSelector(name string, configs []WeightedPathConfig) *WeightedSelector {
	paths := make([]string, 0, len(configs))
	weights := make([]float64, 0, len(configs))

	for _, cfg := range configs {
		paths = append(paths, cfg.Path)
		weights = append(weights, cfg.Weight)
	}

	return &WeightedSelector{
		name:    name,
		paths:   paths,
		weights: weights,
		counts:  make([]int, len(configs)),
		total:   0,
	}
}

// Select 实现加权选择算法
// 基于权重和历史选择次数，尽量使实际选择比例接近理论权重
// 算法规则：
// 1. 第一次选择：选择权重最大的项
// 2. 后续选择：选择 actual_i / theory_i 最小的项
//   - actual_i = count_i / total (实际占比)
//   - theory_i = weight_i / total_weight (理论占比)
//
// 3. 权重<=0的项在创建时已被排除
func (ws *WeightedSelector) Select() string {
	if len(ws.paths) == 0 {
		return ""
	}

	// 第一次选择：选择权重最大的项
	if ws.total == 0 {
		maxWeightIdx := 0
		maxWeight := ws.weights[0]
		for i := 1; i < len(ws.weights); i++ {
			if ws.weights[i] > maxWeight {
				maxWeight = ws.weights[i]
				maxWeightIdx = i
			}
		}
		ws.counts[maxWeightIdx]++
		ws.total++
		return ws.paths[maxWeightIdx]
	}

	// 计算总权重
	totalWeight := ws.totalWeight()

	// 找到 actual_i / theory_i 最小的项
	minRatio := 0.0
	minIdx := 0
	first := true

	for i := range ws.paths {
		// 计算实际占比
		actualRatio := float64(ws.counts[i]) / float64(ws.total)
		// 计算理论占比
		theoryRatio := ws.weights[i] / totalWeight
		// 计算比值 ratio = actual / theory
		// 注意：当theoryRatio为0时不应该发生（权重<=0的项已被排除）
		ratio := actualRatio / theoryRatio

		if first || ratio < minRatio {
			minRatio = ratio
			minIdx = i
			first = false
		}
	}

	// 更新统计并返回选择的路径
	ws.counts[minIdx]++
	ws.total++
	return ws.paths[minIdx]
}

// Name 返回选择器名称
func (ws *WeightedSelector) Name() string {
	return ws.name
}

// totalWeight 计算总权重
func (ws *WeightedSelector) totalWeight() float64 {
	total := 0.0
	for _, w := range ws.weights {
		total += w
	}
	return total
}

// Stats 返回统计信息
func (ws *WeightedSelector) Stats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["total_selections"] = ws.total
	stats["paths"] = ws.paths

	pathStats := make([]map[string]interface{}, len(ws.paths))
	for i, path := range ws.paths {
		pathStats[i] = map[string]interface{}{
			"path":         path,
			"weight":       ws.weights[i],
			"count":        ws.counts[i],
			"actual_ratio": float64(ws.counts[i]) / float64(ws.total),
			"theory_ratio": ws.weights[i] / ws.totalWeight(),
		}
	}
	stats["path_stats"] = pathStats

	return stats
}
