package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

const (
	NonceSize = 12
	KeyLen    = 32 // 256 位密钥
)

var (
	ErrCiphertextTooShort = errors.New("密文数据长度不足或已损坏")
)

// Encryptor 提供 AES-256-GCM 本地加密与解密支持
type Encryptor struct {
	key []byte
}

// NewEncryptorFromKey 使用 32 字节密钥初始化加密器
func NewEncryptorFromKey(key []byte) (*Encryptor, error) {
	if len(key) != KeyLen {
		return nil, fmt.Errorf("密钥长度必须为 %d 字节，当前为 %d", KeyLen, len(key))
	}
	k := make([]byte, KeyLen)
	copy(k, key)
	return &Encryptor{key: k}, nil
}

// GenerateRandomBytes 生成密码学安全强随机字节
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

// Encrypt 使用 AES-256-GCM 加密明文，自动生成 Nonce 并拼接在头部
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt 使用 AES-256-GCM 解密密文
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// SHA256String 计算数据的 SHA-256 哈希串用于剪贴板查重
func SHA256String(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
