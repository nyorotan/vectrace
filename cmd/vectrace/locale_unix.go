//go:build !windows

package main

// getWindowsLocale は Windows 以外のプラットフォームでは空文字を返します。
func getWindowsLocale() string {
	return ""
}
