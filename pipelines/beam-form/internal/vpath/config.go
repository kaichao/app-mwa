package vpath

import (
	"os"

	"github.com/kaichao/gopkg/errors"
	"gopkg.in/yaml.v3"
)

// loadConfig 从 YAML 文件加载配置。
func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.WrapE(err, "read config file")
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.WrapE(err, "parse YAML")
	}

	return &config, nil
}

// validate 校验配置合法性。
func (c *Config) validate() error {
	if len(c.Classes) == 0 {
		return errors.E("no classes defined")
	}

	for name, class := range c.Classes {
		if len(class.Targets) == 0 {
			return errors.E("class has no targets", "class", name)
		}

		// 检查是否至少有一个 pool target（需要 unit_size_gb）
		hasPool := false

		totalWeight := 0.0
		for i, t := range class.Targets {
			if t.Path == "" && t.Pool == "" {
				return errors.E("target must have path or pool", "class", name, "index", i)
			}
			if t.Path != "" && t.Pool != "" {
				return errors.E("target cannot have both path and pool", "class", name, "index", i)
			}
			if t.Weight < 0 {
				return errors.E("target weight must be >= 0", "class", name, "index", i)
			}
			if t.Pool != "" {
				hasPool = true
				if _, ok := c.Pools[t.Pool]; !ok {
					return errors.E("pool not defined", "class", name, "pool", t.Pool)
				}
			}
			totalWeight += t.Weight
		}

		if totalWeight <= 0 {
			return errors.E("total weight must be > 0", "class", name)
		}

		if hasPool && class.UnitSizeGB <= 0 {
			return errors.E("unit_size_gb must be > 0 when class has pool target", "class", name)
		}
	}

	for name, pool := range c.Pools {
		// 池的成员不在配置里，故除策略外没有可校验的字段
		if pool.Strategy != "max-free" {
			return errors.E("unknown strategy", "pool", name, "strategy", pool.Strategy,
				"hint", "池级策略只有 max-free 一种")
		}
	}

	return nil
}
