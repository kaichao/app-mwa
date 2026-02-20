package vpath

import "github.com/kaichao/gopkg/errors"

// MemoryAggregator 内存聚合器（用于测试）
// 注意：这不是生产就绪的实现，不持久化状态
type MemoryAggregator struct {
	name         string
	capacity     int // 总容量（GB）
	members      []string
	allocated    map[string]string // key -> memberPath
	usedCapacity map[string]int    // memberPath -> 已使用容量
}

// NewMemoryAggregator 创建内存聚合器
func NewMemoryAggregator(config AggregatedPathConfig) *MemoryAggregator {
	return &MemoryAggregator{
		name:         config.Name,
		capacity:     config.CapacityGB,
		members:      config.Members,
		allocated:    make(map[string]string),
		usedCapacity: make(map[string]int),
	}
}

// Allocate 分配路径
func (ma *MemoryAggregator) Allocate(key string, capacityGB int) (string, error) {
	// 生成唯一标识符
	allocKey := ma.allocationKey(ma.name, key)

	// 检查是否已分配
	if path, ok := ma.allocated[allocKey]; ok {
		return path, nil
	}

	// 查找有足够容量的成员路径
	for _, member := range ma.members {
		used := ma.usedCapacity[member]
		if used+capacityGB <= ma.capacityPerMember() {
			// 分配这个成员路径
			ma.allocated[allocKey] = member
			ma.usedCapacity[member] = used + capacityGB
			return member, nil
		}
	}

	return "", errors.E("no enough capacity",
		"pool", ma.name,
		"capacity_needed", capacityGB,
		"total_capacity", ma.capacity)
}

// Release 释放路径
func (ma *MemoryAggregator) Release(key string, capacityGB int) error {
	allocKey := ma.allocationKey(ma.name, key)

	path, ok := ma.allocated[allocKey]
	if !ok {
		// 路径未分配，直接返回成功
		return nil
	}

	// 减少已使用容量
	if used, ok := ma.usedCapacity[path]; ok && used >= capacityGB {
		ma.usedCapacity[path] = used - capacityGB
	} else if ok {
		// 如果已使用容量小于要释放的容量，设为0
		ma.usedCapacity[path] = 0
	}

	// 删除分配记录
	delete(ma.allocated, allocKey)

	return nil
}

// Name 返回聚合器名称
func (ma *MemoryAggregator) Name() string {
	return ma.name
}

// Stats 返回统计信息
func (ma *MemoryAggregator) Stats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["name"] = ma.name
	stats["total_capacity"] = ma.capacity
	stats["member_count"] = len(ma.members)
	stats["allocated_count"] = len(ma.allocated)

	// 计算每个成员的容量使用情况
	memberStats := make([]map[string]interface{}, len(ma.members))
	for i, member := range ma.members {
		used := ma.usedCapacity[member]
		memberStats[i] = map[string]interface{}{
			"path":  member,
			"used":  used,
			"free":  ma.capacityPerMember() - used,
			"total": ma.capacityPerMember(),
		}
	}
	stats["member_stats"] = memberStats

	return stats
}

// allocationKey 生成分配键
func (ma *MemoryAggregator) allocationKey(pool, key string) string {
	return pool + ":" + key
}

// capacityPerMember 计算每个成员的平均容量
func (ma *MemoryAggregator) capacityPerMember() int {
	if len(ma.members) == 0 {
		return 0
	}
	return ma.capacity / len(ma.members)
}
