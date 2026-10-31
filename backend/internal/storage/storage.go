// Package storage 抽象对象存储，业务代码不绑定具体云厂商。
package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ObjectStorage 对象存储接口。
type ObjectStorage interface {
	// Put 写入对象。
	Put(ctx context.Context, key string, data []byte, contentType string) error
	// Delete 删除对象。
	Delete(ctx context.Context, key string) error
	// SignedURL 生成带有效期的访问 URL。
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
	// Kind 返回驱动名称。
	Kind() string
}

// ---- 本地文件系统实现（开发/测试） ----

// FSStorage 基于本地磁盘的实现，签名 URL 指向 API 的媒体代理路由。
type FSStorage struct {
	root    string
	baseURL string
	secret  []byte
}

// NewFS 创建本地存储。
func NewFS(root, baseURL, secret string) (*FSStorage, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}
	return &FSStorage{root: abs, baseURL: strings.TrimRight(baseURL, "/"), secret: []byte(secret)}, nil
}

// Kind 实现 ObjectStorage。
func (f *FSStorage) Kind() string { return "fs" }

func (f *FSStorage) path(key string) (string, error) {
	clean := filepath.Clean(filepath.Join(f.root, filepath.FromSlash(key)))
	if !strings.HasPrefix(clean, f.root) {
		return "", fmt.Errorf("非法对象 key")
	}
	return clean, nil
}

// Put 实现 ObjectStorage。
func (f *FSStorage) Put(_ context.Context, key string, data []byte, _ string) error {
	p, err := f.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// Delete 实现 ObjectStorage。
func (f *FSStorage) Delete(_ context.Context, key string) error {
	p, err := f.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// SignedURL 实现 ObjectStorage。
func (f *FSStorage) SignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	if _, err := f.path(key); err != nil {
		return "", err
	}
	exp := time.Now().Add(ttl).Unix()
	sig := f.sign(key, exp)
	q := url.Values{}
	q.Set("exp", strconv.FormatInt(exp, 10))
	q.Set("sig", sig)
	return fmt.Sprintf("%s/api/v1/media/%s?%s", f.baseURL, key, q.Encode()), nil
}

func (f *FSStorage) sign(key string, exp int64) string {
	mac := hmac.New(sha256.New, f.secret)
	fmt.Fprintf(mac, "%s|%d", key, exp)
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 校验媒体代理签名。
func (f *FSStorage) Verify(key, sig string, exp int64) bool {
	if time.Now().Unix() > exp {
		return false
	}
	want := f.sign(key, exp)
	return hmac.Equal([]byte(want), []byte(sig))
}

// Read 读取对象内容，供媒体代理路由使用。
func (f *FSStorage) Read(key string) ([]byte, error) {
	p, err := f.path(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(p)
}

// ContentTypeOf 依据扩展名推断 Content-Type。
func ContentTypeOf(key string) string {
	switch strings.ToLower(filepath.Ext(key)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	}
	return "application/octet-stream"
}

// ---- 内存实现（单元测试） ----

// MemoryStorage 内存对象存储。
type MemoryStorage struct {
	objects map[string][]byte
}

// NewMemory 创建内存存储。
func NewMemory() *MemoryStorage { return &MemoryStorage{objects: map[string][]byte{}} }

// Kind 实现 ObjectStorage。
func (m *MemoryStorage) Kind() string { return "memory" }

// Put 实现 ObjectStorage。
func (m *MemoryStorage) Put(_ context.Context, key string, data []byte, _ string) error {
	m.objects[key] = append([]byte(nil), data...)
	return nil
}

// Delete 实现 ObjectStorage。
func (m *MemoryStorage) Delete(_ context.Context, key string) error {
	delete(m.objects, key)
	return nil
}

// SignedURL 实现 ObjectStorage。
func (m *MemoryStorage) SignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if _, ok := m.objects[key]; !ok {
		return "", fmt.Errorf("对象不存在")
	}
	return "memory://" + key, nil
}

// Get 读取对象（测试使用）。
func (m *MemoryStorage) Get(key string) ([]byte, bool) {
	b, ok := m.objects[key]
	return b, ok
}

// EncodeBase64 便于测试断言。
func EncodeBase64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

// BytesReader 便于测试。
func BytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
