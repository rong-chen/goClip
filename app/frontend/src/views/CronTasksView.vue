<template>
  <div class="cron-view">
    <!-- 顶部导航 -->
    <div class="view-header">
      <button class="back-btn" @click="$emit('back')">
        <ArrowLeft :size="16" />
        <span>返回剪贴板</span>
      </button>
      <h2 class="view-title">定时任务调度</h2>
      <button class="btn btn-primary-sm" @click="openCreateModal">
        <Plus :size="14" />
        <span>新建任务</span>
      </button>
    </div>

    <!-- 主体内容 -->
    <div class="view-content">
      <!-- 任务列表卡片 -->
      <section class="content-card">
        <div class="card-title-row">
          <Clock :size="18" class="card-icon" />
          <div class="card-titles">
            <h3>任务列表</h3>
            <p>管理后台常驻运行的定时脚本任务</p>
          </div>
        </div>

        <div v-if="cronStore.loading" class="empty-box">
          <span>正在加载任务列表...</span>
        </div>

        <div v-else-if="cronStore.tasks.length === 0" class="empty-box">
          <span>暂无定时任务</span>
        </div>

        <div v-else class="task-list">
          <div 
            v-for="task in cronStore.tasks" 
            :key="task.id" 
            class="task-item"
          >
            <!-- 头部：任务名称与周期 -->
            <div class="task-header">
              <div class="task-title-wrap">
                <span class="task-name">{{ task.name }}</span>
                <span class="cron-badge">{{ formatScheduleText(task.cron_expr) }}</span>
              </div>
            </div>

            <!-- 底部：一整排操作按钮与状态 -->
            <div class="task-footer">
              <div class="task-status-wrap">
                <label class="toggle-switch" title="启用或停用">
                  <input 
                    type="checkbox" 
                    :checked="task.enabled" 
                    @change="onToggleTask(task)"
                  />
                  <span class="slider"></span>
                </label>
                <span class="status-label">{{ task.enabled ? '已启用' : '已停用' }}</span>
              </div>

              <div class="task-actions">
                <!-- 手动立即执行 -->
                <button 
                  class="action-btn" 
                  :disabled="cronStore.runningTaskId === task.id"
                  @click="onRunNow(task)" 
                  title="立即执行一次"
                >
                  <Play :size="13" />
                  <span>运行</span>
                </button>

                <!-- 查看日志 -->
                <button 
                  class="action-btn" 
                  @click="openLogs(task)" 
                  title="查看执行日志"
                >
                  <FileText :size="13" />
                  <span>日志</span>
                </button>

                <!-- 编辑 -->
                <button 
                  class="action-btn" 
                  @click="openEditModal(task)" 
                  title="编辑任务"
                >
                  <Edit3 :size="13" />
                  <span>编辑</span>
                </button>

                <!-- 删除 -->
                <button 
                  class="action-btn danger-btn" 
                  @click="onDeleteTask(task.id)" 
                  title="删除任务"
                >
                  <Trash2 :size="13" />
                  <span>删除</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- 任务编辑弹窗 -->
    <div v-if="showModal" class="modal-backdrop" @click.self="showModal = false">
      <div class="modal-dialog">
        <div class="modal-header">
          <h3>{{ editingId ? '编辑任务' : '新建任务' }}</h3>
          <button class="close-btn" @click="showModal = false">
            <X :size="16" />
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label class="form-label">任务名称</label>
            <input 
              type="text" 
              v-model="form.name" 
              class="form-input" 
              placeholder="请输入任务名称"
            />
          </div>

          <!-- 手选执行周期类型 -->
          <div class="form-group">
            <label class="form-label">执行周期</label>
            <select v-model="form.scheduleType" class="form-select">
              <option value="">请选择执行周期</option>
              <option value="interval">间隔循环</option>
              <option value="daily">每天固定时间</option>
              <option value="weekly">每周固定时间</option>
              <option value="monthly">每月固定日期</option>
            </select>
          </div>

          <!-- 依据选择的周期类型展示具体配置 -->
          <div v-if="form.scheduleType === 'interval'" class="form-row">
            <div class="form-group flex-1">
              <label class="form-label">间隔时长</label>
              <input 
                type="number" 
                v-model.number="form.intervalValue" 
                class="form-input" 
                min="1" 
                placeholder="请输入数字"
              />
            </div>
            <div class="form-group flex-1">
              <label class="form-label">间隔单位</label>
              <select v-model="form.intervalUnit" class="form-select">
                <option value="minute">分钟</option>
                <option value="hour">小时</option>
              </select>
            </div>
          </div>

          <div v-else-if="form.scheduleType === 'daily'" class="form-group">
            <label class="form-label">执行时间</label>
            <input 
              type="text" 
              v-model="form.scheduleTime" 
              class="form-input" 
              placeholder="例如 17:00 或 17"
            />
          </div>

          <div v-else-if="form.scheduleType === 'weekly'" class="form-row">
            <div class="form-group flex-1">
              <label class="form-label">星期</label>
              <select v-model="form.weeklyDay" class="form-select">
                <option value="">请选择星期</option>
                <option value="1">星期一</option>
                <option value="2">星期二</option>
                <option value="3">星期三</option>
                <option value="4">星期四</option>
                <option value="5">星期五</option>
                <option value="6">星期六</option>
                <option value="0">星期日</option>
              </select>
            </div>
            <div class="form-group flex-1">
              <label class="form-label">执行时间</label>
              <input 
                type="text" 
                v-model="form.scheduleTime" 
                class="form-input" 
                placeholder="例如 17:00 或 17"
              />
            </div>
          </div>

          <div v-else-if="form.scheduleType === 'monthly'" class="form-row">
            <div class="form-group flex-1">
              <label class="form-label">每月日期</label>
              <select v-model="form.monthlyDay" class="form-select">
                <option value="">请选择日期</option>
                <option v-for="d in 31" :key="d" :value="d">{{ d }} 号</option>
              </select>
            </div>
            <div class="form-group flex-1">
              <label class="form-label">执行时间</label>
              <input 
                type="text" 
                v-model="form.scheduleTime" 
                class="form-input" 
                placeholder="例如 17:00 或 17"
              />
            </div>
          </div>

          <!-- 脚本内容输入与文件导入 -->
          <div class="form-group">
            <div class="form-label-row">
              <label class="form-label">脚本内容</label>
              <button type="button" class="import-link-btn" @click="triggerFileInput">
                <Upload :size="12" />
                <span>导入脚本文件</span>
              </button>
              <input 
                type="file" 
                ref="fileInputRef" 
                style="display: none" 
                @change="handleFileImport" 
              />
            </div>
            <textarea 
              v-model="form.script_content" 
              class="form-textarea" 
              rows="6" 
              placeholder="请输入脚本内容或通过右上角导入文件"
            ></textarea>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showModal = false">取消</button>
          <button class="btn btn-primary" @click="submitForm">保存</button>
        </div>
      </div>
    </div>

    <!-- 执行日志弹窗 -->
    <div v-if="showLogModal" class="modal-backdrop" @click.self="showLogModal = false">
      <div class="modal-dialog log-dialog">
        <div class="modal-header">
          <h3>{{ activeTask ? activeTask.name : '' }} 执行日志</h3>
          <div class="header-actions">
            <button class="btn btn-danger-sm" @click="onClearLogs">清空日志</button>
            <button class="close-btn" @click="showLogModal = false">
              <X :size="16" />
            </button>
          </div>
        </div>

        <div class="modal-body log-list-wrap">
          <div v-if="cronStore.logs.length === 0" class="empty-box">
            <span>暂无执行日志记录</span>
          </div>
          <div v-else class="log-list">
            <div 
              v-for="logItem in cronStore.logs" 
              :key="logItem.id" 
              class="log-item"
            >
              <div class="log-meta">
                <span :class="['status-tag', logItem.status]">
                  {{ logItem.status === 'success' ? '成功' : '失败' }}
                </span>
                <span class="log-time">{{ formatTime(logItem.executed_at) }}</span>
                <span class="log-duration">耗时 {{ logItem.duration_ms }}ms</span>
                <span class="log-exit">退出码 {{ logItem.exit_code }}</span>
              </div>
              <pre class="log-output">{{ logItem.output || '无输出内容' }}</pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { 
  ArrowLeft, 
  Clock, 
  Plus, 
  Play, 
  FileText, 
  Edit3, 
  Trash2, 
  X,
  Upload 
} from 'lucide-vue-next'
import { useCronStore } from '../stores/cron'

defineEmits(['back'])

const cronStore = useCronStore()

const showModal = ref(false)
const showLogModal = ref(false)
const editingId = ref('')
const activeTask = ref(null)
const fileInputRef = ref(null)

const form = reactive({
  name: '',
  scheduleType: '',
  intervalValue: null,
  intervalUnit: 'minute',
  weeklyDay: '',
  monthlyDay: '',
  scheduleTime: '',
  script_content: '',
  enabled: true
})

// 数字格式化补零
const padZero = (n) => String(n).padStart(2, '0')

// validateTimeInput 时间输入校验方法
const validateTimeInput = (timeStr) => {
  if (!timeStr) return null
  const val = timeStr.trim()

  let hh = 0
  let mm = 0

  if (val.includes(':')) {
    const parts = val.split(':')
    if (parts.length !== 2) return null
    hh = Number(parts[0])
    mm = Number(parts[1])
  } else {
    hh = Number(val)
    mm = 0
  }

  if (isNaN(hh) || isNaN(mm)) return null
  if (hh < 0 || hh > 23 || mm < 0 || mm > 59) return null

  return { hh, mm }
}

onMounted(async () => {
  await cronStore.fetchTasks()
})

// 触发本地文件导入
const triggerFileInput = () => {
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
    fileInputRef.value.click()
  }
}

// 读取用户选中的脚本文件内容
const handleFileImport = (e) => {
  const file = e.target.files[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = (event) => {
    form.script_content = event.target.result || ''
  }
  reader.readAsText(file)
}

// 打开创建弹窗
const openCreateModal = () => {
  editingId.value = ''
  form.name = ''
  form.scheduleType = ''
  form.intervalValue = null
  form.intervalUnit = 'minute'
  form.weeklyDay = ''
  form.monthlyDay = ''
  form.scheduleTime = ''
  form.script_content = ''
  form.enabled = true
  showModal.value = true
}

// 根据 Cron 表达式回显手选表单
const populateScheduleForm = (cronExpr) => {
  const parts = cronExpr.trim().split(/\s+/)
  if (parts.length < 6) return

  const sec = parts[0]
  const min = parts[1]
  const hour = parts[2]
  const day = parts[3]
  const month = parts[4]
  const week = parts[5]

  if (min.startsWith('*/')) {
    form.scheduleType = 'interval'
    form.intervalValue = parseInt(min.replace('*/', ''), 10)
    form.intervalUnit = 'minute'
    return
  }

  if (hour.startsWith('*/')) {
    form.scheduleType = 'interval'
    form.intervalValue = parseInt(hour.replace('*/', ''), 10)
    form.intervalUnit = 'hour'
    return
  }

  form.scheduleTime = `${padZero(hour)}:${padZero(min)}`

  if (week !== '*') {
    form.scheduleType = 'weekly'
    form.weeklyDay = week
    return
  }

  if (day !== '*') {
    form.scheduleType = 'monthly'
    form.monthlyDay = parseInt(day, 10)
    return
  }

  form.scheduleType = 'daily'
}

// 打开编辑弹窗
const openEditModal = (task) => {
  editingId.value = task.id
  form.name = task.name
  form.script_content = task.script_content
  form.enabled = task.enabled
  populateScheduleForm(task.cron_expr)
  showModal.value = true
}

// 根据手选配置生成标准 Cron 表达式
const generateCronExpr = () => {
  if (form.scheduleType === 'interval') {
    if (!form.intervalValue) return ''
    if (form.intervalUnit === 'hour') {
      return `0 0 */${form.intervalValue} * * *`
    }
    return `0 */${form.intervalValue} * * * *`
  }

  const parsedTime = validateTimeInput(form.scheduleTime)
  if (!parsedTime) {
    return ''
  }

  const { hh, mm } = parsedTime

  if (form.scheduleType === 'daily') {
    return `0 ${mm} ${hh} * * *`
  }

  if (form.scheduleType === 'weekly') {
    if (form.weeklyDay === '') return ''
    return `0 ${mm} ${hh} * * ${form.weeklyDay}`
  }

  if (form.scheduleType === 'monthly') {
    if (!form.monthlyDay) return ''
    return `0 ${mm} ${hh} ${form.monthlyDay} * *`
  }

  return ''
}

// 格式化展示文本
const formatScheduleText = (cronExpr) => {
  if (!cronExpr) return ''
  const parts = cronExpr.trim().split(/\s+/)
  if (parts.length < 6) return cronExpr

  const min = parts[1]
  const hour = parts[2]
  const day = parts[3]
  const week = parts[5]

  if (min.startsWith('*/')) {
    return `每隔 ${min.replace('*/', '')} 分钟`
  }

  if (hour.startsWith('*/')) {
    return `每隔 ${hour.replace('*/', '')} 小时`
  }

  const pad = (v) => String(v).padStart(2, '0')
  const timeStr = `${pad(hour)}:${pad(min)}`

  if (week !== '*') {
    const weekMap = {
      '1': '星期一',
      '2': '星期二',
      '3': '星期三',
      '4': '星期四',
      '5': '星期五',
      '6': '星期六',
      '0': '星期日'
    }
    return `每周 ${weekMap[week] || week} ${timeStr}`
  }

  if (day !== '*') {
    return `每月 ${day} 号 ${timeStr}`
  }

  return `每天 ${timeStr}`
}

// 提交保存任务
const submitForm = async () => {
  const cronExpr = generateCronExpr()
  if (!form.name || !cronExpr || !form.script_content) {
    return
  }
  await cronStore.saveTask({
    id: editingId.value,
    name: form.name,
    cron_expr: cronExpr,
    script_content: form.script_content,
    enabled: form.enabled
  })
  showModal.value = false
}

// 切换任务状态
const onToggleTask = async (task) => {
  await cronStore.toggleTask(task.id, !task.enabled)
}

// 删除任务
const onDeleteTask = async (id) => {
  await cronStore.deleteTask(id)
}

// 立即运行
const onRunNow = async (task) => {
  await cronStore.runTaskNow(task.id)
}

// 打开日志
const openLogs = async (task) => {
  activeTask.value = task
  cronStore.currentLogTaskId = task.id
  await cronStore.fetchLogs(task.id)
  showLogModal.value = true
}

// 清空日志
const onClearLogs = async () => {
  const taskId = activeTask.value ? activeTask.value.id : ''
  await cronStore.clearLogs(taskId)
}

// 格式化时间戳
const formatTime = (ts) => {
  if (!ts) return ''
  const d = new Date(ts)
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
</script>

<style scoped>
.cron-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  background-color: var(--bg-primary);
  overflow-y: auto;
}

.view-header {
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  background-color: var(--bg-primary);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  flex-shrink: 0;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-sm);
  color: #fff;
  font-size: 13px;
  cursor: pointer;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.view-title {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
}

.view-content {
  flex: 1;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.content-card {
  background: rgba(18, 18, 24, 0.6);
  border: 1px solid rgba(255, 255, 255, 0.07);
  border-radius: var(--radius-md);
  padding: 18px 20px;
}

.card-title-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.card-icon {
  color: var(--accent);
  margin-top: 2px;
}

.card-titles h3 {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

.card-titles p {
  font-size: 12px;
  color: var(--text-dim);
}

.empty-box {
  padding: 30px;
  text-align: center;
  color: var(--text-dim);
  font-size: 13px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-item {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 14px 16px;
  background: rgba(12, 12, 16, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: var(--radius-sm);
  transition: border-color 0.15s ease;
}

.task-item:hover {
  border-color: rgba(255, 255, 255, 0.1);
}

.task-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.task-title-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.task-name {
  font-weight: 600;
  color: var(--text-main);
  font-size: 14px;
}

.cron-badge {
  font-size: 11px;
  padding: 2px 8px;
  background: rgba(56, 189, 248, 0.15);
  color: var(--accent);
  border-radius: 4px;
  font-family: inherit;
}

.task-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: 10px;
  border-top: 1px solid rgba(255, 255, 255, 0.04);
}

.task-status-wrap {
  display: flex;
  align-items: center;
  gap: 8px;
}

.status-label {
  font-size: 12px;
  color: var(--text-muted);
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 10px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-muted);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.action-btn:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
}

.danger-btn:hover {
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger);
}

/* 开关样式 */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.15);
  transition: 0.2s;
  border-radius: 20px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.2s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--accent);
}

input:checked + .slider:before {
  transform: translateX(16px);
}

/* 按钮基础 */
.btn {
  padding: 6px 14px;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.btn-primary {
  background: var(--accent);
  color: #0f172a;
}

.btn-primary-sm {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: var(--accent);
  color: #0f172a;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 600;
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-main);
}

.btn-danger-sm {
  padding: 4px 8px;
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger);
  border-radius: var(--radius-sm);
  font-size: 12px;
}

/* 弹窗 */
.modal-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-dialog {
  width: 480px;
  background: var(--bg-secondary);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
}

.log-dialog {
  width: 650px;
  max-height: 80vh;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-body {
  padding: 18px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.modal-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 18px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.form-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.flex-1 {
  flex: 1;
}

.form-label {
  font-size: 12px;
  color: var(--text-muted);
}

.form-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.import-link-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--accent);
  font-size: 12px;
  background: transparent;
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
}

.import-link-btn:hover {
  background: rgba(56, 189, 248, 0.1);
}

.form-input, .form-select {
  width: 100%;
  height: 38px;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background-color: rgba(0, 0, 0, 0.3);
  color: var(--text-main);
  border-radius: var(--radius-sm);
  padding: 0 10px;
  font-size: 13px;
  font-family: inherit;
  transition: border-color 0.15s ease;
}

.form-select {
  -webkit-appearance: none;
  -moz-appearance: none;
  appearance: none;
  padding-right: 32px;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%2394a3b8' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 10px center;
  background-size: 14px;
  cursor: pointer;
}

.form-select option {
  background: var(--bg-secondary);
  color: #fff;
}

.form-input:focus, .form-select:focus, .form-textarea:focus {
  border-color: var(--border-focus);
}

.form-textarea {
  width: 100%;
  min-height: 120px;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.1);
  background-color: rgba(0, 0, 0, 0.3);
  color: var(--text-main);
  border-radius: var(--radius-sm);
  padding: 10px;
  font-size: 13px;
  font-family: monospace;
  resize: vertical;
  line-height: 1.5;
}

.log-list-wrap {
  max-height: 60vh;
  overflow-y: auto;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.log-item {
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.log-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
}

.status-tag {
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.status-tag.success {
  background: rgba(16, 185, 129, 0.15);
  color: var(--success);
}

.status-tag.failure {
  background: rgba(239, 68, 68, 0.15);
  color: var(--danger);
}

.log-time {
  color: var(--text-muted);
}

.log-duration, .log-exit {
  color: var(--text-dim);
}

.log-output {
  background: rgba(0, 0, 0, 0.4);
  padding: 8px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 12px;
  color: #e2e8f0;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 180px;
  overflow-y: auto;
}

.close-btn {
  color: var(--text-muted);
}

.close-btn:hover {
  color: #fff;
}
</style>
