package picker

import (
	"fmt"
	"os"
	"testing"
)

func TestPicker(t *testing.T) {
	os.Setenv("CLUSTER", "p419")
	// 用于测试的环境变量
	os.Setenv("DIR_PREFIX", "../../../app")
	picker := NewWeightedPicker("target")
	for i := 0; i < 100; i++ {
		fmt.Println(picker.GetNext())
	}
}
