package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"time"
)

var errInstanceRunning = errors.New("route manager instance already running")

type singleInstance struct {
	lock       io.Closer
	activation *activationServer
}

func (i *singleInstance) Close() error {
	return errors.Join(i.activation.Close(), i.lock.Close())
}

func (i *singleInstance) Activations() <-chan struct{} {
	return i.activation.events
}

func instanceKey(configPath string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(configPath)))
}

func acquireSingleInstance(configPath string, wait time.Duration) (*singleInstance, error) {
	key := instanceKey(configPath)
	deadline := time.Now().Add(wait)
	for {
		lock, acquired, err := tryAcquireInstanceLock(key)
		if err != nil {
			return nil, err
		}
		if acquired {
			activation, err := startActivationServer(key)
			if err != nil {
				_ = lock.Close()
				return nil, err
			}
			return &singleInstance{lock: lock, activation: activation}, nil
		}
		if wait <= 0 || time.Now().After(deadline) {
			return nil, errInstanceRunning
		}
		time.Sleep(100 * time.Millisecond)
	}
}
