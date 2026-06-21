//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// Windowsでコマンド実行時の黒い画面を非表示にする
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
