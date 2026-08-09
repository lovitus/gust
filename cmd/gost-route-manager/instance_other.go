//go:build !darwin && !linux && !windows

package main

import (
	"io"
)

type noopInstanceLock struct{}

func (noopInstanceLock) Close() error { return nil }

func tryAcquireInstanceLock(string) (io.Closer, bool, error) {
	return noopInstanceLock{}, true, nil
}
