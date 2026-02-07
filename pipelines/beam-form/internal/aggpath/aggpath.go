package aggpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kaichao/gopkg/errors"
	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semagroup"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/variable"
	"github.com/sirupsen/logrus"
)

// AggregatedPath 汇聚目录
type AggregatedPath struct {
	Name        string
	AppID       int
	CategoryMap map[string]int
}

//	semaphores : path-free-gb:${my-ap}, num-gb
//	variables: : member-path:${category-name}:${path}

// New 创建AggregatedPath
func New(appID int, confFile string) (*AggregatedPath, error) {
	if confFile == "" {
		return nil, errors.E("null config file")
	}

	// 从配置文件名提取Name（不含扩展名）
	// 提取文件名（不含路径和扩展名）
	baseName := filepath.Base(confFile)
	name := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	lines, err := common.GetTextFileLines(confFile)
	if err != nil {
		return nil, errors.WrapE(err, "read config file", "file-name", confFile)
	}

	// confFile文件，每行为：分类名:容量(GB)，加载到CategoryMap中
	categoryMap := make(map[string]int)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var category string
		var delta int
		n, err := fmt.Sscanf(line, `%q:%d`, &category, &delta)
		if err == nil && n == 2 {
			categoryMap[category] = delta
		}
	}

	return &AggregatedPath{
		Name:        name,
		AppID:       appID,
		CategoryMap: categoryMap,
	}, nil
}

// GetMemberPath 获取当前成员目录
func (ap *AggregatedPath) GetMemberPath(category, path string) (string, error) {
	delta, ok := ap.CategoryMap[category]
	if !ok {
		return "", errors.E("no valid category", "category-name", category)
	}

	// if path对应variable存在，直接返回该目录
	varName := fmt.Sprintf("member-path:%s:%s", category, path)
	varValue, err := variable.GetValue(varName, 0, ap.AppID)
	if err == nil && varValue != "" {
		return varValue, nil
	}

	// 通过category获取semagroup的最大值，若该值超过capacityDB值，减去该值
	semaGName := "path-free-gb:" + ap.Name
	semaName, semaValue, err := semagroup.GetMax(semaGName, ap.AppID)
	if err != nil {
		return "", errors.WrapE(err, "semagroup-getmax",
			"semagroup-name", semaGName, "app-id", ap.AppID)
	}
	if semaValue < delta {
		return "", errors.E("No enough disk space", "category", category)
	}
	_, err = semaphore.AddValue(semaName, 0, ap.AppID, -delta)
	if err != nil {
		return "", errors.WrapE(err, "semaphore-addvalue",
			"sema-name", semaName, "app-id", ap.AppID)
	}
	// 以前述的信号量名，创建对应的variable
	// 去掉前面的path-free-gb: 及 name
	varValue = strings.Split(semaName, ":")[2]
	err = variable.Set(varName, varValue, 0, ap.AppID)
	if err != nil {
		return "", errors.WrapE(err, "variable-set",
			"var-name", varName, "app-id", ap.AppID)
	}
	// 返回前面的目录
	return varValue, nil
}

// ReleaseMemberPath ...
func (ap *AggregatedPath) ReleaseMemberPath(category, path string) error {
	delta, ok := ap.CategoryMap[category]
	if !ok {
		return errors.E("no valid category", "category", category)
	}

	// 获取对应的variable值（即成员目录）
	varName := fmt.Sprintf("member-path:%s:%s", category, path)
	memberPath, err := variable.GetValue(varName, 0, ap.AppID)
	if err != nil {
		// 如果variable不存在，可能已经被释放
		logrus.Warnf("Variable not found when releasing member path: %s", varName)
		return errors.WrapE(err, "variable-get",
			"category", category, "path", path)
	}
	// 删除目录中数据
	if err := os.RemoveAll(memberPath); err != nil {
		logrus.Warnf("Failed to remove directory %s: %v", memberPath, err)
	}

	// 对应信号量增加capacityGB
	semaName := fmt.Sprintf("path-free-gb:%s:%s", ap.Name, memberPath)
	if _, err := semaphore.AddValue(semaName, 0, ap.AppID, delta); err != nil {
		return errors.WrapE(err, "semaphore-add", "category", category)
	}
	// 删除variable
	err = variable.Set(varName, "", 0, ap.AppID)
	return errors.WrapE(err, "variable-set", "var-name", varName)
}
