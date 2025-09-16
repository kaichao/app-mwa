package iopath_test

import (
	"beamform/app/message-router/iopath"
	"fmt"
	"os"
	"testing"
)

func TestGetStagingRoot(t *testing.T) {
	os.Setenv("IOPATH_FILE", "../../io-path.yaml")

	for p := 0; p < 10; p++ {
		fmt.Println(iopath.GetStagingRoot(p))
	}
}
