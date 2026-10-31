package photo

import "fmt"

// fmtSscan 解析字符串为整数（避免在 handler 中引入多重依赖）。
func fmtSscan(s string, out *int64) (int, error) {
	var v int64
	n, err := fmt.Sscanf(s, "%d", &v)
	if err != nil {
		return n, err
	}
	*out = v
	return n, nil
}
