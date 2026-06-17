//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

func getWindowsLocale() string {
	// GetUserDefaultLocaleName を呼び出して "ja-JP" などの文字列を取得する
	mod := syscall.NewLazyDLL("kernel32.dll")
	proc := mod.NewProc("GetUserDefaultLocaleName")
	if proc.Find() != nil {
		return ""
	}

	// LOCALE_NAME_MAX_LENGTH = 85
	var buf [85]uint16
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:])
}
