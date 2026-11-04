package crontab

import (
	"strings"
	"testing"

	"app/storage"
)

// TestExecuteTask 验证脚本执行与状态捕获
func TestExecuteTask(t *testing.T) {
	store, err := storage.NewStorage()
	if err != nil {
		t.Fatalf("初始化存储失败: %v", err)
	}
	defer store.Close()

	scheduler := NewScheduler(store, nil)

	// 测试成功执行
	successTask := &storage.CronTask{
		ID:            "test-success-id",
		Name:          "测试成功任务",
		CronExpr:      "0 * * * * *",
		ScriptContent: "echo 'goclip test run'",
		Enabled:       true,
	}

	successLog := scheduler.ExecuteTask(successTask)
	if successLog.Status != "success" {
		t.Fatalf("期望状态 success, 实际: %s, 输出: %s", successLog.Status, successLog.Output)
	}
	if !strings.Contains(successLog.Output, "goclip test run") {
		t.Fatalf("输出未包含预期内容: %s", successLog.Output)
	}

	// 测试失败退出码与错误捕获
	failTask := &storage.CronTask{
		ID:            "test-fail-id",
		Name:          "测试失败任务",
		CronExpr:      "0 * * * * *",
		ScriptContent: "echo 'err detail' >&2; exit 2",
		Enabled:       true,
	}

	failLog := scheduler.ExecuteTask(failTask)
	if failLog.Status != "failure" {
		t.Fatalf("期望状态 failure, 实际: %s", failLog.Status)
	}
	if failLog.ExitCode != 2 {
		t.Fatalf("期望退出码 2, 实际: %d", failLog.ExitCode)
	}
	if !strings.Contains(failLog.Output, "err detail") {
		t.Fatalf("输出未包含错误信息: %s", failLog.Output)
	}

	// 清理测试日志
	_ = store.ClearCronLogs("test-success-id")
	_ = store.ClearCronLogs("test-fail-id")
}
