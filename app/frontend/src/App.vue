<template>
  <div class="app-layout">
    <!-- macOS 专属一体化同色顶栏：消除系统灰色标题栏色差，并给红绿灯留出完美位置 (Windows下自适应隐藏) -->
    <div v-if="isMac" class="window-topbar">
      <div class="traffic-light-pad"></div>
      <div class="window-title">GoClip - 本地加密剪贴板</div>
    </div>

    <!-- 视图 1: 设置页面 (独立完整配置页，不用弹窗) -->
    <SettingsView 
      v-if="currentView === 'settings'" 
      @back="currentView = 'home'" 
    />

    <!-- 视图 2: 剪贴板主页面 -->
    <template v-else>
      <!-- 顶部状态与搜索栏 (含 Settings 按钮) -->
      <HeaderBar @open-settings="currentView = 'settings'" />

      <!-- 分类筛选与时间范围栏 -->
      <FilterBar />

      <!-- 主列表区域 -->
      <main class="list-container">
        <div v-if="clipStore.loading && clipStore.items.length === 0" class="loading-state">
          <div class="spinner"></div>
          <span>正在读取剪贴板数据...</span>
        </div>

        <!-- 暂无数据提示 (极简直接) -->
        <div v-else-if="clipStore.filteredItems.length === 0" class="empty-state">
          <div class="state-icon-box">
            <ClipboardList :size="24" class="state-svg" />
          </div>
          <h3 class="empty-text">暂无数据</h3>
        </div>

        <div v-else class="cards-list">
          <ClipboardItemCard 
            v-for="item in clipStore.filteredItems" 
            :key="item.id" 
            :item="item" 
          />
        </div>
      </main>
    </template>

    <!-- 全局轻量 Toast 提示 -->
    <div v-if="clipStore.toastMessage" class="toast-popup">
      {{ clipStore.toastMessage }}
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ClipboardList } from 'lucide-vue-next'
import HeaderBar from './components/HeaderBar.vue'
import FilterBar from './components/FilterBar.vue'
import ClipboardItemCard from './components/ClipboardItemCard.vue'
import SettingsView from './views/SettingsView.vue'
import { useClipboardStore } from './stores/clipboard'
import { HideWindow } from '../wailsjs/go/main/App'

const clipStore = useClipboardStore()

// 视图路由状态: 'home' 剪贴板列表 | 'settings' 设置页面
const currentView = ref('home')

// 跨平台检测 (macOS 内嵌顶栏，Windows 原生暗黑标题栏)
const isMac = ref(true)

const onKeyDown = async (e) => {
  if (e.key === 'Escape') {
    if (currentView.value === 'settings') {
      currentView.value = 'home'
    } else {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await HideWindow()
        }
      } catch (err) {
        console.warn('Esc 隐藏窗口异常:', err)
      }
    }
  }
}

onMounted(async () => {
  window.addEventListener('keydown', onKeyDown)

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

  await clipStore.fetchItems()

  // 监听后端 Wails 剪贴板事件通知与设置跳转通知
  if (window.runtime && window.runtime.EventsOn) {
    window.runtime.EventsOn('clipboard:changed', (newItem) => {
      clipStore.handleNewItem(newItem)
    })
    window.runtime.EventsOn('route:settings', () => {
      currentView.value = 'settings'
    })
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  if (window.runtime && window.runtime.EventsOff) {
    window.runtime.EventsOff('clipboard:changed')
    window.runtime.EventsOff('route:settings')
  }
})
</script>

<style scoped>
.app-layout {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-primary);
  overflow: hidden;
}

/* 顶部一体化顶栏 */
.window-topbar {
  height: 32px;
  display: flex;
  align-items: center;
  background-color: var(--bg-primary);
  --wails-draggable: drag;
  user-select: none;
  flex-shrink: 0;
  padding: 0 16px;
}

.traffic-light-pad {
  width: 72px;
  height: 100%;
  flex-shrink: 0;
}

.window-title {
  font-size: 11px;
  color: var(--text-dim);
  font-weight: 500;
  letter-spacing: 0.3px;
  flex: 1;
  text-align: center;
  margin-right: 72px;
}

.list-container {
  flex: 1;
  overflow-y: auto;
  padding: 12px 18px;
  background-color: var(--bg-primary);
}

.cards-list {
  display: flex;
  flex-direction: column;
}

.loading-state, .empty-state, .locked-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 60vh;
  color: var(--text-dim);
  text-align: center;
  padding: 0 32px;
}

.state-icon-box {
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.state-svg {
  color: var(--accent);
}

.empty-text {
  font-size: 14px;
  color: var(--text-dim);
  font-weight: 500;
}

.locked-state h3 {
  font-size: 15px;
  color: #fff;
  margin-bottom: 6px;
  font-weight: 600;
}

.locked-state p {
  font-size: 13px;
  color: var(--text-muted);
  max-width: 420px;
  line-height: 1.5;
  margin-bottom: 16px;
}

.spinner {
  width: 28px;
  height: 28px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.btn {
  padding: 8px 18px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}

.btn-primary {
  background: var(--accent);
  color: #0f172a;
}
.btn-primary:hover {
  background: var(--accent-hover);
}

.toast-popup {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid rgba(56, 189, 248, 0.4);
  color: #fff;
  padding: 8px 18px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.5);
  z-index: 2000;
  animation: popIn 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes popIn {
  from { opacity: 0; transform: translate(-50%, 10px); }
  to { opacity: 1; transform: translate(-50%, 0); }
}
</style>
