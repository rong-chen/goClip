package cloudsync

import (
	"context"
	"errors"
	"time"

	"app/storage"
)

var (
	ErrCloudSyncDisabled = errors.New("云端同步当前未启用（当前运行于本地强加密模式）")
)

type SyncResult struct {
	Success      bool   `json:"success"`
	Uploaded     int    `json:"uploaded"`
	Downloaded   int    `json:"downloaded"`
	Conflicts    int    `json:"conflicts"`
	ErrorMessage string `json:"error_message,omitempty"`
	SyncTime     int64  `json:"sync_time"`
}

type SyncManager struct {
	config storage.CloudSyncConfig
}

func NewSyncManager() *SyncManager {
	return &SyncManager{
		config: storage.CloudSyncConfig{
			Enabled:       false,
			ServerURL:     "https://sync.goclip.internal/api/v1",
			E2EEncryption: true,
			AutoSync:      false,
			SyncInterval:  60,
			LastSyncTime:  0,
		},
	}
}

func (m *SyncManager) GetConfig() storage.CloudSyncConfig {
	return m.config
}

func (m *SyncManager) UpdateConfig(cfg storage.CloudSyncConfig) {
	m.config = cfg
}

func (m *SyncManager) TestSync(ctx context.Context) (*SyncResult, error) {
	if !m.config.Enabled {
		return &SyncResult{
			Success:      false,
			ErrorMessage: "云同步功能处于关闭状态，数据安全保存在本地 SQLite 中",
			SyncTime:     time.Now().UnixMilli(),
		}, ErrCloudSyncDisabled
	}

	m.config.LastSyncTime = time.Now().UnixMilli()
	return &SyncResult{
		Success:    true,
		Uploaded:   0,
		Downloaded: 0,
		Conflicts:  0,
		SyncTime:   m.config.LastSyncTime,
	}, nil
}
