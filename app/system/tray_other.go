//go:build !darwin && !windows

package system

// InitTray 非 macOS/Windows 平台（如 Linux）的系统托盘桩
func InitTray(handler TrayHandler) {
	// 在 Linux 上预留支持
}
