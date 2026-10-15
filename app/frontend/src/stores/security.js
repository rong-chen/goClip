import { defineStore } from 'pinia'
import { GetSecuritySettings, SetMasterPassword, UnlockVault, LockVault } from '../../wailsjs/go/main/App'
import { useClipboardStore } from './clipboard'

export const useSecurityStore = defineStore('security', {
  state: () => ({
    hasMasterPassword: false,
    isLocked: false,
    showModal: false, // 控制安全弹窗展示
  }),

  actions: {
    async fetchStatus() {
      try {
        const settings = await GetSecuritySettings()
        if (settings) {
          this.hasMasterPassword = settings.has_master_password
          this.isLocked = settings.is_locked
        }
      } catch (err) {
        console.error('获取安全状态失败:', err)
      }
    },

    async unlock(password) {
      try {
        await UnlockVault(password)
        this.isLocked = false
        const clipStore = useClipboardStore()
        await clipStore.fetchItems()
        clipStore.showToast('保险库已解锁')
        return { success: true }
      } catch (err) {
        return { success: false, error: err }
      }
    },

    async lock() {
      try {
        await LockVault()
        this.isLocked = true
        const clipStore = useClipboardStore()
        clipStore.items = []
        clipStore.showToast('保险库已锁定')
      } catch (err) {
        console.error('锁定失败:', err)
      }
    },

    async setPassword(password) {
      try {
        await SetMasterPassword(password)
        this.hasMasterPassword = true
        this.isLocked = false
        const clipStore = useClipboardStore()
        await clipStore.fetchItems()
        clipStore.showToast('主加密密码设置成功！')
        return { success: true }
      } catch (err) {
        return { success: false, error: err }
      }
    },

    toggleModal(val) {
      this.showModal = val !== undefined ? val : !this.showModal
    }
  }
})
