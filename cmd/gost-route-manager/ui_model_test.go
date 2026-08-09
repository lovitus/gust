package main

import "testing"

func TestNewTunnelUsesPlaceholdersInsteadOfPrefilledValues(t *testing.T) {
	tunnel := newTunnel("tunnel-test")
	if tunnel.ID != "tunnel-test" {
		t.Fatalf("ID = %q", tunnel.ID)
	}
	if tunnel.Name != "" || tunnel.Routes != "" || tunnel.Target != "" {
		t.Fatalf("new tunnel must be empty so placeholders remain placeholders: %+v", tunnel)
	}
}
