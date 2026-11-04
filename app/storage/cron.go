package storage

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GetCronTasks 查询所有定时任务
func (s *Storage) GetCronTasks() ([]*CronTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, cron_expr, script_content, enabled, created_at, updated_at FROM cron_tasks ORDER BY created_at DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询定时任务列表失败: %w", err)
	}
	defer rows.Close()

	var tasks []*CronTask
	for rows.Next() {
		var t CronTask
		var enabledInt int
		err := rows.Scan(&t.ID, &t.Name, &t.CronExpr, &t.ScriptContent, &enabledInt, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("解析任务记录失败: %w", err)
		}
		t.Enabled = enabledInt == 1
		tasks = append(tasks, &t)
	}
	return tasks, nil
}

// GetCronTaskByID 根据主键查询单个任务
func (s *Storage) GetCronTaskByID(id string) (*CronTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `SELECT id, name, cron_expr, script_content, enabled, created_at, updated_at FROM cron_tasks WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var t CronTask
	var enabledInt int
	err := row.Scan(&t.ID, &t.Name, &t.CronExpr, &t.ScriptContent, &enabledInt, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询任务失败: %w", err)
	}
	t.Enabled = enabledInt == 1
	return &t, nil
}

// SaveCronTask 保存或更新定时任务
func (s *Storage) SaveCronTask(task *CronTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if task.ID == "" {
		task.ID = uuid.NewString()
		task.CreatedAt = now
	}
	task.UpdatedAt = now

	enabledInt := 0
	if task.Enabled {
		enabledInt = 1
	}

	query := `
	INSERT INTO cron_tasks (id, name, cron_expr, script_content, enabled, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name = excluded.name,
		cron_expr = excluded.cron_expr,
		script_content = excluded.script_content,
		enabled = excluded.enabled,
		updated_at = excluded.updated_at
	`
	_, err := s.db.Exec(query, task.ID, task.Name, task.CronExpr, task.ScriptContent, enabledInt, task.CreatedAt, task.UpdatedAt)
	if err != nil {
		return fmt.Errorf("保存定时任务失败: %w", err)
	}
	return nil
}

// DeleteCronTask 删除定时任务及其关联日志
func (s *Storage) DeleteCronTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM cron_tasks WHERE id = ?`, id)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除任务失败: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM cron_logs WHERE task_id = ?`, id)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("删除任务日志失败: %w", err)
	}

	return tx.Commit()
}

// ToggleCronTask 切换任务启用禁用状态
func (s *Storage) ToggleCronTask(id string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	enabledInt := 0
	if enabled {
		enabledInt = 1
	}

	now := time.Now().UnixMilli()
	query := `UPDATE cron_tasks SET enabled = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, enabledInt, now, id)
	if err != nil {
		return fmt.Errorf("更新任务状态失败: %w", err)
	}
	return nil
}

// SaveCronLog 记录定时任务执行结果
func (s *Storage) SaveCronLog(log *CronLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if log.ID == "" {
		log.ID = uuid.NewString()
	}
	if log.ExecutedAt == 0 {
		log.ExecutedAt = time.Now().UnixMilli()
	}

	query := `
	INSERT INTO cron_logs (id, task_id, task_name, status, output, exit_code, duration_ms, executed_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(query, log.ID, log.TaskID, log.TaskName, log.Status, log.Output, log.ExitCode, log.DurationMs, log.ExecutedAt)
	if err != nil {
		return fmt.Errorf("保存执行日志失败: %w", err)
	}
	return nil
}

// GetCronLogs 查询执行日志列表
func (s *Storage) GetCronLogs(taskID string, limit int) ([]*CronLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var rows *sql.Rows
	var err error

	if taskID != "" {
		query := `SELECT id, task_id, task_name, status, output, exit_code, duration_ms, executed_at FROM cron_logs WHERE task_id = ? ORDER BY executed_at DESC LIMIT ?`
		rows, err = s.db.Query(query, taskID, limit)
	} else {
		query := `SELECT id, task_id, task_name, status, output, exit_code, duration_ms, executed_at FROM cron_logs ORDER BY executed_at DESC LIMIT ?`
		rows, err = s.db.Query(query, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("查询日志失败: %w", err)
	}
	defer rows.Close()

	var logs []*CronLog
	for rows.Next() {
		var l CronLog
		err := rows.Scan(&l.ID, &l.TaskID, &l.TaskName, &l.Status, &l.Output, &l.ExitCode, &l.DurationMs, &l.ExecutedAt)
		if err != nil {
			return nil, fmt.Errorf("解析日志记录失败: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

// ClearCronLogs 清空指定任务或全部日志
func (s *Storage) ClearCronLogs(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if taskID != "" {
		_, err := s.db.Exec(`DELETE FROM cron_logs WHERE task_id = ?`, taskID)
		return err
	}
	_, err := s.db.Exec(`DELETE FROM cron_logs`)
	return err
}
