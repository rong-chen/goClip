<template>
  <div v-if="clipStore.previewImage" class="modal-backdrop" @click="clipStore.closeImageModal">
    <div class="image-modal-box" @click.stop>
      <div class="modal-header">
        <div class="title-wrap">
          <Image :size="16" class="title-icon" />
          <span class="modal-title">高清图片预览</span>
        </div>
        <button class="close-btn" @click="clipStore.closeImageModal">
          <X :size="16" />
        </button>
      </div>

      <div class="image-container">
        <img :src="clipStore.previewImage" alt="剪贴板大图" class="full-img" />
      </div>

      <div class="modal-footer">
        <span class="info">密文保存在本地 SQLite 数据库中</span>
        <button class="btn btn-primary" @click="copyCurrentImage">
          <Copy :size="13" />
          <span>复制到剪贴板</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Image, X, Copy } from 'lucide-vue-next'
import { useClipboardStore } from '../stores/clipboard'

const clipStore = useClipboardStore()

const copyCurrentImage = async () => {
  const item = clipStore.items.find(i => i.content === clipStore.previewImage)
  if (item) {
    await clipStore.copyItem(item.id)
  }
}
</script>

<style scoped>
.image-modal-box {
  background: #1e293b;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  max-width: 85vw;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.7);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--border-color);
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title-icon {
  color: var(--accent);
}

.modal-title {
  font-weight: 600;
  font-size: 14px;
  color: #fff;
}

.close-btn {
  color: var(--text-dim);
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
}
.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.image-container {
  padding: 16px;
  overflow: auto;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
}

.full-img {
  max-width: 100%;
  max-height: 60vh;
  object-fit: contain;
  border-radius: var(--radius-sm);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.4);
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 18px;
  border-top: 1px solid var(--border-color);
  background: rgba(15, 23, 42, 0.5);
}

.info {
  font-size: 11px;
  color: var(--text-dim);
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
}

.btn-primary {
  background: var(--accent);
  color: #0f172a;
  font-weight: 600;
}
.btn-primary:hover {
  background: var(--accent-hover);
}
</style>
