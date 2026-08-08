package main

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-gost/x/backend/singbox"
	"github.com/go-gost/x/config"
)

type singboxConfigSummary struct {
	services       int
	chains         int
	hops           int
	nodes          int
	nativeServices int
	nativeNodes    int
}

// writeSingboxCheck deliberately reports structure only. It must never print
// source URLs, addresses, names, metadata or native JSON because all of those
// may contain credentials.
func writeSingboxCheck(w io.Writer, cfg *config.Config) error {
	manifest := singbox.BuildManifest()
	if !manifest.Compiled {
		return errors.New("embedded sing-box backend is unavailable in this build")
	}
	if cfg == nil {
		return errors.New("configuration is nil")
	}
	summary := summarizeSingboxConfig(cfg)
	fmt.Fprintln(w, "sing-box configuration OK")
	fmt.Fprintf(w, "flavor=%s sing-box=%s features=%s\n", manifest.Flavor, manifest.PinnedVersion, manifest.Features)
	fmt.Fprintf(w, "services=%d chains=%d hops=%d nodes=%d native_services=%d native_nodes=%d\n",
		summary.services, summary.chains, summary.hops, summary.nodes, summary.nativeServices, summary.nativeNodes)
	fmt.Fprintln(w, "startup not attempted; sockets and system resources were not opened")
	return nil
}

func summarizeSingboxConfig(cfg *config.Config) (summary singboxConfigSummary) {
	summary.services = len(cfg.Services)
	summary.chains = len(cfg.Chains)
	summary.hops = len(cfg.Hops)
	for _, service := range cfg.Services {
		if service == nil {
			continue
		}
		if service.Listener != nil && isSingboxType(service.Listener.Type) ||
			service.Handler != nil && isSingboxType(service.Handler.Type) {
			summary.nativeServices++
		}
	}
	for _, hop := range cfg.Hops {
		countSingboxHop(&summary, hop)
	}
	for _, chain := range cfg.Chains {
		if chain == nil {
			continue
		}
		summary.hops += len(chain.Hops)
		for _, hop := range chain.Hops {
			countSingboxHop(&summary, hop)
		}
	}
	return summary
}

func countSingboxHop(summary *singboxConfigSummary, hop *config.HopConfig) {
	if hop == nil {
		return
	}
	summary.nodes += len(hop.Nodes)
	for _, node := range hop.Nodes {
		if node == nil {
			continue
		}
		if node.Connector != nil && isSingboxType(node.Connector.Type) ||
			node.Dialer != nil && isSingboxType(node.Dialer.Type) {
			summary.nativeNodes++
		}
	}
}

func isSingboxType(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "sings", "singsu", "singbox", "sing-box":
		return true
	default:
		return false
	}
}
