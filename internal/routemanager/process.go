package routemanager

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type ProcessManager struct {
	mu    sync.Mutex
	bin   string
	procs map[string]*managedProcess
	logs  *LogBuffer
}

type managedProcess struct {
	cmd     *exec.Cmd
	control processControl
}

type processControl interface {
	stop() error
	kill() error
	close() error
}

func NewProcessManager(binary string) *ProcessManager {
	return &ProcessManager{bin: binary, procs: make(map[string]*managedProcess), logs: NewLogBuffer(100, 1000)}
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
	cmd := exec.Command(m.bin, args...)
	writer := m.logs.Writer(t.ID)
	cmd.Stdout = writer
	cmd.Stderr = writer
	prepareProcess(cmd)
	if err := cmd.Start(); err != nil {
		m.logs.Append(t.ID, fmt.Sprintf("[管理器] 启动失败: %v", err))
		return err
	}
	control, err := ownProcess(cmd)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		m.logs.Append(t.ID, fmt.Sprintf("[管理器] 建立进程所有权失败: %v", err))
		return fmt.Errorf("建立隧道进程所有权: %w", err)
	}
	proc := &managedProcess{cmd: cmd, control: control}
	m.procs[t.ID] = proc
	m.logs.Append(t.ID, fmt.Sprintf("[管理器] 进程已启动，PID=%d", cmd.Process.Pid))
	go func() {
		err := cmd.Wait()
		_ = control.close()
		m.logs.Flush(t.ID)
		if err != nil {
			m.logs.Append(t.ID, fmt.Sprintf("[管理器] 进程已退出: %v", err))
		} else {
			m.logs.Append(t.ID, "[管理器] 进程已正常退出")
		}
		m.mu.Lock()
		if m.procs[t.ID] == proc {
			delete(m.procs, t.ID)
		}
		m.mu.Unlock()
		done(err)
	}()
	return nil
}

func (m *ProcessManager) Running(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.procs[id]
	return ok
}

func (m *ProcessManager) Stop(id string) error {
	m.mu.Lock()
	proc := m.procs[id]
	m.mu.Unlock()
	if proc == nil || proc.cmd.Process == nil {
		return nil
	}
	if err := proc.control.stop(); err != nil && !isProcessDone(err) {
		m.logs.Append(id, fmt.Sprintf("[管理器] 请求停止失败，将强制回收: %v", err))
	}
	deadline := time.Now().Add(3 * time.Second)
	if m.waitStopped(id, proc, deadline) {
		return nil
	}
	m.logs.Append(id, "[管理器] 优雅停止超时，强制回收本任务进程树")
	if err := proc.control.kill(); err != nil && !isProcessDone(err) {
		return fmt.Errorf("强制回收隧道进程树: %w", err)
	}
	if !m.waitStopped(id, proc, time.Now().Add(2*time.Second)) {
		return errors.New("强制回收后进程仍未退出")
	}
	return nil
}

func (m *ProcessManager) waitStopped(id string, proc *managedProcess, deadline time.Time) bool {
	for time.Now().Before(deadline) {
		m.mu.Lock()
		running := m.procs[id] == proc
		m.mu.Unlock()
		if !running {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func (m *ProcessManager) StopAll() error {
	m.mu.Lock()
	ids := make([]string, 0, len(m.procs))
	for id := range m.procs {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	errByID := make([]error, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := m.Stop(id); err != nil {
				errByID[i] = fmt.Errorf("%s: %w", id, err)
			}
		}()
	}
	wg.Wait()
	return errors.Join(errByID...)
}

func (m *ProcessManager) Output(id string) string {
	lines := m.logs.Lines(id, 100)
	texts := make([]string, 0, len(lines))
	for _, line := range lines {
		texts = append(texts, line.Text)
	}
	return strings.Join(texts, "\n")
}

func (m *ProcessManager) Logs(id string, limit int) []LogLine {
	return m.logs.Lines(id, limit)
}

func (m *ProcessManager) AllLogs(limit int) []LogLine {
	return m.logs.All(limit)
}

func (m *ProcessManager) LogVersion(id string) uint64 {
	return m.logs.Version(id)
}

func (m *ProcessManager) AllLogVersion() uint64 {
	return m.logs.AllVersion()
}

func (m *ProcessManager) LogsSnapshot(id string, limit int) ([]LogLine, uint64) {
	return m.logs.LinesSnapshot(id, limit)
}

func (m *ProcessManager) AllLogsSnapshot(limit int) ([]LogLine, uint64) {
	return m.logs.AllSnapshot(limit)
}

func (m *ProcessManager) AppendLog(id, text string) {
	m.logs.Append(id, "[管理器] "+text)
}
