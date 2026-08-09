//go:build !windows

package routemanager

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

func prepareProcess(cmd *exec.Cmd) {
	// Each tunnel is the leader of a private process group. Signals sent to the
	// negative PID therefore reach only this tunnel and any helpers it spawned.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

type unixProcessControl struct {
	pid int
}

func ownProcess(cmd *exec.Cmd) (processControl, error) {
	return &unixProcessControl{pid: cmd.Process.Pid}, nil
}

func (c *unixProcessControl) stop() error {
	return syscall.Kill(-c.pid, syscall.SIGINT)
}

func (c *unixProcessControl) kill() error {
	return syscall.Kill(-c.pid, syscall.SIGKILL)
}

func (c *unixProcessControl) close() error {
	// The leader may exit before a helper it spawned. Reap any remaining member
	// before ProcessManager forgets this owned group.
	err := syscall.Kill(-c.pid, syscall.SIGKILL)
	if isProcessDone(err) {
		return nil
	}
	return err
}

func isProcessDone(err error) bool {
	return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
}
