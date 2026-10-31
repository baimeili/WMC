package calculator

import "context"

// InMemoryCounter 基于内存记录的计数器，用于单元测试与离线校验。
type InMemoryCounter struct {
	records []Record
}

// NewInMemoryCounter 构造内存计数器。
func NewInMemoryCounter(records []Record) *InMemoryCounter {
	return &InMemoryCounter{records: records}
}

// Total 返回全部样本权重之和。
func (c *InMemoryCounter) Total(context.Context) (float64, error) {
	var total float64
	for _, r := range c.records {
		total += r.Weight
	}
	return total, nil
}

// Count 返回满足条件的样本权重之和。
func (c *InMemoryCounter) Count(_ context.Context, f Filter) (float64, error) {
	var total float64
	for _, r := range c.records {
		if Matches(r, f) {
			total += r.Weight
		}
	}
	return total, nil
}
