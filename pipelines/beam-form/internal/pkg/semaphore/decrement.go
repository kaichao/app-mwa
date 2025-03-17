package semaphore

import (
	"errors"
	"strconv"

	"github.com/kaichao/scalebox/pkg/misc"
)

// Decrement ...
func Decrement(sema string) (int, error) {
	cmd := "scalebox semaphore decrement " + sema
	s := misc.ExecCommandReturnStdout(cmd, 5)
	if s == "-32768" {
		// error while decrement semaphore
		return 0, errors.New("semaphore-decrement")
	}
	v, err := strconv.Atoi(s)
	if err == nil {
		return v, nil
	}
	return 0, err
}

// DecrementExpr ...
func DecrementExpr(semExpr string) (map[string]int, error) {
	return map[string]int{}, nil
}
