package aggpath

import (
	"fmt"
	"os"
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
	AppID           int
	StorageGroupMap map[string]int
}

//	semaphores : path-free-gb:${storage_group_name}, num-gb
//	variables: : member-path:${path_name}

// New 创建AggregatedPath
func New(appID int, storageGroupFile string) (*AggregatedPath, error) {
	// storageGroupFile文件，每行为：存储组名:容量(GB)，加载到StorageGroupMap中
	storageGroupMap := make(map[string]int)

	if storageGroupFile == "" {
		return &AggregatedPath{
			AppID:           appID,
			StorageGroupMap: storageGroupMap,
		}, nil
	}

	lines, err := common.GetTextFileLines(storageGroupFile)
	if err != nil {
		return nil, errors.WrapE(err, "read storage group file", "file", storageGroupFile)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var storageGroup string
		var delta int
		n, err := fmt.Sscanf(line, `%q:%d`, &storageGroup, &delta)
		if err == nil && n == 2 {
			storageGroupMap[storageGroup] = delta
		}

	}

	return &AggregatedPath{
		AppID:           appID,
		StorageGroupMap: storageGroupMap,
	}, nil
}

// GetMemberPath 获取当前成员目录
func (ap *AggregatedPath) GetMemberPath(storageGroup, path string) (string, error) {
	delta, ok := ap.StorageGroupMap[storageGroup]
	if !ok {
		return "", errors.E("no valid storage group", "storageGroup", storageGroup)
	}

	// if path对应variable存在，直接返回该目录
	varName := fmt.Sprintf("member-path:%s:%s", storageGroup, path)
	varValue, err := variable.GetValue(varName, 0, ap.AppID)
	if err == nil && varValue != "" {
		return varValue, nil
	}

	// 通过storageGroup获取semagroup的最大值，若该值超过capacityDB值，减去该值
	semaGName := "path-free-gb:" + storageGroup
	fmt.Println("stor-group:", semaGName)
	semaName, semaValue, err := semagroup.GetMax(semaGName, ap.AppID)
	if err != nil {
		return "", errors.WrapE(err, "semagroup-getmax",
			"storageGroup", storageGroup, "app-id", ap.AppID)
	}
	if semaValue < delta {
		return "", errors.E("No enough disk space", "storageGroup", storageGroup)
	}
	_, err = semaphore.AddValue(semaName, 0, ap.AppID, -delta)
	if err != nil {
		return "", errors.WrapE(err, "semaphore-addvalue",
			"sema-name", semaName, "app-id", ap.AppID)
	}
	// 以前述的信号量名，创建对应的variable
	// 去掉前面的path-free-gb:
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
func (ap *AggregatedPath) ReleaseMemberPath(storageGroup, path string) error {
	delta, ok := ap.StorageGroupMap[storageGroup]
	if !ok {
		return errors.E("no valid storage group", "storageGroup", storageGroup)
	}

	// 获取对应的variable值（即成员目录）
	varName := fmt.Sprintf("member-path:%s:%s", storageGroup, path)
	memberPath, err := variable.GetValue(varName, 0, ap.AppID)
	if err != nil {
		// 如果variable不存在，可能已经被释放
		logrus.Warnf("Variable not found when releasing member path: %s", varName)
		return errors.WrapE(err, "variable-get",
			"storage-group", storageGroup, "path", path)
	}
	// 删除目录中数据
	if err := os.RemoveAll(memberPath); err != nil {
		logrus.Warnf("Failed to remove directory %s: %v", memberPath, err)
	}

	// 对应信号量增加capacityGB
	semaName := fmt.Sprintf("path-free-gb:%s:%s", storageGroup, memberPath)
	if _, err := semaphore.AddValue(semaName, 0, ap.AppID, delta); err != nil {
		return errors.WrapE(err, "semaphore-add", "storageGroup", storageGroup)
	}
	// 删除variable
	err = variable.Set(varName, "", 0, ap.AppID)
	return errors.WrapE(err, "variable-set", "var-name", varName)
}
