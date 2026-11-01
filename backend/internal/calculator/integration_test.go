//go:build integration

package calculator_test

import (
	"context"
	"math"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/calculator"
	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/seed"
	"wmc/backend/internal/testutil/pgtest"
)

// seedDataset 直接写入一个演示数据集并激活，返回数据集 ID。
func seedDataset(t *testing.T, env *pgtest.Env, records []calculator.Record) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	ds, err := env.Store.Q.CreateDataset(ctx, sqlcdb.CreateDatasetParams{
		Name: "集成测试数据集", Version: "it-1", SourceName: "测试来源",
		Description: "用于集成测试", Methodology: "测试", Disclaimer: "测试数据", Status: 1, IsDemo: false,
	})
	if err != nil {
		t.Fatalf("创建数据集失败: %v", err)
	}
	rows := make([][]any, 0, len(records))
	for _, r := range records {
		rows = append(rows, []any{
			ds.ID, r.Gender, r.Age, r.ProvinceCode, r.CityCode, r.HeightCM,
			r.EducationLevel, r.AnnualIncomeBand, r.MaritalStatus, r.ChildrenCount, r.WantsChildren, r.Weight,
		})
	}
	_, err = env.Store.CopyFrom(ctx,
		pgx.Identifier{"baseline_records"},
		[]string{"dataset_id", "gender", "age", "province_code", "city_code", "height_cm",
			"education_level", "annual_income_band", "marital_status", "children_count",
			"wants_children", "sample_weight"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		t.Fatalf("写入基线记录失败: %v", err)
	}
	if err := env.Store.Q.UpdateDatasetStats(ctx, sqlcdb.UpdateDatasetStatsParams{
		ID: ds.ID, RecordCount: int64(len(rows)), Status: 2,
	}); err != nil {
		t.Fatalf("更新统计失败: %v", err)
	}
	if _, err := env.Store.Q.UpsertSetting(ctx, sqlcdb.UpsertSettingParams{
		Key: "calculator_enabled", Value: []byte("true"), Description: "计算器开关",
	}); err != nil {
		t.Fatalf("写入参数失败: %v", err)
	}
	if err := env.Store.Q.DeactivateAllDatasets(ctx); err != nil {
		t.Fatalf("停用数据集失败: %v", err)
	}
	if err := env.Store.Q.SetDatasetActive(ctx, ds.ID); err != nil {
		t.Fatalf("激活数据集失败: %v", err)
	}
	return ds.ID
}

func newService(env *pgtest.Env) *calculator.Service {
	return calculator.NewService(env.Store, env.Pool, env.Settings, redisx.NewMemory())
}

func TestEstimate_SQLCounterMatchesInMemory(t *testing.T) {
	env := pgtest.New(t)
	records := seed.DemoDataset(500, 42)
	seedDataset(t, env, records)

	svc := newService(env)
	ctx := context.Background()

	minAge, maxAge := 26, 35
	height := 170
	education := int16(3)
	req := calculator.EstimateRequest{
		AnonymousSessionID: "integration-test",
		Conditions: calculator.ConditionInput{
			TargetGenders:          []int16{2},
			MinAge:                 &minAge,
			MaxAge:                 &maxAge,
			MinHeightCM:            &height,
			MinEducationLevel:      &education,
			AllowedCityCodes:       []string{"310100", "110100", "440300"},
			AllowedMaritalStatuses: []int16{1},
		},
	}
	res, err := svc.Estimate(ctx, req, nil)
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}

	memFilter := calculator.Filter{
		Genders: []int16{2}, MinAge: &minAge, MaxAge: &maxAge,
		MinHeight: &height, MinEducation: &education,
		Cities: []string{"310100", "110100", "440300"}, MaritalStatuses: []int16{1},
	}
	memRes, err := calculator.NewEngine(calculator.NewInMemoryCounter(records)).Estimate(ctx, memFilter)
	if err != nil {
		t.Fatalf("内存估算失败: %v", err)
	}
	if math.Abs(memRes.TotalPopulation-res.TotalPopulation) > 0.01 {
		t.Fatalf("总权重不一致: SQL=%v 内存=%v", res.TotalPopulation, memRes.TotalPopulation)
	}
	if math.Abs(memRes.MatchedCount-res.MatchedCount) > 0.01 {
		t.Fatalf("匹配权重不一致: SQL=%v 内存=%v", res.MatchedCount, memRes.MatchedCount)
	}
	if res.RunID == uuid.Nil {
		t.Fatal("应生成计算记录")
	}
	if res.Dataset.IsDemo {
		t.Fatal("普通数据集不应标记为 DEMO")
	}
	if res.FunnelNote == "" {
		t.Fatal("必须返回漏斗说明")
	}
	if res.MarketNotice == "" {
		t.Fatal("必须返回理论市场免责说明")
	}
}

func TestEstimate_FunnelAndSensitivityMonotonicInSQL(t *testing.T) {
	env := pgtest.New(t)
	seedDataset(t, env, seed.DemoDataset(800, 7))
	svc := newService(env)

	minAge, maxAge := 25, 34
	income := int16(5)
	education := int16(4)
	height := 168
	res, err := svc.Estimate(context.Background(), calculator.EstimateRequest{
		AnonymousSessionID: "s1",
		Conditions: calculator.ConditionInput{
			TargetGenders: []int16{1}, MinAge: &minAge, MaxAge: &maxAge,
			MinIncomeBand: &income, MinEducationLevel: &education, MinHeightCM: &height,
			AllowedWantsChildren: []int16{2, 3},
		},
	}, nil)
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	prev := res.Funnel.TotalPopulation
	for _, step := range res.Funnel.Steps {
		if step.Count > prev+0.01 {
			t.Fatalf("漏斗第 %s 层候选数增加：%v > %v", step.Key, step.Count, prev)
		}
		prev = step.Count
	}
	for _, s := range res.Sensitivity {
		if s.WithoutCount < s.BaseCount-0.01 {
			t.Fatalf("取消条件 %s 后候选数减少：%v < %v", s.Key, s.WithoutCount, s.BaseCount)
		}
	}
}

func TestEstimate_ShareRoundTripInDB(t *testing.T) {
	env := pgtest.New(t)
	seedDataset(t, env, seed.DemoDataset(100, 3))
	svc := newService(env)
	ctx := context.Background()
	anon := "share-tester"
	res, err := svc.Estimate(ctx, calculator.EstimateRequest{
		AnonymousSessionID: anon, Conditions: calculator.ConditionInput{TargetGenders: []int16{2}},
	}, nil)
	if err != nil {
		t.Fatalf("估算失败: %v", err)
	}
	token, err := svc.CreateShare(ctx, res.RunID, anon, nil)
	if err != nil {
		t.Fatalf("创建分享失败: %v", err)
	}
	shared, err := svc.GetShare(ctx, token)
	if err != nil {
		t.Fatalf("读取分享失败: %v", err)
	}
	if shared.MatchedCount != res.MatchedCount {
		t.Fatal("分享结果与原始结果不一致")
	}
	if _, err := svc.CreateShare(ctx, res.RunID, "other-session", nil); err == nil {
		t.Fatal("非本人不得分享该结果")
	}
}

func TestEstimate_NoActiveDataset(t *testing.T) {
	env := pgtest.New(t)
	svc := newService(env)
	_, err := svc.Estimate(context.Background(), calculator.EstimateRequest{
		AnonymousSessionID: "x", Conditions: calculator.ConditionInput{TargetGenders: []int16{1}},
	}, nil)
	if err == nil {
		t.Fatal("缺少激活数据集时必须报错")
	}
}

func TestEstimate_CacheHitReturnsSameResult(t *testing.T) {
	env := pgtest.New(t)
	seedDataset(t, env, seed.DemoDataset(200, 11))
	svc := newService(env)
	ctx := context.Background()
	req := calculator.EstimateRequest{
		AnonymousSessionID: "cache-tester", Conditions: calculator.ConditionInput{TargetGenders: []int16{2}},
	}
	first, err := svc.Estimate(ctx, req, nil)
	if err != nil {
		t.Fatalf("首次估算失败: %v", err)
	}
	second, err := svc.Estimate(ctx, req, nil)
	if err != nil {
		t.Fatalf("二次估算失败: %v", err)
	}
	if first.RunID == second.RunID {
		t.Fatal("每次估算都应生成新的计算记录")
	}
	if first.MatchedCount != second.MatchedCount {
		t.Fatal("缓存命中时结果应一致")
	}
}