package iopath_test

import (
	"beamform/app/message-router/iopath"
	"fmt"
	"os"
	"testing"
)

func TestGetPreloadRoot(t *testing.T) {
	os.Setenv("IOPATH_FILE", "../../io-path.yaml")

	for p := 2881; p < 3001; p++ {
		fmt.Println(iopath.GetStagingRoot(-1))
	}
}
