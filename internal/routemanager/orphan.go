package routemanager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

type OrphanProcess struct {
	PID           int32
	StartedAt     int64
	Executable    string
	CommandLine   string
	CleanupAction string
}

func ScanOrphanProcesses(binary string, owned map[int32]struct{}) ([]OrphanProcess, error) {
	if !isManagedBackendPath(binary) {
		return nil, nil
	}
	managerExe, _ := os.Executable()
	processes, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("扫描 gost-qt 进程: %w", err)
	}
	orphans := make([]OrphanProcess, 0)
	for _, candidate := range processes {
		if _, ok := owned[candidate.Pid]; ok || !isManagedBackendProcess(candidate) {
			continue
		}
		if hasLiveManagerParent(candidate, managerExe) {
			continue
		}
		startedAt, err := candidate.CreateTime()
		if err != nil || startedAt <= 0 {
			continue
		}
		executable, _ := candidate.Exe()
		argv, _ := candidate.CmdlineSlice()
		orphans = append(orphans, OrphanProcess{
			PID: candidate.Pid, StartedAt: startedAt,
			Executable: executable, CommandLine: FormatCommand(argv),
			CleanupAction: orphanCleanupAction(candidate),
		})
	}
	return orphans, nil
}

func CleanupOrphanProcesses(targets []OrphanProcess) error {
	var errs []error
	for _, target := range targets {
		candidate, err := process.NewProcess(target.PID)
		if err != nil {
			continue
		}
		startedAt, err := candidate.CreateTime()
		if err != nil || startedAt != target.StartedAt || !isManagedBackendProcess(candidate) {
			continue
		}
		if err := cleanupOrphanProcess(candidate); err != nil && !errors.Is(err, os.ErrProcessDone) {
			errs = append(errs, fmt.Errorf("PID %d: %w", target.PID, err))
		}
	}
	return errors.Join(errs...)
}

func isManagedBackendPath(path string) bool {
	name := strings.ToLower(filepath.Base(strings.ReplaceAll(path, `\`, "/")))
	return name == ManagedBackendName || name == ManagedBackendName+".exe"
}

func isManagedBackendProcess(candidate *process.Process) bool {
	name, err := candidate.Name()
	return err == nil && isManagedBackendPath(name)
}

func hasLiveManagerParent(candidate *process.Process, managerExe string) bool {
	parent, err := candidate.Parent()
	if err != nil || parent == nil {
		return false
	}
	parentExe, err := parent.Exe()
	if err != nil {
		return false
	}
	return sameExecutablePath(parentExe, managerExe)
}

func sameExecutablePath(left, right string) bool {
	left, leftErr := filepath.Abs(left)
	right, rightErr := filepath.Abs(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(left); err == nil {
		left = resolved
	}
	if resolved, err := filepath.EvalSymlinks(right); err == nil {
		right = resolved
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(left), filepath.Clean(right))
	}
	return filepath.Clean(left) == filepath.Clean(right)
}
