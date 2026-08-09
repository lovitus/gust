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
