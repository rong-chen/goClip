//go:build darwin

package system

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Carbon

extern int registerCarbonHotKey(unsigned int keyCode, unsigned int modifiers);
extern void unregisterCarbonHotKey();
*/
import "C"
import (
	"fmt"
	"log"
	"strings"
	"sync"
)

var (
	darwinCallback   HotkeyCallback
	darwinCallbackMu sync.Mutex
)

//export goHotkeyTriggered
func goHotkeyTriggered() {
	darwinCallbackMu.Lock()
	cb := darwinCallback
	darwinCallbackMu.Unlock()

	if cb != nil {
		cb()
	}
}

type darwinPlatformHotkey struct {
	mu sync.Mutex
}

func newPlatformHotkey(cb HotkeyCallback) platformHotkeyImpl {
	darwinCallbackMu.Lock()
	darwinCallback = cb
	darwinCallbackMu.Unlock()
	return &darwinPlatformHotkey{}
}

func (p *darwinPlatformHotkey) Register(keyStr string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	keyCode, mods, err := parseCarbonHotkey(keyStr)
	if err != nil {
		return err
	}

	res := C.registerCarbonHotKey(C.uint(keyCode), C.uint(mods))
	if res != 0 {
		return fmt.Errorf("Carbon 注册快捷键失败，错误码: %d", res)
	}

	log.Printf("macOS 原生 Carbon 热键注册成功（免辅助功能权限）: %s", keyStr)
	return nil
}

func (p *darwinPlatformHotkey) Unregister() {
	p.mu.Lock()
	defer p.mu.Unlock()
	C.unregisterCarbonHotKey()
}

// parseCarbonHotkey 解析类似 Cmd+Shift+V
func parseCarbonHotkey(s string) (uint32, uint32, error) {
	parts := strings.Split(s, "+")
	var mods uint32
	var keyCode uint32
	hasKey := false

	// Carbon 修饰键常量定义
	const (
		cmdKey     uint32 = 0x0100
		shiftKey   uint32 = 0x0200
		optionKey  uint32 = 0x0800
		controlKey uint32 = 0x1000
	)

	for _, part := range parts {
		token := strings.ToUpper(strings.TrimSpace(part))
		switch token {
		case "CMD", "COMMAND":
			mods |= cmdKey
		case "CTRL", "CONTROL":
			mods |= controlKey
		case "SHIFT":
			mods |= shiftKey
		case "ALT", "OPTION", "OPT":
			mods |= optionKey
		default:
			code, err := getMacKeyCode(token)
			if err != nil {
				return 0, 0, err
			}
			keyCode = code
			hasKey = true
		}
	}

	if !hasKey {
		return 0, 0, fmt.Errorf("快捷键必须包含一个按键，如 V: %s", s)
	}

	return keyCode, mods, nil
}

// 常见 Mac 虚拟键码对照表 (HIToolbox/Events.h)
func getMacKeyCode(k string) (uint32, error) {
	keyMap := map[string]uint32{
		"A": 0x00, "S": 0x01, "D": 0x02, "F": 0x03, "H": 0x04, "G": 0x05, "Z": 0x06, "X": 0x07,
		"C": 0x08, "V": 0x09, "B": 0x0B, "Q": 0x0C, "W": 0x0D, "E": 0x0E, "R": 0x0F, "Y": 0x10,
		"T": 0x11, "1": 0x12, "2": 0x13, "3": 0x14, "4": 0x15, "6": 0x16, "5": 0x17, "9": 0x19,
		"7": 0x1A, "8": 0x1C, "0": 0x1D, "O": 0x1F, "U": 0x20, "I": 0x22, "P": 0x23, "L": 0x25,
		"J": 0x26, "K": 0x28, "M": 0x2E, "N": 0x2D,
		"SPACE": 0x31, "RETURN": 0x24, "ENTER": 0x24, "TAB": 0x30,
		"F1": 0x7A, "F2": 0x78, "F3": 0x63, "F4": 0x76, "F5": 0x60, "F6": 0x61,
		"F7": 0x62, "F8": 0x64, "F9": 0x65, "F10": 0x6D, "F11": 0x67, "F12": 0x6F,
	}

	if code, ok := keyMap[k]; ok {
		return code, nil
	}
	return 0, fmt.Errorf("不支持的 macOS 按键: %s", k)
}
