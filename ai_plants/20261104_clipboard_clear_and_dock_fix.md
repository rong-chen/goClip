# 剪贴板清空与窗口唤醒退出修复记录

## 1. 剪贴板清空功能
- 后端在 monitor.go 中新增 ClearSystemClipboard 方法，写入空文本并清空历史哈希。
- 后端在 app.go 中扩展 ClearHistory 方法，并在前端触发清空时同步清理系统剪贴板。
- 前端 HeaderBar 增加自定义二次确认弹窗，支持保留置顶或全部清空两种操作。

## 2. 窗口唤醒与退出修复
- 移除 app/main.go 中的 OnBeforeClose 拦截，允许系统正常处理右键退出与关闭事件。
- 移除 app/app.go 中对系统 Hide 的调用，避免窗口陷入系统级强行隐藏状态。
- 在 tray_darwin.m 中增加 NSApplicationDidBecomeActiveNotification 监听，点击 Dock 图标时自动重新激活主窗口。

## 3. 顶栏按钮高度统一
- 将 HeaderBar 中的所有按钮包括图标按钮与文字状态按钮的高度统一设定为 32px，外框与内边距严格对齐，解决高低不一的问题。
