//go:build !windows

package main

import "os/exec"

// Windows以外では何もしない
func hideConsoleWindow(cmd *exec.Cmd) {
	// No-op
}

// Linux等での実行ファイル名
const vectraceCmd = "./vectrace"
