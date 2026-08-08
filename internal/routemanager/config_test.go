package routemanager

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParseRouteOptions(t *testing.T) {
	opts, err := ParseRouteOptions("10.233.0.0/16, 10.27.0.0/16, dns=1.1.1.1,8.8.8.8, mtu=1380")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(opts.Routes, []string{"10.233.0.0/16", "10.27.0.0/16"}) {
		t.Fatalf("routes = %#v", opts.Routes)
	}
	if !reflect.DeepEqual(opts.DNS, []string{"1.1.1.1", "8.8.8.8"}) {
		t.Fatalf("dns = %#v", opts.DNS)
	}
	if opts.MTU != 1380 {
		t.Fatalf("mtu = %d", opts.MTU)
	}
}

func TestBuildArgs(t *testing.T) {
	args, err := BuildArgs(Tunnel{ID: "1", Name: "zwy", Routes: "10.233.0.0/16,dns=1.1.1.1,mtu=1400", Target: "192.168.1.37:5555"})
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 4 || args[0] != "-L" || args[2] != "-F" || args[3] != "socks5://192.168.1.37:5555" {
		t.Fatalf("args = %#v", args)
	}
	u, err := url.Parse(args[1])
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "tungo" || u.Query().Get("routes") != "10.233.0.0/16" || u.Query().Get("dns") != "1.1.1.1" || u.Query().Get("mtu") != "1400" {
		t.Fatalf("listener = %s", args[1])
	}
}

func TestParseRouteOptionsRejectsBadInput(t *testing.T) {
	for _, input := range []string{"", "10.0.0.1", "10.0.0.0/8,mtu=100", "10.0.0.0/8,dns=nope"} {
		if _, err := ParseRouteOptions(input); err == nil {
			t.Errorf("ParseRouteOptions(%q) succeeded", input)
		}
	}
}
