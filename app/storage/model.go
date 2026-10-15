package storage

// ClipboardItem 剪贴板条目模型（包含本地强加密存储与未来云同步所需扩展字段）
type ClipboardItem struct {
	ID            string `json:"id"`             // 全局唯一 UUID
	Type          string `json:"type"`           // "text" 或 "image"
	EncryptedData []byte `json:"-"`              // 本地 SQLite 中 AES-256-GCM 加密存储的原始数据
	Content       string `json:"content"`        // 前端展示内容：纯文本明文，或图片 data:image/png;base64,...
	Preview       string `json:"preview"`        // 列表轻量摘要
	Hash          string `json:"hash"`           // 原文内容 SHA-256，用于去重
	ByteSize      int64  `json:"byte_size"`      // 字节大小
	Width         int    `json:"width"`          // 图片宽度
	Height        int    `json:"height"`         // 图片高度
	IsPinned      bool   `json:"is_pinned"`      // 是否置顶
	DeviceID      string `json:"device_id"`      // 本地设备标识（预留多端同步）
	SyncID        string `json:"sync_id"`        // 云端ID（预留云端同步）
	SyncStatus    int    `json:"sync_status"`    // 0:未同步, 1:已同步, 2:冲突
	Version       int64  `json:"version"`        // 乐观锁版本号
	CreatedAt     int64  `json:"created_at"`     // 毫秒时间戳
	UpdatedAt     int64  `json:"updated_at"`     // 毫秒时间戳
	IsDeleted     bool   `json:"is_deleted"`     // 软删除（支持云同步同步删除）
}

// SecuritySettings 存储本地加密设置
type SecuritySettings struct {
	HasMasterPassword bool `json:"has_master_password"` // 是否设置了主密码
	IsLocked          bool `json:"is_locked"`           // 是否处于锁定状态
	MaxHistoryCount   int  `json:"max_history_count"`   // 最大历史保留条数 (默认 500)
}

// CloudSyncConfig 预留云端同步配置
type CloudSyncConfig struct {
	Enabled       bool   `json:"enabled"`        // 是否启用云端同步
	ServerURL     string `json:"server_url"`     // 云端同步服务地址
	APIKey        string `json:"api_key"`        // 凭证 Token
	E2EEncryption bool   `json:"e2e_encryption"` // 是否开启端到端加密
	AutoSync      bool   `json:"auto_sync"`      // 是否自动同步
	SyncInterval  int    `json:"sync_interval"`  // 同步间隔秒数
	LastSyncTime  int64  `json:"last_sync_time"` // 上次成功同步时间戳
}

// AppSettings 应用偏好设置（开机自启、静默模式、唤醒快捷键）
type AppSettings struct {
	AutoStart  bool   `json:"auto_start"`
	SilentMode bool   `json:"silent_mode"`
	Hotkey     string `json:"hotkey"`
}

