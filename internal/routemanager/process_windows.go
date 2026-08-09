//go:build windows

package routemanager

import (
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func prepareProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

type windowsProcessControl struct {
	mu  sync.Mutex
	job windows.Handle
}

func ownProcess(cmd *exec.Cmd) (processControl, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	info.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	_, err = windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uint32(unsafe.Sizeof(info)),
	)
	if err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	process, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	defer windows.CloseHandle(process)
	if err := windows.AssignProcessToJobObject(job, process); err != nil {
		_ = windows.CloseHandle(job)
		return nil, err
	}
	return &windowsProcessControl{job: job}, nil
}

func (c *windowsProcessControl) stop() error { return c.terminate() }

func (c *windowsProcessControl) kill() error { return c.terminate() }

func (c *windowsProcessControl) terminate() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.job == 0 {
		return os.ErrProcessDone
	}
	return windows.TerminateJobObject(c.job, 1)
}

func (c *windowsProcessControl) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.job == 0 {
		return nil
	}
	err := windows.CloseHandle(c.job)
	c.job = 0
	return err
}

func isProcessDone(err error) bool {
	return errors.Is(err, os.ErrProcessDone) || errors.Is(err, windows.ERROR_INVALID_HANDLE)
}
