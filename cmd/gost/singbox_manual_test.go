package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSingboxManualEmbedded(t *testing.T) {
	var output bytes.Buffer
	if err := writeSingboxManual(&output); err != nil {
		t.Fatal(err)
	}

	manual := output.String()
	for _, required := range []string{
		"# Gust embedded sing-box manual",
		"## CLI configuration",
		"## Gust JSON configuration",
		"## Mixed CLI and JSON configuration",
		"## Proxy chains",
		"vless+singbox://",
		"config=/etc/sing-box/config.json",
	} {
		if !strings.Contains(manual, required) {
			t.Errorf("embedded manual is missing %q", required)
		}
	}
	if !strings.HasSuffix(manual, "\n") {
		t.Error("embedded manual must end with a newline")
	}
}
