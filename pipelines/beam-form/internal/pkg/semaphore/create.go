package semaphore

import (
	"errors"
	"os"

	"github.com/kaichao/scalebox/pkg/misc"
)

// Create
func Create(semaLines string) error {
	misc.AppendToFile("my-sema.txt", semaLines)
	defer os.Remove("my-sema.txt")

	cmd := "scalebox semaphore create --sema-file my-sema.txt"
	if code := misc.ExecCommandReturnExitCode(cmd, 600); code > 0 {
		return errors.New("semaphore-create")
	}
	return nil
}
