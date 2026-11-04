# 修复系统托盘右键退出失效记录

## 1. 问题根因
- 在 main.go 中配置的 OnBeforeClose 回调无条件返回了 true 阻止关闭，导致托盘右键点击退出时触发的关闭流程被直接拦截，程序无法真正退出。
- 在 tray_darwin.m 中使用的 popUpStatusItemMenu 菜单呼出方法在新版本 macOS 容易与左键动作冲突。

## 2. 修复内容
- 托盘控制器与原生菜单弹出优化：
  - 在 app/system/tray_darwin.m 中优化右键点击判定，支持 Control 与左右键复合事件。
  - 在 onMenuQuit 中调用 goTrayQuitApp 并在短延迟后调度 NSApp terminate 执行原生终止。
- 退出流程放行：
  - 在 app/app.go 中为 App 结构体新增 isQuitting 状态标识与 Quit 显式退出方法。
  - 在 app/main.go 的 OnBeforeClose 回调中加入判定，若处于 isQuitting 状态则返回 false 允许正常退出，否则仅隐藏窗口常驻。
