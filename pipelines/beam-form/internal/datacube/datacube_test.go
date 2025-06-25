package datacube_test

import (
	"beamform/internal/datacube"
	"fmt"
	"testing"
)

func TestGetDataCubeFromFile(t *testing.T) {
	cube := datacube.GetDataCubeFromFile("1257010784")
	fmt.Println(cube)
}

func getMyDataCube() *datacube.DataCube {
	datacube := &datacube.DataCube{
		DatasetID:     "1257010784",
		ChannelBegin:  109,
		NumOfChannels: 24,

		TimeBegin:    1257010786,
		NumOfSeconds: 4798,
		TimeUnit:     30,
		TimeStep:     150,

		PointingBegin: 1,
		PointingEnd:   12972,
		PointingStep:  24,
	}

	return datacube
}
