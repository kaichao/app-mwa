package task

import (
	"fmt"

	"github.com/kaichao/scalebox/pkg/misc"
)

// Add ...
func Add(sinkJob string, message string, headers string) int {
	if headers == "" {
		headers = "{}"
	}
	cmd := fmt.Sprintf(`scalebox task add --sink-job=%s --headers='%s' %s`,
		sinkJob, headers, message)
	code := misc.ExecCommandReturnExitCode(cmd, 15)
	return code
}

// AddWithMapHeaders ...
func AddWithMapHeaders(sinkJob string, message string, headers map[string]string) int {
	return 0
}

// AddTasks ...
func AddTasks(sinkJob string, messages []string, headers string) int {
	return 0
}
