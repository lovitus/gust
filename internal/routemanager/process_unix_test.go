//go:build !windows

package routemanager

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestUnixProcessControlStopsOwnedGroup(t *testing.T) {
	cmd := exec.Command("sh", "-c", "trap 'exit 0' INT; sleep 30 & wait")
	prepareProcess(cmd)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	control, err := ownProcess(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		t.Fatal(err)
	}
	finished := false
	defer func() {
		if !finished {
			_ = control.kill()
		}
	}()

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	if err := control.stop(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = control.kill()
		<-done
		t.Fatal("process group did not stop after SIGINT")
	}
	finished = true

	if err := syscall.Kill(-cmd.Process.Pid, 0); err != syscall.ESRCH {
		t.Fatalf("owned process group still exists: %v", err)
	}
}
