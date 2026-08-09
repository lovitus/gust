package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type activationEndpoint struct {
	Address string `json:"address"`
	Token   string `json:"token"`
}

type activationServer struct {
	listener net.Listener
	path     string
	token    string
	events   chan struct{}
	done     chan struct{}
	once     sync.Once
	wg       sync.WaitGroup
}

func activationEndpointPath(key string) string {
	return filepath.Join(os.TempDir(), "gust-route-manager-"+key+".activate")
}

func startActivationServer(key string) (*activationServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("启动窗口激活服务: %w", err)
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("生成窗口激活令牌: %w", err)
	}
	server := &activationServer{
		listener: listener,
		path:     activationEndpointPath(key),
		token:    hex.EncodeToString(tokenBytes),
		events:   make(chan struct{}, 1),
		done:     make(chan struct{}),
	}
	data, err := json.Marshal(activationEndpoint{Address: listener.Addr().String(), Token: server.token})
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	if err := os.WriteFile(server.path, data, 0o600); err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("写入窗口激活端点: %w", err)
	}
	server.wg.Add(1)
	go server.serve()
	return server, nil
}

func (s *activationServer) serve() {
	defer s.wg.Done()
	defer close(s.events)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}
		s.handle(conn)
	}
}

func (s *activationServer) handle(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	token, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || strings.TrimSpace(token) != s.token {
		return
	}
	select {
	case s.events <- struct{}{}:
	default:
	}
	_, _ = fmt.Fprintln(conn, "ok")
}

func (s *activationServer) Close() error {
	var closeErr error
	s.once.Do(func() {
		close(s.done)
		closeErr = s.listener.Close()
		s.wg.Wait()
		data, err := os.ReadFile(s.path)
		if err == nil {
			var endpoint activationEndpoint
			if json.Unmarshal(data, &endpoint) == nil && endpoint.Token == s.token {
				_ = os.Remove(s.path)
			}
		}
	})
	if errors.Is(closeErr, net.ErrClosed) {
		return nil
	}
	return closeErr
}

func activateExistingInstance(configPath string, wait time.Duration) error {
	path := activationEndpointPath(instanceKey(configPath))
	deadline := time.Now().Add(wait)
	var lastErr error
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			var endpoint activationEndpoint
			if err = json.Unmarshal(data, &endpoint); err == nil {
				err = sendActivation(endpoint)
			}
		}
		if err == nil {
			return nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return fmt.Errorf("通知已运行窗口失败: %w", lastErr)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func sendActivation(endpoint activationEndpoint) error {
	conn, err := net.DialTimeout("tcp", endpoint.Address, 300*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := fmt.Fprintln(conn, endpoint.Token); err != nil {
		return err
	}
	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(response) != "ok" {
		return errors.New("窗口激活服务未确认请求")
	}
	return nil
}
