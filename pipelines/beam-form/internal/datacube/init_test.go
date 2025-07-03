package datacube_test

import (
	"beamform/internal/datacube"
	"os"
)

func init() {
	os.Setenv("DATACUBE_FILE", "../../dataset.yaml")
	cube = datacube.NewDataCube("1257010784")
}

var cube *datacube.DataCube
