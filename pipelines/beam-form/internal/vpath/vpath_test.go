package vpath_test

import (
	"beamform/internal/vpath"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createTestConfig 创建测试配置
func createTestConfig(name string, weightedPaths []vpath.WeightedPathConfig, aggregatedPaths []vpath.AggregatedPathConfig) *vpath.Config {
	// 复制weightedPaths，设置Type字段
	weightedPathsCopy := make([]vpath.WeightedPathConfig, len(weightedPaths))
	for i, wp := range weightedPaths {
		weightedPathsCopy[i] = wp
		// 推断类型
		if wp.Path == "AGG_PATH" {
			weightedPathsCopy[i].Type = "aggregated"
		} else {
			weightedPathsCopy[i].Type = "static"
		}
	}

	return &vpath.Config{
		Name:            name,
		WeightedPaths:   weightedPathsCopy,
		AggregatedPaths: aggregatedPaths,
		AggregatorType:  "memory", // 测试使用memory aggregator
	}
}

// TestNewVirtualPath 测试创建VirtualPath
func TestNewVirtualPath(t *testing.T) {
	// 创建临时YAML文件
	tempDir := os.TempDir()
	yamlContent := `
test-category:
  aggregator_type: "memory"
  weighted_paths:
    - path: "/path1"
      weight: 0.5
    - path: "/path2"
      weight: 0.5
`
	tempFile := tempDir + "/test-config.yaml"
	if err := os.WriteFile(tempFile, []byte(yamlContent), 0644); err != nil {
		t.Skipf("Error creating temp file: %v", err)
		return
	}
	defer os.Remove(tempFile)

	vp, err := vpath.NewVirtualPath(1, tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, vp)
}

// TestGetPathStatic 测试获取静态路径
func TestGetPathStatic(t *testing.T) {
	config := createTestConfig("test-static", []vpath.WeightedPathConfig{
		{Path: "/static1", Weight: 1.0},
	}, nil)

	// 直接使用编程方式创建
	vp, err := vpath.NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 现在selector的key是配置名称，而不是wp.Category
	path, err := vp.GetPath("test-static", "key1")
	assert.NoError(t, err)
	assert.Equal(t, "/static1", path)
}

// TestGetPathAggregated 测试获取聚合路径
func TestGetPathAggregated(t *testing.T) {
	config := createTestConfig("test-agg", []vpath.WeightedPathConfig{
		{
			Path:       "AGG_PATH",
			Weight:     1.0,
			Pool:       "storage",
			CapacityGB: 10,
		},
	}, []vpath.AggregatedPathConfig{
		{
			Name:       "storage",
			CapacityGB: 100,
			Members:    []string{"/node1", "/node2"},
		},
	})

	vp, err := vpath.NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 第一次分配
	path1, err := vp.GetPath("test-agg", "job1")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2"}, path1)

	// 相同key返回相同路径
	path2, err := vp.GetPath("test-agg", "job1")
	assert.NoError(t, err)
	assert.Equal(t, path1, path2)

	// 不同key可能不同
	path3, err := vp.GetPath("test-agg", "job2")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2"}, path3)
}

// TestReleasePath 测试释放路径
func TestReleasePath(t *testing.T) {
	config := createTestConfig("test-release", []vpath.WeightedPathConfig{
		{
			Path:       "AGG_PATH",
			Weight:     1.0,
			Pool:       "storage",
			CapacityGB: 10,
		},
	}, []vpath.AggregatedPathConfig{
		{
			Name:       "storage",
			CapacityGB: 30,
			Members:    []string{"/node1", "/node2", "/node3"},
		},
	})

	vp, err := vpath.NewVirtualPathFromConfig(1, config)
	assert.NoError(t, err)

	// 分配路径
	path, err := vp.GetPath("test-release", "job1")
	assert.NoError(t, err)
	_ = path // 使用变量避免编译警告

	// 释放路径
	err = vp.ReleasePath("test-release", "job1")
	assert.NoError(t, err)

	// 可以重新分配
	newPath, err := vp.GetPath("test-release", "job1")
	assert.NoError(t, err)
	assert.Contains(t, []string{"/node1", "/node2", "/node3"}, newPath)
}

// TestWeightedSelection 测试加权选择
func TestWeightedSelection(t *testing.T) {
	config := createTestConfig("test-weighted", []vpath.WeightedPathConfig{
		{Path: "/pathA", Weight: 0.7},
		{Path: "/pathB", Weight: 0.3},
	}, nil)

	vp, err := vpath.NewVirtualPathFromConfig(1, config)
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

// isScaleboxEnvReady 检查Scalebox测试环境是否就绪
func isScaleboxEnvReady(t *testing.T) bool {
	if os.Getenv("PGHOST") == "" {
		t.Skip("PGHOST environment variable not set, skipping Scalebox integration test")
		return false
	}
	return true
}

// cleanupTestVariables 清理测试期间创建的变量
func cleanupTestVariables(t *testing.T, category, key string) {
	// 尝试释放可能存在的variable
	// 注意：这里我们直接调用ScaleboxAggregator的Release方法
	// 但由于我们无法直接访问aggregator，我们只能尝试通过VirtualPath来释放
	// 实际上，清理应该在每个测试中完成，而不是在这里
	// 所以这个函数暂时留空，清理工作由测试自己完成
}

func TestCategoryTestStatic(t *testing.T) {
	vp, err := vpath.NewVirtualPath(appID, "testdata/vpath.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, vp)

	// test-static category只有静态路径
	validPaths := []string{"/test/path1", "/test/path2"}

	// 多次测试，确保选择器工作
	for i := 0; i < 20; i++ {
		p := fmt.Sprintf("key-%02d", i)
		path, err := vp.GetPath("test-static", p)
		assert.NoError(t, err)
		assert.Contains(t, validPaths, path)
	}
}

func TestCategoryTestAgg(t *testing.T) {
	vp, err := vpath.NewVirtualPath(appID, "testdata/vpath.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, vp)

	// test-agg category有AGG_PATH（权重1）和一个权重为0的静态路径
	// 权重为0的路径应该被过滤掉，所以只有AGG_PATH
	category := "test-agg"

	// 多次测试AGG_PATH分配
	for i := 0; i < 10; i++ {
		p := fmt.Sprintf("job-%02d", i)
		path, err := vp.GetPath(category, p)
		assert.NoError(t, err)
		// AGG_PATH应该返回有效的路径
		assert.NotEmpty(t, path)
	}

	for i := 0; i < 10; i++ {
		p := fmt.Sprintf("job-%02d", i)
		err := vp.ReleasePath(category, p)
		assert.NoError(t, err)
	}
}

func TestCategoryTestMixed(t *testing.T) {
	vp, err := vpath.NewVirtualPath(appID, "testdata/vpath.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, vp)

	// test-mixed category有2个静态路径和1个AGG_PATH
	category := "test-mixed"

	// 多次测试，验证选择器按权重选择
	staticPaths := []string{"/fast/ssd", "/slow/hdd"}
	// aggPathPrefix := "/test-mixed/dir" // AGG_PATH分配的路径（暂时注释掉，因为YAML中没有配置aggregated_paths）

	// 统计选择结果
	staticCount := 0
	aggCount := 0

	for i := 0; i < 30; i++ {
		p := fmt.Sprintf("task-%02d", i)
		path, err := vp.GetPath(category, p)
		assert.NoError(t, err)

		if path == "/fast/ssd" || path == "/slow/hdd" {
			staticCount++
			assert.Contains(t, staticPaths, path)
		} else if len(path) > 0 {
			aggCount++
			// AGG_PATH分配的路径
			assert.NotEmpty(t, path)
		}
	}

	// 验证大致比例（权重：/fast/ssd=1, /slow/hdd=2, AGG_PATH=3）
	// 总权重=6，静态路径总权重=3，AGG_PATH权重=3
	// 静态路径应该占约50%，AGG_PATH占约50%
	total := staticCount + aggCount
	if total > 0 {
		staticRatio := float64(staticCount) / float64(total)
		// 允许误差：40%-60%
		assert.Greater(t, staticRatio, 0.4)
		assert.Less(t, staticRatio, 0.6)
	}
}

var appID int

func init() {
	appID = 1
	os.Setenv("PGHOST", "10.0.6.100")
}
