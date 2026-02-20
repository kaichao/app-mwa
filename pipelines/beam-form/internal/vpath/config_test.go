package vpath

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfigFromYAML 测试从YAML加载配置
func TestLoadConfigFromYAML(t *testing.T) {
	// 创建临时YAML文件
	tempDir := os.TempDir()
	yamlContent := `
test-category:
  weighted_paths:
    - path: "/path1"
      weight: 0.5
      capacity_gb: 100
    - path: "/path2"
      weight: 0.5
      capacity_gb: 200
`
	tempFile := filepath.Join(tempDir, "test-config.yaml")
	if err := os.WriteFile(tempFile, []byte(yamlContent), 0644); err != nil {
		t.Skipf("Error creating temp file: %v", err)
		return
	}
	defer os.Remove(tempFile)

	config, err := loadConfigFromYAML(tempFile, "test-category")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-category", config.Name)
	assert.Len(t, config.WeightedPaths, 2)
	assert.Equal(t, "/path1", config.WeightedPaths[0].Path)
	assert.Equal(t, 0.5, config.WeightedPaths[0].Weight)
	assert.Equal(t, "/path2", config.WeightedPaths[1].Path)
	assert.Equal(t, 0.5, config.WeightedPaths[1].Weight)
}

// TestLoadConfigFromYAMLWithAGG_PATH 测试加载包含AGG_PATH的配置
func TestLoadConfigFromYAMLWithAGG_PATH(t *testing.T) {
	// 创建临时YAML文件
	tempDir := os.TempDir()
	yamlContent := `test-agg:
  weighted_paths:
    - path: AGG_PATH
      weight: 1.0
      pool: storage
      need_gb: 10
`
	tempFile := filepath.Join(tempDir, "test-agg-config.yaml")
	if err := os.WriteFile(tempFile, []byte(yamlContent), 0644); err != nil {
		t.Skipf("Error creating temp file: %v", err)
		return
	}
	defer os.Remove(tempFile)

	config, err := loadConfigFromYAML(tempFile, "test-agg")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "test-agg", config.Name)
	assert.Len(t, config.WeightedPaths, 1)
	assert.Equal(t, "AGG_PATH", config.WeightedPaths[0].Path)
	assert.Equal(t, "aggregated", config.WeightedPaths[0].Type)
	assert.Equal(t, "storage", config.WeightedPaths[0].Pool)
	assert.Equal(t, 10, config.WeightedPaths[0].CapacityGB)
}

// TestLoadAllConfigsFromYAML 测试加载所有配置
func TestLoadAllConfigsFromYAML(t *testing.T) {
	// 创建临时YAML文件
	tempDir := os.TempDir()
	yamlContent := `
test-category1:
  weighted_paths:
    - path: "/path1"
      weight: 1.0
      capacity_gb: 100
test-category2:
  weighted_paths:
    - path: "/path2"
      weight: 1.0
      capacity_gb: 200
`
	tempFile := filepath.Join(tempDir, "test-all-configs.yaml")
	if err := os.WriteFile(tempFile, []byte(yamlContent), 0644); err != nil {
		t.Skipf("Error creating temp file: %v", err)
		return
	}
	defer os.Remove(tempFile)

	configs, err := loadAllConfigsFromYAML(tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, configs)
	assert.Len(t, configs, 2)
	assert.Contains(t, configs, "test-category1")
	assert.Contains(t, configs, "test-category2")
}

// TestValidateConfigWeightZero 测试权重为0的配置
func TestValidateConfigWeightZero(t *testing.T) {
	config := &Config{
		Name: "test-weight-zero",
		WeightedPaths: []WeightedPathConfig{
			{Path: "/path1", Weight: 0.0, Type: "static", Pool: "default"},
			{Path: "/path2", Weight: 1.0, Type: "static", Pool: "default"},
		},
	}

	err := validateConfig(config)
	assert.NoError(t, err)
	// 权重为0的项应该被过滤掉
	assert.Len(t, config.WeightedPaths, 1)
	assert.Equal(t, "/path2", config.WeightedPaths[0].Path)
}

// TestValidateConfigTotalWeightZero 测试总权重为0的配置
func TestValidateConfigTotalWeightZero(t *testing.T) {
	config := &Config{
		Name: "test-total-weight-zero",
		WeightedPaths: []WeightedPathConfig{
			{Path: "/path1", Weight: 0.0, Type: "static", Pool: "default"},
			{Path: "/path2", Weight: 0.0, Type: "static", Pool: "default"},
		},
	}

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total weight of all WeightedPaths must be > 0")
}

// TestValidateConfigWeightNegative 测试权重为负的配置
func TestValidateConfigWeightNegative(t *testing.T) {
	config := &Config{
		Name: "test-weight-negative",
		WeightedPaths: []WeightedPathConfig{
			{Path: "/path1", Weight: -1.0, Type: "static", Pool: "default"},
		},
	}

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Weight must be >= 0")
}
