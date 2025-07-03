package datacube_test

import (
	"beamform/internal/datacube"
	"fmt"
	"testing"
)

func TestNewDataCube(t *testing.T) {
	cube := datacube.NewDataCube("1265983624")
	fmt.Println(cube.ToCubeString())

	cube = datacube.NewDataCube("1265983624/p00001_00960")
	fmt.Println(cube.ToCubeString())

	cube = datacube.NewDataCube("1265983624/p_00960")
	fmt.Println(cube.ToCubeString())

	cube = datacube.NewDataCube("1265983624/p00961_")
	fmt.Println(cube.ToCubeString())

	cube = datacube.NewDataCube("1265983624/p00001_00960/t1265983626_1265988429")
	fmt.Println(cube.ToCubeString())

	cube = datacube.NewDataCube("1265983624/p00001_00960/t1265983826_")
	fmt.Println(cube.ToCubeString())

	cube = datacube.NewDataCube("1265983624/p00001_00960/t_1265983825")
	fmt.Println(cube.ToCubeString())
}

func TestGetDataCubeFromFile(t *testing.T) {
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
