package crontab

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"app/storage"

	"github.com/robfig/cron/v3"
)

// Scheduler 定时任务调度管理组件
type Scheduler struct {
	cronEngine *cron.Cron
	storage    *storage.Storage
	entryMap   map[string]cron.EntryID
	mu         sync.Mutex
	onLogSaved func(log *storage.CronLog)
}

// NewScheduler 创建调度器实例
func NewScheduler(store *storage.Storage, onLogSaved func(log *storage.CronLog)) *Scheduler {
	return &Scheduler{
		cronEngine: cron.New(cron.WithSeconds()),
		storage:    store,
		entryMap:   make(map[string]cron.EntryID),
		onLogSaved: onLogSaved,
	}
}

// Start 启动调度器并载入启用任务
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.storage.GetCronTasks()
	if err != nil {
		return fmt.Errorf("读取定时任务失败: %w", err)
	}

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		s.scheduleTaskLocked(task)
	}

	s.cronEngine.Start()
	return nil
}

// Stop 停止调度器运行
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cronEngine != nil {
		s.cronEngine.Stop()
	}
}

// Reload 重新加载所有启用的任务
func (s *Scheduler) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, entryID := range s.entryMap {
		s.cronEngine.Remove(entryID)
	}
	s.entryMap = make(map[string]cron.EntryID)

	tasks, err := s.storage.GetCronTasks()
	if err != nil {
		return fmt.Errorf("读取定时任务失败: %w", err)
	}

	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		s.scheduleTaskLocked(task)
	}

	return nil
}

// scheduleTaskLocked 在内部注册单个定时任务
func (s *Scheduler) scheduleTaskLocked(task *storage.CronTask) {
	t := *task
	entryID, err := s.cronEngine.AddFunc(t.CronExpr, func() {
		s.ExecuteTask(&t)
	})
	if err != nil {
		log.Printf("注册定时任务失败: %s, 规则: %s, 错误: %v", t.Name, t.CronExpr, err)
		return
	}
	s.entryMap[t.ID] = entryID
}

// RunTaskNow 手动立即触发执行任务
func (s *Scheduler) RunTaskNow(taskID string) (*storage.CronLog, error) {
	task, err := s.storage.GetCronTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务信息失败: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("任务不存在")
	}

	return s.ExecuteTask(task), nil
}

// ExecuteTask 执行指定任务脚本并持久化日志
func (s *Scheduler) ExecuteTask(task *storage.CronTask) *storage.CronLog {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", task.ScriptContent)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", task.ScriptContent)
	}

	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	runErr := cmd.Run()
	duration := time.Since(startTime).Milliseconds()

	status := "success"
	exitCode := 0

	if runErr != nil {
		status = "failure"
		exitCode = 1
		if exitError, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
		if outputBuf.Len() == 0 {
			outputBuf.WriteString(runErr.Error())
		}
	}

	logRecord := &storage.CronLog{
		TaskID:     task.ID,
		TaskName:   task.Name,
		Status:     status,
		Output:     outputBuf.String(),
		ExitCode:   exitCode,
		DurationMs: duration,
		ExecutedAt: startTime.UnixMilli(),
	}

	if err := s.storage.SaveCronLog(logRecord); err != nil {
		log.Printf("保存任务日志失败: %v", err)
	}

	if s.onLogSaved != nil {
		s.onLogSaved(logRecord)
	}

	return logRecord
}
