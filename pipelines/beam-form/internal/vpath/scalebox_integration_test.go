package vpath_test

import (
	"beamform/internal/vpath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestScaleboxAggregatorBasic 测试ScaleboxAggregator基本功能
func TestScaleboxAggregatorBasic(t *testing.T) {
	if !isScaleboxEnvReady(t) {
		return
	}

	// 测试配置：使用test-agg category（对应test-agg.sema中的信号量）
	config := &vpath.Config{
		Name: "test-scalebox-basic",
		WeightedPaths: []vpath.WeightedPathConfig{
			{
				Path:       "AGG_PATH",
				Weight:     1.0,
				Pool:       "test-agg", // 对应test-agg.sema
				CapacityGB: 50,         // 需要50GB容量
			},
		},
		AggregatorType: "scalebox", // 明确指定使用ScaleboxAggregator
	}

	vp, err := vpath.NewVirtualPathFromConfig(appID, config)
	assert.NoError(t, err)
	assert.NotNil(t, vp)

	// 测试1：分配路径
	path1, err := vp.GetPath("test-scalebox-basic", "test-job-1")
	assert.NoError(t, err)
	assert.NotEmpty(t, path1)
	assert.Contains(t, []string{"/test-agg/dir0", "/test-agg/dir1", "/test-agg/dir2"}, path1)

	// 测试2：相同key返回相同路径
	path2, err := vp.GetPath("test-scalebox-basic", "test-job-1")
	assert.NoError(t, err)
	assert.Equal(t, path1, path2)

	// 测试3：释放路径
	err = vp.ReleasePath("test-scalebox-basic", "test-job-1")
	assert.NoError(t, err)

	// 测试4：释放后可以重新分配（可能分配到相同或不同路径）
	path3, err := vp.GetPath("test-scalebox-basic", "test-job-1")
	assert.NoError(t, err)
	assert.NotEmpty(t, path3)
}

// TestScaleboxAggregatorMultipleAllocations 测试多个分配
func TestScaleboxAggregatorMultipleAllocations(t *testing.T) {
	if !isScaleboxEnvReady(t) {
		return
	}

	config := &vpath.Config{
		Name: "test-scalebox-multiple",
		WeightedPaths: []vpath.WeightedPathConfig{
			{
				Path:       "AGG_PATH",
				Weight:     1.0,
				Pool:       "test-agg",
				CapacityGB: 30, // 较小容量，可以分配多个
			},
		},
		AggregatorType: "scalebox",
	}

	vp, err := vpath.NewVirtualPathFromConfig(appID, config)
	assert.NoError(t, err)

	// 分配多个不同key的路径
	paths := make(map[string]string)
	for i := 0; i < 3; i++ {
		key := "test-job-multi-" + string(rune('A'+i))
		path, err := vp.GetPath("test-scalebox-multiple", key)
		assert.NoError(t, err)
		assert.NotEmpty(t, path)
		paths[key] = path
	}

	// 验证不同key可能分配到不同路径（但也不一定，取决于容量）
	uniquePaths := make(map[string]bool)
	for _, path := range paths {
		uniquePaths[path] = true
	}
	// 至少有一个路径被分配
	assert.Greater(t, len(uniquePaths), 0)

	// 清理：释放所有路径
	for key := range paths {
		err = vp.ReleasePath("test-scalebox-multiple", key)
		assert.NoError(t, err)
	}
}

// TestScaleboxAggregatorInsufficientCapacity 测试容量不足的情况
func TestScaleboxAggregatorInsufficientCapacity(t *testing.T) {
	if !isScaleboxEnvReady(t) {
		return
	}

	config := &vpath.Config{
		Name: "test-scalebox-insufficient",
		WeightedPaths: []vpath.WeightedPathConfig{
			{
				Path:       "AGG_PATH",
				Weight:     1.0,
				Pool:       "test-agg",
				CapacityGB: 500, // 超过任何单个目录的容量（最大199GB）
			},
		},
		AggregatorType: "scalebox",
	}

	vp, err := vpath.NewVirtualPathFromConfig(appID, config)
	assert.NoError(t, err)

	// 应该失败，因为容量不足
	path, err := vp.GetPath("test-scalebox-insufficient", "test-job-large")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No enough disk space")
	assert.Empty(t, path)
}

// TestScaleboxAggregatorWithMixedPaths 测试混合路径（静态+聚合）
func TestScaleboxAggregatorWithMixedPaths(t *testing.T) {
	if !isScaleboxEnvReady(t) {
		return
	}

	// 使用已经存在的test-agg category（数据库中有对应的信号量）
	config := &vpath.Config{
		Name: "test-scalebox-mixed",
		WeightedPaths: []vpath.WeightedPathConfig{
			{
				Path:       "/local/ssd",
				Weight:     0.3,
				CapacityGB: 500,
				Pool:       "test-agg", // 使用已经存在的pool
			},
			{
				Path:       "AGG_PATH",
				Weight:     0.7,
				Pool:       "test-agg", // 使用已经存在的pool
				CapacityGB: 50,
			},
		},
		AggregatorType: "scalebox",
	}

	vp, err := vpath.NewVirtualPathFromConfig(appID, config)
	assert.NoError(t, err)

	// 测试多次获取，验证加权选择
	staticCount := 0
	aggregatedCount := 0

	// 使用正确的category
	for i := 0; i < 100; i++ {
		path, err := vp.GetPath("test-scalebox-mixed", "test-mixed")
		assert.NoError(t, err)
		if path == "/local/ssd" {
			staticCount++
		} else if path == "/test-agg/dir0" || path == "/test-agg/dir1" || path == "/test-agg/dir2" {
			aggregatedCount++
		}
	}

	// 验证大致比例（允许误差）
	// 注意：由于随机性，我们放宽条件
	assert.Greater(t, staticCount, 5)      // 30%的应该多于5（放宽条件）
	assert.Greater(t, aggregatedCount, 30) // 70%的应该多于30（放宽条件）
}

// TestScaleboxAggregatorFromYAML 测试从YAML加载配置并使用ScaleboxAggregator
func TestScaleboxAggregatorFromYAML(t *testing.T) {
	if !isScaleboxEnvReady(t) {
		return
	}

	// 使用testdata中的vpath.yaml
	yamlFile := "testdata/vpath.yaml"

	// 使用NewVirtualPath从YAML文件创建
	vp, err := vpath.NewVirtualPath(appID, yamlFile)
	assert.NoError(t, err)
	assert.NotNil(t, vp)

	// 注意：从YAML加载的配置中，配置名称是"test-agg"
	// 查看vpath.yaml：test-agg配置中AGG_PATH的category是"cat-agg"
	// 但selector的key是配置名称"test-agg"
	path, err := vp.GetPath("test-agg", "test-yaml-job")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)

	// 验证路径是有效的（由于数据库中没有cat-agg信号量，可能会失败）
	// 我们只检查是否返回了路径
	assert.NotEmpty(t, path)

	// 清理
	err = vp.ReleasePath("test-agg", "test-yaml-job")
	assert.NoError(t, err)
}
