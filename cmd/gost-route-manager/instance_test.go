//go:build darwin || linux

package main

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestSingleInstanceRejectsDuplicateAndReleases(t *testing.T) {
	config := filepath.Join(t.TempDir(), "route-manager.json")
	first, err := acquireSingleInstance(config, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := acquireSingleInstance(config, 0); !errors.Is(err, errInstanceRunning) {
		t.Fatalf("duplicate error = %v, want %v", err, errInstanceRunning)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := acquireSingleInstance(config, 0)
	if err != nil {
		t.Fatalf("reacquire after close: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestSingleInstanceWaitsForElevationHandoff(t *testing.T) {
	config := filepath.Join(t.TempDir(), "handoff.json")
	first, err := acquireSingleInstance(config, 0)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(150 * time.Millisecond)
		_ = first.Close()
	}()
	second, err := acquireSingleInstance(config, 2*time.Second)
	if err != nil {
		t.Fatalf("handoff lock: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}
