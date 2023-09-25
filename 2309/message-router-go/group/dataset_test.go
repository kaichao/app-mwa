package group

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// 水平分组
func TestAddEntity_Horizonal(t *testing.T) {
	ds := NewDataSet("mwa-dat:1257010784", "H", 3, 4600, 1, 60)

	assert.Equal(t, false, ds.AddEntity("file1", 1, 20))
	assert.Equal(t, false, ds.AddEntity("file2", 2, 20))
	assert.Equal(t, true, ds.AddEntity("file3", 3, 20))
}

// 垂直分组
func TestAddEntity_Vertical(t *testing.T) {
	ds := NewDataSet("mwa-dat:1257010784", "V", 10, 8, 1257010781, 3)

	assert.Equal(t, false, ds.AddEntity("file1", 1, 1257010783))
	assert.Equal(t, false, ds.AddEntity("file2", 1, 1257010784))
	assert.Equal(t, false, ds.AddEntity("file3", 1, 1257010785))
	assert.Equal(t, true, ds.AddEntity("file4", 1, 1257010786))
	assert.Equal(t, false, ds.AddEntity("file5", 1, 1257010782))
	assert.Equal(t, true, ds.AddEntity("file6", 1, 1257010781))
	assert.Equal(t, false, ds.AddEntity("file7", 1, 1257010788))
	assert.Equal(t, true, ds.AddEntity("file8", 1, 1257010787))
	// assert.Equal(t, true, ds.AddEntity("file9", 1, 1257010789))
}

func TestNewDataSet(t *testing.T) {
	NewDataSet("mwa-dat:1257010784", "V", 24, 4600, 1, 60)
	ds := mapDataset["mwa-dat:1257010784"]
	if ds.groupType != "V" || ds.horizontalSize != 24 || ds.verticalSize != 4600 || ds.groupLength != 60 {
		t.Errorf("dataset assertion.")
	}
}
