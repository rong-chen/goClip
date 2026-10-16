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
          <div class="pref-item hotkey-card-item">
            <div class="pref-left">
              <div class="pref-icon-box">
                <Keyboard :size="16" />
              </div>
              <div class="pref-info">
                <span class="pref-title">快捷键唤醒</span>
                <span class="pref-desc">在任意第三方应用程序中按下该组合键，快速呼出或隐藏 GoClip 主窗口</span>
              </div>
            </div>
            <div class="pref-action hotkey-action-col">
              <!-- 录制模式中 -->
              <div v-if="isRecording" class="hotkey-recording-box">
                <div class="recording-indicator">
                  <span class="recording-dot"></span>
                  <span class="recording-text">
                    {{ currentRecordingText || '请直接在键盘上按下新组合键...' }}
                  </span>
                </div>
                <button class="btn btn-xs btn-cancel" @click="stopRecording">
                  <X :size="12" />
                  <span>取消 (Esc)</span>
                </button>
              </div>

              <!-- 手动文本编辑模式 -->
              <div v-else-if="isManualEdit" class="hotkey-manual-wrap">
                <input 
                  type="text" 
                  v-model="hotkeyInput" 
                  class="hotkey-input"
                  placeholder="如 Ctrl+Shift+V"
                />
                <button class="btn btn-primary-sm" @click="() => handleSaveHotkey(hotkeyInput)" :disabled="appSettings.saving">
                  保存
                </button>
                <button class="btn btn-secondary-sm" @click="isManualEdit = false">
                  返回
                </button>
              </div>

              <!-- 默认展示：按键徽标卡片与操作按钮 -->
              <div v-else class="hotkey-display-wrap">
                <div class="hotkey-badge-list" @click="startRecording" title="点击开始录制新快捷键">
                  <template v-for="(key, idx) in parsedKeys" :key="idx">
                    <kbd class="key-badge">{{ formatKeySymbol(key) }}</kbd>
                    <span v-if="idx < parsedKeys.length - 1" class="key-plus">+</span>
                  </template>
                </div>

                <div class="hotkey-btns">
                  <button class="btn btn-record" @click="startRecording">
                    <CircleDot :size="13" class="record-icon" />
                    <span>录制按键</span>
                  </button>
                  <button class="btn btn-icon-subtle" @click="restoreDefault" title="恢复默认快捷键">
                    <RotateCcw :size="13" />
                  </button>
                  <button class="btn btn-icon-subtle" @click="isManualEdit = true" title="手动输入">
                    <Edit3 :size="13" />
                  </button>
                </div>
              </div>

              <!-- 快捷推荐预设 -->
              <div class="hotkey-presets-row">
                <span class="preset-label">一键预设:</span>
                <div class="preset-pills">
                  <button 
                    v-for="preset in availablePresets" 
                    :key="preset.value"
                    class="preset-pill"
                    :class="{ active: hotkeyInput === preset.value }"
                    @click="applyPreset(preset.value)"
                  >
                    <span>{{ preset.label }}</span>
                  </button>
                </div>
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
import { reactive, ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { 
  ArrowLeft, 
  Sliders,
  Power,
  Minimize2,
  Keyboard,
  Database, 
  Trash2,
  AlertTriangle,
  RotateCcw,
  CircleDot,
  Edit3,
  X,
  Check
} from 'lucide-vue-next'
import { useClipboardStore } from '../stores/clipboard'
import { useAppSettingsStore } from '../stores/appSettings'

const emit = defineEmits(['back'])

const clipStore = useClipboardStore()
const appSettings = useAppSettingsStore()

const hotkeyInput = ref('Ctrl+Shift+V')
const isMac = ref(false)
const isRecording = ref(false)
const recordingModifiers = ref([])
const isManualEdit = ref(false)

// 拆解当前快捷键为按键数组
const parsedKeys = computed(() => {
  if (!hotkeyInput.value) return []
  return hotkeyInput.value.split('+').map(k => k.trim()).filter(Boolean)
})

// 录制过程中的动态文案
const currentRecordingText = computed(() => {
  if (recordingModifiers.value.length === 0) {
    return isMac.value ? '请按下组合键 (如 Cmd + Shift + V)...' : '请按下组合键 (如 Ctrl + Alt + V)...'
  }
  return `${recordingModifiers.value.join(' + ')} + 请按下主按键...`
})

// 对应系统的推荐预设
const availablePresets = computed(() => {
  if (isMac.value) {
    return [
      { label: 'Cmd + Shift + V', value: 'Cmd+Shift+V' },
      { label: 'Cmd + Option + V', value: 'Cmd+Option+V' },
      { label: 'Ctrl + Shift + V', value: 'Ctrl+Shift+V' }
    ]
  }
  return [
    { label: 'Ctrl + Shift + V', value: 'Ctrl+Shift+V' },
    { label: 'Alt + Shift + V', value: 'Alt+Shift+V' },
    { label: 'Alt + V', value: 'Alt+V' },
    { label: 'Ctrl + Alt + V', value: 'Ctrl+Alt+V' }
  ]
})

// 按键符号美化
const formatKeySymbol = (key) => {
  const upper = key.toUpperCase()
  if (isMac.value) {
    if (upper === 'CMD' || upper === 'COMMAND') return '⌘ Cmd'
    if (upper === 'OPTION' || upper === 'OPT' || upper === 'ALT') return '⌥ Option'
    if (upper === 'SHIFT') return '⇧ Shift'
    if (upper === 'CTRL' || upper === 'CONTROL') return '⌃ Ctrl'
  } else {
    if (upper === 'CTRL' || upper === 'CONTROL' || upper === 'CMD') return 'Ctrl'
    if (upper === 'ALT' || upper === 'OPTION' || upper === 'OPT') return 'Alt'
    if (upper === 'SHIFT') return 'Shift'
    if (upper === 'WIN' || upper === 'SUPER') return '⊞ Win'
  }
  return key
}

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
  hotkeyInput.value = appSettings.hotkey || (isMac.value ? 'Cmd+Shift+V' : 'Ctrl+Shift+V')
})

onBeforeUnmount(() => {
  stopRecording()
})

const startRecording = () => {
  isRecording.value = true
  recordingModifiers.value = []
  window.addEventListener('keydown', onRecordKeyDown, true)
}

const stopRecording = () => {
  if (isRecording.value) {
    isRecording.value = false
    recordingModifiers.value = []
    window.removeEventListener('keydown', onRecordKeyDown, true)
  }
}

const onRecordKeyDown = (e) => {
  e.preventDefault()
  e.stopPropagation()

  if (e.key === 'Escape') {
    stopRecording()
    return
  }

  // 捕获修饰键
  const mods = []
  if (isMac.value) {
    if (e.metaKey) mods.push('Cmd')
    if (e.ctrlKey) mods.push('Ctrl')
    if (e.altKey) mods.push('Option')
    if (e.shiftKey) mods.push('Shift')
  } else {
    if (e.ctrlKey) mods.push('Ctrl')
    if (e.altKey) mods.push('Alt')
    if (e.shiftKey) mods.push('Shift')
    if (e.metaKey) mods.push('Win')
  }

  // 若当前仅按下了修饰键本身，更新实时提示
  const modifierKeys = ['Control', 'Shift', 'Alt', 'Meta', 'AltGraph']
  if (modifierKeys.includes(e.key)) {
    recordingModifiers.value = mods
    return
  }

  // 提取有效的主键名
  let keyName = ''
  if (e.code && e.code.startsWith('Key')) {
    keyName = e.code.replace('Key', '').toUpperCase()
  } else if (e.code && e.code.startsWith('Digit')) {
    keyName = e.code.replace('Digit', '')
  } else if (e.key && /^F\d+$/i.test(e.key)) {
    keyName = e.key.toUpperCase()
  } else if (e.code === 'Space' || e.key === ' ') {
    keyName = 'Space'
  } else if (e.key && e.key.length === 1 && /[a-zA-Z0-9]/.test(e.key)) {
    keyName = e.key.toUpperCase()
  }

  if (!keyName) {
    return
  }

  if (mods.length === 0) {
    clipStore.showToast('全局快捷键必须包含至少一个修饰键 (如 Ctrl, Alt, Shift)')
    return
  }

  const finalHotkey = [...mods, keyName].join('+')
  hotkeyInput.value = finalHotkey
  stopRecording()
  handleSaveHotkey(finalHotkey)
}

const applyPreset = async (preset) => {
  hotkeyInput.value = preset
  await handleSaveHotkey(preset)
}

const restoreDefault = async () => {
  const def = isMac.value ? 'Cmd+Shift+V' : 'Ctrl+Shift+V'
  hotkeyInput.value = def
  await handleSaveHotkey(def)
}

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

const handleSaveHotkey = async (keyVal) => {
  const val = (typeof keyVal === 'string' ? keyVal : hotkeyInput.value).trim()
  if (!val) return
  hotkeyInput.value = val

  const res = await appSettings.saveSettings({
    autoStart: appSettings.autoStart,
    silentMode: appSettings.silentMode,
    hotkey: val
  })
  if (res.success) {
    clipStore.showToast(`全局快捷键已更新为: ${val}`)
  } else {
    clipStore.showToast(res.error || '快捷键更新失败，可能与其他软件冲突')
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

.hotkey-card-item {
  align-items: flex-start;
}

.hotkey-action-col {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10px;
  flex-shrink: 0;
}

.hotkey-display-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 键盘实体键帽徽标 */
.hotkey-badge-list {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(15, 23, 42, 0.7);
  padding: 5px 10px;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(255, 255, 255, 0.12);
  cursor: pointer;
  transition: all 0.2s ease;
}

.hotkey-badge-list:hover {
  border-color: var(--accent);
  background: rgba(56, 189, 248, 0.06);
}

.key-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 7px;
  font-size: 11px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-weight: 700;
  color: #f8fafc;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.12) 0%, rgba(255, 255, 255, 0.05) 100%);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.4), inset 0 1px 0 rgba(255, 255, 255, 0.15);
  border-radius: 4px;
}

.key-plus {
  font-size: 11px;
  color: var(--text-dim);
  font-weight: 600;
  user-select: none;
}

.hotkey-btns {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-record {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 11px;
  background: rgba(56, 189, 248, 0.12);
  border: 1px solid rgba(56, 189, 248, 0.3);
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-record:hover {
  background: var(--accent);
  color: #0f172a;
  border-color: var(--accent);
}

.btn-icon-subtle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-dim);
  cursor: pointer;
  transition: all 0.15s ease;
}

.btn-icon-subtle:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

/* 录制状态中的呼吸灯交互盒 */
.hotkey-recording-box {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(245, 158, 11, 0.12);
  border: 1.5px solid #f59e0b;
  box-shadow: 0 0 12px rgba(245, 158, 11, 0.25);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  animation: pulseRecording 1.8s infinite ease-in-out;
}

@keyframes pulseRecording {
  0%, 100% { border-color: #f59e0b; box-shadow: 0 0 8px rgba(245, 158, 11, 0.2); }
  50% { border-color: #fbbf24; box-shadow: 0 0 16px rgba(245, 158, 11, 0.45); }
}

.recording-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
}

.recording-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #f59e0b;
  animation: blinkDot 1s infinite alternate;
}

@keyframes blinkDot {
  from { opacity: 0.3; }
  to { opacity: 1; }
}

.recording-text {
  font-size: 12px;
  font-weight: 600;
  color: #fef3c7;
}

.btn-cancel {
  background: rgba(255, 255, 255, 0.1);
  color: #f8fafc;
  padding: 3px 8px;
  border-radius: 4px;
  font-size: 11px;
}
.btn-cancel:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* 手动输入模式 */
.hotkey-manual-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.hotkey-input {
  width: 140px;
  padding: 6px 10px;
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.2);
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

.btn-secondary-sm {
  background: rgba(255, 255, 255, 0.08);
  color: #f8fafc;
  padding: 6px 10px;
  font-size: 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.btn-secondary-sm:hover {
  background: rgba(255, 255, 255, 0.15);
}

/* 推荐预设胶囊 */
.hotkey-presets-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preset-label {
  font-size: 11px;
  color: var(--text-dim);
}

.preset-pills {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.preset-pill {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: var(--text-dim);
  font-size: 11px;
  padding: 3px 8px;
  border-radius: 12px;
  cursor: pointer;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  transition: all 0.15s ease;
}

.preset-pill:hover {
  background: rgba(56, 189, 248, 0.1);
  border-color: rgba(56, 189, 248, 0.3);
  color: #fff;
}

.preset-pill.active {
  background: rgba(56, 189, 248, 0.15);
  border-color: var(--accent);
  color: var(--accent);
  font-weight: 600;
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
