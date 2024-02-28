package main

import (
	"fmt"
	"regexp"
	"strconv"
)

// 性能测试专用，绕过将全局存储拷贝到本机存储的unpack模块
func fromDirListTest(message string, headers map[string]string) int {
	fmt.Println("input-message:", message)

	re := regexp.MustCompile(".+~([0-9]+)/([0-9]+)_([0-9]+)_ch([0-9]{3}).+")
	ss := re.FindStringSubmatch(message)
	ds := ss[1]
	begin := ss[2]
	end := ss[3]
	ch := ss[4]
	fmt.Println("channel:", ch)
	if ch > "110" {
		return 0
	}

	b, _ := strconv.Atoi(begin)
	e, _ := strconv.Atoi(end)
	// if e > 1257011025 {
	// 	// 240秒
	// }
	sinkJob := "data-grouping-main"
	// channel := 109
	// for i := b; i <= e; i++ {
	// 	m := fmt.Sprintf("dat,%s/%s_%d_ch%d.dat", ds, ds, i, channel)
	// 	sendChannelAwareMessage(m, sinkJob, channel)
	// }
	channel := 110
	for i := b; i <= e; i++ {
		m := fmt.Sprintf("dat,%s/%s_%d_ch%d.dat", ds, ds, i, channel)
		sendNodeAwareMessage(m, make(map[string]string), sinkJob, channel-109)
	}
	return 0
}

func fromBeamMakerTest(message string, headers map[string]string) int {
	// 1257010784/1257010786_1257010795/00001/ch123.fits

	return 0
}
