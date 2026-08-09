//go:build !windows

package routemanager

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

func cleanupOrphanProcess(candidate *process.Process) error {
	pid := int(candidate.Pid)
	pgid, err := syscall.Getpgid(pid)
	if err == nil && pgid == pid {
		err = syscall.Kill(-pid, syscall.SIGINT)
	} else {
		err = candidate.SendSignal(syscall.SIGINT)
	}
	if err != nil && !errors.Is(err, os.ErrProcessDone) && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	if waitForProcessInstance(candidate, 2*time.Second) {
		if pgid == pid {
			return killOrphanGroup(pid)
		}
		return nil
	}
	if pgid == pid {
		return killOrphanGroup(pid)
	} else {
		err = candidate.Kill()
	}
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func orphanCleanupAction(candidate *process.Process) string {
	pid := int(candidate.Pid)
	pgid, err := syscall.Getpgid(pid)
	if err == nil && pgid == pid {
		return fmt.Sprintf("kill(PGID=%d, SIGINT); 2 秒超时后 kill(PGID=%d, SIGKILL)", pgid, pgid)
	}
	return fmt.Sprintf("kill(PID=%d, SIGINT); 2 秒超时后 kill(PID=%d, SIGKILL)", pid, pid)
}

func killOrphanGroup(pid int) error {
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}

func waitForProcessInstance(candidate *process.Process, timeout time.Duration) bool {
	startedAt, err := candidate.CreateTime()
	if err != nil {
		return true
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		current, err := process.NewProcess(candidate.Pid)
		if err != nil {
			return true
		}
		currentStartedAt, err := current.CreateTime()
		if err != nil || currentStartedAt != startedAt {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}
