//go:build darwin

package system

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

extern void initMacTray(const void *bytes, int len);
extern int isMacWindowFrontAndActive();
*/
import "C"
import (
	"sync"
	"unsafe"
)

var (
	trayMu      sync.Mutex
	trayHandler TrayHandler
)

//export goTrayToggleWindow
func goTrayToggleWindow() {
	trayMu.Lock()
	fn := trayHandler.OnToggle
	trayMu.Unlock()
	if fn != nil {
		fn()
	}
}

//export goTrayShowWindow
func goTrayShowWindow() {
	trayMu.Lock()
	fn := trayHandler.OnShow
	trayMu.Unlock()
	if fn != nil {
		fn()
	}
}

//export goTrayHideWindow
func goTrayHideWindow() {
	trayMu.Lock()
	fn := trayHandler.OnHide
	trayMu.Unlock()
	if fn != nil {
		fn()
	}
}

//export goTrayOpenSettings
func goTrayOpenSettings() {
	trayMu.Lock()
	fn := trayHandler.OnSettings
	trayMu.Unlock()
	if fn != nil {
		fn()
	}
}

//export goTrayQuitApp
func goTrayQuitApp() {
	trayMu.Lock()
	fn := trayHandler.OnQuit
	trayMu.Unlock()
	if fn != nil {
		fn()
	}
}

// IsWindowActive 检查 macOS 下当前应用主窗口是否处于前台活跃展示状态
func IsWindowActive() bool {
	return int(C.isMacWindowFrontAndActive()) != 0
}

// InitTray 在 macOS 顶部菜单栏状态栏常驻纯白剪贴板图标
func InitTray(handler TrayHandler) {
	trayMu.Lock()
	trayHandler = handler
	trayMu.Unlock()

	if len(TrayIconData) > 0 {
		C.initMacTray(unsafe.Pointer(&TrayIconData[0]), C.int(len(TrayIconData)))
	}
}
