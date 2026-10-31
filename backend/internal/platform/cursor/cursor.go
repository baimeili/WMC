// Package cursor 实现候选列表的 keyset 游标编解码。
// 禁止 offset 深分页，游标中只包含排序键（分数 + 用户 ID）。
package cursor

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Cursor 排序键。
type Cursor struct {
	Score  float64
	UserID string
}

// Encode 编码游标。
func Encode(c Cursor) string {
	raw := strconv.FormatFloat(c.Score, 'f', 6, 64) + "|" + c.UserID
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// Decode 解码游标。
func Decode(s string) (Cursor, error) {
	if strings.TrimSpace(s) == "" {
		return Cursor{}, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, fmt.Errorf("游标格式非法")
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return Cursor{}, errors.New("游标内容非法")
	}
	score, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || math.IsNaN(score) || math.IsInf(score, 0) {
		return Cursor{}, errors.New("游标分数非法")
	}
	return Cursor{Score: score, UserID: parts[1]}, nil
}

// Valid 判断游标是否非空。
func (c Cursor) Valid() bool { return c.UserID != "" }
