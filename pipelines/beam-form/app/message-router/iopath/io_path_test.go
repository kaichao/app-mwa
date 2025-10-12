package iopath_test

import (
	"beamform/app/message-router/iopath"
	"fmt"
	"os"
	"testing"
)

func TestGetPreloadRoot(t *testing.T) {
	os.Setenv("IOPATH_FILE", "../../io-path.yaml")

	for p := 0; p < 100; p++ {
		fmt.Println(iopath.GetPreloadRoot(p))
	}
}
