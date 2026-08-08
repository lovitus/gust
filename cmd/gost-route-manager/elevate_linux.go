//go:build linux

package main

import (
	"errors"
	"os"
	"os/exec"
)

func elevateSelf(args []string) error {
	pkexec, err := exec.LookPath("pkexec")
	if err != nil {
		return errors.New("系统未安装 pkexec，无法弹出图形化提权窗口；也可以用 sudo 启动本程序")
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	// pkexec intentionally starts with a restricted environment. Pass only the
	// desktop-session variables needed by the elevated Fyne window.
	command := []string{"env"}
	for _, name := range []string{"DISPLAY", "XAUTHORITY", "WAYLAND_DISPLAY", "XDG_RUNTIME_DIR", "DBUS_SESSION_BUS_ADDRESS"} {
		if value := os.Getenv(name); value != "" {
			command = append(command, name+"="+value)
		}
	}
	command = append(command, exe)
	command = append(command, args...)
	return exec.Command(pkexec, command...).Start()
}
