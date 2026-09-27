package vpath

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	config, err := loadConfig("testdata/vpath.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 5, len(config.Classes))
	assert.Equal(t, 2, len(config.Pools))
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := loadConfig("testdata/nonexistent.yaml")
	assert.Error(t, err)
}

func TestValidateOK(t *testing.T) {
	config, err := loadConfig("testdata/vpath.yaml")
	assert.NoError(t, err)

	err = config.validate()
	assert.NoError(t, err)
}

func TestValidateNoClasses(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no classes")
}

func TestValidateEmptyTargets(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"empty": {Targets: []Target{}},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no targets")
}

func TestValidateTargetNoPathOrPool(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"bad": {Targets: []Target{
				{Weight: 1.0},
			}},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must have path or pool")
}

func TestValidateTargetBothPathAndPool(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"bad": {Targets: []Target{
				{Path: "/foo", Pool: "bar", Weight: 1.0},
			}},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot have both")
}

func TestValidateTargetNegativeWeight(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"bad": {Targets: []Target{
				{Path: "/foo", Weight: -1.0},
			}},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weight must be >= 0")
}

func TestValidateTargetPoolNotFound(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"bad": {UnitSizeGB: 10, Targets: []Target{
				{Pool: "no-such-pool", Weight: 1.0},
			}},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pool not defined")
}

func TestValidateUnitSizeGBMissing(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"bad": {Targets: []Target{
				{Pool: "test-pool", Weight: 1.0},
			}},
		},
		Pools: map[string]*Pool{
			"test-pool": {Strategy: "max-free"},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unit_size_gb must be > 0")
}

func TestValidateTotalWeightZero(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"bad": {Targets: []Target{
				{Path: "/foo", Weight: 0.0},
			}},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total weight must be > 0")
}

// 池的成员不在配置里声明——只写分配策略即可
func TestValidatePoolStrategyOnly(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"ok": {Targets: []Target{
				{Path: "/foo", Weight: 1.0},
			}},
		},
		Pools: map[string]*Pool{
			"p": {Strategy: "max-free"},
		},
	}
	assert.NoError(t, config.validate())
}

func TestValidatePoolUnknownStrategy(t *testing.T) {
	config := &Config{
		Classes: map[string]*StorageClass{
			"ok": {Targets: []Target{
				{Path: "/foo", Weight: 1.0},
			}},
		},
		Pools: map[string]*Pool{
			"bad": {Strategy: "random"},
		},
	}
	err := config.validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown strategy")
}

func TestLoadFromTempFile(t *testing.T) {
	yamlContent := `
classes:
  test:
    unit_size_gb: 10
    targets:
      - path: "/tmp/data"
        weight: 1.0
`
	tmpFile := t.TempDir() + "/test-vpath.yaml"
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0644)
	assert.NoError(t, err)

	config, err := loadConfig(tmpFile)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(config.Classes))
}
