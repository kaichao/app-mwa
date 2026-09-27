package vpath

import (
	"fmt"
	"os"
	"strings"

	"github.com/kaichao/gopkg/errors"
	"github.com/sirupsen/logrus"
)

// ── 命名约定 ──────────────────────────────────────────────
//
// 键名一律以 vpath: 为前缀，集中在此构造，避免散落各文件手写字符串。

// memberVarName 是「数据单元 → 所在空间」的位置记录键。
func memberVarName(poolName, key string) string {
	return fmt.Sprintf("vpath:member-path:%s:%s", poolName, key)
}

// memberSemaName 是池内某成员剩余容量的信号量键。
func memberSemaName(poolName, member string) string {
	return fmt.Sprintf("vpath:free-gb:%s:%s", poolName, member)
}

// poolSemaGroup 是池的信号量分组名。
func poolSemaGroup(poolName string) string {
	return fmt.Sprintf("vpath:free-gb:%s", poolName)
}

// ── Pool 层：空间分配 ─────────────────────────────────────

// allocate 从池中分配空间。
func (p *Pool) allocate(s store, poolName, key string, needGB int) (string, error) {
	// 检查是否已分配
	varName := memberVarName(poolName, key)
	memberPath, ok, err := s.LookupVar(varName)
	if err != nil {
		return "", errors.WrapE(err, "get variable", "var", varName)
	}
	if ok {
		return memberPath, nil
	}

	// 池级策略只有 max-free 一种，合法值已由配置校验保证
	picked, err := p.pickMaxFree(s, poolName, needGB)
	if err != nil {
		return "", err
	}

	// 扣减容量
	semaName := memberSemaName(poolName, picked)
	if _, err := s.AddSema(semaName, -needGB); err != nil {
		return "", errors.WrapE(err, "deduct capacity", "sema", semaName)
	}

	// 记录映射
	if err := s.SetVar(varName, picked); err != nil {
		return "", errors.WrapE(err, "set variable", "var", varName)
	}

	return picked, nil
}

// release 释放已分配的空间。
// release 释放已分配的空间，返回是否确有释放。
// 该 key 在本池没有分配记录时返回 released=false——重复释放是幂等的。
func (p *Pool) release(s store, poolName, key string, needGB int) (bool, error) {
	varName := memberVarName(poolName, key)
	memberPath, ok, err := s.LookupVar(varName)
	if err != nil {
		return false, errors.WrapE(err, "get variable", "var", varName)
	}
	if !ok {
		return false, nil
	}

	// 删除数据
	if err := os.RemoveAll(memberPath); err != nil {
		logrus.Warnf("failed to remove %s: %v", memberPath, err)
	}

	// 归还容量
	semaName := memberSemaName(poolName, memberPath)
	if _, err := s.AddSema(semaName, needGB); err != nil {
		return false, errors.WrapE(err, "add capacity", "sema", semaName)
	}

	// 清除映射
	if err := s.SetVar(varName, ""); err != nil {
		return false, errors.WrapE(err, "clear variable", "var", varName)
	}
	return true, nil
}

// pickMaxFree 选剩余空间最大的 member。
func (p *Pool) pickMaxFree(s store, poolName string, needGB int) (string, error) {
	semaGroup := poolSemaGroup(poolName)
	semaName, semaValue, err := s.GetSemaMax(semaGroup)
	if err != nil {
		return "", errors.WrapE(err, "semagroup.GetMax", "group", semaGroup)
	}
	if semaValue < needGB {
		return "", errors.E("no enough disk space", "pool", poolName, "need", needGB, "max_free", semaValue)
	}

	// semaName 格式: "vpath:free-gb:<pool>:<memberPath>"
	// memberPath 自身可能含冒号（如 astro@10.100.1.30:10022/...），故限次切分
	parts := strings.SplitN(semaName, ":", 4)
	if len(parts) < 4 {
		return "", errors.E("invalid sema name", "sema", semaName)
	}
	return parts[3], nil
}

