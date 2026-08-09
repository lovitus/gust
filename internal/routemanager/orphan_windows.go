//go:build windows

package routemanager

import (
	"errors"
	"fmt"
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

func orphanCleanupAction(candidate *process.Process) string {
	return fmt.Sprintf("TerminateProcess(PID=%d)，先递归终止其子进程", candidate.Pid)
}

func cleanupOrphanProcess(candidate *process.Process) error {
	children, _ := candidate.Children()
	var errs []error
	for _, child := range children {
		if err := cleanupOrphanProcess(child); err != nil && !errors.Is(err, os.ErrProcessDone) {
			errs = append(errs, err)
		}
	}
	if err := candidate.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}
