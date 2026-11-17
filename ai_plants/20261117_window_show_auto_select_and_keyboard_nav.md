# 支持唤醒自动选中首项、键盘上下键导航与回车自动复制

## 修改内容
- 在 app/app.go 中为 ShowWindow 与 ToggleWindow 增加 window:shown 事件通知，在窗口展示置顶时同步推送到前端。
- 在 app/frontend/src/stores/clipboard.js 中新增 selectedIndex 状态与 selectFirst、selectNext、selectPrev、setSelectedIndex、copySelected 等方法，并在搜索、过滤以及新数据到达时自动重置选中第一项。
- 在 app/frontend/src/components/ClipboardItemCard.vue 中支持 is-selected 属性与选中高亮边框和发光样式，点击卡片时同步更新选中状态。
- 在 app/frontend/src/components/HeaderBar.vue 中在窗口获得焦点或收到唤醒通知时自动聚焦搜索输入框。
- 在 app/frontend/src/App.vue 中监听 window:shown 与窗口获得焦点事件，触发首项选中与滚动定位；监听 ArrowDown 与 ArrowUp 执行上下导航，监听 Enter 执行选定项自动复制回剪贴板。
