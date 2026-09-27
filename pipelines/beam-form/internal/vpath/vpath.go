// Package vpath 路径映射引擎
//
// 两层模型：
//   StorageClass — 存储类别，定义单位容量 + 加权目标（路径 或 Pool）
//   Pool        — 存储池，按剩余量最大的策略分配空间
//
// 三类存储虚拟化。类型不在配置里声明，而由调用方选择的 API 决定：
//   a  — 聚合：多个小空间合成大空间，写时按剩余量选（每次分配都选）
//   b1 — 复制：多个空间存同一份，读时按权重哈希选一个副本（完全无状态）
//   b2 — 合并：多个空间容量相加，首次写入时按权重哈希选
//
// API：
//   engine, _ := vpath.Load("/vpath.yaml", appID)
//   path, err := engine.Allocate(class, key)     // 确定位置，必要时新建
//   path, ok := engine.Locate(class, key)        // 只查不写
//   paths, err := engine.AllocateAll(class, key) // b1 写多份：全部副本路径
//   err := engine.Release(class, key)            // 按实际分配池释放
package vpath

import (
	"github.com/kaichao/gopkg/errors"
)

// ── 配置类型 ──────────────────────────────────────────────

// Config 顶层配置，对应 YAML 文件结构。
type Config struct {
	Classes map[string]*StorageClass `yaml:"classes"`
	Pools   map[string]*Pool         `yaml:"pools"`
}

// StorageClass 存储类别：定义该类数据的单位容量及存储目标。
type StorageClass struct {
	UnitSizeGB int      `yaml:"unit_size_gb,omitempty"` // 该类数据每份的大小（GB）
	Targets    []Target `yaml:"targets"`
}

// Target 存储目标。Path 和 Pool 二选一。
type Target struct {
	Path   string  `yaml:"path,omitempty"`
	Pool   string  `yaml:"pool,omitempty"`
	Weight float64 `yaml:"weight"`
}

// Pool 存储池：分配策略。
//
// 池级策略只有 max-free 一种（类型 a 用）；类型 b2 的固定比例落在
// class 的 targets 权重上，不在池上。
//
// 池有哪些成员**不在配置里声明**——完全由信号量组 vpath:free-gb:<pool>
// 决定。管理员用 sema 文件导入（scalebox semaphore create --sema-file），
// 分配时取组内剩余量最大的那个。成员路径从信号量名里切出，见
// memberSemaName。
type Pool struct {
	Strategy string `yaml:"strategy"` // 仅 "max-free"
}

// ── 运行时类型 ────────────────────────────────────────────

// Engine 路径映射引擎。
type Engine struct {
	config *Config
	store  store
}

// ── 公开 API ──────────────────────────────────────────────

// Load 从 YAML 文件加载配置并创建 Engine，使用 scalebox 平台作存储后端。
func Load(configFile string, appID int) (*Engine, error) {
	return LoadWithStore(configFile, newScaleboxStore(appID))
}

// LoadWithStore 从 YAML 文件加载配置并创建 Engine，使用指定的存储后端。
// 测试可传入内存实现（见 store_test.go），使池逻辑脱离 scalebox 单测。
func LoadWithStore(configFile string, s store) (*Engine, error) {
	config, err := loadConfig(configFile)
	if err != nil {
		return nil, errors.WrapE(err, "load config")
	}
	if err := config.validate(); err != nil {
		return nil, errors.WrapE(err, "validate config")
	}

	// 与 vpath 一样：加载只读配置文件建内存结构，不查数据库
	return &Engine{config: config, store: s}, nil
}

// ── 路径查询与分配 ────────────────────────────────────────
//
// 查与分是两个动词：Locate 只查不写，Allocate / AllocateAll 确定位置。
// 分开是为了避免「查一个尚未分配的东西，却凭空分配一块空间并扣掉容量」。
//
// 容量一律取 class 的 unit_size_gb，不由调用方传入——避免分配与释放
// 各传一个 size 而账本对不上的隐患。

// Locate 查询数据单元当前所在的位置。只查不写，未分配时返回 ("", false)。
//
// 先查位置记录（类型 a / b2 在分配时写在这里）；无记录时按权重哈希选
// 一个 target——选中的是 path，说明该类别的副本处处都在（类型 b1）。
func (e *Engine) Locate(className, key string) (string, bool) {
	class, ok := e.config.Classes[className]
	if !ok {
		return "", false
	}

	for _, t := range class.Targets {
		if t.Pool == "" {
			continue
		}
		if path, found := e.lookup(t.Pool, key); found {
			return path, true
		}
	}

	target, ok := selectTarget(class.Targets, key)
	if !ok || target.Path == "" {
		return "", false
	}
	return target.Path, true
}

// lookup 查某池中该 key 的位置记录；查询失败一律视同未分配。
func (e *Engine) lookup(poolName, key string) (string, bool) {
	path, ok, err := e.store.LookupVar(memberVarName(poolName, key))
	if err != nil || !ok {
		return "", false
	}
	return path, true
}

// Allocate 确定数据单元的位置：已分配则返回原位置，否则选出空间并记下。
func (e *Engine) Allocate(className, key string) (string, error) {
	class, ok := e.config.Classes[className]
	if !ok {
		return "", errors.E("class not found", "class", className)
	}

	// 幂等：同一 key 重复调用返回同一位置
	if path, found := e.Locate(className, key); found {
		return path, nil
	}

	target, ok := selectTarget(class.Targets, key)
	if !ok {
		return "", errors.E("no selectable target", "class", className)
	}
	if target.Path != "" {
		return target.Path, nil
	}

	pool, ok := e.config.Pools[target.Pool]
	if !ok {
		return "", errors.E("pool not found", "pool", target.Pool)
	}
	return pool.allocate(e.store, target.Pool, key, class.UnitSizeGB)
}

// AllocateAll 返回该数据单元的全部副本位置（类型 b1 的写多份）。
//
// 按组成元素逐个处理：path 直接返回，pool 在池内各分配一块，共 n 个。
// 只对类型 b1 有意义——对以 pool 为主的类别调用会造成 N 倍容量分配。
func (e *Engine) AllocateAll(className, key string) ([]string, error) {
	class, ok := e.config.Classes[className]
	if !ok {
		return nil, errors.E("class not found", "class", className)
	}

	paths := make([]string, 0, len(class.Targets))
	for _, t := range class.Targets {
		if t.Path != "" {
			paths = append(paths, t.Path)
			continue
		}

		pool, ok := e.config.Pools[t.Pool]
		if !ok {
			return nil, errors.E("pool not found", "pool", t.Pool)
		}
		path, err := pool.allocate(e.store, t.Pool, key, class.UnitSizeGB)
		if err != nil {
			return nil, errors.WrapE(err, "allocate", "pool", t.Pool)
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// Release 释放该数据单元占用的空间。
//
// 按**实际分配**的池释放——遍历 class 的 pool target 逐个查位置记录，
// 命中即释放，多池 class 的次生池不会被漏掉。未分配时静默返回（幂等）。
// 对类型 b1（targets 全为 path）无需释放。
func (e *Engine) Release(className, key string) error {
	class, ok := e.config.Classes[className]
	if !ok {
		return errors.E("class not found", "class", className)
	}

	for _, t := range class.Targets {
		if t.Pool == "" {
			continue
		}
		pool, ok := e.config.Pools[t.Pool]
		if !ok {
			return errors.E("pool not found", "pool", t.Pool)
		}
		if _, err := pool.release(e.store, t.Pool, key, class.UnitSizeGB); err != nil {
			return errors.WrapE(err, "release", "pool", t.Pool, "key", key)
		}
	}

	return nil
}

// ValidateConfig 验证 YAML 配置文件的合法性（包外可访问）。
func ValidateConfig(filename string) error {
	config, err := loadConfig(filename)
	if err != nil {
		return err
	}
	return config.validate()
}
