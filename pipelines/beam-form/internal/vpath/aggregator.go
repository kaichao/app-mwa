package vpath

// Aggregator 聚合目录管理器接口
type Aggregator interface {
	// Allocate 分配路径
	// key: 唯一标识符
	// capacityGB: 需要的容量（GB）
	Allocate(key string, capacityGB int) (string, error)

	// Release 释放路径
	// key: 唯一标识符
	// capacityGB: 释放的容量（GB）
	Release(key string, capacityGB int) error

	// Name 返回聚合器名称
	Name() string

	// Stats 返回统计信息
	Stats() map[string]interface{}
}
