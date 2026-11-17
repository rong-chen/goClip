# 修复 macOS 下点击两次才显示窗口的生命周期问题

## 修改内容
- 在 app/system/tray.go 中为 TrayHandler 扩展 OnHide 回调。
- 在 app/system/tray_darwin.go 中新增 goTrayHideWindow 导出方法与 IsWindowActive 实时前台状态检测方法。
- 在 app/system/tray_other.go 与 app/system/tray_windows.go 中补充 IsWindowActive 跨平台接口定义。
- 在 app/system/tray_darwin.m 中新增 isMacWindowFrontAndActive 原生检测，判断窗口是否可见且处于前台活跃状态。
- 在 app/system/tray_darwin.m 中通过 NSAppleEventManager 注册 kAEReopenApplication 事件处理器，保证从 Dock 栏点击时直接唤醒展示窗口。
- 在 app/system/tray_darwin.m 中监听 NSApplicationDidHideNotification 与 NSWindowDidMiniaturizeNotification 通知，同步触发隐藏状态回调。
- 在 app/system/tray_darwin.m 中重构托盘点击与菜单切换逻辑，仅在窗口已处于前台活跃状态时执行隐藏，非活跃或隐藏状态下执行置顶展示。
- 在 app/app.go 中调整 ToggleWindow 与 HideWindow，在 macOS 下联动原生状态检测并同步处理应用级别隐藏与展示，避免状态失步导致需要点击两次。
