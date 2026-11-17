//go:build !darwin && !windows

package system

// IsWindowActive 非 macOS 平台默认状态
func IsWindowActive() bool {
	return false
}

// InitTray 非 macOS/Windows 平台（如 Linux）的系统托盘桩
func InitTray(handler TrayHandler) {
	// 在 Linux 上预留支持
}
