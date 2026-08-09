package routemanager

import "testing"

func TestManagedBackendPathIsExact(t *testing.T) {
	for _, path := range []string{"gost-qt", "/opt/gust/gost-qt", `C:\\Gust\\gost-qt.exe`} {
		if !isManagedBackendPath(path) {
			t.Fatalf("expected managed backend path: %s", path)
		}
	}
	for _, path := range []string{"gost", "gost.exe", "my-gost-qt", "gost-qt-old"} {
		if isManagedBackendPath(path) {
			t.Fatalf("unexpected managed backend path: %s", path)
		}
	}
}
