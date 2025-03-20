package semaphore

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kaichao/scalebox/pkg/misc"
	"github.com/sirupsen/logrus"
)

// Decrement ...
func Decrement(sema string) (int, error) {
	cmd := "scalebox semaphore decrement " + sema
	code, stdout, stderr := misc.ExecCommandReturnAll(cmd, 20)
	logrus.Errorf("stcerr:\n%s\n", stderr)
	fmt.Printf("stdout:\n%s\n", stdout)
	if code > 0 {
		return code, errors.New("Exec semaphore-decrement")
	}
	v, err := strconv.Atoi(strings.TrimSpace(stdout))
	if err != nil {
		logrus.Errorf("semaphore-value not a integer, value=%s\n", stdout)
		return -1, err
	}
	return v, nil
}

// DecrementExpr ...
func DecrementExpr(semExpr string) (map[string]int, error) {
	return map[string]int{}, nil
}
