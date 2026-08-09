package main

import (
	"strings"
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

func TestFormatOrphanPreviewShowsIdentityAndCommand(t *testing.T) {
	preview := formatOrphanPreview([]routemanager.OrphanProcess{{
		PID: 42, StartedAt: time.Date(2026, 8, 9, 12, 34, 56, 0, time.Local).UnixMilli(),
		Executable: "/opt/gust/gost-qt", CommandLine: "gost-qt -L tcp://:8080",
		CleanupAction: "kill(PGID=42, SIGINT)",
	}})
	for _, want := range []string{"PID: 42", "2026-08-09 12:34:56", "/opt/gust/gost-qt", "gost-qt -L tcp://:8080", "kill(PGID=42, SIGINT)"} {
		if !strings.Contains(preview, want) {
			t.Fatalf("preview %q does not contain %q", preview, want)
		}
	}
}

func TestTunnelFailureSummaryUsesLatestStructuredMessage(t *testing.T) {
	lines := []routemanager.LogLine{
		{Text: `[管理器] 进程已启动，PID=42`},
		{Text: `{"level":"error","msg":"older error"}`},
		{Text: `{"level":"fatal","msg":"listen tcp :12988: bind: Only one usage of each socket address is normally permitted."}`},
		{Text: `[管理器] 进程已退出: exit status 1`},
	}
	got := tunnelFailureSummary(lines, nil)
	want := "listen tcp :12988: bind: Only one usage of each socket address is normally permitted."
	if got != want {
		t.Fatalf("tunnelFailureSummary() = %q, want %q", got, want)
	}
	if !isPortBindingFailure(got) {
		t.Fatalf("expected port binding failure: %q", got)
	}
}

func TestTunnelFailureSummaryIsBounded(t *testing.T) {
	got := tunnelFailureSummary([]routemanager.LogLine{{Text: strings.Repeat("x", 1000)}}, nil)
	if len([]rune(got)) != 600 || !strings.HasSuffix(got, "…") {
		t.Fatalf("unexpected bounded summary length=%d suffix=%q", len([]rune(got)), got[len(got)-3:])
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
