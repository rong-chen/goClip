import { defineStore } from 'pinia'
import {
  GetCronTasks,
  SaveCronTask,
  DeleteCronTask,
  ToggleCronTask,
  RunCronTaskNow,
  GetCronLogs,
  ClearCronLogs
} from '../../wailsjs/go/main/App'

export const useCronStore = defineStore('cron', {
  state: () => ({
    tasks: [],
    logs: [],
    loading: false,
    runningTaskId: null,
    currentLogTaskId: null
  }),

  actions: {
    // 拉取所有定时任务
    async fetchTasks() {
      this.loading = true
      try {
        const res = await GetCronTasks()
        this.tasks = res || []
      } catch (err) {
        console.error('获取定时任务失败:', err)
      } finally {
        this.loading = false
      }
    },

    // 保存任务
    async saveTask(task) {
      await SaveCronTask(task)
      await this.fetchTasks()
    },

    // 删除任务
    async deleteTask(id) {
      await DeleteCronTask(id)
      await this.fetchTasks()
    },

    // 启停任务
    async toggleTask(id, enabled) {
      await ToggleCronTask(id, enabled)
      await this.fetchTasks()
    },

    // 手动立即触发执行
    async runTaskNow(id) {
      this.runningTaskId = id
      try {
        const logRecord = await RunCronTaskNow(id)
        if (this.currentLogTaskId === id || !this.currentLogTaskId) {
          await this.fetchLogs(this.currentLogTaskId)
        }
        return logRecord
      } finally {
        this.runningTaskId = null
      }
    },

    // 拉取执行日志
    async fetchLogs(taskId = '') {
      try {
        const res = await GetCronLogs(taskId, 50)
        this.logs = res || []
      } catch (err) {
        console.error('获取执行日志失败:', err)
      }
    },

    // 清空日志
    async clearLogs(taskId = '') {
      await ClearCronLogs(taskId)
      await this.fetchLogs(taskId)
    }
  }
})
