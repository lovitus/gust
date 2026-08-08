package compat

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type synchronizedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *synchronizedBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(data)
}

func (b *synchronizedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}

type testProcess struct {
	name string
	cmd  *exec.Cmd
	done chan error
	log  synchronizedBuffer
}

func startProcess(t *testing.T, name, executable string, args ...string) *testProcess {
	t.Helper()
	process := &testProcess{
		name: name,
		cmd:  exec.Command(executable, args...),
		done: make(chan error, 1),
	}
	process.cmd.Stdout = &process.log
	process.cmd.Stderr = &process.log
	if err := process.cmd.Start(); err != nil {
		t.Fatalf("start %s: %v", name, err)
	}
	go func() { process.done <- process.cmd.Wait() }()
	t.Cleanup(func() {
		select {
		case <-process.done:
			return
		default:
		}
		_ = process.cmd.Process.Signal(os.Interrupt)
		select {
		case <-process.done:
		case <-time.After(3 * time.Second):
			_ = process.cmd.Process.Kill()
			<-process.done
		}
	})
	return process
}

func freePort(t *testing.T, network string) int {
	t.Helper()
	address := "127.0.0.1:0"
	if network == "tcp" {
		listener, err := net.Listen(network, address)
		if err != nil {
			t.Fatal(err)
		}
		defer listener.Close()
		return listener.Addr().(*net.TCPAddr).Port
	}
	connection, err := net.ListenPacket(network, address)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	return connection.LocalAddr().(*net.UDPAddr).Port
}

func waitTCP(t *testing.T, address string, process *testProcess) {
	t.Helper()
	deadline := time.Now().Add(12 * time.Second)
	for time.Now().Before(deadline) {
		connection, err := net.DialTimeout("tcp", address, 250*time.Millisecond)
		if err == nil {
			_ = connection.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("%s did not listen on %s:\n%s", process.name, address, process.log.String())
}

func runUDPResponder(t *testing.T, reply func([]byte) []byte) string {
	t.Helper()
	connection, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.Close() })
	go func() {
		buffer := make([]byte, 4096)
		for {
			count, source, err := connection.ReadFrom(buffer)
			if err != nil {
				return
			}
			response := reply(append([]byte(nil), buffer[:count]...))
			_, _ = connection.WriteTo(response, source)
		}
	}()
	return connection.LocalAddr().String()
}

func relayDatagram(t *testing.T, address string, payload []byte) ([]byte, error) {
	t.Helper()
	connection, err := net.DialTimeout("udp", address, 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	buffer := make([]byte, 4096)
	var lastError error
	for range 10 {
		if _, err = connection.Write(payload); err != nil {
			return nil, err
		}
		_ = connection.SetReadDeadline(time.Now().Add(time.Second))
		count, err := connection.Read(buffer)
		if err == nil {
			return append([]byte(nil), buffer[:count]...), nil
		}
		lastError = err
		time.Sleep(200 * time.Millisecond)
	}
	return nil, fmt.Errorf("UDP relay did not return a response: %w", lastError)
}

func dnsQuery() []byte {
	return []byte{
		0x47, 0x55, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00,
		0x04, 'e', 'c', 'h', 'o', 0x04, 't', 'e', 's', 't', 0x00,
		0x00, 0x01, 0x00, 0x01,
	}
}

func dnsReply(query []byte) []byte {
	if len(query) < 12 {
		return query
	}
	response := append([]byte(nil), query...)
	response[2], response[3] = 0x81, 0x80
	response[6], response[7] = 0x00, 0x01
	return append(response,
		0xc0, 0x0c, 0x00, 0x01, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x3c, 0x00, 0x04,
		192, 0, 2, 1,
	)
}

func TestMihomoUoTV2UDPAndDNS(t *testing.T) {
	gostBinary := os.Getenv("GOST_BIN")
	mihomoBinary := os.Getenv("MIHOMO_BIN")
	if gostBinary == "" || mihomoBinary == "" {
		t.Skip("set GOST_BIN and MIHOMO_BIN to run external compatibility")
	}
	for _, binary := range []string{gostBinary, mihomoBinary} {
		if _, err := os.Stat(binary); err != nil {
			t.Fatalf("compatibility binary %s: %v", binary, err)
		}
	}

	serverPort := freePort(t, "tcp")
	mihomoPort := freePort(t, "tcp")
	serverAddress := fmt.Sprintf("127.0.0.1:%d", serverPort)
	mihomoAddress := fmt.Sprintf("127.0.0.1:%d", mihomoPort)

	server := startProcess(
		t, "Gust sings server", gostBinary,
		"-L", fmt.Sprintf("sings://aes-128-gcm:e2e-password@%s?mux=false", serverAddress),
	)
	waitTCP(t, serverAddress, server)

	configDirectory := t.TempDir()
	config := fmt.Sprintf(`mixed-port: %d
allow-lan: false
mode: rule
log-level: warning
proxies:
  - name: gust-sings
    type: ss
    server: 127.0.0.1
    port: %d
    cipher: aes-128-gcm
    password: e2e-password
    udp: true
    udp-over-tcp: true
    udp-over-tcp-version: 2
rules:
  - MATCH,gust-sings
`, mihomoPort, serverPort)
	configPath := filepath.Join(configDirectory, "config.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	mihomo := startProcess(t, "Mihomo", mihomoBinary, "-d", configDirectory, "-f", configPath)
	waitTCP(t, mihomoAddress, mihomo)

	echoAddress := runUDPResponder(t, func(packet []byte) []byte { return packet })
	dnsAddress := runUDPResponder(t, dnsReply)

	startForward := func(name, target string) (string, *testProcess) {
		entryPort := freePort(t, "udp")
		entryAddress := fmt.Sprintf("127.0.0.1:%d", entryPort)
		process := startProcess(
			t, name, gostBinary,
			"-L", fmt.Sprintf("udp://%s/%s", entryAddress, target),
			"-F", "socks5://"+mihomoAddress+"?relay=udp",
		)
		return entryAddress, process
	}

	t.Run("arbitrary UDP", func(t *testing.T) {
		payload := []byte("mihomo-uot-v2-data-plane")
		entryAddress, forward := startForward("UDP forward", echoAddress)
		response, err := relayDatagram(t, entryAddress, payload)
		if err != nil {
			t.Fatalf("%v\nforward:\n%s\nmihomo:\n%s\nserver:\n%s", err, forward.log.String(), mihomo.log.String(), server.log.String())
		}
		if !bytes.Equal(response, payload) {
			t.Fatalf("UDP payload mismatch: got %q want %q", response, payload)
		}
	})

	t.Run("DNS", func(t *testing.T) {
		query := dnsQuery()
		entryAddress, forward := startForward("DNS forward", dnsAddress)
		response, err := relayDatagram(t, entryAddress, query)
		if err != nil {
			t.Fatalf("%v\nforward:\n%s\nmihomo:\n%s\nserver:\n%s", err, forward.log.String(), mihomo.log.String(), server.log.String())
		}
		if len(response) < len(query)+16 {
			t.Fatalf("short DNS response: %x", response)
		}
		if response[0] != query[0] || response[1] != query[1] || response[2]&0x80 == 0 {
			t.Fatalf("invalid DNS response header: %x", response[:12])
		}
		if response[6] != 0 || response[7] != 1 {
			t.Fatalf("DNS answer count is not one: %x", response[:12])
		}
	})

}
