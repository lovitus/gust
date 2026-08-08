//go:build !linux && !darwin && !windows

package main

import "errors"

func elevateSelf(args []string) error {
	return errors.New("当前平台不支持从 UI 提权，请以管理员身份启动")
}
