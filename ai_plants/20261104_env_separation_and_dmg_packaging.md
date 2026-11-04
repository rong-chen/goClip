# 修改记录

## 1. 生产环境与开发环境数据隔离
在 storage 模块中新增环境判定与目录分发逻辑：
- 开发环境统一使用用户主目录下的 .goclip-dev 目录，避免本地调试时产生的历史记录污染正式环境。
- 生产环境统一使用用户主目录下的 .goclip 目录。
- 通过 Go 条件编译标签 dev 以及环境变量 APP_ENV 与 WAILS_ENV 共同识别当前模式。
- 修改相关文件：
  - app/storage/env.go
  - app/storage/env_dev.go
  - app/storage/env_prod.go
  - app/storage/db.go
  - app/storage/db_test.go
  - app/frontend/src/views/SettingsView.vue

## 2. GitHub Actions 流水线增加 macOS DMG 镜像打包
在 .github/workflows/build.yml 的 build-macos 作业中增加 DMG 打包步骤：
- 利用 macOS 原生 hdiutil 命令将编译生成的 GoClip.app 封装为可直接拖拽安装的 GoClip-macOS-Universal.dmg 文件。
- 将生成的 DMG 安装镜像同步加入 artifact 与 release 挂载附件中。
