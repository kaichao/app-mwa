package vpath

import (
	"encoding/binary"
	"hash/fnv"
	"math"
)

// ── 无状态加权选择：rendezvous hashing（HRW） ────────────────
//
//	member = argmax_i ( -weight_i / ln(hash(key, i)) )
//
// 三类存储虚拟化共用这一段实现：target 层选出目标（path 或 pool），
// 类型 b1 的读分摊与 b2 的写分摊由此得到。
//
// 与计数器方案的区别在于**无状态**：同一 key 恒选同一目标（可重放、
// 多进程天然一致），key 数量大时各目标被选中的频次 ∝ 权重。代价是
// 比例由精确变为统计趋近。

// hashPrecision 为哈希值取高位后的分母，2^53 使 float64 可精确表示。
const hashPrecision = 1 << 53

// selectTarget 按权重哈希从 targets 中选出一个。
//
// 权重 <= 0 的目标不参与选择；若没有可用的目标，返回 (Target{}, false)。
func selectTarget(targets []Target, key string) (Target, bool) {
	bestIdx := -1
	bestScore := 0.0

	for i, t := range targets {
		if t.Weight <= 0 {
			continue
		}
		score := -t.Weight / math.Log(hashToUnit(key, i))
		if bestIdx < 0 || score > bestScore {
			bestIdx, bestScore = i, score
		}
	}

	if bestIdx < 0 {
		return Target{}, false
	}
	return targets[bestIdx], true
}

// hashToUnit 把 (key, 目标下标) 映射到 [0, 1)。
//
// 取哈希值的高 53 位——float64 的尾数只有 53 位，取更多位会因舍入而丢精度。
//
// FNV-1a 之后必须再走一遍 mix64：本选择式对 h 接近 1 的尾部极其敏感
// （|ln h| → 0 时得分趋于无穷），而 FNV-1a 对短输入的雪崩不足，高位
// 之间残留的相关性会被这个尾部放大——实测 7:3 的权重会偏离成 8.8:1。
// 过一遍 finalizer 后回到 7:3。
func hashToUnit(key string, idx int) float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(idx))
	_, _ = h.Write(buf[:])

	return float64(mix64(h.Sum64())>>11) / hashPrecision
}

// mix64 是 splitmix64 的 finalizer，把 64 位输入的各位充分混合。
func mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}
