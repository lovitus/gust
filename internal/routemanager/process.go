package routemanager

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type ProcessManager struct {
	mu     sync.Mutex
	bin    string
	procs  map[string]*exec.Cmd
	output map[string]*bytes.Buffer
}

func NewProcessManager(binary string) *ProcessManager {
	return &ProcessManager{bin: binary, procs: make(map[string]*exec.Cmd), output: make(map[string]*bytes.Buffer)}
}

func FindGost(explicit string) (string, error) {
	if explicit != "" {
		return executableFile(explicit)
	}
	if env := os.Getenv("GUST_GOST_BINARY"); env != "" {
		return executableFile(env)
	}
	exe, _ := os.Executable()
	name := "gost"
	if filepath.Ext(exe) == ".exe" {
		name += ".exe"
	}
	for _, candidate := range []string{filepath.Join(filepath.Dir(exe), name), filepath.Join(".", name)} {
		if path, err := executableFile(candidate); err == nil {
			return path, nil
		}
	}
	if path, err := exec.LookPath("gost"); err == nil {
		return path, nil
	}
	return "", errors.New("找不到 gost；请把 gost 放在本程序同目录，或使用 --gost 指定")
}

func executableFile(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s 不是可执行文件", abs)
	}
	return abs, nil
}

func (m *ProcessManager) Start(t Tunnel, done func(error)) error {
	args, err := BuildArgs(t)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.procs[t.ID]; ok {
		return errors.New("隧道已经在运行")
	}
	buf := &bytes.Buffer{}
	cmd := exec.Command(m.bin, args...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	prepareProcess(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	m.procs[t.ID] = cmd
	m.output[t.ID] = buf
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		delete(m.procs, t.ID)
		m.mu.Unlock()
		done(err)
	}()
	return nil
}

func (m *ProcessManager) Stop(id string) error {
	m.mu.Lock()
	cmd := m.procs[id]
	m.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = stopProcess(cmd.Process)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		_, running := m.procs[id]
		m.mu.Unlock()
		if !running {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return cmd.Process.Kill()
}

func (m *ProcessManager) StopAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.procs))
	for id := range m.procs {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		_ = m.Stop(id)
	}
}

func (m *ProcessManager) Output(id string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if b := m.output[id]; b != nil {
		return b.String()
	}
	return ""
}
