package vpath

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ── selectTarget 无状态哈希测试 ────────────────────────────

func TestSelectTargetDeterministic(t *testing.T) {
	targets := []Target{
		{Path: "/a", Weight: 1.0},
		{Path: "/b", Weight: 2.0},
		{Path: "/c", Weight: 1.0},
	}

	for _, key := range []string{"cube-1", "cube-2", "cube-3", "cube-1000"} {
		first, ok := selectTarget(targets, key)
		assert.True(t, ok)

		for i := 0; i < 10; i++ {
			again, _ := selectTarget(targets, key)
			assert.Equal(t, first.Path, again.Path, "key %s 应恒选同一 target", key)
		}
	}
}

func TestSelectTargetDistribution(t *testing.T) {
	targets := []Target{
		{Path: "/a", Weight: 7.0},
		{Path: "/b", Weight: 3.0},
	}

	n := 10000
	countA := 0
	for i := 0; i < n; i++ {
		picked, ok := selectTarget(targets, fmt.Sprintf("cube-%d", i))
		assert.True(t, ok)
		if picked.Path == "/a" {
			countA++
		}
	}

	// 期望 7000；容差 500（5%）远大于抽样标准差（约 46）
	assert.InDelta(t, 7000, countA, 500)
}

func TestSelectTargetComplexWeights(t *testing.T) {
	targets := []Target{
		{Path: "/a", Weight: 5.0},
		{Path: "/b", Weight: 3.0},
		{Path: "/c", Weight: 2.0},
	}

	counts := map[string]int{}
	for i := 0; i < 10000; i++ {
		picked, ok := selectTarget(targets, fmt.Sprintf("cube-%d", i))
		assert.True(t, ok)
		counts[picked.Path]++
	}

	assert.InDelta(t, 5000, counts["/a"], 500)
	assert.InDelta(t, 3000, counts["/b"], 500)
	assert.InDelta(t, 2000, counts["/c"], 500)
}

func TestSelectTargetSkipsZeroWeight(t *testing.T) {
	targets := []Target{
		{Path: "/active", Weight: 1.0},
		{Path: "/disabled", Weight: 0.0},
	}

	for i := 0; i < 100; i++ {
		picked, ok := selectTarget(targets, fmt.Sprintf("cube-%d", i))
		assert.True(t, ok)
		assert.Equal(t, "/active", picked.Path)
	}
}

func TestSelectTargetNoneAvailable(t *testing.T) {
	allZero := []Target{
		{Path: "/a", Weight: 0.0},
		{Path: "/b", Weight: 0.0},
	}
	_, ok := selectTarget(allZero, "cube-1")
	assert.False(t, ok, "权重全为 0 时应返回 false")

	_, ok = selectTarget(nil, "cube-1")
	assert.False(t, ok, "空 targets 应返回 false")
}

// ── Engine Load + path target 测试 ─────────────────────────

func TestLoad(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)
	assert.NotNil(t, engine)
}

func TestAllocateReadClass(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	validPaths := map[string]bool{
		"/data/partition-a": true,
		"/data/partition-b": true,
	}

	// 等权重的两个路径都应被选到（哈希把不同 key 分摊开）
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		path, err := engine.Allocate("read-class", fmt.Sprintf("key-%d", i))
		assert.NoError(t, err)
		assert.True(t, validPaths[path], "unexpected path: %s", path)
		seen[path] = true
	}
	assert.Len(t, seen, 2)
}

func TestAllocateDisabledTarget(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		path, err := engine.Allocate("disabled-target", fmt.Sprintf("key-%d", i))
		assert.NoError(t, err)
		assert.Equal(t, "/data/primary", path)
	}
}

func TestAllocateClassNotFound(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	_, err = engine.Allocate("no-such-class", "key")
	assert.Error(t, err)
}

func TestReleaseNoPoolTarget(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	err = engine.Release("read-class", "key")
	assert.NoError(t, err)
}

// ── ValidateConfig 测试 ──────────────────────────────────

func TestValidateConfigOK(t *testing.T) {
	err := ValidateConfig("testdata/vpath.yaml")
	assert.NoError(t, err)
}

func TestValidateConfigBadFile(t *testing.T) {
	err := ValidateConfig("testdata/nonexistent.yaml")
	assert.Error(t, err)
}

func TestValidateConfigBadYAML(t *testing.T) {
	yamlContent := `
classes:
  bad:
    targets:
      - path: "/x"
        weight: 0.0
`
	tmpFile := t.TempDir() + "/bad.yaml"
	os.WriteFile(tmpFile, []byte(yamlContent), 0644)

	err := ValidateConfig(tmpFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "total weight must be > 0")
}

// ── 池分配测试（内存后端，不依赖 scalebox server） ──────────

// newTestStore 预置各池成员的信号量，模拟管理员用
// scalebox semaphore create --sema-file 导入的初始剩余量。
func newTestStore() *memoryStore {
	s := newMemoryStore()
	for _, m := range []string{"/pool/node1", "/pool/node2", "/pool/node3"} {
		s.setSema("vpath:free-gb:test-pool:"+m, 100)
	}
	for _, m := range []string{"/large/disk1", "/large/disk2", "/small/disk3"} {
		s.setSema("vpath:free-gb:big-pool:"+m, 1000)
	}
	return s
}

func TestAllocatePoolTarget(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	path, err := engine.Allocate("write-class", "test-key-1")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)

	// 幂等：同一 key 重复分配返回同一位置
	path2, err := engine.Allocate("write-class", "test-key-1")
	assert.NoError(t, err)
	assert.Equal(t, path, path2)

	err = engine.Release("write-class", "test-key-1")
	assert.NoError(t, err)
}

func TestAllocatePoolMultipleKeys(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	paths := make(map[string]string)
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("multi-key-%d", i)
		path, err := engine.Allocate("write-class", key)
		assert.NoError(t, err)
		paths[key] = path
	}

	for key, expected := range paths {
		path, err := engine.Allocate("write-class", key)
		assert.NoError(t, err)
		assert.Equal(t, expected, path)
	}

	for key := range paths {
		err = engine.Release("write-class", key)
		assert.NoError(t, err)
	}
}

// 多池 class 的释放必须按**实际分配**的池进行——旧实现在循环内 return
// 第一个 pool，次生池的位置记录与容量会永久泄漏。
func TestReleaseSecondaryPool(t *testing.T) {
	s := newTestStore()
	engine, err := LoadWithStore("testdata/vpath.yaml", s)
	assert.NoError(t, err)

	var key, member string
	for i := 0; i < 100 && member == ""; i++ {
		k := fmt.Sprintf("multi-%d", i)
		_, err := engine.Allocate("multi-pool", k)
		assert.NoError(t, err)
		if v, ok := s.vars["vpath:member-path:big-pool:"+k]; ok {
			key, member = k, v
		}
	}
	if key == "" {
		t.Fatal("100 个 key 中应至少有一个落在 big-pool（multi-pool 的次生池）")
	}

	before := s.semas["vpath:free-gb:big-pool:"+member]

	assert.NoError(t, engine.Release("multi-pool", key))

	assert.Empty(t, s.vars["vpath:member-path:big-pool:"+key], "次生池的位置记录应被清除")
	assert.Equal(t, before+10, s.semas["vpath:free-gb:big-pool:"+member],
		"容量应按 unit_size_gb 归还")
}

// ── Locate / AllocateAll ──────────────────────────────────

func TestLocateUnallocated(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	// write-class 只有 pool target，未分配时没有位置可查
	path, ok := engine.Locate("write-class", "never-allocated")
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestLocateAfterAllocate(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	allocated, err := engine.Allocate("write-class", "k1")
	assert.NoError(t, err)

	path, ok := engine.Locate("write-class", "k1")
	assert.True(t, ok)
	assert.Equal(t, allocated, path)
}

func TestLocatePathClass(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	// read-class 全是 path target：副本处处都在，无需分配即命中
	for i := 0; i < 10; i++ {
		path, ok := engine.Locate("read-class", fmt.Sprintf("k-%d", i))
		assert.True(t, ok)
		assert.NotEmpty(t, path)
	}
}

func TestAllocateAll(t *testing.T) {
	engine, err := LoadWithStore("testdata/vpath.yaml", newTestStore())
	assert.NoError(t, err)

	paths, err := engine.AllocateAll("read-class", "k1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"/data/partition-a", "/data/partition-b"}, paths)
}

// 容量一律取 class 的 unit_size_gb，调用方不再传入——原先的 sizeGB
// 覆盖参数已按设计移除（见 DESIGN 三.API）。
