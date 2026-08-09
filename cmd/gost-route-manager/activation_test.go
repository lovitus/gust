package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestActivationServerNotifiesRunningInstanceAndCleansUp(t *testing.T) {
	config := filepath.Join(t.TempDir(), "route-manager.json")
	key := instanceKey(config)
	server, err := startActivationServer(key)
	if err != nil {
		t.Fatal(err)
	}
	path := activationEndpointPath(key)
	if err := activateExistingInstance(config, time.Second); err != nil {
		t.Fatal(err)
	}
	select {
	case <-server.events:
	case <-time.After(time.Second):
		t.Fatal("activation event was not delivered")
	}
	if err := server.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("activation endpoint remains after close: %v", err)
	}
}

func TestActivationServerRejectsWrongToken(t *testing.T) {
	server, err := startActivationServer(instanceKey(filepath.Join(t.TempDir(), "route-manager.json")))
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if err := sendActivation(activationEndpoint{Address: server.listener.Addr().String(), Token: "wrong"}); err == nil {
		t.Fatal("wrong activation token was accepted")
	}
	select {
	case <-server.events:
		t.Fatal("wrong token emitted an activation event")
	default:
	}
}
