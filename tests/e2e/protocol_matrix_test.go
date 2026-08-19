package e2e

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type proxyPathCase struct {
	name               string
	handler            string
	connector          string
	listener           string
	dialer             string
	serverPort         string
	requireAuth        bool
	password           string
	expectFailure      bool
	concurrentRequests int
}

func writeE2EConfig(t *testing.T, body string) string {
	t.Helper()
	file, err := os.CreateTemp("", "gust-e2e-protocol-*.yaml")
	require.NoError(t, err)
	path := file.Name()
	t.Cleanup(func() { _ = os.Remove(path) })
	_, err = file.WriteString(body)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return path
}

func protocolServerConfig(tc proxyPathCase) string {
	config := fmt.Sprintf(`services:
- name: protocol-server
  addr: :%s
  handler:
    type: %s
  listener:
    type: %s
`, tc.serverPort, tc.handler, tc.listener)
	if tc.requireAuth {
		config = fmt.Sprintf(`services:
- name: protocol-server
  addr: :%s
  handler:
    type: %s
    auther: protocol-auther
  listener:
    type: %s
authers:
- name: protocol-auther
  auths:
  - username: e2e-user
    password: correct-password
`, tc.serverPort, tc.handler, tc.listener)
	}
	return config
}

func protocolClientConfig(tc proxyPathCase, serverAddress string) string {
	auth := ""
	if tc.password != "" {
		auth = fmt.Sprintf(`
        auth:
          username: e2e-user
          password: %s`, tc.password)
	}
	return fmt.Sprintf(`services:
- name: local-http-proxy
  addr: :8080
  handler:
    type: http
    chain: protocol-chain
  listener:
    type: tcp
chains:
- name: protocol-chain
  hops:
  - name: protocol-hop
    nodes:
    - name: protocol-node
      addr: %s
      connector:
        type: %s%s
      dialer:
        type: %s
`, serverAddress, tc.connector, auth, tc.dialer)
}

func executeProxyPathCase(t *testing.T, ctx context.Context, echoIP string, tc proxyPathCase) (int, string) {
	t.Helper()
	alias := "protocol-" + strings.ReplaceAll(tc.name, "_", "-")
	serverConfig := writeE2EConfig(t, protocolServerConfig(tc))
	server, err := RunGostContainerWithOptions(
		ctx,
		SharedNetworkName,
		serverConfig,
		[]string{alias},
		[]string{tc.serverPort + transportPortSuffix(tc.listener)},
	)
	require.NoError(t, err)
	defer server.Terminate(ctx)

	clientConfig := writeE2EConfig(
		t,
		protocolClientConfig(tc, alias+":"+tc.serverPort),
	)
	client, err := RunGostContainerWithPorts(
		ctx,
		SharedNetworkName,
		clientConfig,
		"8080/tcp",
	)
	require.NoError(t, err)
	defer client.Terminate(ctx)

	command := []string{
		"curl", "-fsS", "--connect-timeout", "3", "--max-time", "15",
		"-x", "http://127.0.0.1:8080",
		fmt.Sprintf("http://%s:5678", echoIP),
	}
	if !tc.expectFailure {
		command = append(command[:2], append([]string{"--retry", "5", "--retry-all-errors"}, command[2:]...)...)
	}
	requestCount := tc.concurrentRequests
	if requestCount < 1 {
		requestCount = 1
	}
	type curlResult struct {
		code int
		body string
		err  error
	}
	results := make(chan curlResult, requestCount)
	var waitGroup sync.WaitGroup
	for range requestCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			code, output, execErr := client.Exec(ctx, command)
			if execErr != nil {
				results <- curlResult{code: code, err: execErr}
				return
			}
			body, readErr := io.ReadAll(output)
			results <- curlResult{code: code, body: string(body), err: readErr}
		}()
	}
	waitGroup.Wait()
	close(results)
	combinedBody := strings.Builder{}
	for result := range results {
		if result.err != nil {
			return result.code, result.err.Error()
		}
		if result.code != 0 || !strings.Contains(result.body, "hello-gost") {
			return result.code, result.body
		}
		combinedBody.WriteString(result.body)
	}
	return 0, combinedBody.String()
}

func runProxyPathCase(t *testing.T, ctx context.Context, echoIP string, tc proxyPathCase) {
	t.Helper()
	code, body := executeProxyPathCase(t, ctx, echoIP, tc)
	if code != 0 || !strings.Contains(string(body), "hello-gost") {
		t.Logf("%s did not return the expected echo payload: exit=%d body=%q", tc.name, code, body)
	}
	require.Equal(t, 0, code)
	require.Contains(t, body, "hello-gost")
}

func transportPortSuffix(listener string) string {
	switch listener {
	case "quic", "kcp", "dtls", "http3", "h3", "wt":
		return "/udp"
	default:
		return "/tcp"
	}
}

func TestCoreProxyProtocolMatrix(t *testing.T) {
	ctx := context.Background()
	echo, err := RunEchoContainer(ctx, SharedNetworkName)
	require.NoError(t, err)
	defer echo.Terminate(ctx)
	echoIP, err := echo.ContainerIP(ctx)
	require.NoError(t, err)

	cases := []proxyPathCase{
		{name: "socks5_tcp", handler: "socks5", connector: "socks5", listener: "tcp", dialer: "tcp", serverPort: "1080"},
		{name: "socks4_tcp", handler: "socks4", connector: "socks4", listener: "tcp", dialer: "tcp", serverPort: "1080"},
		{name: "socks4a_tcp", handler: "socks4a", connector: "socks4a", listener: "tcp", dialer: "tcp", serverPort: "1080"},
		{name: "relay_tcp", handler: "relay", connector: "relay", listener: "tcp", dialer: "tcp", serverPort: "8421"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runProxyPathCase(t, ctx, echoIP, tc)
		})
	}
}

func TestCoreProxyAuthenticationMatrix(t *testing.T) {
	ctx := context.Background()
	echo, err := RunEchoContainer(ctx, SharedNetworkName)
	require.NoError(t, err)
	defer echo.Terminate(ctx)
	echoIP, err := echo.ContainerIP(ctx)
	require.NoError(t, err)

	for _, protocol := range []struct {
		name string
		port string
	}{
		{name: "socks5", port: "1080"},
		{name: "relay", port: "8421"},
	} {
		t.Run(protocol.name+"_correct_credentials", func(t *testing.T) {
			runProxyPathCase(t, ctx, echoIP, proxyPathCase{
				name: protocol.name + "-auth-ok", handler: protocol.name,
				connector: protocol.name, listener: "tcp", dialer: "tcp",
				serverPort: protocol.port, requireAuth: true, password: "correct-password",
			})
		})
		t.Run(protocol.name+"_wrong_credentials", func(t *testing.T) {
			code, body := executeProxyPathCase(t, ctx, echoIP, proxyPathCase{
				name: protocol.name + "-auth-reject", handler: protocol.name,
				connector: protocol.name, listener: "tcp", dialer: "tcp",
				serverPort: protocol.port, requireAuth: true, password: "wrong-password",
				expectFailure: true,
			})
			// Curl may report a successful process exit when an authenticated
			// upstream closes without returning an application response. The
			// invariant is that rejected credentials never reach the echo data
			// plane; the paired positive case above proves the path is live.
			require.NotContains(t, body, "hello-gost",
				"wrong credentials reached the echo payload (curl exit=%d)", code)
		})
	}
}

func TestSOCKS5UDPAssociation(t *testing.T) {
	ctx := context.Background()
	udpEcho, err := RunUDPEchoContainer(ctx, SharedNetworkName)
	require.NoError(t, err)
	defer udpEcho.Terminate(ctx)

	serverConfig := writeE2EConfig(t, `services:
- name: socks5-server
  addr: :1080
  handler:
    type: socks5
    metadata:
      udp: true
  listener:
    type: tcp
`)
	server, err := RunGostContainerWithOptions(
		ctx, SharedNetworkName, serverConfig,
		[]string{"socks5-udp-server"}, []string{"1080/tcp"},
	)
	require.NoError(t, err)
	defer server.Terminate(ctx)

	clientConfig := writeE2EConfig(t, `services:
- name: udp-entry
  addr: :9000
  handler:
    type: udp
    chain: socks5-chain
  listener:
    type: udp
  forwarder:
    nodes:
    - name: udp-echo
      addr: udp-echo:5679
chains:
- name: socks5-chain
  hops:
  - name: socks5-hop
    nodes:
    - name: socks5-node
      addr: socks5-udp-server:1080
      connector:
        type: socks5
      dialer:
        type: tcp
`)
	client, err := RunGostContainerWithPorts(
		ctx, SharedNetworkName, clientConfig, "9000/udp",
	)
	require.NoError(t, err)
	defer client.Terminate(ctx)

	host, err := client.Host(ctx)
	require.NoError(t, err)
	port, err := client.MappedPort(ctx, "9000/udp")
	require.NoError(t, err)
	connection, err := net.DialTimeout(
		"udp", net.JoinHostPort(host, port.Port()), 5*time.Second,
	)
	require.NoError(t, err)
	defer connection.Close()

	buffer := make([]byte, 2048)
	for _, payload := range []string{"first-datagram", "second-datagram", "third-datagram"} {
		var readErr error
		for attempt := 0; attempt < 5; attempt++ {
			_, err = connection.Write([]byte(payload))
			require.NoError(t, err)
			require.NoError(t, connection.SetReadDeadline(time.Now().Add(2*time.Second)))
			var count int
			count, readErr = connection.Read(buffer)
			if readErr == nil {
				require.Contains(t, string(buffer[:count]), payload)
				break
			}
		}
		if readErr != nil {
			DumpLogs(t, ctx, "socks5 udp client", client)
			DumpLogs(t, ctx, "socks5 udp server", server)
		}
		require.NoError(t, readErr)
	}
}

func TestStreamTransportMatrix(t *testing.T) {
	ctx := context.Background()
	echo, err := RunEchoContainer(ctx, SharedNetworkName)
	require.NoError(t, err)
	defer echo.Terminate(ctx)
	echoIP, err := echo.ContainerIP(ctx)
	require.NoError(t, err)

	cases := []proxyPathCase{
		{name: "websocket", handler: "socks5", connector: "socks5", listener: "ws", dialer: "ws", serverPort: "8443"},
		{name: "websocket_tls", handler: "socks5", connector: "socks5", listener: "wss", dialer: "wss", serverPort: "8443"},
		{name: "grpc", handler: "socks5", connector: "socks5", listener: "grpc", dialer: "grpc", serverPort: "8443"},
		{name: "h2c", handler: "socks5", connector: "socks5", listener: "h2c", dialer: "h2c", serverPort: "8443"},
		{name: "h2_tls", handler: "socks5", connector: "socks5", listener: "h2", dialer: "h2", serverPort: "8443"},
		{name: "multiplexed_websocket", handler: "socks5", connector: "socks5", listener: "mws", dialer: "mws", serverPort: "8443", concurrentRequests: 6},
		{name: "multiplexed_websocket_tls", handler: "socks5", connector: "socks5", listener: "mwss", dialer: "mwss", serverPort: "8443", concurrentRequests: 6},
		{name: "multiplexed_tcp", handler: "socks5", connector: "socks5", listener: "mtcp", dialer: "mtcp", serverPort: "8443", concurrentRequests: 6},
		{name: "multiplexed_tls", handler: "socks5", connector: "socks5", listener: "mtls", dialer: "mtls", serverPort: "8443", concurrentRequests: 6},
		{name: "obfuscated_http", handler: "socks5", connector: "socks5", listener: "ohttp", dialer: "ohttp", serverPort: "8443"},
		{name: "obfuscated_tls", handler: "socks5", connector: "socks5", listener: "otls", dialer: "otls", serverPort: "8443"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runProxyPathCase(t, ctx, echoIP, tc)
		})
	}
}

func TestUDPBackedStreamTransportMatrix(t *testing.T) {
	ctx := context.Background()
	echo, err := RunEchoContainer(ctx, SharedNetworkName)
	require.NoError(t, err)
	defer echo.Terminate(ctx)
	echoIP, err := echo.ContainerIP(ctx)
	require.NoError(t, err)

	cases := []proxyPathCase{
		{name: "quic", handler: "socks5", connector: "socks5", listener: "quic", dialer: "quic", serverPort: "8443"},
		{name: "kcp", handler: "socks5", connector: "socks5", listener: "kcp", dialer: "kcp", serverPort: "8443"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runProxyPathCase(t, ctx, echoIP, tc)
		})
	}
}
