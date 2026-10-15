<template>
  <div class="settings-view">
    <!-- 顶部返回导航 -->
    <div class="settings-header">
      <button class="back-btn" @click="$emit('back')">
        <ArrowLeft :size="16" />
        <span>返回剪贴板</span>
      </button>
      <h2 class="view-title">应用偏好与设置</h2>
    </div>

    <!-- 设置主要内容区 -->
    <div class="settings-content">
      <!-- 模块 1: 系统偏好与快捷键唤醒 -->
      <section class="settings-card">
        <div class="card-title-row">
          <Sliders :size="18" class="card-icon" />
          <div class="card-titles">
            <h3>系统偏好与全局快捷键</h3>
            <p>配置开机自启、后台静默监听与全局呼出热键</p>
          </div>
        </div>

        <div class="card-body preference-list">
          <!-- 项 1: 开机自动启动 -->
          <div class="pref-item">
            <div class="pref-left">
              <div class="pref-icon-box">
                <Power :size="16" />
              </div>
              <div class="pref-info">
                <span class="pref-title">开机自动启动</span>
                <span class="pref-desc">系统开机或登录后自动启动 GoClip 并开启剪贴板监听</span>
              </div>
            </div>
            <div class="pref-action">
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="appSettings.autoStart" 
                  @change="handleSavePreference" 
                />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <!-- 项 2: 静默模式 -->
          <div class="pref-item">
            <div class="pref-left">
              <div class="pref-icon-box">
                <Minimize2 :size="16" />
              </div>
              <div class="pref-info">
                <span class="pref-title">开机静默模式 (后台常驻)</span>
                <span class="pref-desc">开机自启时不弹窗，自动隐入后台监听；随时按 Esc 键或顶栏收起按钮可退入后台，按快捷键呼出</span>
              </div>
            </div>
            <div class="pref-action">
              <label class="toggle-switch">
                <input 
                  type="checkbox" 
                  v-model="appSettings.silentMode" 
                  @change="handleSavePreference" 
                />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <!-- 项 3: 全局唤醒快捷键 -->
          <div class="pref-item">
            <div class="pref-left">
              <div class="pref-icon-box">
                <Keyboard :size="16" />
              </div>
              <div class="pref-info">
                <span class="pref-title">快捷键唤醒</span>
                <span class="pref-desc">在任意第三方应用程序中按下该组合键，快速呼出或隐藏 GoClip 主窗口</span>
              </div>
            </div>
            <div class="pref-action hotkey-action">
              <div class="hotkey-input-wrap">
                <input 
                  type="text" 
                  v-model="hotkeyInput" 
                  class="hotkey-input"
                  placeholder="如 Cmd+Shift+V"
                />
                <button class="btn btn-primary-sm" @click="handleSaveHotkey" :disabled="appSettings.saving">
                  保存快捷键
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- 模块 2: 本地存储与历史管理 -->
      <section class="settings-card">
        <div class="card-title-row">
          <Database :size="18" class="card-icon" />
          <div class="card-titles">
            <h3>本地存储与历史管理</h3>
            <p>SQLite 数据库存储路径: ~/.goclip/clipboard.db (采用 AES-256 默认加密)</p>
          </div>
        </div>

        <div class="card-body">
          <div class="storage-actions">
            <button class="btn btn-danger-outline" @click="handleClearUnpinned">
              <Trash2 :size="14" />
              <span>清空未置顶历史</span>
            </button>
            <button class="btn btn-danger-solid" @click="handleClearAll">
              <Trash2 :size="14" />
              <span>清空所有历史 (含置顶)</span>
            </button>
          </div>
        </div>
      </section>
    </div>

    <!-- 自定义二次确认模态对话框 -->
    <div v-if="confirmDialog.visible" class="modal-backdrop" @click="closeConfirm">
      <div class="confirm-dialog" @click.stop>
        <div class="confirm-header">
          <div class="confirm-icon-wrap">
            <AlertTriangle :size="22" class="warning-icon" />
          </div>
          <div class="confirm-title-wrap">
            <h4>{{ confirmDialog.title }}</h4>
            <p>{{ confirmDialog.message }}</p>
          </div>
        </div>
        <div class="confirm-actions">
          <button class="btn btn-secondary" @click="closeConfirm">
            取消
          </button>
          <button class="btn btn-danger-solid" @click="executeConfirm">
            {{ confirmDialog.confirmText }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { 
  ArrowLeft, 
  Sliders,
  Power,
  Minimize2,
  Keyboard,
  Database, 
  Trash2,
  AlertTriangle
} from 'lucide-vue-next'
import { useClipboardStore } from '../stores/clipboard'
import { useAppSettingsStore } from '../stores/appSettings'

const emit = defineEmits(['back'])

const clipStore = useClipboardStore()
const appSettings = useAppSettingsStore()

const hotkeyInput = ref('Cmd+Shift+V')
const isMac = ref(true)

onMounted(async () => {
  try {
    if (window.runtime && window.runtime.Environment) {
      const env = await window.runtime.Environment()
      isMac.value = env.platform === 'darwin'
    } else {
      isMac.value = /Mac|iPhone|iPad|iPod/i.test(navigator.userAgent)
    }
  } catch (e) {
    isMac.value = /Mac|iPhone|iPad|iPod/i.test(navigator.userAgent)
  }

  await appSettings.fetchSettings()
  hotkeyInput.value = appSettings.hotkey
})

const handleSavePreference = async () => {
  const res = await appSettings.saveSettings({
    autoStart: appSettings.autoStart,
    silentMode: appSettings.silentMode,
    hotkey: hotkeyInput.value
  })
  if (res.success) {
    clipStore.showToast('设置已自动保存并生效')
  } else {
    clipStore.showToast(res.error || '保存失败')
  }
}

const handleSaveHotkey = async () => {
  if (!hotkeyInput.value.trim()) return
  const res = await appSettings.saveSettings({
    autoStart: appSettings.autoStart,
    silentMode: appSettings.silentMode,
    hotkey: hotkeyInput.value.trim()
  })
  if (res.success) {
    clipStore.showToast(`全局快捷键已更新为: ${hotkeyInput.value}`)
  } else {
    clipStore.showToast(res.error || '快捷键更新失败')
  }
}

// 二次确认对话框状态
const confirmDialog = reactive({
  visible: false,
  title: '',
  message: '',
  confirmText: '确认清空',
  action: null
})

const openConfirm = ({ title, message, confirmText = '确认清空', onConfirm }) => {
  confirmDialog.title = title
  confirmDialog.message = message
  confirmDialog.confirmText = confirmText
  confirmDialog.action = onConfirm
  confirmDialog.visible = true
}

const closeConfirm = () => {
  confirmDialog.visible = false
  confirmDialog.action = null
}

const executeConfirm = async () => {
  if (confirmDialog.action) {
    await confirmDialog.action()
  }
  closeConfirm()
}

const handleClearUnpinned = () => {
  openConfirm({
    title: '确认清空未置顶历史？',
    message: '此操作将清理所有普通历史记录，已置顶的内容将安全保留。该操作无法撤销。',
    confirmText: '确认清空',
    onConfirm: async () => {
      await clipStore.clearAll(true)
      clipStore.showToast('已清空所有未置顶历史')
    }
  })
}

const handleClearAll = () => {
  openConfirm({
    title: '高危操作：清空所有历史（含置顶）？',
    message: '此操作将彻底抹除本地数据库中的全部剪贴板记录，包括所有已置顶内容！此操作不可逆，请确认是否继续？',
    confirmText: '彻底清空全部',
    onConfirm: async () => {
      await clipStore.clearAll(false)
      clipStore.showToast('已清空全部历史记录')
    }
  })
}
</script>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: var(--bg-primary);
  overflow-y: auto;
}

.settings-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 22px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background-color: var(--bg-primary);
  flex-shrink: 0;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}

.view-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
}

.settings-content {
  flex: 1;
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.settings-card {
  background: rgba(30, 41, 59, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-md);
  padding: 18px 20px;
}

.card-title-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.card-icon {
  color: var(--accent);
  margin-top: 2px;
  flex-shrink: 0;
}

.card-titles h3 {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 2px;
}

.card-titles p {
  font-size: 12px;
  color: var(--text-dim);
}

.card-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.preference-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.pref-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  gap: 20px;
}

.pref-item:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.pref-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
}

.pref-icon-box {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: rgba(56, 189, 248, 0.1);
  color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pref-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.pref-title {
  font-size: 13px;
  font-weight: 600;
  color: #fff;
}

.pref-desc {
  font-size: 12px;
  color: var(--text-dim);
}

/* 现代 Toggle 开关样式 */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 42px;
  height: 22px;
  flex-shrink: 0;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background-color: rgba(255, 255, 255, 0.15);
  transition: 0.25s;
  border-radius: 22px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background-color: #cbd5e1;
  transition: 0.25s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--accent);
  border-color: var(--accent);
}

input:checked + .slider:before {
  transform: translateX(20px);
  background-color: #0f172a;
}

.hotkey-action {
  flex-shrink: 0;
}

.hotkey-input-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.hotkey-input {
  width: 150px;
  padding: 6px 10px;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: var(--radius-sm);
  color: var(--accent);
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 600;
  text-align: center;
}

.btn-primary-sm {
  background: var(--accent);
  color: #0f172a;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: background 0.15s;
}

.btn-primary-sm:hover {
  background: var(--accent-hover);
}

.storage-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: #fff;
}
.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.14);
}

.btn-danger-outline {
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: #f87171;
}
.btn-danger-outline:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #fff;
}

.btn-danger-solid {
  background: rgba(239, 68, 68, 0.2);
  border: 1px solid rgba(239, 68, 68, 0.4);
  color: #fca5a5;
}
.btn-danger-solid:hover {
  background: var(--danger);
  color: #fff;
}

/* 二次确认模态框样式 */
.modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  animation: fadeIn 0.15s ease-out;
}

.confirm-dialog {
  width: 100%;
  max-width: 440px;
  background: #1e293b;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5), 0 0 0 1px rgba(255, 255, 255, 0.05);
  display: flex;
  flex-direction: column;
  gap: 20px;
  animation: scaleIn 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.confirm-header {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.confirm-icon-wrap {
  width: 42px;
  height: 42px;
  border-radius: 10px;
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.25);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.warning-icon {
  color: #f87171;
}

.confirm-title-wrap h4 {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 6px;
}

.confirm-title-wrap p {
  font-size: 13px;
  color: var(--text-dim);
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style>
