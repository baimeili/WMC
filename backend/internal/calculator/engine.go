package calculator

import (
	"context"
	"sort"
	"time"
)

// Counter 条件计数抽象：
//   - InMemoryCounter：单元测试使用，遍历内存记录；
//   - SQLCounter：生产路径使用，PostgreSQL 加权聚合。
type Counter interface {
	Total(ctx context.Context) (float64, error)
	Count(ctx context.Context, f Filter) (float64, error)
}

// Engine 理论市场估算引擎。
type Engine struct {
	counter Counter
}

// NewEngine 构造引擎。
func NewEngine(counter Counter) *Engine { return &Engine{counter: counter} }

// FunnelNote 漏斗固定说明文案。
const FunnelNote = "按当前条件顺序计算，改变条件顺序会改变每一步的损失展示"

// Estimate 计算理论市场估算。
func (e *Engine) Estimate(ctx context.Context, f Filter) (*Estimate, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	total, err := e.counter.Total(ctx)
	if err != nil {
		return nil, err
	}

	// 顺序漏斗
	funnel := Funnel{TotalPopulation: total, Steps: make([]Step, 0, len(TheoreticalDimensions))}
	prev := total
	for _, d := range TheoreticalDimensions {
		applied := f.enabled(d.Key)
		count := prev
		if applied {
			count, err = e.counter.Count(ctx, f.upto(d.Key))
			if err != nil {
				return nil, err
			}
		}
		ratio := 0.0
		if total > 0 {
			ratio = count / total
		}
		drop := 0.0
		if prev > 0 {
			drop = (prev - count) / prev
		}
		funnel.Steps = append(funnel.Steps, Step{
			Key: d.Key, Label: d.Label, Count: count, Ratio: ratio, DropRatio: drop, Applied: applied,
		})
		prev = count
	}
	matched, err := e.counter.Count(ctx, f)
	if err != nil {
		return nil, err
	}
	funnel.FinalCount = matched
	if total > 0 {
		funnel.FinalRatio = matched / total
	}

	// Leave-One-Out 敏感性分析：逐条取消条件后重新计数
	sens := make([]Sensitivity, 0, len(TheoreticalDimensions))
	for _, d := range TheoreticalDimensions {
		if !f.enabled(d.Key) {
			continue
		}
		without, err := e.counter.Count(ctx, f.Without(d.Key))
		if err != nil {
			return nil, err
		}
		delta := 0.0
		if matched > 0 {
			delta = (without - matched) / matched
		} else if without > 0 {
			delta = 1
		}
		sens = append(sens, Sensitivity{
			Key: d.Key, Label: d.Label, BaseCount: matched, WithoutCount: without,
			DeltaRatio: delta, AddedCount: without - matched,
		})
	}
	sort.SliceStable(sens, func(i, j int) bool {
		if sens[i].DeltaRatio != sens[j].DeltaRatio {
			return sens[i].DeltaRatio > sens[j].DeltaRatio
		}
		return sens[i].Key < sens[j].Key
	})
	top := sens
	if len(top) > 3 {
		top = top[:3]
	}

	return &Estimate{
		TotalPopulation: total,
		MatchedCount:    matched,
		MatchedRatio:    funnel.FinalRatio,
		Funnel:          funnel,
		Sensitivity:     sens,
		TopFactors:      top,
		FunnelNote:      FunnelNote,
		GeneratedAt:     time.Now().UTC(),
	}, nil
}

// upto 返回只应用前 n 个维度（含当前维度）的条件副本。
func (f Filter) upto(key string) Filter {
	limit := len(TheoreticalDimensions)
	for i, d := range TheoreticalDimensions {
		if d.Key == key {
			limit = i + 1
			break
		}
	}
	out := Filter{}
	for _, d := range TheoreticalDimensions[:limit] {
		switch d.Key {
		case "gender":
			out.Genders = f.Genders
		case "age":
			out.MinAge, out.MaxAge = f.MinAge, f.MaxAge
		case "city":
			out.Cities = f.Cities
		case "marital_status":
			out.MaritalStatuses = f.MaritalStatuses
		case "children_state":
			out.ChildrenStates = f.ChildrenStates
		case "wants_children":
			out.WantsChildren = f.WantsChildren
		case "height":
			out.MinHeight, out.MaxHeight = f.MinHeight, f.MaxHeight
		case "education":
			out.MinEducation = f.MinEducation
		case "income":
			out.MinIncome = f.MinIncome
		}
	}
	return out
}
