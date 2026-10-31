// Package secure 提供手机号加密、Token 摘要等安全相关工具。
package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// PhoneHash 计算手机号的 HMAC-SHA256 摘要，用于唯一检索。数据库不保存手机号明文。
func PhoneHash(secret, phone string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(phone))
	return mac.Sum(nil)
}

// EncryptPhone 使用 AES-256-GCM 加密手机号，输出 nonce||ciphertext。
func EncryptPhone(key []byte, phone string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("初始化加密器失败: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("初始化 GCM 失败: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("生成随机数失败: %w", err)
	}
	return gcm.Seal(nonce, nonce, []byte(phone), nil), nil
}

// DecryptPhone 解密手机号，仅供有权限的管理操作使用。
func DecryptPhone(key, ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("密文长度非法")
	}
	nonce, data := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// TokenHash 计算 Refresh Token 摘要。
func TokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// RandomToken 生成 base64url 随机字符串。
func RandomToken(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// MaskPhone 手机号脱敏：138****8000。
func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// ConstantTimeEqual 常量时间比较。
func ConstantTimeEqual(a, b []byte) bool { return hmac.Equal(a, b) }
