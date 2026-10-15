# GoClip - 跨平台本地加密剪贴板管理器

GoClip 是一款基于 **Go 跨平台桌面 GUI（Wails） + Vue 3 + Pinia + SQLite + 本地 AES 密钥库** 实现的高颜值现代化剪贴板历史管理工具。

纯单机运行，**无需远程后端服务器**，数据全部落地在本地 SQLite 数据库中，并通过 **AES-256-GCM** 全文强加密存储。支持文本与图片的无感知复制捕获、搜索筛选、大图预览、一键回写系统剪贴板，并全面预留了未来多端云同步架构。

---

## 🌟 核心特性

1. **跨平台桌面原生体验 (Cross-Platform GUI)**
   - 适配 **macOS、Windows、Linux** 三大主流桌面操作系统。
   - 前端采用 **Vue 3 + Pinia + Vite** 构建，拥有现代毛玻璃质感、流畅微动画与卡片式布局。

2. **本地单机 SQLite + 强加密模式 (Local AES-256-GCM Vault)**
   - **零远程后端**：纯本地数据库 `~/.goclip/clipboard.db`，无数据外泄风险。
   - **透明硬件密钥 / 主密码保护**：
     - 默认模式：自动生成 256 位本地安全设备密钥，入库的每条文本、每张图片均以密文保存，即使数据库文件被拷走也无法明文解密。
     - 安全锁模式：支持用户设置主密码（Master Password），通过 **PBKDF2（SHA-256，100,000 次哈希迭代）** 派生主密钥，支持随时加锁/解锁。

3. **文本与图片全格式支持 (Text & Image)**
   - **智能剪贴板监听**：后台自动捕获 `Cmd+C` / `Ctrl+C` 复制动作。
   - **图片支持**：自动捕获图片位图并转为 PNG 加密保存，列表展示分辨率与缩略图，支持点击全屏高清查看及一键复制回剪贴板。
   - **防循环录入（Self-Write Suppression）**：从应用内点击复制回剪贴板时，自动识别自身哈希，避免重复录入。
   - **自动去重与修剪**：相同内容自动提升为最新时间戳，超额记录自动清理并永久保护置顶项。

4. **预留后期云同步架构 (Future Cloud Sync Ready)**
   - 数据表已设计 **`sync_id`、`device_id`、`sync_status`（未同步/已同步/冲突）、`version`（乐观锁版本号）与 `is_deleted`（软删除机制）**。
   - 代码中抽象封装了 `CloudSyncProvider` 与预留配置界面（支持云端端点配置、Token 认证、E2EE 端到端零知识加密），后期开启云端服务即可无缝扩展。

---

## 📁 项目目录结构 (`app/`)

```
app/
├── main.go                     # 应用启动入口，配置 Wails 跨平台桌面窗口与系统托盘
├── app.go                      # Go 后端与 Vue3 前端桥接服务（剪贴板读写、本地加密、云同步预留方法）
├── clipboard/                  # 跨平台剪贴板监听器
│   └── monitor.go              # 文本与图片监听、自写回抑制哈希、事件推送
├── crypto/                     # 本地加密密钥库
│   └── crypto.go               # AES-256-GCM 加解密、PBKDF2 密码派生、随机 Salt 与 SHA-256 查重
├── storage/                    # 本地数据持久化
│   ├── db.go                   # SQLite (现代纯 Go 驱动 modernc.org/sqlite) 存取、去重与软删除
│   └── model.go                # 剪贴板条目、安全配置与云同步扩展模型
├── cloudsync/                  # 预留云端同步模块
│   └── sync.go                 # 多端同步协议接口、状态与连接测试桩
├── frontend/                   # Vue 3 + Pinia 现代化前端
│   ├── src/
│   │   ├── App.vue             # 主界面布局与实时事件订阅
│   │   ├── main.js             # Vue 3 与 Pinia 挂载
│   │   ├── style.css           # 科技感深色毛玻璃样式系统
│   │   ├── stores/             # Pinia 状态管理
│   │   │   ├── clipboard.js    # 剪贴板历史、筛选、搜索、回写与删除
│   │   │   ├── security.js     # 本地主密码、加锁/解锁状态管理
│   │   │   └── sync.js         # 预留云同步配置与测试
│   │   └── components/         # 界面组件
│   │       ├── HeaderBar.vue   # 顶部导航、搜索栏、密码锁与云端入口
│   │       ├── FilterBar.vue   # 全部/文本/图片/置顶过滤
│   │       ├── ClipboardItemCard.vue # 记录卡片（文本预览、图片缩略图、快捷操作）
│   │       ├── ImageModal.vue  # 图片高清全屏预览弹窗
│   │       ├── SecurityModal.vue # 保险库主密码设置与解锁弹窗
│   │       └── CloudSyncModal.vue# 云端同步预留配置面板
│   ├── package.json
│   └── vite.config.js
└── go.mod                      # Go 模块定义
```

---

## 🚀 启动与构建指南

### 1. 启动开发模式（实时热重载）
在 `app` 目录下执行：
```bash
wails dev
```
前端 Vue 代码修改会自动热重载（HMR），Go 代码修改会自动触发重新编译。

### 2. 打包为生产应用
```bash
# 构建当前平台可执行文件
wails build

# 生成精简产物
wails build -clean
```
- **macOS**：生成产物位于 `build/bin/app.app`
- **Windows**：生成产物位于 `build/bin/app.exe`
- **Linux**：生成产物位于 `build/bin/app`
