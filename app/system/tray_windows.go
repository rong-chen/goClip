//go:build windows

package system

import (
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

//go:embed tray_icon_windows.ico
var trayIcoData []byte

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	registerClassExW     = user32.NewProc("RegisterClassExW")
	createWindowExW      = user32.NewProc("CreateWindowExW")
	defWindowProcW       = user32.NewProc("DefWindowProcW")
	destroyWindow        = user32.NewProc("DestroyWindow")
	postQuitMessage      = user32.NewProc("PostQuitMessage")
	getMessageW          = user32.NewProc("GetMessageW")
	translateMessage     = user32.NewProc("TranslateMessage")
	dispatchMessageW     = user32.NewProc("DispatchMessageW")
	createPopupMenu      = user32.NewProc("CreatePopupMenu")
	appendMenuW          = user32.NewProc("AppendMenuW")
	trackPopupMenu       = user32.NewProc("TrackPopupMenu")
	destroyMenu          = user32.NewProc("DestroyMenu")
	setForegroundWindow  = user32.NewProc("SetForegroundWindow")
	getCursorPos         = user32.NewProc("GetCursorPos")
	postMessageW         = user32.NewProc("PostMessageW")
	loadImageW           = user32.NewProc("LoadImageW")
	loadIconW            = user32.NewProc("LoadIconW")
	getModuleHandleW     = kernel32.NewProc("GetModuleHandleW")
	shellNotifyIconW     = shell32.NewProc("Shell_NotifyIconW")

	winTrayMu      sync.Mutex
	winTrayHandler TrayHandler
)

const (
	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	WM_DESTROY       = 0x0002
	WM_COMMAND       = 0x0111
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_USER          = 0x0400
	WM_TRAY          = WM_USER + 100

	IMAGE_ICON       = 1
	LR_LOADFROMFILE  = 0x0010
	LR_DEFAULTSIZE   = 0x0040
	IDI_APPLICATION  = 32512

	TPM_BOTTOMALIGN = 0x0020
	TPM_LEFTALIGN   = 0x0000

	MF_STRING    = 0x00000000
	MF_SEPARATOR = 0x00000800

	ID_TRAY_SHOW     = 2001
	ID_TRAY_SETTINGS = 2002
	ID_TRAY_QUIT     = 2003
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

func loadTrayIcon() uintptr {
	if len(trayIcoData) > 0 {
		tempIco := filepath.Join(os.TempDir(), "goclip_tray.ico")
		if err := os.WriteFile(tempIco, trayIcoData, 0644); err == nil {
			pathPtr, err := syscall.UTF16PtrFromString(tempIco)
			if err == nil {
				hIcon, _, _ := loadImageW.Call(
					0,
					uintptr(unsafe.Pointer(pathPtr)),
					IMAGE_ICON,
					0,
					0,
					LR_LOADFROMFILE|LR_DEFAULTSIZE,
				)
				if hIcon != 0 {
					return hIcon
				}
			}
		}
	}
	// Fallback 使用应用程序默认图标
	hIcon, _, _ := loadIconW.Call(0, uintptr(IDI_APPLICATION))
	return hIcon
}

func winTrayProc(hWnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAY:
		switch lParam {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
			winTrayMu.Lock()
			fn := winTrayHandler.OnToggle
			winTrayMu.Unlock()
			if fn != nil {
				go fn()
			}
		case WM_RBUTTONUP:
			var pt POINT
			getCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
			setForegroundWindow.Call(hWnd)

			hMenu, _, _ := createPopupMenu.Call()
			if hMenu != 0 {
				titleStr, _ := syscall.UTF16PtrFromString("显示主窗口")
				settingsStr, _ := syscall.UTF16PtrFromString("设置")
				quitStr, _ := syscall.UTF16PtrFromString("退出 GoClip")

				appendMenuW.Call(hMenu, MF_STRING, ID_TRAY_SHOW, uintptr(unsafe.Pointer(titleStr)))
				appendMenuW.Call(hMenu, MF_STRING, ID_TRAY_SETTINGS, uintptr(unsafe.Pointer(settingsStr)))
				appendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
				appendMenuW.Call(hMenu, MF_STRING, ID_TRAY_QUIT, uintptr(unsafe.Pointer(quitStr)))

				trackPopupMenu.Call(hMenu, TPM_BOTTOMALIGN|TPM_LEFTALIGN, uintptr(pt.X), uintptr(pt.Y), 0, hWnd, 0)
				postMessageW.Call(hWnd, 0, 0, 0)
				destroyMenu.Call(hMenu)
			}
		}
		return 0

	case WM_COMMAND:
		cmdID := wParam & 0xffff
		switch cmdID {
		case ID_TRAY_SHOW:
			winTrayMu.Lock()
			fn := winTrayHandler.OnToggle
			winTrayMu.Unlock()
			if fn != nil {
				go fn()
			}
		case ID_TRAY_SETTINGS:
			winTrayMu.Lock()
			fn := winTrayHandler.OnSettings
			winTrayMu.Unlock()
			if fn != nil {
				go fn()
			}
		case ID_TRAY_QUIT:
			winTrayMu.Lock()
			fn := winTrayHandler.OnQuit
			winTrayMu.Unlock()
			if fn != nil {
				go fn()
			}
		}
		return 0

	case WM_DESTROY:
		postQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := defWindowProcW.Call(hWnd, uintptr(msg), wParam, lParam)
	return ret
}

// InitTray 在 Windows 系统任务栏通知区域常驻 GoClip 托盘图标
func InitTray(handler TrayHandler) {
	winTrayMu.Lock()
	winTrayHandler = handler
	winTrayMu.Unlock()

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		hInstance, _, _ := getModuleHandleW.Call(0)
		className, _ := syscall.UTF16PtrFromString("GoClipTrayWindowClass")

		wcex := WNDCLASSEXW{
			CbSize:        uint32(unsafe.Sizeof(WNDCLASSEXW{})),
			LpfnWndProc:   syscall.NewCallback(winTrayProc),
			HInstance:     hInstance,
			LpszClassName: className,
		}
		registerClassExW.Call(uintptr(unsafe.Pointer(&wcex)))

		windowName, _ := syscall.UTF16PtrFromString("GoClipTray")
		hWnd, _, _ := createWindowExW.Call(
			0,
			uintptr(unsafe.Pointer(className)),
			uintptr(unsafe.Pointer(windowName)),
			0,
			0, 0, 0, 0,
			0, 0, hInstance, 0,
		)

		if hWnd == 0 {
			return
		}

		hIcon := loadTrayIcon()

		var nid NOTIFYICONDATAW
		nid.CbSize = uint32(unsafe.Sizeof(nid))
		nid.HWnd = hWnd
		nid.UID = 1
		nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
		nid.UCallbackMessage = WM_TRAY
		nid.HIcon = hIcon

		tipStr := "GoClip - 剪贴板历史"
		tipChars, _ := syscall.UTF16FromString(tipStr)
		copy(nid.SzTip[:], tipChars)

		shellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
		defer shellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))

		// 运行消息循环
		var msg MSG
		for {
			r, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
			if int32(r) <= 0 {
				break
			}
			translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
			dispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
		}
	}()
}
