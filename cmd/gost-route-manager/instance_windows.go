//go:build windows

package main

import (
	"errors"
	"fmt"
	"io"

	"golang.org/x/sys/windows"
)

type windowsInstanceLock struct {
	handle windows.Handle
}

func (l *windowsInstanceLock) Close() error {
	return windows.CloseHandle(l.handle)
}

func tryAcquireInstanceLock(key string) (io.Closer, bool, error) {
	name, err := windows.UTF16PtrFromString("Local\\GustRouteManager-" + key)
	if err != nil {
		return nil, false, err
	}
	handle, err := windows.CreateMutex(nil, false, name)
	if errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if handle != 0 {
			_ = windows.CloseHandle(handle)
		}
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("创建单实例互斥体: %w", err)
	}
	return &windowsInstanceLock{handle: handle}, true, nil
}
