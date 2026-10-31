package settings

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeLoader struct {
	data  map[string]json.RawMessage
	err   error
	calls int
}

func (f *fakeLoader) LoadSettings(context.Context) (map[string]json.RawMessage, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.data, nil
}

func raw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func TestDefaultsWhenMissing(t *testing.T) {
	svc := NewService(&fakeLoader{data: map[string]json.RawMessage{}}, time.Second)
	ctx := context.Background()
	if !svc.Bool(ctx, KeyCalculatorEnabled, false) {
		t.Fatal("缺省应回落到默认值 true")
	}
	if got := svc.Int(ctx, KeyKAnonymityThreshold, 0); got != 10 {
		t.Fatalf("K 匿名阈值默认应为 10，实际 %d", got)
	}
	if got := svc.Float(ctx, KeyCompatibilityWeight, 0); got != 0.6 {
		t.Fatalf("兼容度权重默认应为 0.6，实际 %v", got)
	}
}

func TestReadsStoredValues(t *testing.T) {
	loader := &fakeLoader{data: map[string]json.RawMessage{
		KeyLikesEnabled:         raw(false),
		KeyKAnonymityThreshold:  raw(20),
		KeyPreferenceWeight:     raw(0.5),
		KeyPolicyPrivacy:        raw("2027.03"),
	}}
	svc := NewService(loader, time.Second)
	ctx := context.Background()
	if svc.Bool(ctx, KeyLikesEnabled, true) {
		t.Fatal("应读取到 false")
	}
	if got := svc.Int(ctx, KeyKAnonymityThreshold, 10); got != 20 {
		t.Fatalf("期望 20，实际 %d", got)
	}
	if got := svc.String(ctx, KeyPolicyPrivacy, ""); got != "2027.03" {
		t.Fatalf("期望 2027.03，实际 %s", got)
	}
}

func TestCacheTTL(t *testing.T) {
	loader := &fakeLoader{data: map[string]json.RawMessage{KeyCalculatorEnabled: raw(true)}}
	svc := NewService(loader, 50*time.Millisecond)
	now := time.Now()
	svc.SetClock(func() time.Time { return now })
	ctx := context.Background()

	svc.Bool(ctx, KeyCalculatorEnabled, true)
	svc.Bool(ctx, KeyCalculatorEnabled, true)
	if loader.calls != 1 {
		t.Fatalf("缓存期内只应加载一次，实际 %d", loader.calls)
	}
	now = now.Add(time.Second)
	svc.Bool(ctx, KeyCalculatorEnabled, true)
	if loader.calls != 2 {
		t.Fatalf("缓存过期后应重新加载，实际 %d", loader.calls)
	}
}

func TestFallbackOnLoaderError(t *testing.T) {
	loader := &fakeLoader{data: map[string]json.RawMessage{KeyCalculatorEnabled: raw(false)}}
	svc := NewService(loader, time.Millisecond)
	ctx := context.Background()
	if svc.Bool(ctx, KeyCalculatorEnabled, true) {
		t.Fatal("首次应读到 false")
	}
	loader.err = errors.New("数据库不可用")
	time.Sleep(2 * time.Millisecond)
	if svc.Bool(ctx, KeyCalculatorEnabled, true) {
		t.Fatal("加载失败时应使用旧缓存")
	}
}

func TestFeaturesAndPublicConfig(t *testing.T) {
	loader := &fakeLoader{data: map[string]json.RawMessage{
		KeyChatEnabled:   raw(false),
		KeyMatchesEnabled: raw(true),
	}}
	svc := NewService(loader, time.Second)
	ctx := context.Background()
	features := svc.Features(ctx)
	if features[KeyChatEnabled] {
		t.Fatal("chat_enabled 默认关闭")
	}
	pub := svc.PublicConfig(ctx)
	if _, ok := pub["features"]; !ok {
		t.Fatal("公开配置应包含 features")
	}
	if _, ok := pub["settings"]; !ok {
		t.Fatal("公开配置应包含 settings")
	}
}
