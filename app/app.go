package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"app/clipboard"
	"app/cloudsync"
	"app/storage"
	"app/system"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 结构体，管理应用程序的生命周期与前端交互方法
type App struct {
	ctx             context.Context
	storage         *storage.Storage
	monitor         *clipboard.Monitor
	syncManager     *cloudsync.SyncManager
	hotkeyMgr       *system.HotkeyManager
	mu              sync.Mutex
	isWindowVisible bool
	isSilentStart   bool
}

// NewApp 初始化 App
func NewApp() *App {
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	syncMgr := cloudsync.NewSyncManager()

	// 检查命令行是否包含 --silent 参数
	silentArg := false
	for _, arg := range os.Args {
		if strings.ToLower(arg) == "--silent" {
			silentArg = true
			break
		}
	}

	a := &App{
		storage:         store,
		syncManager:     syncMgr,
		isWindowVisible: !silentArg,
		isSilentStart:   silentArg,
	}

	// 初始化热键管理器，绑定全局唤醒切换
	a.hotkeyMgr = system.NewHotkeyManager(func() {
		a.ToggleWindow()
	})

	// 初始化剪贴板监听器：每当有新剪贴板变动，通过 Wails 事件系统推送到前端
	a.monitor = clipboard.NewMonitor(store, func(item *storage.ClipboardItem) {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "clipboard:changed", item)
		}
	})

	return a
}

// startup 在 Wails 启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 启动剪贴板监听器
	if err := a.monitor.Start(); err != nil {
		log.Printf("启动剪贴板监听器异常: %v", err)
	}

	// 读取应用设置并注册全局快捷键
	settings := a.storage.GetAppSettings(system.DefaultHotkey())
	if err := a.hotkeyMgr.RegisterKey(settings.Hotkey); err != nil {
		log.Printf("初始化全局快捷键提示: %v", err)
	}

	// 初始化系统顶部状态栏托盘图标（Mac 顶部 Wi-Fi 行白色剪贴板图标）
	system.InitTray(system.TrayHandler{
		OnToggle: func() {
			a.ToggleWindow()
		},
		OnSettings: func() {
			a.ShowWindow()
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "route:settings")
			}
		},
		OnQuit: func() {
			if a.ctx != nil {
				wailsRuntime.Quit(a.ctx)
			}
		},
	})

	// 静默模式启动处理：若带 --silent 启动，则开机不弹出窗口，直接后台常驻监听
	if a.isSilentStart {
		wailsRuntime.WindowHide(ctx)
		wailsRuntime.Hide(ctx)
		a.isWindowVisible = false
	} else {
		a.isWindowVisible = true
	}
}

// ShowWindow 强制在前台展示窗口并获得焦点
func (a *App) ShowWindow() {
	if a.ctx == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowUnminimise(a.ctx)
	wailsRuntime.Show(a.ctx)
	a.isWindowVisible = true
}

// shutdown 在程序关闭时调用
func (a *App) shutdown(ctx context.Context) {
	if a.hotkeyMgr != nil {
		a.hotkeyMgr.Unregister()
	}
	if a.monitor != nil {
		a.monitor.Stop()
	}
	if a.storage != nil {
		a.storage.Close()
	}
}

// ToggleWindow 全局热键触发或前端调用：切换主窗口的显示与隐藏
func (a *App) ToggleWindow() {
	if a.ctx == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isWindowVisible {
		wailsRuntime.WindowHide(a.ctx)
		wailsRuntime.Hide(a.ctx)
		a.isWindowVisible = false
	} else {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
		wailsRuntime.Show(a.ctx)
		a.isWindowVisible = true
	}
}

// HideWindow 最小化/隐藏窗口至后台
func (a *App) HideWindow() {
	if a.ctx == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	wailsRuntime.WindowHide(a.ctx)
	wailsRuntime.Hide(a.ctx)
	a.isWindowVisible = false
}

// GetAppSettings 获取应用偏好设置（开机自启、静默模式、全局快捷键）
func (a *App) GetAppSettings() storage.AppSettings {
	settings := a.storage.GetAppSettings(system.DefaultHotkey())
	// 同步检查系统真实的自启动状态
	systemAuto := system.IsAutoStartEnabled()
	if systemAuto != settings.AutoStart {
		settings.AutoStart = systemAuto
		_ = a.storage.SaveAppSettings(settings)
	}
	return settings
}

// SaveAppSettings 保存应用偏好设置并应用系统级更改
func (a *App) SaveAppSettings(settings storage.AppSettings) error {
	// 1. 持久化存储到 SQLite settings 表
	if err := a.storage.SaveAppSettings(settings); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	// 2. 配置系统级开机自启 (macOS LaunchAgent / Windows Registry / Linux Desktop)
	if err := system.SetAutoStart(settings.AutoStart, settings.SilentMode); err != nil {
		log.Printf("配置开机自启提示: %v", err)
	}

	// 3. 动态重新注册全局唤醒快捷键
	if err := a.hotkeyMgr.RegisterKey(settings.Hotkey); err != nil {
		return fmt.Errorf("快捷键生效提示: %w", err)
	}

	return nil
}

// GetItems 查询剪贴板历史列表并解密
func (a *App) GetItems(search string, filterType string, pinnedOnly bool, limit int, offset int) ([]*storage.ClipboardItem, error) {
	return a.storage.GetItems(search, filterType, pinnedOnly, limit, offset)
}

// WriteText 将纯文本写回系统剪贴板
func (a *App) WriteText(text string) error {
	return a.monitor.WriteText(text)
}

// WriteItem 将指定 ID 的记录（文本或图片）写回系统剪贴板
func (a *App) WriteItem(id string) error {
	itemType, raw, err := a.storage.GetItemRawData(id)
	if err != nil {
		return fmt.Errorf("读取数据失败: %w", err)
	}

	if itemType == "text" {
		return a.monitor.WriteText(string(raw))
	} else if itemType == "image" {
		return a.monitor.WriteImage(raw)
	}

	return fmt.Errorf("未知条目类型: %s", itemType)
}

// TogglePin 切换条目置顶
func (a *App) TogglePin(id string) (bool, error) {
	return a.storage.TogglePin(id)
}

// DeleteItem 软删除单条记录
func (a *App) DeleteItem(id string) error {
	return a.storage.DeleteItem(id)
}

// ClearHistory 清空历史记录
func (a *App) ClearHistory(keepPinned bool) error {
	return a.storage.ClearHistory(keepPinned)
}

// OpenImageInSystemViewer 解密图片并写入临时缓存，调用操作系统默认图片查看器原生打开
func (a *App) OpenImageInSystemViewer(id string) error {
	itemType, raw, err := a.storage.GetItemRawData(id)
	if err != nil {
		return fmt.Errorf("读取图片数据失败: %w", err)
	}
	if itemType != "image" {
		return fmt.Errorf("该条目不是图片")
	}

	tempDir := filepath.Join(os.TempDir(), "goclip_preview")
	_ = os.MkdirAll(tempDir, 0700)

	ext := ".png"
	if len(raw) > 3 && raw[0] == 0xFF && raw[1] == 0xD8 && raw[2] == 0xFF {
		ext = ".jpg"
	} else if len(raw) > 6 && (string(raw[:6]) == "GIF87a" || string(raw[:6]) == "GIF89a") {
		ext = ".gif"
	}

	filePath := filepath.Join(tempDir, fmt.Sprintf("clip_%s%s", id, ext))
	if err := os.WriteFile(filePath, raw, 0600); err != nil {
		return fmt.Errorf("写入临时图片文件失败: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}

	return cmd.Start()
}

// 兼容接口桩（避免前端旧绑定报错）
func (a *App) GetSecuritySettings() storage.SecuritySettings {
	return a.storage.GetSecuritySettings()
}
func (a *App) SetMasterPassword(password string) error { return nil }
func (a *App) UnlockVault(password string) error        { return nil }
func (a *App) LockVault()                               {}
func (a *App) GetCloudSyncConfig() storage.CloudSyncConfig {
	return a.syncManager.GetConfig()
}
func (a *App) SaveCloudSyncConfig(cfg storage.CloudSyncConfig) {}
func (a *App) TestCloudSync() (*cloudsync.SyncResult, error) {
	return a.syncManager.TestSync(a.ctx)
}
