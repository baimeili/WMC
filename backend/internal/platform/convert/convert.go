// Package convert 提供数据库类型与业务类型之间的转换辅助。
package convert

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// Time 将 pgtype.Timestamptz 转为 time.Time（无效值返回零值）。
func Time(ts pgtype.Timestamptz) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

// TimePtr 将 pgtype.Timestamptz 转为 *time.Time。
func TimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

// ToTime 将 time.Time 转为 pgtype.Timestamptz。
func ToTime(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// ToTimePtr 将 *time.Time 转为 pgtype.Timestamptz。
func ToTimePtr(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// JSONStrings 将 JSONB 内容解析为字符串数组（无效或空返回空切片）。
func JSONStrings(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return []string{}
	}
	return out
}

// MustJSON 序列化为 JSON，失败返回 "null"。
func MustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}
	return b
}

// JSONMap 将 JSONB 内容解析为 map。
func JSONMap(raw []byte) map[string]any {
	out := map[string]any{}
	if len(raw) == 0 {
		return out
	}
	_ = json.Unmarshal(raw, &out)
	return out
}
