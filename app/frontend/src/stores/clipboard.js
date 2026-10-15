import { defineStore } from 'pinia'
import { GetItems, WriteItem, TogglePin, DeleteItem, ClearHistory, OpenImageInSystemViewer } from '../../wailsjs/go/main/App'

export const useClipboardStore = defineStore('clipboard', {
  state: () => ({
    items: [],
    search: '',
    filterType: 'all', // 'all' | 'text' | 'image'
    pinnedOnly: false,
    timeFilter: 'all', // 'all' | 'today' | '7days' | '30days'
    loading: false,
    selectedId: null,
    previewImage: null,
    toastMessage: '',
    searchRequestId: 0,
    debounceTimer: null,
  }),

  getters: {
    // 响应式多维综合过滤器 (搜索词即时过滤 + 类型过滤 + 置顶过滤 + 时间过滤)
    filteredItems: (state) => {
      let list = state.items

      // 1. 类型过滤 (文本 / 图片)
      if (state.filterType !== 'all') {
        list = list.filter(item => item.type === state.filterType)
      }

      // 2. 仅看置顶
      if (state.pinnedOnly) {
        list = list.filter(item => item.is_pinned)
      }

      // 3. 搜索关键词即时全文匹配 (0ms 即时响应，绝不误漏不匹配项)
      const q = state.search ? state.search.trim().toLowerCase() : ''
      if (q) {
        list = list.filter(item => {
          if (item.type === 'text') {
            const content = (item.content || '').toLowerCase()
            const preview = (item.preview || '').toLowerCase()
            return content.includes(q) || preview.includes(q)
          } else if (item.type === 'image') {
            // 用户搜索“图片”或“image”时命中图片，绝不因图片宽高分辨率数字而误搜
            return q === '图片' || q === 'image' || q === 'img'
          }
          return false
        })
      }

      // 4. 时间范围过滤
      if (state.timeFilter !== 'all') {
        const now = Date.now()
        if (state.timeFilter === 'today') {
          const todayStart = new Date().setHours(0, 0, 0, 0)
          list = list.filter(item => item.created_at >= todayStart)
        } else if (state.timeFilter === '7days') {
          const sevenDaysAgo = now - 7 * 86400 * 1000
          list = list.filter(item => item.created_at >= sevenDaysAgo)
        } else if (state.timeFilter === '30days') {
          const thirtyDaysAgo = now - 30 * 86400 * 1000
          list = list.filter(item => item.created_at >= thirtyDaysAgo)
        }
      }

      return list
    },
    totalCount: (state) => state.filteredItems.length,
    pinnedCount: (state) => state.filteredItems.filter(i => i.is_pinned).length,
  },

  actions: {
    async fetchItems() {
      const currentReq = ++this.searchRequestId
      this.loading = true
      try {
        const res = await GetItems(this.search, this.filterType, this.pinnedOnly, 200, 0)
        // 防竞态竞争：只有最新一次请求的结果才允许生效，杜绝旧请求覆盖新输入
        if (currentReq === this.searchRequestId) {
          this.items = res || []
        }
      } catch (err) {
        console.error('获取剪贴板历史失败:', err)
      } finally {
        if (currentReq === this.searchRequestId) {
          this.loading = false
        }
      }
    },

    async copyItem(id) {
      try {
        await WriteItem(id)
        this.showToast('已复制回系统剪贴板！')
      } catch (err) {
        this.showToast('复制失败: ' + err, true)
      }
    },

    async togglePin(id) {
      try {
        await TogglePin(id)
        await this.fetchItems()
      } catch (err) {
        this.showToast('置顶操作失败', true)
      }
    },

    async deleteItem(id) {
      try {
        await DeleteItem(id)
        await this.fetchItems()
        this.showToast('已删除记录')
      } catch (err) {
        this.showToast('删除失败', true)
      }
    },

    async clearAll(keepPinned = true) {
      try {
        await ClearHistory(keepPinned)
        await this.fetchItems()
        this.showToast(keepPinned ? '已清理未置顶历史' : '已清空全部历史')
      } catch (err) {
        this.showToast('清理历史失败', true)
      }
    },

    // 搜索输入触发：立即响应前端计算属性，并防抖向后端发起同步
    setSearch(text) {
      this.search = text
      if (this.debounceTimer) {
        clearTimeout(this.debounceTimer)
      }
      this.debounceTimer = setTimeout(() => {
        this.fetchItems()
      }, 200)
    },

    setFilter(type) {
      this.filterType = type
      this.pinnedOnly = false // 点击切换类型时自动重置“仅看置顶”，恢复完整浏览
      this.fetchItems()
    },

    setPinnedOnly(val) {
      this.pinnedOnly = val
      if (val) {
        this.filterType = 'all'
      }
      this.fetchItems()
    },

    async openSystemViewer(id) {
      try {
        await OpenImageInSystemViewer(id)
      } catch (err) {
        this.showToast('调起系统查看器失败: ' + err, true)
      }
    },

    openImageModal(imgSrc) {
      this.previewImage = imgSrc
    },

    closeImageModal() {
      this.previewImage = null
    },

    showToast(msg, isError = false) {
      this.toastMessage = msg
      setTimeout(() => {
        if (this.toastMessage === msg) {
          this.toastMessage = ''
        }
      }, 2500)
    },

    // 接收到后端 Wails 实时剪切板变更事件
    handleNewItem(newItem) {
      const idx = this.items.findIndex(i => i.id === newItem.id)
      if (idx !== -1) {
        this.items.splice(idx, 1)
      }
      this.items.unshift(newItem)
    }
  }
})
