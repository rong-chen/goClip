import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GetAppSettings, SaveAppSettings } from '../../wailsjs/go/main/App'

export const useAppSettingsStore = defineStore('appSettings', () => {
  const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad|iPod/i.test(navigator.userAgent)
  const defaultHK = isMac ? 'Cmd+Shift+V' : 'Ctrl+Shift+V'
  const autoStart = ref(false)
  const silentMode = ref(false)
  const hotkey = ref(defaultHK)
  const loading = ref(false)
  const saving = ref(false)
  const message = ref('')

  const fetchSettings = async () => {
    loading.value = true
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const res = await GetAppSettings()
        if (res) {
          autoStart.value = res.auto_start
          silentMode.value = res.silent_mode
          hotkey.value = res.hotkey || defaultHK
        }
      }
    } catch (e) {
      console.warn('获取应用设置异常:', e)
    } finally {
      loading.value = false
    }
  }

  const saveSettings = async (newSettings) => {
    saving.value = true
    message.value = ''
    try {
      if (window.go && window.go.main && window.go.main.App) {
        const payload = {
          auto_start: newSettings.autoStart ?? autoStart.value,
          silent_mode: newSettings.silentMode ?? silentMode.value,
          hotkey: newSettings.hotkey ?? hotkey.value
        }
        await SaveAppSettings(payload)
        autoStart.value = payload.auto_start
        silentMode.value = payload.silent_mode
        hotkey.value = payload.hotkey
        return { success: true }
      }
      return { success: true }
    } catch (e) {
      console.error('保存设置异常:', e)
      return { success: false, error: e.toString() }
    } finally {
      saving.value = false
    }
  }

  return {
    autoStart,
    silentMode,
    hotkey,
    loading,
    saving,
    message,
    fetchSettings,
    saveSettings
  }
})
