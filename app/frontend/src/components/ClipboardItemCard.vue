<template>
  <div 
    :class="['card', item.is_pinned ? 'pinned' : '', item.type]"
    @click="copy"
  >
    <!-- 卡片主体内容 -->
    <div class="card-body">
      <!-- 顶部元数据行 -->
      <div class="card-meta">
        <div class="meta-left">
          <span v-if="item.is_pinned" class="pin-tag">
            <Star :size="11" fill="currentColor" />
            <span>已置顶</span>
          </span>
          <span class="type-badge">
            <component :is="item.type === 'image' ? Image : FileText" :size="12" />
            <span>{{ item.type === 'image' ? '图片' : '文本' }}</span>
          </span>
          <span class="size-badge">{{ formatBytes(item.byte_size) }}</span>
          <span v-if="item.width > 0" class="res-badge">{{ item.width }}×{{ item.height }}</span>
        </div>
        <div class="meta-right">
          <span class="time-text">{{ formatTime(item.created_at) }}</span>
        </div>
      </div>

      <!-- 纯文本展示 -->
      <div v-if="item.type === 'text'" class="content-text">
        {{ item.content }}
      </div>

      <!-- 图片展示 -->
      <div v-else-if="item.type === 'image'" class="content-image-wrapper">
        <img 
          :src="item.content" 
          alt="剪贴板图片" 
          class="thumbnail" 
          @click.stop="openImageViewer"
          title="点击使用系统图片查看器打开"
        />
        <div class="image-hint" @click.stop="openImageViewer">
          <ZoomIn :size="12" />
          <span>系统查看器打开</span>
        </div>
      </div>
    </div>

    <!-- 右侧悬浮操作栏 -->
    <div class="card-actions" @click.stop>
      <button @click="copy" class="action-btn copy-btn" title="回写到系统剪贴板">
        <Copy :size="14" />
      </button>

      <button @click="togglePin" :class="['action-btn', item.is_pinned ? 'pinned-btn' : '']" title="置顶/取消置顶">
        <Star :size="14" :fill="item.is_pinned ? 'currentColor' : 'none'" />
      </button>

      <button @click="deleteItem" class="action-btn del-btn" title="删除该条记录">
        <Trash2 :size="14" />
      </button>
    </div>
  </div>
</template>

<script setup>
import { 
  Star, 
  Image, 
  FileText, 
  ZoomIn, 
  Copy, 
  Trash2 
} from 'lucide-vue-next'
import { useClipboardStore } from '../stores/clipboard'

const props = defineProps({
  item: {
    type: Object,
    required: true
  }
})

const clipStore = useClipboardStore()

const copy = () => {
  clipStore.copyItem(props.item.id)
}

const togglePin = () => {
  clipStore.togglePin(props.item.id)
}

const deleteItem = () => {
  clipStore.deleteItem(props.item.id)
}

const openImageViewer = () => {
  clipStore.openSystemViewer(props.item.id)
}

const formatBytes = (bytes) => {
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const formatTime = (ts) => {
  if (!ts) return ''
  const diff = Date.now() - ts
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return Math.floor(diff / 60000) + ' 分钟前'
  if (diff < 86400000) return Math.floor(diff / 3600000) + ' 小时前'
  const date = new Date(ts)
  return `${date.getMonth() + 1}-${date.getDate()} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`
}
</script>

<style scoped>
.card {
  position: relative;
  display: flex;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  margin-bottom: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.card:hover {
  background: var(--bg-card-hover);
  border-color: rgba(255, 255, 255, 0.15);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.card.pinned {
  border-color: rgba(245, 158, 11, 0.35);
  background: rgba(30, 41, 59, 0.85);
}

.card-body {
  flex: 1;
  padding: 10px 14px;
  min-width: 0;
}

.card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
  font-size: 11px;
}

.meta-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pin-tag {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  color: #fbbf24;
  font-weight: 600;
}

.type-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
  font-weight: 500;
}

.size-badge, .res-badge {
  color: var(--text-dim);
  background: rgba(255, 255, 255, 0.05);
  padding: 1px 5px;
  border-radius: 4px;
}

.time-text {
  color: var(--text-dim);
}

.content-text {
  font-size: 13px;
  color: #e2e8f0;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 84px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  text-overflow: ellipsis;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.content-image-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 4px;
}

.thumbnail {
  max-width: 180px;
  max-height: 100px;
  object-fit: cover;
  border-radius: var(--radius-sm);
  border: 1px solid rgba(255, 255, 255, 0.1);
  background: #0f172a;
}

.thumbnail:hover {
  opacity: 0.9;
  border-color: var(--accent);
}

.image-hint {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--accent);
  cursor: pointer;
}

.card-actions {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 4px;
  padding: 0 10px;
  opacity: 0.4;
  transition: opacity 0.15s ease;
}

.card:hover .card-actions {
  opacity: 1;
}

.action-btn {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.copy-btn:hover {
  color: var(--accent);
}

.pinned-btn {
  color: #fbbf24 !important;
}

.del-btn:hover {
  color: var(--danger) !important;
}
</style>
