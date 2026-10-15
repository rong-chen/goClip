package clipboard

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"strings"
	"sync"
	"time"

	"app/crypto"
	"app/storage"

	"github.com/google/uuid"
	"golang.design/x/clipboard"
)

type EventCallback func(item *storage.ClipboardItem)

type Monitor struct {
	storage       *storage.Storage
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	lastHash      string
	lastWriteHash string
	onNewItem     EventCallback
	isPaused      bool
}

// NewMonitor 创建跨平台剪贴板监听器
func NewMonitor(store *storage.Storage, callback EventCallback) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Monitor{
		storage:   store,
		ctx:       ctx,
		cancel:    cancel,
		onNewItem: callback,
	}
}

// Start 启动跨平台剪贴板监听
func (m *Monitor) Start() error {
	if err := clipboard.Init(); err != nil {
		log.Printf("系统剪贴板监听器初始化提示: %v", err)
	}

	// 统一 Watch 管道
	go m.watchUnified()
	go m.pollingFallback()

	return nil
}

// Stop 停止监听
func (m *Monitor) Stop() {
	if m.cancel != nil {
		m.cancel()
	}
}

// WriteText 将文本写回系统剪贴板
func (m *Monitor) WriteText(text string) error {
	raw := []byte(text)
	hash := crypto.SHA256String(raw)

	m.mu.Lock()
	m.lastWriteHash = hash
	m.lastHash = hash
	m.mu.Unlock()

	_, err := clipboard.Write(m.ctx, clipboard.FmtText, raw)
	return err
}

// WriteImage 将 PNG 图片写回系统剪贴板
func (m *Monitor) WriteImage(rawPngBytes []byte) error {
	hash := crypto.SHA256String(rawPngBytes)

	m.mu.Lock()
	m.lastWriteHash = hash
	m.lastHash = hash
	m.mu.Unlock()

	_, err := clipboard.Write(m.ctx, clipboard.FmtImage, rawPngBytes)
	return err
}

func (m *Monitor) watchUnified() {
	ch := clipboard.Watch(m.ctx)
	for data := range ch {
		if data.Format == clipboard.FmtText {
			m.handleNewText(string(data.Bytes))
		} else if data.Format == clipboard.FmtImage {
			m.handleNewImage(data.Bytes)
		}
	}
}

func (m *Monitor) pollingFallback() {
	ticker := time.NewTicker(600 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			paused := m.isPaused
			m.mu.Unlock()
			if paused {
				continue
			}

			// 优先尝试读取图片
			imgBytes, err := clipboard.Read(m.ctx, clipboard.FmtImage)
			if err == nil && len(imgBytes) > 0 {
				m.handleNewImage(imgBytes)
				continue
			}

			// 读取文本
			textBytes, err := clipboard.Read(m.ctx, clipboard.FmtText)
			if err == nil && len(textBytes) > 0 {
				m.handleNewText(string(textBytes))
			}
		}
	}
}

const (
	// 单条文本最大入库限制 (5 MB，相当于数百个中篇小说的字数，足以容纳超大代码片段)
	MaxTextBytes = 5 * 1024 * 1024
	// 单张图片最大入库限制 (25 MB，足以容纳 4K/8K 超高清全屏截图)
	MaxImageBytes = 25 * 1024 * 1024
)

func (m *Monitor) handleNewText(text string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}

	rawBytes := []byte(text)
	// 针对超大文本（如几百MB日志、巨大数据导出）做内存安全保护
	if len(rawBytes) > MaxTextBytes {
		log.Printf("剪贴板文本过大 (%.2f MB)，超出 5MB 安全上限，已自动跳过以保护内存", float64(len(rawBytes))/(1024*1024))
		return
	}

	hash := crypto.SHA256String(rawBytes)

	m.mu.Lock()
	if m.isPaused || hash == m.lastHash || hash == m.lastWriteHash {
		m.mu.Unlock()
		return
	}
	m.lastHash = hash
	m.mu.Unlock()

	preview := text
	if len([]rune(preview)) > 120 {
		preview = string([]rune(preview)[:120]) + "..."
	}

	item := &storage.ClipboardItem{
		ID:         uuid.NewString(),
		Type:       "text",
		Content:    text,
		Preview:    preview,
		Hash:       hash,
		ByteSize:   int64(len(rawBytes)),
		DeviceID:   "local-device",
		SyncStatus: 0,
		Version:    1,
	}

	if err := m.storage.SaveItem(item, rawBytes); err != nil {
		return
	}

	if m.onNewItem != nil {
		m.onNewItem(item)
	}
}

func (m *Monitor) handleNewImage(data []byte) {
	if len(data) == 0 {
		return
	}

	// 针对超大图片（如未压缩的高清 RAW 或 8K 巨图）做安全拦截保护
	if len(data) > MaxImageBytes {
		log.Printf("剪贴板图片过大 (%.2f MB)，超出 25MB 安全上限，已自动跳过以保护内存与数据库稳定", float64(len(data))/(1024*1024))
		return
	}

	hash := crypto.SHA256String(data)

	m.mu.Lock()
	if m.isPaused || hash == m.lastHash || hash == m.lastWriteHash {
		m.mu.Unlock()
		return
	}
	m.lastHash = hash
	m.mu.Unlock()

	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	width, height := 0, 0
	if err == nil {
		width = cfg.Width
		height = cfg.Height
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	content := fmt.Sprintf("data:image/png;base64,%s", b64)
	preview := fmt.Sprintf("[图片] %dx%d (%.1f KB)", width, height, float64(len(data))/1024.0)

	item := &storage.ClipboardItem{
		ID:         uuid.NewString(),
		Type:       "image",
		Content:    content,
		Preview:    preview,
		Hash:       hash,
		ByteSize:   int64(len(data)),
		Width:      width,
		Height:     height,
		DeviceID:   "local-device",
		SyncStatus: 0,
		Version:    1,
	}

	if err := m.storage.SaveItem(item, data); err != nil {
		return
	}

	if m.onNewItem != nil {
		m.onNewItem(item)
	}
}
