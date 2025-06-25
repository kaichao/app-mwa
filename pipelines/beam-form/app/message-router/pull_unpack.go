package main

import (
	"beamform/internal/datacube"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/kaichao/scalebox/pkg/common"
	"github.com/kaichao/scalebox/pkg/semaphore"
	"github.com/kaichao/scalebox/pkg/task"
	"github.com/sirupsen/logrus"
)

func fromPullUnpack(msg string, headers map[string]string) int {
	defer func() {
		common.AddTimeStamp("leave-fromPullUnpack()")
	}()
	common.AddTimeStamp("enter-fromPullUnpack()")
	// input message: 1257617424/p00001_00096/1257617426_1257617465_ch112.dat.tar.zst
	// - target_dir:1257617424/t1257617426_1257617505/ch111
	// semaphore: dat-ready:1257010784/p00001_00960/t1257010786_1257010985/ch109
	re := regexp.MustCompile(`^(([0-9]+)/p([0-9]+)_([0-9]+))/([0-9]+)_[0-9]+_(ch[0-9]+).dat.tar.zst$`)
	ss := re.FindStringSubmatch(msg)
	if len(ss) == 0 {
		logrus.Errorf("Invalid Message Format, body=%s\n", msg)
		return 1
	}
	prefix := ss[1]
	obsID := ss[2]
	p0, _ := strconv.Atoi(ss[3])
	p1, _ := strconv.Atoi(ss[4])
	t, _ := strconv.Atoi(ss[5])
	ch := ss[6]
	cube := datacube.GetDataCube(obsID)
	t0, t1 := cube.GetTimeRange(t)

	sema := fmt.Sprintf(`dat-ready:%s/t%d_%d/%s`, prefix, t0, t1, ch)
	v, err := semaphore.AddValue(sema, appID, -1)
	if err != nil {
		logrus.Errorf("semaphore-decrement, sema=%s\n", sema)
		return 2
	}
	semaVal, _ := strconv.Atoi(v)
	if semaVal > 0 {
		return 0
	}
	common.AddTimeStamp("prepare-messages")
	ps := cube.GetPointingRangesByInterval(p0, p1)
	messages := []string{}
	for k := 0; k < len(ps); k += 2 {
		body := fmt.Sprintf(`%s/p%05d_%05d/t%d_%d/%s`,
			obsID, ps[k], ps[k+1], t0, t1, ch)
		// 加上排序标签
		if os.Getenv("POINTING_FIRST") == "yes" {
			body = fmt.Sprintf(`%s,{"sort_tag":"p%05d:t%d"}`,
				body, ps[k], t0)
		}
		messages = append(messages, body)
	}
	fmt.Printf("num-of-messages in fromPullUnpack():%d\n", len(messages))
	common.AddTimeStamp("before-send-messages")
	envVars := map[string]string{
		"SINK_JOB":        "beam-make",
		"TIMEOUT_SECONDS": "600",
	}
	return task.AddTasks(messages, "{}", envVars)
}
