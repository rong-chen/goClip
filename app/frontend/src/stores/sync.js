import { defineStore } from 'pinia'
import { GetCloudSyncConfig, SaveCloudSyncConfig, TestCloudSync } from '../../wailsjs/go/main/App'
import { useClipboardStore } from './clipboard'

export const useSyncStore = defineStore('sync', {
  state: () => ({
    enabled: false,
    serverUrl: 'https://sync.goclip.internal/api/v1',
    apiKey: '',
    e2eEncryption: true,
    autoSync: false,
    syncInterval: 60,
    lastSyncTime: 0,
    showModal: false,
    testing: false,
  }),

  actions: {
    async fetchConfig() {
      try {
        const cfg = await GetCloudSyncConfig()
        if (cfg) {
          this.enabled = cfg.enabled
          this.serverUrl = cfg.server_url
          this.apiKey = cfg.api_key
          this.e2eEncryption = cfg.e2e_encryption
          this.autoSync = cfg.auto_sync
          this.syncInterval = cfg.sync_interval
          this.lastSyncTime = cfg.last_sync_time
        }
      } catch (err) {
        console.error('获取云同步配置失败:', err)
      }
    },

    async saveConfig(newConfig) {
      try {
        await SaveCloudSyncConfig({
          enabled: newConfig.enabled,
          server_url: newConfig.serverUrl,
          api_key: newConfig.apiKey,
          e2e_encryption: newConfig.e2eEncryption,
          auto_sync: newConfig.autoSync,
          sync_interval: Number(newConfig.syncInterval),
          last_sync_time: this.lastSyncTime
        })
        this.enabled = newConfig.enabled
        this.serverUrl = newConfig.serverUrl
        this.apiKey = newConfig.apiKey
        this.e2eEncryption = newConfig.e2eEncryption
        this.autoSync = newConfig.autoSync
        this.syncInterval = Number(newConfig.syncInterval)

        const clipStore = useClipboardStore()
        clipStore.showToast('云同步配置已保存')
      } catch (err) {
        console.error('保存云同步配置失败:', err)
      }
    },

    async testSync() {
      this.testing = true
      try {
        const res = await TestCloudSync()
        return res
      } catch (err) {
        return { success: false, error_message: String(err) }
      } finally {
        this.testing = false
      }
    },

    toggleModal(val) {
      this.showModal = val !== undefined ? val : !this.showModal
    }
  }
})
