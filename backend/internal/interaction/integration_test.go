//go:build integration

package interaction_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/interaction"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/testutil/pgtest"
)

func createUser(t *testing.T, env *pgtest.Env, phone string, gender int16, age int) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	u, err := env.Store.Q.CreateUser(ctx, sqlcdb.CreateUserParams{
		PhoneHash: append([]byte(phone), 1, 2), PhoneEncrypted: []byte("enc"), Status: 2, Role: 1,
	})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	birth := time.Now().AddDate(-age, 0, 0)
	h := int16(170)
	_, err = env.Store.Q.CreateProfile(ctx, sqlcdb.CreateProfileParams{
		UserID: u.ID, Nickname: "用户" + phone[len(phone)-3:], Gender: gender, BirthDate: &birth,
		ProvinceCode: "310000", CityCode: "310100", HeightCm: &h, EducationLevel: 3,
		OccupationCategory: 1, EmploymentStatus: 1, AnnualIncomeBand: 4, DebtBand: 0,
		HouseStatus: 2, VehicleStatus: 1, MaritalStatus: 1, ChildrenCount: 0, WantsChildren: 2,
		MarriageIntent: 2, MarriageTimeline: 2, SmokingStatus: 1, DrinkingStatus: 1, Bio: "测试",
		ProfileCompletion: 100,
	})
	if err != nil {
		t.Fatalf("创建资料失败: %v", err)
	}
	target := int16(1)
	if gender == 1 {
		target = 2
	}
	_, err = env.Store.Q.CreatePreferences(ctx, sqlcdb.CreatePreferencesParams{
		UserID: u.ID, TargetGenders: []int16{target}, MinAge: 18, MaxAge: 99, AgeHard: true,
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
	return u.ID
}

func TestMatchCreatedOnceAndBlockedOnBlock(t *testing.T) {
	env := pgtest.New(t)
	ctx := context.Background()
	// 打开 Match 功能（默认迁移值已为 true，这里再确保一次）
	if _, err := env.Store.Q.UpsertSetting(ctx, sqlcdb.UpsertSettingParams{
		Key: "matches_enabled", Value: []byte("true"), Description: "Match 开关",
	}); err != nil {
		t.Fatalf("写入参数失败: %v", err)
	}
	a := createUser(t, env, "13600000011", 1, 30)
	b := createUser(t, env, "13600000012", 2, 28)

	svc := interaction.NewService(env.Store, env.Settings, storage.NewMemory())

	res1, err := svc.Act(ctx, a, b, 1, 1)
	if err != nil {
		t.Fatalf("A Like B 失败: %v", err)
	}
	if res1.Matched {
		t.Fatal("单向 Like 不应产生 Match")
	}
	res2, err := svc.Act(ctx, b, a, 1, 1)
	if err != nil {
		t.Fatalf("B Like A 失败: %v", err)
	}
	if !res2.Matched || res2.MatchID == nil {
		t.Fatal("双向 Like 应产生 Match")
	}
	// 重复 Like 不得产生第二条 Match
	res3, err := svc.Act(ctx, a, b, 1, 1)
	if err != nil {
		t.Fatalf("重复 Like 失败: %v", err)
	}
	if res3.MatchID == nil || *res3.MatchID != *res2.MatchID {
		t.Fatal("重复 Like 必须返回同一条 Match")
	}
	matches, err := svc.ListMatches(ctx, a, 10, 0)
	if err != nil {
		t.Fatalf("Match 列表失败: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("Match 列表应为 1 条，实际 %d", len(matches))
	}

	// 双方都反馈线下见面 -> confirmed_meeting
	if err := svc.AddOutcome(ctx, a, *res2.MatchID, 3, nil, ""); err != nil {
		t.Fatalf("A 反馈失败: %v", err)
	}
	detail, err := svc.GetMatch(ctx, a, *res2.MatchID)
	if err != nil {
		t.Fatalf("Match 详情失败: %v", err)
	}
	if detail.ConfirmedMeeting {
		t.Fatal("仅一方反馈时不应确认见面")
	}
	if err := svc.AddOutcome(ctx, b, *res2.MatchID, 3, nil, ""); err != nil {
		t.Fatalf("B 反馈失败: %v", err)
	}
	detail, err = svc.GetMatch(ctx, a, *res2.MatchID)
	if err != nil {
		t.Fatalf("Match 详情失败: %v", err)
	}
	if !detail.ConfirmedMeeting {
		t.Fatal("双方均反馈线下见面后应确认见面")
	}

	// Block 后 Match 应被置为已屏蔽且不再出现在列表
	if err := svc.Block(ctx, a, b); err != nil {
		t.Fatalf("Block 失败: %v", err)
	}
	matches, err = svc.ListMatches(ctx, a, 10, 0)
	if err != nil {
		t.Fatalf("Block 后 Match 列表失败: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("Block 后 Match 列表应为空，实际 %d", len(matches))
	}
	blocked, err := env.Store.Q.GetMatchByID(ctx, *res2.MatchID)
	if err != nil {
		t.Fatalf("读取 Match 失败: %v", err)
	}
	if blocked.Status != 3 {
		t.Fatalf("Block 后 Match 状态应为 3（已屏蔽），实际 %d", blocked.Status)
	}
}

func TestActRejectsSelfAndBlocked(t *testing.T) {
	env := pgtest.New(t)
	ctx := context.Background()
	a := createUser(t, env, "13600000021", 1, 30)
	b := createUser(t, env, "13600000022", 2, 28)
	svc := interaction.NewService(env.Store, env.Settings, storage.NewMemory())
	if _, err := svc.Act(ctx, a, a, 1, 1); err == nil {
		t.Fatal("不能对自己 Like")
	}
	if err := svc.Block(ctx, b, a); err != nil {
		t.Fatalf("Block 失败: %v", err)
	}
	if _, err := svc.Act(ctx, a, b, 1, 1); err == nil {
		t.Fatal("被 Block 后不应允许互动")
	}
}
