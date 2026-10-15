package system

import (
	"runtime"
	"strings"
	"sync"
)

type HotkeyCallback func()

type platformHotkeyImpl interface {
	Register(keyStr string) error
	Unregister()
}

type HotkeyManager struct {
	mu         sync.Mutex
	impl       platformHotkeyImpl
	currentKey string
}

func NewHotkeyManager(onToggle HotkeyCallback) *HotkeyManager {
	impl := newPlatformHotkey(onToggle)
	return &HotkeyManager{
		impl: impl,
	}
}

// DefaultHotkey 返回当前平台的推荐快捷键
func DefaultHotkey() string {
	if runtime.GOOS == "darwin" {
		return "Cmd+Shift+V"
	}
	return "Ctrl+Shift+V"
}

// RegisterKey 注册全局唤醒热键
func (m *HotkeyManager) RegisterKey(keyStr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	keyStr = strings.TrimSpace(keyStr)
	if keyStr == "" {
		keyStr = DefaultHotkey()
	}

	if err := m.impl.Register(keyStr); err != nil {
		return err
	}

	m.currentKey = keyStr
	return nil
}

// Unregister 注销当前热键
func (m *HotkeyManager) Unregister() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.impl != nil {
		m.impl.Unregister()
	}
}

// CurrentHotkey 获取当前注册的热键字符串
func (m *HotkeyManager) CurrentHotkey() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.currentKey == "" {
		return DefaultHotkey()
	}
	return m.currentKey
}
