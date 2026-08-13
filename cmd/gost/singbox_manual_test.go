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
		"## Nine-protocol example menu",
		"## CLI configuration",
		"## Gust JSON configuration",
		"## Mixed CLI and JSON configuration",
		"## Architecture and request path",
		"## Proxy chains",
		"## Performance and trade-offs",
		"vless+singbox://",
		"config=/etc/sing-box/config.json",
		"udp=true",
		"gost-json-only-nine-server.json",
		"gost-json-only-nine-client.json",
	} {
		if !strings.Contains(manual, required) {
			t.Errorf("embedded manual is missing %q", required)
		}
	}
	menuStart := strings.Index(manual, "## Nine-protocol example menu")
	menuEnd := strings.Index(manual[menuStart+1:], "\n## ")
	if menuStart < 0 || menuEnd < 0 {
		t.Fatal("embedded manual is missing the bounded nine-protocol menu")
	}
	menu := manual[menuStart : menuStart+1+menuEnd]
	for _, required := range []string{
		"VLESS Reality Vision",
		"Hysteria2",
		"TUIC",
		"ShadowTLS v3 + Shadowsocks 2022",
		"Shadowsocks 2022",
		"Trojan TLS",
		"VLESS gRPC Reality",
		"AnyTLS",
		"VMess WebSocket TLS",
		"CLI-only form",
		"CLI + native JSON files",
		"Gust JSON-only client",
	} {
		if !strings.Contains(menu, required) {
			t.Errorf("embedded nine-protocol menu is missing %q", required)
		}
	}
	for _, required := range []string{
		"Naive outbound is not an embedded capability",
		"Tailscale endpoints are rejected in all modes",
		"WireGuard endpoint | PASS | PASS | Unscoped only",
	} {
		if !strings.Contains(manual, required) {
			t.Errorf("embedded manual is missing current capability boundary %q", required)
		}
	}
	if !strings.HasSuffix(manual, "\n") {
		t.Error("embedded manual must end with a newline")
	}
	for _, obsolete := range []string{
		"SSH, Naive | PASS",
		"WireGuard or Tailscale endpoint",
		"For Naive, verify the Cronet library",
	} {
		if strings.Contains(manual, obsolete) {
			t.Errorf("embedded manual retains obsolete capability claim %q", obsolete)
		}
	}
}
