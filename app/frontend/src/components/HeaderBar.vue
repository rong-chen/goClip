<template>
  <header class="header-bar">
    <div class="brand">
      <div class="logo-badge">
        <ClipboardList :size="17" />
      </div>
      <div class="brand-text">
        <span class="title">GoClip</span>
        <span class="badge">AES-256 加密</span>
      </div>
    </div>

    <!-- 快速搜索栏 (支持中文输入法实时全文过滤) -->
    <div class="search-container">
      <Search :size="15" class="search-icon" />
      <input
        type="text"
        v-model="searchInput"
        @input="onInput"
        @compositionend="onCompositionEnd"
        placeholder="搜索文本或图片... (全文实时匹配)"
        class="search-input"
      />
      <button v-if="searchInput" @click="clearSearch" class="clear-search-btn">
        <X :size="12" />
      </button>
    </div>

    <!-- 右侧功能按钮 -->
    <div class="actions">
      <!-- 隐藏至后台静默运行 -->
      <button 
        @click="hideToBackground" 
        class="icon-btn" 
        title="隐藏至后台静默运行"
      >
        <Minimize2 :size="14" />
      </button>

      <!-- 定时任务入口按钮 -->
      <button 
        @click="$emit('open-cron')" 
        class="status-btn setting-btn" 
        title="进入定时任务管理"
      >
        <Clock :size="14" />
        <span>定时任务</span>
      </button>

      <!-- 设置页面入口按钮 -->
      <button 
        @click="$emit('open-settings')" 
        class="status-btn setting-btn" 
        title="进入应用设置页面"
      >
        <Settings :size="14" />
        <span>设置</span>
      </button>

      <!-- 清空剪贴板 -->
      <button @click="showClearModal = true" class="icon-btn danger-btn" title="清空剪贴板">
        <Trash2 :size="14" />
      </button>
    </div>

    <!-- 清空剪贴板自定义确认弹窗 -->
    <div v-if="showClearModal" class="modal-backdrop" @click="showClearModal = false">
      <div class="confirm-dialog" @click.stop>
        <div class="confirm-header">
          <div class="confirm-icon-wrap">
            <Trash2 :size="20" class="danger-icon" />
          </div>
          <div class="confirm-title-wrap">
            <h4>清空剪贴板</h4>
            <p>选择清空方式，操作将同步清空操作系统当前剪贴板内容</p>
          </div>
        </div>
        <div class="confirm-actions">
          <button class="btn btn-secondary" @click="showClearModal = false">
            取消
          </button>
          <button class="btn btn-secondary-danger" @click="handleClear(true)">
            清空未置顶
          </button>
          <button class="btn btn-danger-solid" @click="handleClear(false)">
            全部清空
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, watch } from 'vue'
import { 
  ClipboardList, 
  Search, 
  X, 
  Clock,
  Settings, 
  Trash2,
  Minimize2
} from 'lucide-vue-next'
import { useClipboardStore } from '../stores/clipboard'
import { HideWindow } from '../../wailsjs/go/main/App'

const emit = defineEmits(['open-settings', 'open-cron'])
const clipStore = useClipboardStore()

const searchInput = ref(clipStore.search)

watch(() => clipStore.search, (val) => {
  if (searchInput.value !== val) {
    searchInput.value = val
  }
})

const onInput = (e) => {
  if (!e.isComposing) {
    clipStore.setSearch(e.target.value)
  }
}

const onCompositionEnd = (e) => {
  clipStore.setSearch(e.target.value)
}

const clearSearch = () => {
  searchInput.value = ''
  clipStore.setSearch('')
}

const hideToBackground = async () => {
  try {
    if (window.go && window.go.main && window.go.main.App) {
      await HideWindow()
    }
  } catch (e) {
    console.warn('隐藏窗口异常:', e)
  }
}

const showClearModal = ref(false)

const handleClear = async (keepPinned) => {
  showClearModal.value = false
  await clipStore.clearAll(keepPinned)
}
</script>

<style scoped>
.header-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px 10px 18px;
  background-color: var(--bg-primary);
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  gap: 16px;
  user-select: none;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-badge {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  background: rgba(56, 189, 248, 0.12);
  color: var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(56, 189, 248, 0.25);
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-text .title {
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  letter-spacing: -0.3px;
}

.brand-text .badge {
  font-size: 10px;
  color: var(--accent);
  font-weight: 500;
}

.search-container {
  flex: 1;
  max-width: 440px;
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: var(--text-dim);
  pointer-events: none;
}

.search-input {
  width: 100%;
  height: 34px;
  padding: 0 34px 0 36px;
  border-radius: var(--radius-sm);
  background: rgba(15, 23, 42, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  font-size: 13px;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent);
  background: rgba(15, 23, 42, 0.9);
  box-shadow: 0 0 0 2px rgba(56, 189, 248, 0.2);
}

.clear-search-btn {
  position: absolute;
  right: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.15);
  color: var(--text-muted);
}

.clear-search-btn:hover {
  background: rgba(255, 255, 255, 0.25);
  color: #fff;
}

.actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-btn {
  height: 32px;
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 0 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  transition: all 0.15s ease;
  cursor: pointer;
  line-height: 1;
  flex-shrink: 0;
}

.setting-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: var(--text-main);
}

.setting-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  border-color: rgba(255, 255, 255, 0.2);
}

.icon-btn {
  width: 32px;
  height: 32px;
  box-sizing: border-box;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
  padding: 0;
}

.icon-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.danger-btn:hover {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.3);
  color: #f87171;
}

.confirm-dialog {
  width: 100%;
  max-width: 420px;
  background: var(--bg-secondary);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.confirm-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.confirm-icon-wrap {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.confirm-title-wrap h4 {
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 4px;
}

.confirm-title-wrap p {
  font-size: 12px;
  color: var(--text-dim);
  line-height: 1.5;
}

.confirm-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.btn {
  padding: 6px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}

.btn-secondary-danger {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  color: var(--danger);
}

.btn-danger-solid {
  background: var(--danger);
  color: #fff;
}
</style>
