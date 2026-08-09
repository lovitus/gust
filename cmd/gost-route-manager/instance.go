package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"
)

var errInstanceRunning = errors.New("route manager instance already running")

func acquireSingleInstance(configPath string, wait time.Duration) (io.Closer, error) {
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(configPath)))
	deadline := time.Now().Add(wait)
	for {
		lock, acquired, err := tryAcquireInstanceLock(key)
		if err != nil {
			return nil, err
		}
		if acquired {
			return lock, nil
		}
		if wait <= 0 || time.Now().After(deadline) {
			return nil, errInstanceRunning
		}
		time.Sleep(100 * time.Millisecond)
	}
}
