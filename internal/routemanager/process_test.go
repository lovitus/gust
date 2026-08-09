package routemanager

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindGostUsesDedicatedBackendName(t *testing.T) {
	t.Setenv("GUST_GOST_BINARY", "")
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	if err := os.WriteFile(filepath.Join(dir, "gost"+ext), []byte("ordinary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := FindGost(""); err == nil {
		t.Fatal("ordinary gost must not be selected as the QtUI backend")
	}
	want := filepath.Join(dir, ManagedBackendName+ext)
	if err := os.WriteFile(want, []byte("managed"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := FindGost("")
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("FindGost() = %q, want %q", got, want)
	}
}

func TestFormatCommandPreservesArgumentBoundaries(t *testing.T) {
	got := FormatCommand([]string{"gost-qt", "-L", "tcp://:8080", "-F", "socks5://user name@example.test:1080", ""})
	want := `gost-qt -L tcp://:8080 -F "socks5://user name@example.test:1080" ""`
	if got != want {
		t.Fatalf("FormatCommand() = %q, want %q", got, want)
	}
}

func TestCommandPreviewShowsExpandedArguments(t *testing.T) {
	got, err := CommandPreview("/opt/gust/gost-qt", Tunnel{
		Name: "free", Mode: TunnelModeFree, Args: `-L tcp://:8080 -F "socks5://user name@example.test:1080"`,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `/opt/gust/gost-qt -L tcp://:8080 -F "socks5://user name@example.test:1080"`
	if got != want {
		t.Fatalf("CommandPreview() = %q, want %q", got, want)
	}
}
