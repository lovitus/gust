package main

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-gost/x/config"
	"github.com/go-gost/x/config/parsing/parser"
	xmetrics "github.com/go-gost/x/metrics"
)

func runtimeTestAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate test address: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()
	return addr
}

func runtimeAssertBound(t *testing.T, addr string) {
	t.Helper()
	ln, err := net.Listen("tcp", addr)
	if err == nil {
		ln.Close()
		t.Fatalf("expected %s to be bound", addr)
	}
}

func cleanupRuntimeProgram(t *testing.T, p *program) {
	t.Helper()
	t.Cleanup(func() {
		_ = p.Stop()
		config.Set(&config.Config{})
		xmetrics.Enable(false)
	})
}

func TestActivateConfigInvalidCandidateKeepsOldRuntime(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	addr := runtimeTestAddr(t)
	old := &config.Config{API: &config.APIConfig{Addr: addr, PathPrefix: "/old"}}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old config: %v", err)
	}
	original := p.srvApi

	candidate := &config.Config{
		API: &config.APIConfig{Addr: addr, PathPrefix: "/new"},
		Services: []*config.ServiceConfig{{
			Name:     "invalid",
			Addr:     "127.0.0.1:0",
			Listener: &config.ListenerConfig{Type: "missing-listener"},
		}},
	}
	if err := p.activateConfig(candidate); err == nil {
		t.Fatal("expected candidate validation failure")
	}
	if p.srvApi != original {
		t.Fatal("static candidate failure replaced the old API service")
	}
	if got := config.Global().API.PathPrefix; got != "/old" {
		t.Fatalf("published config changed on failure: %q", got)
	}
	runtimeAssertBound(t, addr)
}

func TestActivateConfigRejectsMissingProcessAutherWithoutReplacingRuntime(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	addr := runtimeTestAddr(t)
	old := &config.Config{
		Authers: []*config.AutherConfig{{
			Name:  "admin",
			Auths: []*config.AuthConfig{{Username: "old", Password: "secret"}},
		}},
		API: &config.APIConfig{Addr: addr, Auther: "admin"},
	}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old config: %v", err)
	}
	original := p.srvApi

	candidate := &config.Config{API: &config.APIConfig{Addr: addr, Auther: "missing"}}
	if err := p.activateConfig(candidate); err == nil {
		t.Fatal("expected missing API auther to reject candidate")
	}
	if p.srvApi != original {
		t.Fatal("missing auther candidate replaced the authenticated API runtime")
	}
	if got := config.Global(); got.API == nil || got.API.Auther != "admin" || len(got.Authers) != 1 {
		t.Fatalf("missing auther candidate changed published config: %#v", got)
	}
	runtimeAssertBound(t, addr)
}

func TestActivateConfigRejectsMissingMetricsAuther(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	cfg := &config.Config{Metrics: &config.MetricsConfig{
		Addr:   "127.0.0.1:0",
		Auther: "missing",
	}}
	if err := p.activateConfig(cfg); err == nil {
		t.Fatal("expected missing metrics auther to be rejected")
	}
	if p.srvMetrics != nil {
		t.Fatal("missing metrics auther started an unauthenticated metrics service")
	}
	if config.Global().Metrics != nil {
		t.Fatal("missing metrics auther candidate was published")
	}
}

func TestActivateConfigRuntimeFailureRestoresAPIAndMetrics(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	apiAddr, metricsAddr := runtimeTestAddr(t), runtimeTestAddr(t)
	old := &config.Config{
		API:     &config.APIConfig{Addr: apiAddr, PathPrefix: "/old"},
		Metrics: &config.MetricsConfig{Addr: metricsAddr, Path: "/old-metrics"},
	}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old config: %v", err)
	}

	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	candidate := &config.Config{
		API:     &config.APIConfig{Addr: apiAddr, PathPrefix: "/candidate"},
		Metrics: &config.MetricsConfig{Addr: occupied.Addr().String(), Path: "/candidate-metrics"},
	}
	if err := p.activateConfig(candidate); err == nil {
		t.Fatal("expected occupied metrics bind to reject candidate")
	}

	global := config.Global()
	if global.API == nil || global.API.PathPrefix != "/old" || global.Metrics == nil || global.Metrics.Addr != metricsAddr {
		t.Fatalf("old published config not retained: %#v", global)
	}
	if p.srvApi == nil || p.srvMetrics == nil {
		t.Fatal("old runtime services were not rebuilt")
	}
	runtimeAssertBound(t, apiAddr)
	runtimeAssertBound(t, metricsAddr)
}

func TestActivateConfigLoaderFailureRestoresMetricsState(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	old := &config.Config{Metrics: &config.MetricsConfig{Addr: runtimeTestAddr(t)}}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old config: %v", err)
	}
	if !xmetrics.IsEnabled() {
		t.Fatal("old metrics state was not enabled")
	}

	candidate := &config.Config{Hops: []*config.HopConfig{{
		Name: "bad-hop",
		Nodes: []*config.NodeConfig{{
			Name:      "bad-node",
			Addr:      "127.0.0.1:1",
			Connector: &config.ConnectorConfig{Type: "missing-connector"},
		}},
	}}}
	if err := p.activateConfig(candidate); err == nil {
		t.Fatal("expected loader activation failure")
	}
	if !xmetrics.IsEnabled() {
		t.Fatal("loader failure did not restore the predecessor metrics state")
	}
	if p.srvMetrics == nil {
		t.Fatal("loader failure replaced the predecessor metrics service")
	}
	if got := config.Global(); got.Metrics == nil || got.Metrics.Addr != old.Metrics.Addr {
		t.Fatalf("loader failure published candidate config: %#v", got)
	}
}

func TestActivateConfigProfilingBindFailureIsSynchronous(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	oldAddr := runtimeTestAddr(t)
	old := &config.Config{Profiling: &config.ProfilingConfig{Addr: oldAddr}}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old profiling: %v", err)
	}

	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	candidate := &config.Config{Profiling: &config.ProfilingConfig{Addr: occupied.Addr().String()}}
	if err := p.activateConfig(candidate); err == nil {
		t.Fatal("expected profiling bind failure from activateConfig")
	}
	if got := config.Global().Profiling.Addr; got != oldAddr {
		t.Fatalf("profiling config changed after failed bind: got %s want %s", got, oldAddr)
	}
	if p.srvProfiling == nil {
		t.Fatal("old profiling service was not restored")
	}
	runtimeAssertBound(t, oldAddr)
}

func TestActivateConfigSamePortCommitAndUnchangedRetention(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	addr := runtimeTestAddr(t)
	old := &config.Config{API: &config.APIConfig{Addr: addr, PathPrefix: "/one"}}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old API: %v", err)
	}
	first := p.srvApi

	candidate := &config.Config{API: &config.APIConfig{Addr: addr, PathPrefix: "/two"}}
	if err := p.activateConfig(candidate); err != nil {
		t.Fatalf("same-port reload: %v", err)
	}
	if p.srvApi == first {
		t.Fatal("changed API config unexpectedly retained old service")
	}
	second := p.srvApi
	runtimeAssertBound(t, addr)

	unchanged := &config.Config{
		API: candidate.API,
		Log: &config.LogConfig{Level: "warn"},
	}
	if err := p.activateConfig(unchanged); err != nil {
		t.Fatalf("reload with unchanged API: %v", err)
	}
	if p.srvApi != second {
		t.Fatal("unchanged API service was unnecessarily rebound")
	}
	if got := config.Global().API.PathPrefix; got != "/two" {
		t.Fatalf("candidate config not published: %q", got)
	}
}

func TestAPIReloadCanReplaceItsOwnFixedPortWithoutDeadlock(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	addr := runtimeTestAddr(t)
	old := &config.Config{API: &config.APIConfig{Addr: addr}}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old API: %v", err)
	}

	parser.Init(parser.Args{CfgFiles: []string{fmt.Sprintf(
		`{"api":{"addr":%q,"pathPrefix":"/new"}}`, addr,
	)}})
	t.Cleanup(func() { parser.Init(parser.Args{}) })

	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	resp, err := client.Post("http://"+addr+"/config/reload", "application/json", nil)
	if err != nil {
		t.Fatalf("self-hosted reload request failed or deadlocked: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("self-hosted reload status = %d", resp.StatusCode)
	}
	if got := config.Global().API.PathPrefix; got != "/new" {
		t.Fatalf("candidate config not published: %q", got)
	}

	resp, err = client.Get("http://" + addr + "/new/config")
	if err != nil {
		t.Fatalf("replacement API is not serving on the fixed port: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("replacement API status = %d", resp.StatusCode)
	}
}

func TestAPIReloadFailureRestoresItsOwnFixedPortAndOldConfig(t *testing.T) {
	p := &program{}
	cleanupRuntimeProgram(t, p)
	addr := runtimeTestAddr(t)
	old := &config.Config{API: &config.APIConfig{Addr: addr, PathPrefix: "/old"}}
	if err := p.activateConfig(old); err != nil {
		t.Fatalf("activate old API: %v", err)
	}

	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	parser.Init(parser.Args{CfgFiles: []string{fmt.Sprintf(
		`{"api":{"addr":%q,"pathPrefix":"/new"},"metrics":{"addr":%q}}`,
		addr, occupied.Addr().String(),
	)}})
	t.Cleanup(func() { parser.Init(parser.Args{}) })

	client := &http.Client{Timeout: 3 * time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Post("http://"+addr+"/old/config/reload", "application/json", nil)
	if err != nil {
		t.Fatalf("failed self-reload request deadlocked or lost its response: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("failed self-reload status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if got := config.Global().API.PathPrefix; got != "/old" {
		t.Fatalf("failed self-reload published candidate path %q", got)
	}

	resp, err = client.Get("http://" + addr + "/old/config")
	if err != nil {
		t.Fatalf("restored API is not serving on the original fixed port: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("restored API status = %d", resp.StatusCode)
	}
}
