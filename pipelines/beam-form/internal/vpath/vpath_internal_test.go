package vpath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewVirtualPathFromConfigBasic 测试基本功能
func TestNewVirtualPathFromConfigBasic(t *testing.T) {
	config := &Config{
		Name: "test-basic",
		WeightedPaths: []WeightedPathConfig{
			{Path: "/path1", Weight: 1.0, Type: "static", Category: "default"},
		},
	}

	vp, err := NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)
	assert.NotNil(t, vp)

	// 现在selector的key是配置名称，而不是wp.Category
	path, err := vp.GetPath("test-basic", "test-key")
	assert.NoError(t, err)
	assert.Equal(t, "/path1", path)
}

// TestNewVirtualPathFromConfigWeightedSelection 测试加权选择
func TestNewVirtualPathFromConfigWeightedSelection(t *testing.T) {
	config := &Config{
		Name: "test-weighted",
		WeightedPaths: []WeightedPathConfig{
			{Path: "/pathA", Weight: 0.7, Type: "static", Category: "default"},
			{Path: "/pathB", Weight: 0.3, Type: "static", Category: "default"},
		},
	}

	vp, err := NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 多次选择，验证大致比例
	countA, countB := 0, 0
	for i := 0; i < 100; i++ {
		path, err := vp.GetPath("test-weighted", "test")
		assert.NoError(t, err)
		if path == "/pathA" {
			countA++
		} else if path == "/pathB" {
			countB++
		}
	}

	// 验证大致比例（允许误差）
	assert.Greater(t, countA, 50) // 70%的应该多于50
	assert.Less(t, countB, 50)    // 30%的应该少于50
}

// TestNewVirtualPathFromConfigWithMemoryAggregator 测试内存聚合器
func TestNewVirtualPathFromConfigWithMemoryAggregator(t *testing.T) {
	config := &Config{
		Name: "test-memory-aggregator",
		WeightedPaths: []WeightedPathConfig{
			{
				Path:       "AGG_PATH",
				Weight:     1.0,
				Type:       "aggregated",
				Category:   "storage",
				CapacityGB: 10,
			},
		},
		AggregatedPaths: []AggregatedPathConfig{
			{
				Name:       "storage",
				CapacityGB: 100,
				Members:    []string{"/node1", "/node2"},
			},
		},
		AggregatorType: "memory",
	}

	vp, err := NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 第一次分配
	path1, err := vp.GetPath("test-memory-aggregator", "job1")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2"}, path1)

	// 相同key返回相同路径
	path2, err := vp.GetPath("test-memory-aggregator", "job1")
	assert.NoError(t, err)
	assert.Equal(t, path1, path2)

	// 不同key可能不同
	path3, err := vp.GetPath("test-memory-aggregator", "job2")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2"}, path3)
}

// TestNewVirtualPathFromConfigMixedPaths 测试混合路径
func TestNewVirtualPathFromConfigMixedPaths(t *testing.T) {
	config := &Config{
		Name: "test-mixed-paths",
		WeightedPaths: []WeightedPathConfig{
			{Path: "/fast/ssd", Weight: 1.0, Type: "static", Category: "default"},
			{Path: "/slow/hdd", Weight: 2.0, Type: "static", Category: "default"},
			{
				Path:       "AGG_PATH",
				Weight:     3.0,
				Type:       "aggregated",
				Category:   "default",
				CapacityGB: 20,
			},
		},
		AggregatedPaths: []AggregatedPathConfig{
			{
				Name:       "default",
				CapacityGB: 1000,
				Members:    []string{"/dir0", "/dir1", "/dir2"},
			},
		},
		AggregatorType: "memory",
	}

	vp, err := NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 多次获取，验证路径是有效的
	validPaths := []string{"/fast/ssd", "/slow/hdd", "/dir0", "/dir1", "/dir2"}
	for i := 0; i < 20; i++ {
		path, err := vp.GetPath("test-mixed-paths", fmt.Sprintf("key-%d", i))
		assert.NoError(t, err)
		assert.Contains(t, validPaths, path)
	}
}

// TestReleasePath 测试释放路径
func TestReleasePath(t *testing.T) {
	config := &Config{
		Name: "test-release",
		WeightedPaths: []WeightedPathConfig{
			{
				Path:       "AGG_PATH",
				Weight:     1.0,
				Type:       "aggregated",
				Category:   "storage",
				CapacityGB: 10,
			},
		},
		AggregatedPaths: []AggregatedPathConfig{
			{
				Name:       "storage",
				CapacityGB: 30,
				Members:    []string{"/node1", "/node2", "/node3"},
			},
		},
		AggregatorType: "memory",
	}

	vp, err := NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 分配路径
	path, err := vp.GetPath("test-release", "job1")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2", "/node3"}, path)

	// 释放路径
	err = vp.ReleasePath("test-release", "job1")
	assert.NoError(t, err)

	// 可以重新分配
	newPath, err := vp.GetPath("test-release", "job1")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2", "/node3"}, newPath)
}

// TestNewSelectorAlgorithm 测试新的选择器算法公平性
func TestNewSelectorAlgorithm(t *testing.T) {
	// 测试用例1：简单权重 {0.7, 0.3}
	t.Run("SimpleWeights_7_3", func(t *testing.T) {
		config := &Config{
			Name: "test-simple-weights",
			WeightedPaths: []WeightedPathConfig{
				{Path: "/pathA", Weight: 0.7, Type: "static", Category: "default"},
				{Path: "/pathB", Weight: 0.3, Type: "static", Category: "default"},
			},
		}

		vp, err := NewVirtualPathFromConfig(1, config)
		assert.NoError(t, err)

		// 多次选择，统计结果
		countA, countB := 0, 0
		for i := 0; i < 1000; i++ {
			path, err := vp.GetPath("test-simple-weights", fmt.Sprintf("key-%d", i))
			assert.NoError(t, err)
			if path == "/pathA" {
				countA++
			} else if path == "/pathB" {
				countB++
			}
		}

		total := countA + countB
		ratioA := float64(countA) / float64(total)
		ratioB := float64(countB) / float64(total)

		// 验证实际占比接近理论权重（允许5%误差）
		assert.InDelta(t, 0.7, ratioA, 0.05, "PathA ratio should be close to 0.7")
		assert.InDelta(t, 0.3, ratioB, 0.05, "PathB ratio should be close to 0.3")

		// 验证算法确定性：重新运行应该得到相同结果
		vp2, _ := NewVirtualPathFromConfig(1, config)
		firstPath, _ := vp2.GetPath("test-simple-weights", "test-key")
		// 第一次选择应该是权重最大的路径（/pathA，权重0.7）
		assert.Equal(t, "/pathA", firstPath)
	})

	// 测试用例2：相等权重 {1.0, 1.0, 1.0}
	t.Run("EqualWeights", func(t *testing.T) {
		config := &Config{
			Name: "test-equal-weights",
			WeightedPaths: []WeightedPathConfig{
				{Path: "/path1", Weight: 1.0, Type: "static", Category: "default"},
				{Path: "/path2", Weight: 1.0, Type: "static", Category: "default"},
				{Path: "/path3", Weight: 1.0, Type: "static", Category: "default"},
			},
		}

		vp, err := NewVirtualPathFromConfig(1, config)
		assert.NoError(t, err)

		// 多次选择，统计结果
		counts := make(map[string]int)
		for i := 0; i < 999; i++ { // 使用999确保能被3整除的测试
			path, err := vp.GetPath("test-equal-weights", fmt.Sprintf("key-%d", i))
			assert.NoError(t, err)
			counts[path]++
		}

		// 验证每个路径被选择大约333次（允许10次误差）
		for _, path := range []string{"/path1", "/path2", "/path3"} {
			assert.InDelta(t, 333, counts[path], 10, fmt.Sprintf("Path %s should be selected about 333 times", path))
		}
	})

	// 测试用例3：复杂权重 {0.1, 0.2, 0.3, 0.4}
	t.Run("ComplexWeights", func(t *testing.T) {
		config := &Config{
			Name: "test-complex-weights",
			WeightedPaths: []WeightedPathConfig{
				{Path: "/pathA", Weight: 0.1, Type: "static", Category: "default"},
				{Path: "/pathB", Weight: 0.2, Type: "static", Category: "default"},
				{Path: "/pathC", Weight: 0.3, Type: "static", Category: "default"},
				{Path: "/pathD", Weight: 0.4, Type: "static", Category: "default"},
			},
		}

		vp, err := NewVirtualPathFromConfig(1, config)
		assert.NoError(t, err)

		// 多次选择，统计结果
		counts := make(map[string]int)
		for i := 0; i < 1000; i++ {
			path, err := vp.GetPath("test-complex-weights", fmt.Sprintf("key-%d", i))
			assert.NoError(t, err)
			counts[path]++
		}

		// 验证实际占比接近理论权重（允许5%误差）
		total := 1000.0
		assert.InDelta(t, 0.1, float64(counts["/pathA"])/total, 0.05, "PathA ratio")
		assert.InDelta(t, 0.2, float64(counts["/pathB"])/total, 0.05, "PathB ratio")
		assert.InDelta(t, 0.3, float64(counts["/pathC"])/total, 0.05, "PathC ratio")
		assert.InDelta(t, 0.4, float64(counts["/pathD"])/total, 0.05, "PathD ratio")
	})

	// 测试用例4：权重为0的项被排除
	t.Run("ZeroWeightExcluded", func(t *testing.T) {
		config := &Config{
			Name: "test-zero-weight",
			WeightedPaths: []WeightedPathConfig{
				{Path: "/pathA", Weight: 1.0, Type: "static", Category: "default"},
				{Path: "/pathB", Weight: 0.0, Type: "static", Category: "default"}, // 权重为0，应该被排除
				{Path: "/pathC", Weight: 2.0, Type: "static", Category: "default"},
			},
		}

		vp, err := NewVirtualPathFromConfig(1, config)
		assert.NoError(t, err)

		// 多次选择，验证权重为0的路径永远不会被选择
		for i := 0; i < 100; i++ {
			path, err := vp.GetPath("test-zero-weight", fmt.Sprintf("key-%d", i))
			assert.NoError(t, err)
			assert.NotEqual(t, "/pathB", path, "Zero weight path should never be selected")
		}
	})
}
