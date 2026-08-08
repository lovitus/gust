//go:build !windows

package routemanager

import (
	"os"
	"os/exec"
)

func prepareProcess(cmd *exec.Cmd) {}

func stopProcess(process *os.Process) error {
	return process.Signal(os.Interrupt)
}
