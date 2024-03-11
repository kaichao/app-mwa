package main

import (
	"fmt"
	"testing"
)

func TestGetTimeRanges(t *testing.T) {
	datacube := &DataCube{
		DatasetID:  "1257010784",
		TimeBegin:  1257010786,
		TimeLength: 47,
	}
	fmt.Println(datacube.getTimeRanges())
}

func TestGetTimeRange(t *testing.T) {
	datacube := &DataCube{
		DatasetID:  "1257010784",
		TimeBegin:  1257010786,
		TimeLength: 45,
	}

	fmt.Println(datacube.getTimeRange(1257010786))

	fmt.Println(datacube.getTimeRange(1257010815))

	fmt.Println(datacube.getTimeRange(1257010792))

	fmt.Println(datacube.getTimeRange(1257010816))

	fmt.Println(datacube.getTimeRange(1257010833))

	fmt.Println(datacube.getTimeRange(1257010784))
}

// func TestParseDataSet(t *testing.T) {
// 	// s := `{"datasetID":"1257010784","verticalStart":"1257010786","verticalHeight":"60","horizontalWidth":"24"}`
// 	s := `{"dataset_id":"1257010784","vertical_start":"1257010786","vertical_height":"60","horizontal_width":"24"}`
// 	ds := parseDataSetX(s)
// 	fmt.Println(ds)
// }

func TestGetDataCube(t *testing.T) {
	datasetID := "1257010784"
	ds := getDataCube(datasetID)
	fmt.Println(ds)
}
