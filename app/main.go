package main

import (
	"context"
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "GoClip - 本地加密剪贴板",
		Width:             960,
		Height:            680,
		MinWidth:          800,
		MinHeight:         520,
		StartHidden:       app.isSilentStart,
		HideWindowOnClose: true, // 点击关闭按钮自动隐藏至后台静默运行，绝不退出程序
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 15, G: 23, B: 42, A: 1}, // 与前端统一的深色底色 #0f172a
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			app.HideWindow()
			return true // 阻止真正退出，转入静默后台常驻监听
		},
		Bind: []interface{}{
			app,
		},
		// macOS 专属优化：内嵌红绿灯按钮，消除系统灰色标题栏色差
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "GoClip 剪贴板",
				Message: "本地 AES-256 加密存储，支持快捷键唤醒与系统菜单栏静默常驻",
			},
		},
		// Windows 专属优化：启用 Windows 10/11 原生深色暗黑主题标题栏与原生缩放拖拽
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			Theme:                windows.Dark, // Windows 原生深色主题，同色系一体化
		},
	})

	if err != nil {
		println("启动失败:", err.Error())
	}
}
