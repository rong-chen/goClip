<template>
  <div class="filter-bar">
    <!-- 左侧类型分类 -->
    <div class="tabs">
      <button 
        @click="clipStore.setFilter('all')" 
        :class="['tab-btn', clipStore.filterType === 'all' && !clipStore.pinnedOnly ? 'active' : '']"
      >
        <Layers :size="13" />
        <span>全部历史</span>
      </button>

      <button 
        @click="clipStore.setFilter('text')" 
        :class="['tab-btn', clipStore.filterType === 'text' && !clipStore.pinnedOnly ? 'active' : '']"
      >
        <FileText :size="13" />
        <span>文本</span>
      </button>

      <button 
        @click="clipStore.setFilter('image')" 
        :class="['tab-btn', clipStore.filterType === 'image' && !clipStore.pinnedOnly ? 'active' : '']"
      >
        <Image :size="13" />
        <span>图片</span>
      </button>

      <button 
        @click="togglePinnedOnly" 
        :class="['tab-btn', clipStore.pinnedOnly ? 'active' : '']"
      >
        <Star :size="13" />
        <span>仅看置顶</span>
        <span v-if="clipStore.pinnedCount > 0" class="count-badge highlight">{{ clipStore.pinnedCount }}</span>
      </button>
    </div>

    <!-- 右侧时间范围过滤 -->
    <div class="time-filter-wrap">
      <Clock :size="12" class="time-icon" />
      <select v-model="clipStore.timeFilter" class="time-select">
        <option value="all">全部时间</option>
        <option value="today">今天</option>
        <option value="7days">近 7 天</option>
        <option value="30days">近 30 天</option>
      </select>
    </div>
  </div>
</template>

<script setup>
import { 
  Layers, 
  FileText, 
  Image, 
  Star, 
  Clock 
} from 'lucide-vue-next'
import { useClipboardStore } from '../stores/clipboard'

const clipStore = useClipboardStore()

const togglePinnedOnly = () => {
  clipStore.setPinnedOnly(!clipStore.pinnedOnly)
}
</script>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 18px;
  background-color: var(--bg-primary);
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  user-select: none;
}

.tabs {
  display: flex;
  align-items: center;
  gap: 6px;
}

.tab-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
  cursor: pointer;
  background: transparent;
  transition: all 0.15s ease;
}

.tab-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.05);
}

.tab-btn.active {
  color: var(--accent);
  background: rgba(56, 189, 248, 0.1);
  font-weight: 600;
}

.count-badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-muted);
}

.count-badge.highlight {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

/* 右侧时间过滤器样式 */
.time-filter-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-sm);
  padding: 2px 8px;
  transition: all 0.15s ease;
}

.time-filter-wrap:hover {
  background: rgba(255, 255, 255, 0.07);
  border-color: rgba(255, 255, 255, 0.15);
}

.time-icon {
  color: var(--text-dim);
  flex-shrink: 0;
}

.time-select {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 12px;
  font-family: inherit;
  outline: none;
  cursor: pointer;
  padding: 3px 0;
}

.time-select:hover {
  color: #fff;
}

.time-select option {
  background: #1e293b;
  color: #fff;
}
</style>
