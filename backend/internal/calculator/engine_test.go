package calculator

import (
	"context"
	"math"
	"testing"
)

func ptr[T any](v T) *T { return &v }

func sampleRecords() []Record {
	h165, h180, h172 := int16(165), int16(180), int16(172)
	return []Record{
		{Gender: 2, Age: 28, CityCode: "310100", HeightCM: &h165, EducationLevel: 3, AnnualIncomeBand: 4, MaritalStatus: 1, ChildrenCount: 0, WantsChildren: 2, Weight: 100},
		{Gender: 2, Age: 31, CityCode: "310100", HeightCM: &h172, EducationLevel: 4, AnnualIncomeBand: 6, MaritalStatus: 1, ChildrenCount: 0, WantsChildren: 2, Weight: 50},
		{Gender: 2, Age: 35, CityCode: "110100", HeightCM: &h165, EducationLevel: 2, AnnualIncomeBand: 3, MaritalStatus: 1, ChildrenCount: 0, WantsChildren: 1, Weight: 30},
		{Gender: 1, Age: 29, CityCode: "310100", HeightCM: &h180, EducationLevel: 3, AnnualIncomeBand: 5, MaritalStatus: 1, ChildrenCount: 0, WantsChildren: 2, Weight: 200},
		{Gender: 1, Age: 33, CityCode: "440300", HeightCM: nil, EducationLevel: 5, AnnualIncomeBand: 7, MaritalStatus: 2, ChildrenCount: 1, WantsChildren: 4, Weight: 20},
	}
}

func TestEstimate_WeightedTotalAndFilter(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))

	est, err := eng.Estimate(ctx, Filter{Genders: []int16{2}, MinAge: ptr(25), MaxAge: ptr(32)})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if est.TotalPopulation != 400 {
		t.Fatalf("总权重应为 400，实际 %v", est.TotalPopulation)
	}
	// 女性 28/31 岁，权重 100+50
	if est.MatchedCount != 150 {
		t.Fatalf("匹配权重应为 150，实际 %v", est.MatchedCount)
	}
	if math.Abs(est.MatchedRatio-0.375) > 1e-9 {
		t.Fatalf("匹配比例应为 0.375，实际 %v", est.MatchedRatio)
	}
}

func TestEstimate_AgeBoundaryInclusive(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	tests := []struct {
		name string
		min  int
		max  int
		want float64
	}{
		{"下界含等号", 28, 28, 100},
		{"上界含等号", 31, 31, 50},
		{"区间覆盖", 28, 31, 150},
		{"区间为空", 36, 40, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			est, err := eng.Estimate(ctx, Filter{Genders: []int16{2}, MinAge: ptr(tc.min), MaxAge: ptr(tc.max)})
			if err != nil {
				t.Fatalf("估算失败: %v", err)
			}
			if est.MatchedCount != tc.want {
				t.Fatalf("期望 %v，实际 %v", tc.want, est.MatchedCount)
			}
		})
	}
}

func TestEstimate_FunnelMonotonic(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	f := Filter{
		Genders:         []int16{2},
		MinAge:          ptr(25),
		MaxAge:          ptr(40),
		Cities:          []string{"310100", "110100"},
		MaritalStatuses: []int16{1},
		WantsChildren:   []int16{2, 3},
		MinEducation:    ptr(int16(3)),
	}
	est, err := eng.Estimate(ctx, f)
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	prev := est.Funnel.TotalPopulation
	for _, step := range est.Funnel.Steps {
		if step.Count > prev+1e-9 {
			t.Fatalf("漏斗第 %s 层候选数增加：%v > %v", step.Key, step.Count, prev)
		}
		prev = step.Count
	}
	if len(est.Funnel.Steps) != len(TheoreticalDimensions) {
		t.Fatalf("漏斗层数应为 %d，实际 %d", len(TheoreticalDimensions), len(est.Funnel.Steps))
	}
}

func TestEstimate_SensitivityNeverDecreases(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	f := Filter{
		Genders:         []int16{2},
		MinAge:          ptr(25),
		Cities:          []string{"310100"},
		MinEducation:    ptr(int16(3)),
		MinIncome:       ptr(int16(4)),
		MinHeight:       ptr(160),
	}
	est, err := eng.Estimate(ctx, f)
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if len(est.Sensitivity) == 0 {
		t.Fatal("敏感性分析结果不应为空")
	}
	for _, s := range est.Sensitivity {
		if s.WithoutCount < s.BaseCount-1e-9 {
			t.Fatalf("取消条件 %s 后候选数减少：%v < %v", s.Key, s.WithoutCount, s.BaseCount)
		}
		if s.AddedCount < 0 {
			t.Fatalf("取消条件 %s 后新增候选数为负", s.Key)
		}
	}
	// TopFactors 必须按影响从大到小
	for i := 1; i < len(est.TopFactors); i++ {
		if est.TopFactors[i-1].DeltaRatio < est.TopFactors[i].DeltaRatio {
			t.Fatal("TopFactors 未按影响排序")
		}
	}
}

func TestEstimate_DisabledDimensionsNotInSensitivity(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	est, err := eng.Estimate(ctx, Filter{Genders: []int16{2}})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if len(est.Sensitivity) != 1 || est.Sensitivity[0].Key != "gender" {
		t.Fatalf("仅性别条件时应只有 1 项敏感性，实际 %+v", est.Sensitivity)
	}
	for _, step := range est.Funnel.Steps {
		if step.Key == "gender" && !step.Applied {
			t.Fatal("性别维度应标记为已应用")
		}
		if step.Key == "age" && step.Applied {
			t.Fatal("年龄维度不应标记为已应用")
		}
	}
}

func TestEstimate_UnknownHeightExcludedByHeightFilter(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	est, err := eng.Estimate(ctx, Filter{Genders: []int16{1}, MinHeight: ptr(170)})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	// 男性 180cm 通过，身高未知的男性被排除
	if est.MatchedCount != 200 {
		t.Fatalf("期望 200，实际 %v", est.MatchedCount)
	}
}

func TestEstimate_CityArrayCondition(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	est, err := eng.Estimate(ctx, Filter{Cities: []string{"440300"}})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if est.MatchedCount != 20 {
		t.Fatalf("期望 20，实际 %v", est.MatchedCount)
	}
}

func TestEstimate_ChildrenState(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	hasKids, err := eng.Estimate(ctx, Filter{ChildrenStates: []int16{1}})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if hasKids.MatchedCount != 20 {
		t.Fatalf("有子女权重应为 20，实际 %v", hasKids.MatchedCount)
	}
	noKids, err := eng.Estimate(ctx, Filter{ChildrenStates: []int16{0}})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if noKids.MatchedCount != 380 {
		t.Fatalf("无子女权重应为 380，实际 %v", noKids.MatchedCount)
	}
}

func TestFilterValidate(t *testing.T) {
	cases := []struct {
		name    string
		filter  Filter
		wantErr bool
	}{
		{"合法", Filter{MinAge: ptr(22), MaxAge: ptr(30)}, false},
		{"小于18", Filter{MinAge: ptr(17)}, true},
		{"年龄倒挂", Filter{MinAge: ptr(35), MaxAge: ptr(30)}, true},
		{"身高倒挂", Filter{MinHeight: ptr(180), MaxHeight: ptr(170)}, true},
		{"身高过低", Filter{MinHeight: ptr(50)}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.filter.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("期望错误=%v，实际 %v", tc.wantErr, err)
			}
		})
	}
}

func TestEstimate_EmptyDataset(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(nil))
	est, err := eng.Estimate(ctx, Filter{Genders: []int16{2}})
	if err != nil {
		t.Fatalf("空数据集不应报错: %v", err)
	}
	if est.MatchedRatio != 0 || est.MatchedCount != 0 {
		t.Fatalf("空数据集比例应为 0，实际 %+v", est)
	}
	for _, s := range est.Sensitivity {
		if s.DeltaRatio != 0 {
			t.Fatalf("空数据集敏感性变化率应为 0，实际 %v", s.DeltaRatio)
		}
	}
}

func TestEstimate_NoConditionsIsNotSensitive(t *testing.T) {
	ctx := context.Background()
	eng := NewEngine(NewInMemoryCounter(sampleRecords()))
	est, err := eng.Estimate(ctx, Filter{})
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	if est.MatchedCount != 400 {
		t.Fatalf("无条件时匹配全部，实际 %v", est.MatchedCount)
	}
	if len(est.Sensitivity) != 0 {
		t.Fatalf("无条件时不应生成敏感性分析，实际 %+v", est.Sensitivity)
	}
}
