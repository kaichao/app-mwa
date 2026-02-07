package aggpath_test

import (
	"beamform/internal/aggpath"
)

// ExampleNew 展示如何使用New函数从文件创建AggregatedPath
func ExampleNew() {
	// 从文件创建AggregatedPath
	ap, err := aggpath.New(1, "my-group.txt")
	if err != nil {
		// 处理错误
		return
	}

	// 使用AggregatedPath
	_ = ap
	// Output:
}

// ExampleAggregatedPath_GetMemberPath 展示如何使用GetMemberPath函数获取成员路径
func ExampleAggregatedPath_GetMemberPath() {
	// 创建AggregatedPath实例
	ap := &aggpath.AggregatedPath{
		AppID: 1,
		StorageGroupMap: map[string]int{
			"group0": 20,
		},
	}

	// 获取成员路径
	path, err := ap.GetMemberPath("group0", "test-path")
	if err != nil {
		// 处理错误
		return
	}

	// 使用获取的路径
	_ = path
	// Output:
}

// ExampleAggregatedPath_ReleaseMemberPath 展示如何使用ReleaseMemberPath函数释放成员路径
func ExampleAggregatedPath_ReleaseMemberPath() {
	// 创建AggregatedPath实例
	ap := &aggpath.AggregatedPath{
		AppID: 1,
		StorageGroupMap: map[string]int{
			"group0": 20,
		},
	}

	// 释放成员路径
	err := ap.ReleaseMemberPath("group0", "test-path")
	if err != nil {
		// 处理错误
		return
	}
	// Output:
}
