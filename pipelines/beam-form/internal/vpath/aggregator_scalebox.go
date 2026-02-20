package vpath

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/semagroup"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

// ScaleboxAggregator 存储聚合器（生产用）
// 使用Scalebox的信号量和变量管理聚合目录状态
type ScaleboxAggregator struct {
	name      string
	appID     int
	semaName  string // 信号量名称，如"path-free-gb:mwa"
	varPrefix string // 变量前缀，如"member-path:mwa"
}

// NewScaleboxAggregator 创建Scalebox存储聚合器
// pool: 存储池标识名，如"mwa"
// appID: 应用ID
func NewScaleboxAggregator(pool string, appID int) (*ScaleboxAggregator, error) {
	if pool == "" {
		return nil, errors.E("pool is empty")
	}

	// 构建信号量和变量名称（遵循aggpath的命名约定）
	semaName := "path-free-gb:" + pool
	varPrefix := "member-path:" + pool

	return &ScaleboxAggregator{
		name:      pool,
		appID:     appID,
		semaName:  semaName,
		varPrefix: varPrefix,
	}, nil
}

// Allocate 分配路径
func (sa *ScaleboxAggregator) Allocate(key string, capacityGB int) (string, error) {
	// if path对应variable存在，直接返回该目录
	varName := fmt.Sprintf("member-path:%s:%s", sa.name, key)
	varValue, err := variable.GetValue(varName, 0, sa.appID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", errors.WrapE(err, "variable.GetValue()", "var-name", varName)
	}
	if varValue != "" {
		return varValue, nil
	}

	// var-value == ""
	// 通过pool获取semagroup的最大值，若该值超过capacityDB值，减去该值
	semaGName := "path-free-gb:" + sa.name
	semaName, semaValue, err := semagroup.GetMax(semaGName, sa.appID)
	if err != nil {
		return "", errors.WrapE(err, "semagroup-getmax",
			"semagroup-name", semaGName, "app-id", sa.appID)
	}
	if semaValue < capacityGB {
		return "", errors.E("No enough disk space", "pool", sa.name)
	}
	_, err = semaphore.AddValue(semaName, 0, sa.appID, -capacityGB)
	if err != nil {
		return "", errors.WrapE(err, "semaphore-addvalue",
			"sema-name", semaName, "app-id", sa.appID)
	}
	// 以前述的信号量名，创建对应的variable
	// 去掉前面的path-free-gb: 及 name
	varValue = strings.Split(semaName, ":")[2]
	err = variable.Set(varName, varValue, 0, sa.appID)
	if err != nil {
		return "", errors.WrapE(err, "variable-set",
			"var-name", varName, "app-id", sa.appID)
	}
	// 返回前面的目录
	return varValue, nil
}

// Release 释放路径
func (sa *ScaleboxAggregator) Release(key string, capacityGB int) error {
	// 获取对应的variable值（即成员目录）
	varName := fmt.Sprintf("member-path:%s:%s", sa.name, key)
	memberPath, err := variable.GetValue(varName, 0, sa.appID)
	if err != nil {
		// 如果variable不存在，可能已经被释放
		logrus.Warnf("Variable not found when releasing member path: %s", varName)
		return errors.WrapE(err, "variable-get",
			"pool", sa.name, "path", key)
	}
	// 删除目录中数据
	if err := os.RemoveAll(memberPath); err != nil {
		logrus.Warnf("Failed to remove directory %s: %v", memberPath, err)
	}

	// 对应信号量增加capacityGB
	semaName := fmt.Sprintf("path-free-gb:%s:%s", sa.name, memberPath)
	if _, err := semaphore.AddValue(semaName, 0, sa.appID, capacityGB); err != nil {
		return errors.WrapE(err, "semaphore-add", "pool", sa.name)
	}
	// 删除variable
	err = variable.Set(varName, "", 0, sa.appID)
	return errors.WrapE(err, "variable-set", "var-name", varName)
}

// Name 返回聚合器名称
func (sa *ScaleboxAggregator) Name() string {
	return sa.name
}

// Stats 返回统计信息（待实现）
func (sa *ScaleboxAggregator) Stats() map[string]interface{} {
	return map[string]interface{}{
		"name":   sa.name,
		"type":   "scalebox_aggregator",
		"status": "not_implemented",
	}
}
