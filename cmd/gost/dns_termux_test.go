package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseResolvConfAcceptsOnlyIPNameservers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolv.conf")
	content := []byte("nameserver 8.8.8.8\nnameserver 2001:4860:4860::8888\nnameserver resolver.example\nnameserver 0.0.0.0\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	want := []string{"8.8.8.8:53", "[2001:4860:4860::8888]:53"}
	if got := parseResolvConf(path); !reflect.DeepEqual(got, want) {
		t.Fatalf("parseResolvConf()=%v, want %v", got, want)
	}
}
