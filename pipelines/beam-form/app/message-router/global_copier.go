package main

// 全局存储间的拷贝
// - 通过环境变量，实现基于容量的流控
// - Staging -> Final 优先，快速腾出计算存储
//
// Origin -> Preload
// Staging -> Final

func fromGlobalCopier(body string, headers map[string]string) int {
	return 0
}

func toGlobalCopier() int {
	return 0
}
