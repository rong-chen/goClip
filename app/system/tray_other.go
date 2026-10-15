//go:build !darwin

package system

// InitTray 非 macOS 平台的系统托盘桩
func InitTray(handler TrayHandler) {
	// 在 Windows/Linux 上预留支持
}
