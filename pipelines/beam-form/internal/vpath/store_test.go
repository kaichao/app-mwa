package vpath

import (
	"strings"
)

// memoryStore 是 store 的内存实现，让池的逻辑能脱离 scalebox 单测。
type memoryStore struct {
	vars  map[string]string
	semas map[string]int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		vars:  make(map[string]string),
		semas: make(map[string]int),
	}
}

// setSema 预置信号量，模拟人工用 sema 文件导入初始值。
func (m *memoryStore) setSema(name string, value int) {
	m.semas[name] = value
}


func (m *memoryStore) LookupVar(name string) (string, bool, error) {
	value, ok := m.vars[name]
	return value, ok && value != "", nil
}

func (m *memoryStore) SetVar(name, value string) error {
	m.vars[name] = value
	return nil
}

func (m *memoryStore) AddSema(name string, delta int) (int, error) {
	m.semas[name] += delta
	return m.semas[name], nil
}

// GetSemaMax 取组内值最大的成员。组内无成员时返回空名称，与生产语义一致。
func (m *memoryStore) GetSemaMax(group string) (string, int, error) {
	prefix := group + ":"
	bestName, bestValue := "", 0
	found := false

	for name, value := range m.semas {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		if !found || value > bestValue {
			bestName, bestValue, found = name, value, true
		}
	}

	if !found {
		return "", 0, nil
	}
	return bestName, bestValue, nil
}
