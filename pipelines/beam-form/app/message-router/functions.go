package main

import (
	"regexp"

	"github.com/sirupsen/logrus"
)

func fromFitsPush(m string, headers map[string]string) int {
	// mwa/24ch/1257617424/p00021/t1257617426_1257617505.fits.zst
	re := regexp.MustCompile(`^mwa/24ch/([0-9]+/p[0-9]+)/t[0-9]+_[0-9]+`)
	ss := re.FindStringSubmatch(m)
	if ss == nil {
		logrus.Errorf("Invalid format, message:%s\n", m)
		return 1
	}
	return toCrossAppPresto(ss[1])
}
