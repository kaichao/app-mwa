package message

import (
	"fmt"
	"os"
	"testing"
)

func TestProcessMessage(t *testing.T) {
	os.Setenv("DATACUBE_FILE", "../../../dataset.yaml")
	// m := "1257010784/p00001_00960/t1257012766_1257012965"
	m := "1257010784/p00001_00960"
	messages, _ := ParseForPullUnpack(m)
	fmt.Println(messages)
	fmt.Println(len(messages))
}
