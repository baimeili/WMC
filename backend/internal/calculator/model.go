// Package calculator 实现「理论婚恋市场」计算引擎与相关服务。
// 引擎为纯函数：输入条件 + 基线记录（或 SQL 计数），输出估算结果，可独立单元测试。
package calculator

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Record 基准人口加权样本。
type Record struct {
	Gender           int16
	Age              int16
	ProvinceCode     string
	CityCode         string
	HeightCM         *int16
	EducationLevel   int16
	AnnualIncomeBand int16
	MaritalStatus    int16
	ChildrenCount    int16
	WantsChildren    int16
	Weight           float64
}

// Filter 一组筛选条件；字段为 nil/空表示该维度不限制。
type Filter struct {
	Genders         []int16
	MinAge          *int
	MaxAge          *int
	Cities          []string
	MaritalStatuses []int16
	ChildrenStates  []int16
	WantsChildren   []int16
	MinHeight       *int
	MaxHeight       *int
	MinEducation    *int16
	MinIncome       *int16
}

// Dimension 漏斗/敏感性分析使用的条件维度。
type Dimension struct {
	Key   string
	Label string
}

// TheoreticalDimensions 理论市场的固定漏斗顺序。
// 顺序会影响每一步的损失展示，因此页面必须标注「按当前条件顺序计算」。
var TheoreticalDimensions = []Dimension{
	{"gender", "目标性别"},
	{"age", "年龄"},
	{"city", "城市"},
	{"marital_status", "婚姻状态"},
	{"children_state", "子女情况"},
	{"wants_children", "生育意愿"},
	{"height", "身高"},
	{"education", "学历"},
	{"income", "年收入"},
}

// HasGender 是否限定性别。
func (f Filter) HasGender() bool { return len(f.Genders) > 0 }

// HasAge 是否限定年龄。
func (f Filter) HasAge() bool { return f.MinAge != nil || f.MaxAge != nil }

// HasCity 是否限定城市。
func (f Filter) HasCity() bool { return len(f.Cities) > 0 }

// HasMarital 是否限定婚姻状态。
func (f Filter) HasMarital() bool { return len(f.MaritalStatuses) > 0 }

// HasChildren 是否限定子女情况。
func (f Filter) HasChildren() bool { return len(f.ChildrenStates) > 0 }

// HasWantsChildren 是否限定生育意愿。
func (f Filter) HasWantsChildren() bool { return len(f.WantsChildren) > 0 }

// HasHeight 是否限定身高。
func (f Filter) HasHeight() bool { return f.MinHeight != nil || f.MaxHeight != nil }

// HasEducation 是否限定学历。
func (f Filter) HasEducation() bool { return f.MinEducation != nil }

// HasIncome 是否限定收入。
func (f Filter) HasIncome() bool { return f.MinIncome != nil }

// enabled 判断某维度是否生效。
func (f Filter) enabled(key string) bool {
	switch key {
	case "gender":
		return f.HasGender()
	case "age":
		return f.HasAge()
	case "city":
		return f.HasCity()
	case "marital_status":
		return f.HasMarital()
	case "children_state":
		return f.HasChildren()
	case "wants_children":
		return f.HasWantsChildren()
	case "height":
		return f.HasHeight()
	case "education":
		return f.HasEducation()
	case "income":
		return f.HasIncome()
	}
	return false
}

// EnabledDimensions 返回当前实际生效的维度（保持固定顺序）。
func (f Filter) EnabledDimensions() []Dimension {
	out := make([]Dimension, 0, len(TheoreticalDimensions))
	for _, d := range TheoreticalDimensions {
		if f.enabled(d.Key) {
			out = append(out, d)
		}
	}
	return out
}

// Without 返回去掉某个维度约束后的副本，用于 Leave-One-Out 敏感性分析。
func (f Filter) Without(key string) Filter {
	out := f
	switch key {
	case "gender":
		out.Genders = nil
	case "age":
		out.MinAge, out.MaxAge = nil, nil
	case "city":
		out.Cities = nil
	case "marital_status":
		out.MaritalStatuses = nil
	case "children_state":
		out.ChildrenStates = nil
	case "wants_children":
		out.WantsChildren = nil
	case "height":
		out.MinHeight, out.MaxHeight = nil, nil
	case "education":
		out.MinEducation = nil
	case "income":
		out.MinIncome = nil
	}
	return out
}

// WithAge 便捷构造年龄条件。
func WithAge(min, max int) Filter { return Filter{MinAge: &min, MaxAge: &max} }

// Matches 判断单条记录是否满足条件（内存计数实现与 SQL 实现语义必须一致）。
func Matches(r Record, f Filter) bool {
	if len(f.Genders) > 0 && !containsInt16(f.Genders, r.Gender) {
		return false
	}
	if f.MinAge != nil && int(r.Age) < *f.MinAge {
		return false
	}
	if f.MaxAge != nil && int(r.Age) > *f.MaxAge {
		return false
	}
	if len(f.Cities) > 0 && !containsString(f.Cities, r.CityCode) {
		return false
	}
	if len(f.MaritalStatuses) > 0 && !containsInt16(f.MaritalStatuses, r.MaritalStatus) {
		return false
	}
	if len(f.ChildrenStates) > 0 && !containsInt16(f.ChildrenStates, ChildrenState(r.ChildrenCount)) {
		return false
	}
	if len(f.WantsChildren) > 0 && !containsInt16(f.WantsChildren, r.WantsChildren) {
		return false
	}
	if f.MinHeight != nil && (r.HeightCM == nil || int(*r.HeightCM) < *f.MinHeight) {
		return false
	}
	if f.MaxHeight != nil && (r.HeightCM == nil || int(*r.HeightCM) > *f.MaxHeight) {
		return false
	}
	if f.MinEducation != nil && r.EducationLevel < *f.MinEducation {
		return false
	}
	if f.MinIncome != nil && r.AnnualIncomeBand < *f.MinIncome {
		return false
	}
	return true
}

// ChildrenState 由子女数量推导子女状态编码。
func ChildrenState(count int16) int16 {
	if count > 0 {
		return 1
	}
	return 0
}

func containsInt16(list []int16, v int16) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func containsString(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// Step 漏斗中的一层。
type Step struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Count      float64 `json:"count"`
	Ratio      float64 `json:"ratio"`
	DropRatio  float64 `json:"drop_ratio"`
	Applied    bool    `json:"applied"`
}

// Funnel 顺序漏斗结果。
type Funnel struct {
	TotalPopulation float64 `json:"total_population"`
	Steps           []Step  `json:"steps"`
	FinalCount      float64 `json:"final_count"`
	FinalRatio      float64 `json:"final_ratio"`
}

// Sensitivity 单条件 Leave-One-Out 结果。
type Sensitivity struct {
	Key          string  `json:"key"`
	Label        string  `json:"label"`
	BaseCount    float64 `json:"base_count"`
	WithoutCount float64 `json:"without_count"`
	DeltaRatio   float64 `json:"delta_ratio"`
	AddedCount   float64 `json:"added_count"`
}

// Estimate 理论市场估算结果。
type Estimate struct {
	DatasetID       uuid.UUID     `json:"dataset_id"`
	DatasetName     string        `json:"dataset_name"`
	DatasetVersion  string        `json:"dataset_version"`
	SourceName      string        `json:"source_name"`
	Disclaimer      string        `json:"disclaimer"`
	TotalPopulation float64       `json:"total_population"`
	MatchedCount    float64       `json:"matched_count"`
	MatchedRatio    float64       `json:"matched_ratio"`
	Funnel          Funnel        `json:"funnel"`
	Sensitivity     []Sensitivity `json:"sensitivity"`
	TopFactors      []Sensitivity `json:"top_factors"`
	FunnelNote      string        `json:"funnel_note"`
	GeneratedAt     time.Time     `json:"generated_at"`
}

// Validate 校验条件合法性。
func (f Filter) Validate() error {
	if f.MinAge != nil && *f.MinAge < 18 {
		return fmt.Errorf("最小年龄不能小于 18 岁")
	}
	if f.MaxAge != nil && *f.MaxAge > 100 {
		return fmt.Errorf("最大年龄不能大于 100 岁")
	}
	if f.MinAge != nil && f.MaxAge != nil && *f.MinAge > *f.MaxAge {
		return fmt.Errorf("最小年龄不能大于最大年龄")
	}
	if f.MinHeight != nil && *f.MinHeight < 100 {
		return fmt.Errorf("身高条件不合理")
	}
	if f.MaxHeight != nil && *f.MaxHeight > 250 {
		return fmt.Errorf("身高条件不合理")
	}
	if f.MinHeight != nil && f.MaxHeight != nil && *f.MinHeight > *f.MaxHeight {
		return fmt.Errorf("最小身高不能大于最大身高")
	}
	return nil
}
