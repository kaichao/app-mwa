package iopath_test

import (
	"os"
	"testing"
)

func TestGetPreloadRoot(t *testing.T) {
	os.Setenv("IOPATH_FILE", "../../io-path.yaml")
}
