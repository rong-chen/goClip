package storage

import (
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"strings"
	"time"

	"app/crypto"

	_ "modernc.org/sqlite"
)

type Storage struct {
	db        *sql.DB
	mu        sync.RWMutex
	encryptor *crypto.Encryptor
	dbPath    string
}

// NewStorage 初始化 SQLite 数据库与本地默认加密驱动 (AES-256-GCM)
func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	appDir := filepath.Join(homeDir, ".goclip")
	if err := os.MkdirAll(appDir, 0700); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	dbPath := filepath.Join(appDir, "clipboard.db")
	db, err := sql.Open("sqlite", dbPath+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("打开 SQLite 失败: %w", err)
	}
	db.SetMaxOpenConns(1)

	s := &Storage{
		db:     db,
		dbPath: dbPath,
	}

	if err := s.initSchema(); err != nil {
		return nil, fmt.Errorf("初始化表结构失败: %w", err)
	}

	if err := s.initDefaultEncryption(); err != nil {
		return nil, fmt.Errorf("初始化本地默认加密失败: %w", err)
	}

	return s, nil
}

// Close 关闭数据库
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS clipboard_items (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		encrypted_data BLOB NOT NULL,
		preview TEXT,
		hash TEXT NOT NULL,
		byte_size INTEGER NOT NULL,
		width INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		is_pinned INTEGER DEFAULT 0,
		device_id TEXT,
		sync_id TEXT,
		sync_status INTEGER DEFAULT 0,
		version INTEGER DEFAULT 1,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		is_deleted INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_clipboard_hash ON clipboard_items(hash);
	CREATE INDEX IF NOT EXISTS idx_clipboard_created ON clipboard_items(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_clipboard_pinned ON clipboard_items(is_pinned);
	CREATE INDEX IF NOT EXISTS idx_clipboard_deleted ON clipboard_items(is_deleted);

	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// initDefaultEncryption 初始化本地默认 AES-256-GCM 透明加密密钥
func (s *Storage) initDefaultEncryption() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	localKeyHex, err := s.getSetting("local_device_key")
	var keyBytes []byte
	if err != nil || localKeyHex == "" {
		keyBytes, err = crypto.GenerateRandomBytes(32)
		if err != nil {
			return err
		}
		if err := s.setSetting("local_device_key", hex.EncodeToString(keyBytes)); err != nil {
			return err
		}
	} else {
		keyBytes, err = hex.DecodeString(localKeyHex)
		if err != nil {
			return err
		}
	}

	enc, err := crypto.NewEncryptorFromKey(keyBytes)
	if err != nil {
		return err
	}
	s.encryptor = enc
	return nil
}

// SetMasterPassword 预留桩接口
func (s *Storage) SetMasterPassword(password string) error {
	return nil
}

// UnlockWithPassword 预留桩接口
func (s *Storage) UnlockWithPassword(password string) error {
	return nil
}

// Lock 预留桩接口
func (s *Storage) Lock() {}

// GetSecuritySettings 读取安全设置
func (s *Storage) GetSecuritySettings() SecuritySettings {
	return SecuritySettings{
		HasMasterPassword: false,
		IsLocked:          false,
		MaxHistoryCount:   500,
	}
}

// SaveItem 使用 AES-256-GCM 本地加密写入剪贴板条目
func (s *Storage) SaveItem(item *ClipboardItem, rawData []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.encryptor != nil {
		encrypted, err := s.encryptor.Encrypt(rawData)
		if err == nil {
			item.EncryptedData = encrypted
		} else {
			item.EncryptedData = rawData
		}
	} else {
		item.EncryptedData = rawData
	}

	now := time.Now().UnixMilli()
	item.UpdatedAt = now

	// 查重：若同 hash 已经存在，更新其时间避免重复插入
	var existingID string
	err := s.db.QueryRow("SELECT id FROM clipboard_items WHERE hash = ?", item.Hash).Scan(&existingID)
	if err == nil {
		_, err = s.db.Exec(`
			UPDATE clipboard_items 
			SET created_at = ?, updated_at = ?, is_deleted = 0, version = version + 1
			WHERE id = ?
		`, now, now, existingID)
		item.ID = existingID
		item.CreatedAt = now
		return err
	}

	item.CreatedAt = now
	_, err = s.db.Exec(`
		INSERT INTO clipboard_items (
			id, type, encrypted_data, preview, hash, byte_size, width, height, 
			is_pinned, device_id, sync_id, sync_status, version, created_at, updated_at, is_deleted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.Type, item.EncryptedData, item.Preview, item.Hash, item.ByteSize,
		item.Width, item.Height, item.IsPinned, item.DeviceID, item.SyncID,
		item.SyncStatus, item.Version, item.CreatedAt, item.UpdatedAt, 0)

	if err != nil {
		return err
	}

	s.pruneOldHistory(500)
	return nil
}

// GetItems 查询并自动解密剪贴板历史
func (s *Storage) GetItems(search string, filterType string, pinnedOnly bool, limit int, offset int) ([]*ClipboardItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query := `
		SELECT id, type, encrypted_data, preview, hash, byte_size, width, height, 
		       is_pinned, device_id, sync_id, sync_status, version, created_at, updated_at, is_deleted
		FROM clipboard_items 
		WHERE is_deleted = 0
	`
	var args []interface{}

	if filterType != "" && filterType != "all" {
		query += " AND type = ?"
		args = append(args, filterType)
	}

	if pinnedOnly {
		query += " AND is_pinned = 1"
	}

	if search != "" {
		trimmed := strings.TrimSpace(search)
		if trimmed != "" {
			lower := strings.ToLower(trimmed)
			if lower == "图片" || lower == "image" || lower == "img" {
				query += " AND type = 'image'"
			} else {
				query += " AND (type = 'text' AND preview LIKE ?)"
				args = append(args, "%"+trimmed+"%")
			}
		}
	}

	if limit <= 0 {
		limit = 100
	}
	query += " ORDER BY is_pinned DESC, created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ClipboardItem
	for rows.Next() {
		item := &ClipboardItem{}
		var pinnedInt, deletedInt int
		err := rows.Scan(
			&item.ID, &item.Type, &item.EncryptedData, &item.Preview, &item.Hash, &item.ByteSize,
			&item.Width, &item.Height, &pinnedInt, &item.DeviceID, &item.SyncID,
			&item.SyncStatus, &item.Version, &item.CreatedAt, &item.UpdatedAt, &deletedInt,
		)
		if err != nil {
			continue
		}
		item.IsPinned = pinnedInt == 1
		item.IsDeleted = deletedInt == 1

		var plain []byte
		if s.encryptor != nil {
			decrypted, err := s.encryptor.Decrypt(item.EncryptedData)
			if err == nil {
				plain = decrypted
			} else {
				plain = item.EncryptedData
			}
		} else {
			plain = item.EncryptedData
		}

		if item.Type == "image" {
			mimeType := "image/png"
			if len(plain) > 3 && plain[0] == 0xFF && plain[1] == 0xD8 && plain[2] == 0xFF {
				mimeType = "image/jpeg"
			} else if len(plain) > 6 && (string(plain[:6]) == "GIF87a" || string(plain[:6]) == "GIF89a") {
				mimeType = "image/gif"
			}
			item.Content = fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(plain))
		} else {
			item.Content = string(plain)
		}

		items = append(items, item)
	}

	return items, nil
}

// GetItemRawData 解密获取原始明文字节
func (s *Storage) GetItemRawData(id string) (string, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var itemType string
	var data []byte
	err := s.db.QueryRow("SELECT type, encrypted_data FROM clipboard_items WHERE id = ?", id).Scan(&itemType, &data)
	if err != nil {
		return "", nil, err
	}

	if s.encryptor != nil {
		plain, err := s.encryptor.Decrypt(data)
		if err == nil {
			return itemType, plain, nil
		}
	}

	return itemType, data, nil
}

// TogglePin 切换置顶
func (s *Storage) TogglePin(id string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var currentPinned int
	err := s.db.QueryRow("SELECT is_pinned FROM clipboard_items WHERE id = ?", id).Scan(&currentPinned)
	if err != nil {
		return false, err
	}

	newPinned := 1
	if currentPinned == 1 {
		newPinned = 0
	}

	now := time.Now().UnixMilli()
	_, err = s.db.Exec("UPDATE clipboard_items SET is_pinned = ?, updated_at = ?, version = version + 1 WHERE id = ?", newPinned, now, id)
	return newPinned == 1, err
}

// DeleteItem 软删除
func (s *Storage) DeleteItem(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	_, err := s.db.Exec("UPDATE clipboard_items SET is_deleted = 1, updated_at = ?, version = version + 1 WHERE id = ?", now, id)
	return err
}

// ClearHistory 清空历史
func (s *Storage) ClearHistory(keepPinned bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if keepPinned {
		_, err := s.db.Exec("UPDATE clipboard_items SET is_deleted = 1, updated_at = ?, version = version + 1 WHERE is_pinned = 0", now)
		return err
	}
	_, err := s.db.Exec("UPDATE clipboard_items SET is_deleted = 1, updated_at = ?, version = version + 1", now)
	return err
}

func (s *Storage) pruneOldHistory(maxLimit int) {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM clipboard_items WHERE is_deleted = 0 AND is_pinned = 0").Scan(&count)
	if count > maxLimit {
		overflow := count - maxLimit
		s.db.Exec(`
			DELETE FROM clipboard_items 
			WHERE id IN (
				SELECT id FROM clipboard_items 
				WHERE is_deleted = 0 AND is_pinned = 0 
				ORDER BY created_at ASC 
				LIMIT ?
			)
		`, overflow)
	}
}

func (s *Storage) getSetting(key string) (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	return val, err
}

func (s *Storage) setSetting(key string, value string) error {
	_, err := s.db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

// GetAppSettings 读取应用偏好设置
func (s *Storage) GetAppSettings(defaultHotkey string) AppSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	autoStart, _ := s.getSetting("auto_start")
	silentMode, _ := s.getSetting("silent_mode")
	hk, _ := s.getSetting("hotkey")
	if hk == "" {
		hk = defaultHotkey
	}

	return AppSettings{
		AutoStart:  autoStart == "true",
		SilentMode: silentMode == "true",
		Hotkey:     hk,
	}
}

// SaveAppSettings 保存应用偏好设置
func (s *Storage) SaveAppSettings(settings AppSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.setSetting("auto_start", fmt.Sprintf("%t", settings.AutoStart)); err != nil {
		return err
	}
	if err := s.setSetting("silent_mode", fmt.Sprintf("%t", settings.SilentMode)); err != nil {
		return err
	}
	return s.setSetting("hotkey", settings.Hotkey)
}

