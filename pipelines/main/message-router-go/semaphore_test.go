package main

import (
	"fmt"
	"testing"
)

func TestGetRange(t *testing.T) {
	dataset := &DataSet{
		DatasetID:      "1257010784",
		VerticalStart:  1257010786,
		VerticalHeight: 47,
	}
	fmt.Println(getTimeRange(dataset))
}
