//go:build windows

package routemanager

import (
	"os"
	"os/exec"
	"syscall"
)

func prepareProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func stopProcess(process *os.Process) error {
	return process.Kill()
}
