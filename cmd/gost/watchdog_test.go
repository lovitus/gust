package main

import (
	"testing"
	"time"
)

func TestNextWatchdogBackoffCapsAtMaximum(t *testing.T) {
	maximum := 60 * time.Second
	tests := []struct {
		current time.Duration
		want    time.Duration
	}{
		{time.Second, 2 * time.Second},
		{32 * time.Second, maximum},
		{maximum, maximum},
	}
	for _, test := range tests {
		if got := nextWatchdogBackoff(test.current, maximum); got != test.want {
			t.Fatalf("nextWatchdogBackoff(%v)=%v, want %v", test.current, got, test.want)
		}
	}
}
