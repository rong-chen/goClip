<template>
  <div v-if="secStore.showModal" class="modal-backdrop" @click="secStore.toggleModal(false)">
    <div class="modal-box" @click.stop>
      <div class="modal-header">
        <div class="title-wrap">
          <ShieldCheck :size="18" class="header-icon" />
          <span class="modal-title">本地保险库安全与加密</span>
        </div>
        <button class="close-btn" @click="secStore.toggleModal(false)">
          <X :size="16" />
        </button>
      </div>

      <div class="modal-body">
        <!-- 模式 1: 已有密码且处于锁定状态 -> 输入密码解锁 -->
        <div v-if="secStore.hasMasterPassword && secStore.isLocked" class="section">
          <p class="desc">
            剪贴板历史目前处于锁定保护中。请输入主密码解锁：
          </p>
          <div class="form-group">
            <label>主密码:</label>
            <input 
              type="password" 
              v-model="unlockPassword" 
              placeholder="输入解锁密码"
              @keydown.enter="handleUnlock"
              class="input-control"
              autofocus
            />
          </div>
          <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>
          <button class="btn btn-primary btn-block" @click="handleUnlock">
            <Unlock :size="14" />
            <span>立即解锁</span>
          </button>
        </div>

        <!-- 模式 2: 未设置主密码 或 用户需要修改主密码 -->
        <div v-else class="section">
          <div class="info-alert">
            <div class="info-title">
              <Info :size="14" />
              <strong>本地强加密保护说明：</strong>
            </div>
            <p>
              所有存入 SQLite 的文本和图片均已通过 <b>AES-256-GCM</b> 加密存储。
              设置专属主密码后，密钥将由 <b>PBKDF2</b>（100,000 次哈希迭代）实时派生，全面防护隐私泄漏。
            </p>
          </div>

          <div class="form-group">
            <label>{{ secStore.hasMasterPassword ? '修改主密码:' : '设置主密码:' }}</label>
            <input 
              type="password" 
              v-model="newPassword" 
              placeholder="至少 6 位字符"
              class="input-control"
            />
          </div>

          <div class="form-group">
            <label>确认主密码:</label>
            <input 
              type="password" 
              v-model="confirmPassword" 
              placeholder="再次输入以确认"
              @keydown.enter="handleSetPassword"
              class="input-control"
            />
          </div>

          <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>

          <div class="btn-row">
            <button class="btn btn-primary" @click="handleSetPassword">
              {{ secStore.hasMasterPassword ? '更新主密码' : '启用主密码保护' }}
            </button>
            <button v-if="secStore.hasMasterPassword" class="btn btn-secondary" @click="handleLock">
              <Lock :size="14" />
              <span>立即锁定</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { 
  ShieldCheck, 
  Info, 
  Unlock, 
  Lock, 
  X 
} from 'lucide-vue-next'
import { useSecurityStore } from '../stores/security'

const secStore = useSecurityStore()
const unlockPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const errorMsg = ref('')

const handleUnlock = async () => {
  errorMsg.value = ''
  if (!unlockPassword.value) {
    errorMsg.value = '请输入解锁密码'
    return
  }
  const res = await secStore.unlock(unlockPassword.value)
  if (!res.success) {
    errorMsg.value = '密码错误，解锁失败'
  } else {
    unlockPassword.value = ''
    secStore.toggleModal(false)
  }
}

const handleSetPassword = async () => {
  errorMsg.value = ''
  if (!newPassword.value || newPassword.value.length < 6) {
    errorMsg.value = '密码长度至少需为 6 位'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    errorMsg.value = '两次输入的密码不一致'
    return
  }

  const res = await secStore.setPassword(newPassword.value)
  if (!res.success) {
    errorMsg.value = '设置密码失败: ' + res.error
  } else {
    newPassword.value = ''
    confirmPassword.value = ''
    secStore.toggleModal(false)
  }
}

const handleLock = () => {
  secStore.lock()
  secStore.toggleModal(false)
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
  color: var(--accent);
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

.desc {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 14px;
}

.info-alert {
  background: rgba(56, 189, 248, 0.08);
  border: 1px solid rgba(56, 189, 248, 0.2);
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
  color: var(--accent);
  margin-bottom: 4px;
}

.info-alert p {
  line-height: 1.4;
  color: var(--text-muted);
}

.form-group {
  margin-bottom: 12px;
}

.form-group label {
  display: block;
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 4px;
}

.input-control {
  width: 100%;
  font-size: 13px;
}

.error-msg {
  color: var(--danger);
  font-size: 12px;
  margin-bottom: 10px;
}

.btn-row {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}

.btn-block {
  width: 100%;
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
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
