package vpath

import (
	"beamform/internal/varerr"

	"github.com/kaichao/scalebox/pkg/semagroup"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/variable"
)

// store 抽象 vpath 依赖的存储后端——共享变量与信号量。
//
// 抽出接口的目的是让池的分配逻辑能脱离 scalebox 单测：生产用
// scaleboxStore，测试用 memoryStore（见 store_test.go）。
type store interface {
	// LookupVar 读取共享变量。变量未设置、或值已被清空
	// （SetVar(name, "") 表示位置可重新分配）时返回 ok=false。
	LookupVar(name string) (value string, ok bool, err error)

	// SetVar 写入共享变量；写空串表示清除。
	SetVar(name, value string) error

	// AddSema 调整信号量的值（delta 可为负），返回调整后的值。
	AddSema(name string, delta int) (int, error)

	// GetSemaMax 取信号量组中值最大的成员，返回其完整名称与值。
	// 组内无成员时返回空名称。
	GetSemaMax(group string) (name string, value int, err error)
}

// ── 生产实现 ──────────────────────────────────────────────

// scaleboxStore 通过 scalebox 平台的 gRPC 接口读写变量与信号量。
type scaleboxStore struct {
	appID int
}

func newScaleboxStore(appID int) *scaleboxStore {
	return &scaleboxStore{appID: appID}
}

// LookupVar 把「变量不存在」与「值为空」都归一到 ok=false。
//
// 变量不存在是正常语义（尚未分配），不该当成致命错误——平台升级为
// gRPC 后端后，server 端返回 codes.Internal 且只在消息里保留 SQL 原文，
// 原先的 errors.Is(err, sql.ErrNoRows) 判断会失效，故统一走 varerr。
func (s *scaleboxStore) LookupVar(name string) (string, bool, error) {
	value, err := variable.GetValue(name, s.appID)
	if err != nil {
		if varerr.IsNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}
	if value == "" {
		return "", false, nil
	}
	return value, true, nil
}

func (s *scaleboxStore) SetVar(name, value string) error {
	return variable.Set(name, value, s.appID)
}

func (s *scaleboxStore) AddSema(name string, delta int) (int, error) {
	return semaphore.AddValue(name, delta, s.appID)
}

func (s *scaleboxStore) GetSemaMax(group string) (string, int, error) {
	return semagroup.GetMax(group, s.appID)
}
