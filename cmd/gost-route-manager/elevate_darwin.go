//go:build darwin

package main

import (
	"os"
	"os/exec"
	"strings"
)

func elevateSelf(args []string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	parts := append([]string{exe}, args...)
	for i := range parts {
		parts[i] = shellQuote(parts[i])
	}
	// Background the elevated process so osascript returns after launch.
	command := strings.Join(parts, " ") + " >/dev/null 2>&1 &"
	script := `do shell script ` + appleScriptQuote(command) + ` with administrator privileges`
	return exec.Command("osascript", "-e", script).Start()
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func appleScriptQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}
