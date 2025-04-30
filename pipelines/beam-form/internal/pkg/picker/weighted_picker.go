package picker

import (
	"encoding/json"
	"fmt"
	"os"
)

type WeightedPicker struct {
	TheoryPercent map[string]float64
	HistoryCounts map[string]int
	TotalCount    int
}

func NewWeightedPicker(obj string) *WeightedPicker {
	jsonFile := fmt.Sprintf("%s/%s-%s.json",
		os.Getenv("DIR_PREFIX"), os.Getenv("CLUSTER"), obj)
	fmt.Println("json-file:", jsonFile)
	data, _ := os.ReadFile(jsonFile)
	weights := map[string]float64{}
	json.Unmarshal(data, &weights)

	// 删除value为0的项
	for key, value := range weights {
		if value == 0 {
			delete(weights, key)
		}
	}

	theoryPercent := make(map[string]float64)
	historyCounts := make(map[string]int)
	totalWeight := 0.0
	for _, w := range weights {
		totalWeight += w
	}
	for key, w := range weights {
		theoryPercent[key] = w / totalWeight
		historyCounts[key] = 0
	}
	wp := &WeightedPicker{
		TheoryPercent: theoryPercent,
		HistoryCounts: historyCounts,
		TotalCount:    0,
	}
	fmt.Println("wp", *wp)
	return wp
}

// GetNext 返回一个根据权重随机选择的字符串
func (wp *WeightedPicker) GetNext() string {
	var firstKey string
	for key := range wp.TheoryPercent {
		firstKey = key
		if wp.TotalCount == 0 || float64(wp.HistoryCounts[key])/float64(wp.TotalCount) < wp.TheoryPercent[key] {
			wp.HistoryCounts[key]++
			wp.TotalCount++
			return key
		}
	}
	wp.HistoryCounts[firstKey]++
	wp.TotalCount++
	return firstKey
}
