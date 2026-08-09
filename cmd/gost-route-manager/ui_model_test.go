package main

import (
	"testing"
	"time"

	"github.com/go-gost/gost/internal/routemanager"
)

func TestNewTunnelUsesPlaceholdersInsteadOfPrefilledValues(t *testing.T) {
	tunnel := newTunnel("tunnel-test")
	if tunnel.ID != "tunnel-test" {
		t.Fatalf("ID = %q", tunnel.ID)
	}
	if tunnel.Name != "" || tunnel.Routes != "" || tunnel.Target != "" || tunnel.Mode != "" || tunnel.Args != "" {
		t.Fatalf("new tunnel must be empty so placeholders remain placeholders: %+v", tunnel)
	}
}

func TestRestartBackoffIsBounded(t *testing.T) {
	want := []time.Duration{time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second, 16 * time.Second, 16 * time.Second}
	for i, expected := range want {
		if got := restartBackoff(i + 1); got != expected {
			t.Fatalf("attempt %d: got %s, want %s", i+1, got, expected)
		}
	}
}

func TestFormatLogsUsesNamesForGlobalView(t *testing.T) {
	when := time.Date(2026, 8, 9, 12, 34, 56, 0, time.Local)
	lines := []routemanager.LogLine{{Time: when, TunnelID: "id-1", Text: "connected"}}
	if got, want := formatLogs(lines, map[string]string{"id-1": "zwy"}), "[12:34:56] [zwy] connected"; got != want {
		t.Fatalf("formatLogs() = %q, want %q", got, want)
	}
}
