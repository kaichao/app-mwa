package main

import (
	"fmt"
	"testing"
)

func TestGetTimeRanges(t *testing.T) {
	dataset := &DataSet{
		DatasetID:      "1257010784",
		VerticalStart:  1257010786,
		VerticalHeight: 47,
	}
	fmt.Println(dataset.getTimeRanges())
}

func TestGetTimeRange(t *testing.T) {
	dataset := &DataSet{
		DatasetID:      "1257010784",
		VerticalStart:  1257010786,
		VerticalHeight: 45,
	}

	fmt.Println(dataset.getTimeRange(1257010786))

	fmt.Println(dataset.getTimeRange(1257010815))

	fmt.Println(dataset.getTimeRange(1257010792))

	fmt.Println(dataset.getTimeRange(1257010816))

	fmt.Println(dataset.getTimeRange(1257010833))

	fmt.Println(dataset.getTimeRange(1257010784))
}

// func TestParseDataSet(t *testing.T) {
// 	// s := `{"datasetID":"1257010784","verticalStart":"1257010786","verticalHeight":"60","horizontalWidth":"24"}`
// 	s := `{"dataset_id":"1257010784","vertical_start":"1257010786","vertical_height":"60","horizontal_width":"24"}`
// 	ds := parseDataSetX(s)
// 	fmt.Println(ds)
// }

func TestGetDataSet(t *testing.T) {
	datasetID := "1257010784"
	ds := getDataSet(datasetID)
	fmt.Println(ds)
}
