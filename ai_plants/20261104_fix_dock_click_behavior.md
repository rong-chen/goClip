# 修复点击 Dock 应用栏窗口隐藏与无响应问题

## 修改内容
- 在 app/system/tray.go 中为 TrayHandler 扩展 OnShow 纯展示回调，严格区分 OnToggle 切换逻辑与 OnShow 展示逻辑。
- 在 app/app.go 中将 ShowWindow 注册至 OnShow 回调，确保唤醒主窗口时仅执行展示与置顶，不执行隐藏。
- 在 app/system/tray_darwin.go 中调整 goTrayShowWindow 调用 OnShow。
- 在 app/system/tray_darwin.m 中通过运行时动态接入 applicationShouldHandleReopen:hasVisibleWindows: 代理方法，支持应用处于活跃状态时点击 Dock 栏直接唤醒并置顶主窗口。
- 在激活通知与重开代理中直接调用 makeKeyAndOrderFront 与 activateIgnoringOtherApps，保证任意次数点击应用栏均稳定唤醒显示窗口。
