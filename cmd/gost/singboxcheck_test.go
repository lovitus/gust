package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-gost/x/backend/singbox"
	"github.com/go-gost/x/config"
)

func TestSummarizeSingboxConfig(t *testing.T) {
	cfg := &config.Config{
		Services: []*config.ServiceConfig{
			{Listener: &config.ListenerConfig{Type: "sings"}, Handler: &config.HandlerConfig{Type: "sings"}},
			{Listener: &config.ListenerConfig{Type: "tcp"}, Handler: &config.HandlerConfig{Type: "auto"}},
		},
		Hops: []*config.HopConfig{{Nodes: []*config.NodeConfig{
			{Connector: &config.ConnectorConfig{Type: "sings"}},
			{Connector: &config.ConnectorConfig{Type: "http"}},
		}}},
		Chains: []*config.ChainConfig{{Hops: []*config.HopConfig{{Nodes: []*config.NodeConfig{
			{Dialer: &config.DialerConfig{Type: "singsu"}},
		}}}}},
	}
	summary := summarizeSingboxConfig(cfg)
	if summary.services != 2 || summary.chains != 1 || summary.hops != 2 || summary.nodes != 3 ||
		summary.nativeServices != 1 || summary.nativeNodes != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestWriteSingboxCheckDoesNotExposeConfigurationValues(t *testing.T) {
	manifest := singbox.BuildManifest()
	if !manifest.Compiled {
		t.Skip("requires an embedded sing-box build")
	}
	const secret = "must-not-appear"
	cfg := &config.Config{Services: []*config.ServiceConfig{{
		Name: secret, Addr: "198.51.100.20:1080",
		Listener: &config.ListenerConfig{Type: "sings", Metadata: map[string]any{"password": secret}},
		Handler:  &config.HandlerConfig{Type: "sings"},
	}}}
	var output bytes.Buffer
	if err := writeSingboxCheck(&output, cfg); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, forbidden := range []string{secret, "198.51.100.20", "password"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("safe summary leaked %q: %s", forbidden, text)
		}
	}
	for _, expected := range []string{"sing-box configuration OK", "native_services=1", "startup not attempted"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("safe summary missing %q: %s", expected, text)
		}
	}
}
