package storage

import (
	"bytes"
	"testing"
)

func TestDefaultTransparentEncryption(t *testing.T) {
	s, err := NewStorage()
	if err != nil {
		t.Fatalf("初始化存储失败: %v", err)
	}
	defer s.Close()

	plainText := "Hello AES-256 Default Transparent Encryption!"
	raw := []byte(plainText)

	item := &ClipboardItem{
		ID:       "test-id-1",
		Type:     "text",
		Preview:  plainText,
		Hash:     "test-hash-1",
		ByteSize: int64(len(raw)),
	}

	if err := s.SaveItem(item, raw); err != nil {
		t.Fatalf("SaveItem 失败: %v", err)
	}

	// 1. 直接检查数据库中的原始 BLOB 字段，确认它已经被 AES-256-GCM 强加密（绝不是明文）
	var dbBlob []byte
	err = s.db.QueryRow("SELECT encrypted_data FROM clipboard_items WHERE id = ?", "test-id-1").Scan(&dbBlob)
	if err != nil {
		t.Fatalf("查询数据库 BLOB 失败: %v", err)
	}

	if bytes.Equal(dbBlob, raw) {
		t.Fatalf("严重漏洞: 数据库中的 BLOB 竟然是原始明文，没有加密！")
	}

	// 2. 检查通过 GetItems 读取，是否透明解密还原为正确明文
	items, err := s.GetItems("", "all", false, 10, 0)
	if err != nil {
		t.Fatalf("GetItems 失败: %v", err)
	}

	found := false
	for _, it := range items {
		if it.ID == "test-id-1" {
			found = true
			if it.Content != plainText {
				t.Fatalf("解密内容不匹配: 期望 %s, 实际 %s", plainText, it.Content)
			}
			break
		}
	}

	if !found {
		t.Fatalf("未能查到测试条目")
	}

	// 清理测试数据
	s.DeleteItem("test-id-1")
}
