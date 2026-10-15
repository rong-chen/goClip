<template>
  <div v-if="syncStore.showModal" class="modal-backdrop" @click="syncStore.toggleModal(false)">
    <div class="modal-box" @click.stop>
      <div class="modal-header">
        <div class="title-wrap">
          <Cloud :size="18" class="header-icon" />
          <span class="modal-title">云端同步配置 (架构预留)</span>
        </div>
        <button class="close-btn" @click="syncStore.toggleModal(false)">
          <X :size="16" />
        </button>
      </div>

      <div class="modal-body">
        <div class="cloud-info-badge">
          <div class="info-title">
            <Info :size="14" />
            <strong>后期上云预留说明：</strong>
          </div>
          <p>
            当前版本以<b>本地单机 SQLite 加密</b>模式运行，所有数据绝不上传第三方。
            底层数据表已全面预留 <code>sync_id</code>、<code>device_id</code>、<code>sync_status</code> 及乐观锁版本字段，
            后期开放云端服务后，只需在此填入同步地址即可无缝上云！
          </p>
        </div>

        <div class="form-group">
          <label class="switch-label">
            <span>启用云端同步通道</span>
            <input type="checkbox" v-model="form.enabled" class="checkbox-input" />
          </label>
        </div>

        <div class="form-group">
          <label>云服务端点 (API Endpoint):</label>
          <input 
            type="text" 
            v-model="form.serverUrl" 
            placeholder="例如: https://sync.goclip.internal/v1"
            class="input-control"
            :disabled="!form.enabled"
          />
        </div>

        <div class="form-group">
          <label>认证令牌 (API Token / Key):</label>
          <input 
            type="password" 
            v-model="form.apiKey" 
            placeholder="输入云端访问 Token"
            class="input-control"
            :disabled="!form.enabled"
          />
        </div>

        <div class="form-group">
          <label class="switch-label">
            <span>端到端加密 (E2EE 零知识加密传输)</span>
            <input type="checkbox" v-model="form.e2eEncryption" class="checkbox-input" :disabled="!form.enabled" />
          </label>
        </div>

        <div v-if="testResult" :class="['result-box', testResult.success ? 'success' : 'warn']">
          <component :is="testResult.success ? CheckCircle2 : AlertCircle" :size="14" />
          <span>{{ testResult.success ? '连接测试成功！' : ('状态: ' + (testResult.error_message || '服务未开启')) }}</span>
        </div>

        <div class="btn-row">
          <button class="btn btn-secondary" @click="handleTest" :disabled="syncStore.testing">
            {{ syncStore.testing ? '正在测试...' : '测试云端连接' }}
          </button>
          <button class="btn btn-primary" @click="handleSave">
            保存配置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { 
  Cloud, 
  Info, 
  CheckCircle2, 
  AlertCircle, 
  X 
} from 'lucide-vue-next'
import { useSyncStore } from '../stores/sync'

const syncStore = useSyncStore()
const testResult = ref(null)

const form = reactive({
  enabled: false,
  serverUrl: '',
  apiKey: '',
  e2eEncryption: true,
  autoSync: false,
  syncInterval: 60,
})

onMounted(async () => {
  await syncStore.fetchConfig()
  form.enabled = syncStore.enabled
  form.serverUrl = syncStore.serverUrl
  form.apiKey = syncStore.apiKey
  form.e2eEncryption = syncStore.e2eEncryption
  form.autoSync = syncStore.autoSync
  form.syncInterval = syncStore.syncInterval
})

const handleTest = async () => {
  testResult.value = await syncStore.testSync()
}

const handleSave = async () => {
  await syncStore.saveConfig(form)
  syncStore.toggleModal(false)
}
</script>

<style scoped>
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid var(--border-color);
}

.title-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-icon {
  color: #38bdf8;
}

.modal-title {
  font-weight: 600;
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
}

.modal-body {
  padding: 18px;
}

.cloud-info-badge {
  background: rgba(14, 165, 233, 0.08);
  border: 1px solid rgba(14, 165, 233, 0.25);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  font-size: 12px;
  color: #bae6fd;
  margin-bottom: 16px;
}

.info-title {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #38bdf8;
  margin-bottom: 4px;
}

.cloud-info-badge p {
  color: var(--text-muted);
  line-height: 1.45;
}

code {
  background: rgba(0, 0, 0, 0.3);
  padding: 1px 4px;
  border-radius: 3px;
  color: #38bdf8;
  font-size: 11px;
}

.form-group {
  margin-bottom: 14px;
}

.form-group label {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.switch-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer;
  color: #fff !important;
  font-weight: 500;
}

.checkbox-input {
  width: 16px;
  height: 16px;
  accent-color: var(--accent);
  cursor: pointer;
}

.input-control {
  width: 100%;
  font-size: 13px;
}

.input-control:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.result-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  margin-bottom: 12px;
}

.result-box.success {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success);
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.result-box.warn {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.btn-row {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 18px;
}

.btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  font-size: 13px;
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

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}
.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.15);
}
</style>
