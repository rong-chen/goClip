//go:build !darwin

package system

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"golang.design/x/hotkey"
)

type otherPlatformHotkey struct {
	mu           sync.Mutex
	currentHK    *hotkey.Hotkey
	cancelFunc   context.CancelFunc
	onToggleShow HotkeyCallback
}

func newPlatformHotkey(cb HotkeyCallback) platformHotkeyImpl {
	return &otherPlatformHotkey{
		onToggleShow: cb,
	}
}

func (p *otherPlatformHotkey) Register(keyStr string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	mods, key, err := parseHotkey(keyStr)
	if err != nil {
		return err
	}

	if p.currentHK != nil {
		if p.cancelFunc != nil {
			p.cancelFunc()
		}
		_ = p.currentHK.Unregister()
		p.currentHK = nil
	}

	hk := hotkey.New(mods, key)
	if err := hk.Register(); err != nil {
		return fmt.Errorf("快捷键注册失败 (可能已被系统或其他软件占用): %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.currentHK = hk
	p.cancelFunc = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-hk.Keydown():
				if p.onToggleShow != nil {
					p.onToggleShow()
				}
			}
		}
	}()

	log.Printf("成功注册系统全局唤醒快捷键: %s", keyStr)
	return nil
}

func (p *otherPlatformHotkey) Unregister() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
		p.cancelFunc = nil
	}
	if p.currentHK != nil {
		_ = p.currentHK.Unregister()
		p.currentHK = nil
	}
}

func parseHotkey(s string) ([]hotkey.Modifier, hotkey.Key, error) {
	parts := strings.Split(s, "+")
	var mods []hotkey.Modifier
	var key hotkey.Key
	hasKey := false

	for _, part := range parts {
		token := strings.ToUpper(strings.TrimSpace(part))
		switch token {
		case "CMD", "COMMAND":
			// 在 Windows 上将 Cmd 智能兼容为 Ctrl，避免与系统 Win 键全局冲突
			mods = appendUniqueMod(mods, hotkey.ModCtrl)
		case "WIN", "WINDOWS", "SUPER":
			mods = appendUniqueMod(mods, hotkey.ModWin)
		case "CTRL", "CONTROL":
			mods = appendUniqueMod(mods, hotkey.ModCtrl)
		case "SHIFT":
			mods = appendUniqueMod(mods, hotkey.ModShift)
		case "ALT", "OPTION", "OPT":
			mods = appendUniqueMod(mods, hotkey.ModAlt)
		default:
			k, err := parseKey(token)
			if err != nil {
				return nil, 0, err
			}
			key = k
			hasKey = true
		}
	}

	if !hasKey {
		return nil, 0, fmt.Errorf("快捷键必须包含一个主按键（如字母、数字或功能键）: %s", s)
	}

	return mods, key, nil
}

func appendUniqueMod(mods []hotkey.Modifier, mod hotkey.Modifier) []hotkey.Modifier {
	for _, m := range mods {
		if m == mod {
			return mods
		}
	}
	return append(mods, mod)
}

func parseKey(k string) (hotkey.Key, error) {
	if len(k) == 1 {
		char := k[0]
		switch {
		case char >= 'A' && char <= 'Z':
			mapping := []hotkey.Key{
				hotkey.KeyA, hotkey.KeyB, hotkey.KeyC, hotkey.KeyD, hotkey.KeyE,
				hotkey.KeyF, hotkey.KeyG, hotkey.KeyH, hotkey.KeyI, hotkey.KeyJ,
				hotkey.KeyK, hotkey.KeyL, hotkey.KeyM, hotkey.KeyN, hotkey.KeyO,
				hotkey.KeyP, hotkey.KeyQ, hotkey.KeyR, hotkey.KeyS, hotkey.KeyT,
				hotkey.KeyU, hotkey.KeyV, hotkey.KeyW, hotkey.KeyX, hotkey.KeyY, hotkey.KeyZ,
			}
			return mapping[char-'A'], nil
		case char >= '0' && char <= '9':
			mapping := []hotkey.Key{
				hotkey.Key0, hotkey.Key1, hotkey.Key2, hotkey.Key3, hotkey.Key4,
				hotkey.Key5, hotkey.Key6, hotkey.Key7, hotkey.Key8, hotkey.Key9,
			}
			return mapping[char-'0'], nil
		}
	}

	switch k {
	case "SPACE":
		return hotkey.KeySpace, nil
	case "RETURN", "ENTER":
		return hotkey.KeyReturn, nil
	case "ESCAPE", "ESC":
		return hotkey.KeyEscape, nil
	case "F1":
		return hotkey.KeyF1, nil
	case "F2":
		return hotkey.KeyF2, nil
	case "F3":
		return hotkey.KeyF3, nil
	case "F4":
		return hotkey.KeyF4, nil
	case "F5":
		return hotkey.KeyF5, nil
	case "F6":
		return hotkey.KeyF6, nil
	case "F7":
		return hotkey.KeyF7, nil
	case "F8":
		return hotkey.KeyF8, nil
	case "F9":
		return hotkey.KeyF9, nil
	case "F10":
		return hotkey.KeyF10, nil
	case "F11":
		return hotkey.KeyF11, nil
	case "F12":
		return hotkey.KeyF12, nil
	default:
		return 0, fmt.Errorf("不支持的按键: %s", k)
	}
}
