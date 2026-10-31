//go:build integration

package discovery_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/discovery"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/testutil/pgtest"
)

type testUser struct {
	ID    uuid.UUID
	Phone string
}

func createUser(t *testing.T, env *pgtest.Env, phone string, gender int16, age int, height int16, city string) testUser {
	t.Helper()
	ctx := context.Background()
	u, err := env.Store.Q.CreateUser(ctx, sqlcdb.CreateUserParams{
		PhoneHash: append([]byte(phone), 0, 1), PhoneEncrypted: []byte("enc"), Status: 2, Role: 1,
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	birth := time.Now().AddDate(-age, 0, 0)
	h := height
	_, err = env.Store.Q.CreateProfile(ctx, sqlcdb.CreateProfileParams{
		UserID: u.ID, Nickname: "用户" + phone[len(phone)-3:], Gender: gender, BirthDate: &birth,
		ProvinceCode: city[:2] + "0000", CityCode: city, HeightCm: &h, EducationLevel: 3,
		OccupationCategory: 1, EmploymentStatus: 1, AnnualIncomeBand: 4, DebtBand: 0,
		HouseStatus: 2, VehicleStatus: 1, MaritalStatus: 1, ChildrenCount: 0, WantsChildren: 2,
		MarriageIntent: 2, MarriageTimeline: 2, SmokingStatus: 1, DrinkingStatus: 1, Bio: "集成测试",
		ProfileCompletion: 100,
	})
	if err != nil {
		t.Fatalf("创建资料失败: %v", err)
	}
	targetGender := int16(1)
	if gender == 1 {
		targetGender = 2
	}
	_, err = env.Store.Q.CreatePreferences(ctx, sqlcdb.CreatePreferencesParams{
		UserID: u.ID, TargetGenders: []int16{targetGender}, MinAge: 18, MaxAge: 99, AgeHard: true,
		MinHeightCm: 0, MaxHeightCm: 250, HeightHard: false,
		AllowedCityCodes: []string{}, CityHard: false, MinEducationLevel: 0, EducationHard: false,
		MinIncomeBand: 0, IncomeHard: false, AllowedMaritalStatuses: []int16{}, MaritalStatusHard: false,
		AllowedChildrenStates: []int16{}, ChildrenHard: false,
		AllowedWantsChildren: []int16{}, WantsChildrenHard: false,
		AllowedSmokingStatuses: []int16{}, SmokingHard: false,
		AllowedDrinkingStatuses: []int16{}, DrinkingHard: false,
		MinMarriageIntent: 0, MarriageIntentHard: false, Weights: []byte("{}"),
	})
	if err != nil {
		t.Fatalf("创建偏好失败: %v", err)
	}
	if err := env.Store.Q.SetUserDiscoveryEnabled(ctx, sqlcdb.SetUserDiscoveryEnabledParams{ID: u.ID, DiscoveryEnabled: true}); err != nil {
		t.Fatalf("入池失败: %v", err)
	}
	return testUser{ID: u.ID, Phone: phone}
}

func updateTargetGender(t *testing.T, env *pgtest.Env, userID uuid.UUID, gender int16) {
	t.Helper()
	ctx := context.Background()
	pref, err := env.Store.Q.GetPreferences(ctx, userID)
	if err != nil {
		t.Fatalf("读取偏好失败: %v", err)
	}
	_, err = env.Store.Q.UpdatePreferences(ctx, sqlcdb.UpdatePreferencesParams{
		UserID: userID, TargetGenders: []int16{gender}, MinAge: pref.MinAge, MaxAge: pref.MaxAge,
		AgeHard: pref.AgeHard, MinHeightCm: pref.MinHeightCm, MaxHeightCm: pref.MaxHeightCm,
		HeightHard: pref.HeightHard, AllowedCityCodes: pref.AllowedCityCodes, CityHard: pref.CityHard,
		MinEducationLevel: pref.MinEducationLevel, EducationHard: pref.EducationHard,
		MinIncomeBand: pref.MinIncomeBand, IncomeHard: pref.IncomeHard,
		AllowedMaritalStatuses: pref.AllowedMaritalStatuses, MaritalStatusHard: pref.MaritalStatusHard,
		AllowedChildrenStates: pref.AllowedChildrenStates, ChildrenHard: pref.ChildrenHard,
		AllowedWantsChildren: pref.AllowedWantsChildren, WantsChildrenHard: pref.WantsChildrenHard,
		AllowedSmokingStatuses: pref.AllowedSmokingStatuses, SmokingHard: pref.SmokingHard,
		AllowedDrinkingStatuses: pref.AllowedDrinkingStatuses, DrinkingHard: pref.DrinkingHard,
		MinMarriageIntent: pref.MinMarriageIntent, MarriageIntentHard: pref.MarriageIntentHard,
		Weights: pref.Weights,
	})
	if err != nil {
		t.Fatalf("更新偏好失败: %v", err)
	}
}

func newService(env *pgtest.Env) *discovery.Service {
	return discovery.NewService(env.Store, env.Pool, env.Settings, storage.NewMemory())
}

func TestDiscovery_ReciprocalRequirement(t *testing.T) {
	env := pgtest.New(t)
	ctx := context.Background()
	_ = ctx
	// A：男性 30 岁，目标女性；B：女性 28 岁，目标女性（不接受男性）
	a := createUser(t, env, "13700000001", 1, 30, 178, "310100")
	b := createUser(t, env, "13700000002", 2, 28, 165, "310100")
	updateTargetGender(t, env, b.ID, 2)

	svc := newService(env)
	sum, err := svc.GetSummary(ctx, a.ID)
	if err != nil {
		t.Fatalf("概览失败: %v", err)
	}
	if sum.OneWayCount != 1 {
		t.Fatalf("单向候选应为 1，实际 %d", sum.OneWayCount)
	}
	if sum.ReciprocalCount != 0 {
		t.Fatalf("B 不接受男性时双向候选必须为 0，实际 %d", sum.ReciprocalCount)
	}
	if sum.EligibleCountText != "<10" {
		t.Fatalf("小样本应匿名为 <10，实际 %s", sum.EligibleCountText)
	}

	// B 改为接受男性后，双向应成立
	updateTargetGender(t, env, b.ID, 1)
	sum2, err := svc.GetSummary(ctx, a.ID)
	if err != nil {
		t.Fatalf("二次概览失败: %v", err)
	}
	if sum2.ReciprocalCount != 1 {
		t.Fatalf("双方互选后双向候选应为 1，实际 %d", sum2.ReciprocalCount)
	}
	page, err := svc.Candidates(ctx, a.ID, "", 10)
	if err != nil {
		t.Fatalf("候选列表失败: %v", err)
	}
	if len(page.Items) != 1 || page.Items[0].UserID != b.ID.String() {
		t.Fatalf("候选列表应包含 B：%+v", page.Items)
	}

	detail, err := svc.CandidateDetail(ctx, a.ID, b.ID)
	if err != nil {
		t.Fatalf("候选人详情失败: %v", err)
	}
	if len(detail.Reasons) == 0 {
		t.Fatal("详情必须包含匹配解释")
	}
}

func TestDiscovery_BlockHidesCandidate(t *testing.T) {
	env := pgtest.New(t)
	ctx := context.Background()
	a := createUser(t, env, "13700000011", 1, 30, 178, "310100")
	b := createUser(t, env, "13700000012", 2, 28, 165, "310100")
	_ = ctx

	svc := newService(env)
	if _, err := env.Store.Pool.Exec(ctx,
		"INSERT INTO user_blocks (blocker_user_id, blocked_user_id) VALUES ($1, $2)", b.ID, a.ID); err != nil {
		t.Fatalf("写入 Block 失败: %v", err)
	}
	sum, err := svc.GetSummary(ctx, a.ID)
	if err != nil {
		t.Fatalf("概览失败: %v", err)
	}
	if sum.ReciprocalCount != 0 {
		t.Fatalf("被 Block 后不应出现候选，实际 %d", sum.ReciprocalCount)
	}
	if _, err := svc.CandidateDetail(ctx, a.ID, b.ID); err == nil {
		t.Fatal("被 Block 后详情应不可见")
	}
}

func TestDiscovery_FunnelMonotonicAndSensitivity(t *testing.T) {
	env := pgtest.New(t)
	ctx := context.Background()
	a := createUser(t, env, "13700000021", 1, 30, 178, "310100")
	_ = a
	for i := 0; i < 6; i++ {
		createUser(t, env, "1370000003"+string(rune('0'+i)), 2, 26+i, 160+int16(i), "310100")
	}
	svc := newService(env)
	funnel, err := svc.GetFunnel(ctx, a.ID)
	if err != nil {
		t.Fatalf("漏斗失败: %v", err)
	}
	prev := funnel.Total
	for _, s := range funnel.Steps {
		if s.Count > prev {
			t.Fatalf("漏斗第 %s 层候选数增加：%d > %d", s.Key, s.Count, prev)
		}
		prev = s.Count
	}
	sens, err := svc.GetSensitivity(ctx, a.ID)
	if err != nil {
		t.Fatalf("敏感性失败: %v", err)
	}
	for _, item := range sens.Items {
		if item.WithoutCount < item.BaseCount {
			t.Fatalf("取消条件 %s 后候选数减少：%d < %d", item.Key, item.WithoutCount, item.BaseCount)
		}
	}
	_ = settings.KeyKAnonymityThreshold
	_ = redisx.NewMemory()
}
