package aggpath_test

import (
	"beamform/internal/aggpath"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// 测试1: 空文件路径
	t.Run("empty file path", func(t *testing.T) {
		ap, err := aggpath.New(1, "")
		assert.NoError(t, err)
		assert.NotNil(t, ap)
		assert.Equal(t, 1, ap.AppID)
		assert.Empty(t, ap.StorageGroupMap)
	})

	// 测试2: 有效的storage group文件
	t.Run("valid storage group file", func(t *testing.T) {
		filePath := "my-group.txt"
		ap, err := aggpath.New(1, filePath)
		assert.NoError(t, err)
		assert.NotNil(t, ap)
		assert.Equal(t, 1, ap.AppID)
		assert.Equal(t, 2, len(ap.StorageGroupMap))
		assert.Equal(t, 20, ap.StorageGroupMap["group0"])
		assert.Equal(t, 10, ap.StorageGroupMap["group1"])
	})

	// 测试3: 不存在的文件
	t.Run("non-existent file", func(t *testing.T) {
		ap, err := aggpath.New(1, "/non/existent/file.txt")
		assert.Error(t, err)
		assert.Nil(t, ap)
	})

	// 测试4: 从文件创建并验证解析
	t.Run("create from file and verify parsing", func(t *testing.T) {
		ap, err := aggpath.New(1, "my-group.txt")
		assert.NoError(t, err)
		assert.NotNil(t, ap)
		assert.Greater(t, len(ap.StorageGroupMap), 0)
	})
}

func TestGetMemberPath(t *testing.T) {
	ap := &aggpath.AggregatedPath{
		AppID: 1,
		StorageGroupMap: map[string]int{
			"group0": 20,
			"group1": 30,
		},
	}

	// 测试1: 基本功能 - 不会崩溃
	t.Run("basic functionality - should not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.GetMemberPath("group0", "test-path")
		})
	})

	// 测试2: 无效的storage group
	t.Run("invalid storage group", func(t *testing.T) {
		result, err := ap.GetMemberPath("invalid-storage-group", "test-path")
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "no valid storage group")
	})

	// 测试3: path以"/"开头（应该被正确处理）
	t.Run("path starting with slash", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.GetMemberPath("group0", "/test/path")
		})
	})

	// 测试4: 空path
	t.Run("empty path", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.GetMemberPath("group0", "")
		})
	})

	// 测试5: 不同的storage group
	t.Run("different storage group", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.GetMemberPath("group1", "another-path")
		})
	})
}

func TestReleaseMemberPath(t *testing.T) {
	ap := &aggpath.AggregatedPath{
		AppID: 1,
		StorageGroupMap: map[string]int{
			"group0": 20,
			"group1": 30,
		},
	}

	// 测试1: 基本功能 - 不会崩溃
	t.Run("basic functionality - should not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.ReleaseMemberPath("group0", "test-path")
		})
	})

	// 测试2: 无效的storage group
	t.Run("invalid storage group", func(t *testing.T) {
		err := ap.ReleaseMemberPath("invalid-storage-group", "test-path")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid storage group")
	})

	// 测试3: path以"/"开头
	t.Run("path starting with slash", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.ReleaseMemberPath("group0", "/test/path")
		})
	})

	// 测试4: 空path
	t.Run("empty path", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.ReleaseMemberPath("group0", "")
		})
	})

	// 测试5: 释放不存在的路径（应该不会崩溃）
	t.Run("release non-existent path", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.ReleaseMemberPath("group0", "non-existent-path")
		})
	})

	// 测试6: 不同的storage group
	t.Run("different storage group", func(t *testing.T) {
		assert.NotPanics(t, func() {
			ap.ReleaseMemberPath("group1", "another-path")
		})
	})
}

func init() {
	// 设置环境变量以避免连接错误
	os.Setenv("PGHOST", "10.0.6.100")
}
