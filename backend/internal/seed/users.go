package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/secure"
	"wmc/backend/internal/profile"
	"wmc/backend/internal/store"
)

// Users 生成指定数量的虚构用户（TEST DATA），包含资料、择偶条件与价值观规则。
// 生产环境禁止调用（由 cmd/seed 守卫）。
func Users(ctx context.Context, st *store.Store, phoneHashSecret string, phoneKey []byte, count int, seedValue int64) (int, error) {
	if count <= 0 {
		return 0, nil
	}
	r := rand.New(rand.NewSource(seedValue))
	now := time.Now()

	questions, err := st.Q.ListActiveQuestions(ctx)
	if err != nil {
		return 0, fmt.Errorf("读取价值观问题失败: %w", err)
	}
	if len(questions) == 0 {
		return 0, fmt.Errorf("没有可用的价值观问题")
	}

	cities := []struct{ Province, City string }{
		{"310000", "310100"}, {"110000", "110100"}, {"440000", "440300"}, {"440000", "440100"},
		{"330000", "330100"}, {"320000", "320100"}, {"510000", "510100"}, {"420000", "420100"},
		{"430000", "430100"}, {"500000", "500100"}, {"370000", "370200"}, {"610000", "610100"},
	}
	nicknames := []string{"小林", "阿哲", "小满", "南风", "子墨", "晴川", "沐言", "一诺", "知微", "青禾",
		"小屿", "远山", "清和", "安然", "念安", "亦舒", "楚宁", "洛书", "予安", "星野"}

	created := 0
	for i := 0; i < count; i++ {
		phone := fmt.Sprintf("137%08d", i+1)
		hash := secure.PhoneHash(phoneHashSecret, phone)
		if _, err := st.Q.GetUserByPhoneHash(ctx, hash); err == nil {
			continue // 已存在则跳过，保证可重复执行
		}
		enc, err := secure.EncryptPhone(phoneKey, phone)
		if err != nil {
			return created, err
		}
		user, err := st.Q.CreateUser(ctx, sqlcdb.CreateUserParams{
			PhoneHash: hash, PhoneEncrypted: enc, Status: 2, Role: 1, IsSeed: true,
		})
		if err != nil {
			return created, fmt.Errorf("创建测试用户失败: %w", err)
		}

		gender := int16(1)
		if r.Intn(2) == 0 {
			gender = 2
		}
		age := 23 + r.Intn(16) // 23~38
		birth := now.AddDate(-age, -r.Intn(12), -r.Intn(28))
		height := int16(158 + r.Intn(28))
		loc := cities[r.Intn(len(cities))]
		education := int16(2 + r.Intn(4))
		income := int16(2 + r.Intn(5))
		marital := int16(1)
		if age > 32 && r.Intn(10) < 2 {
			marital = 2
		}
		children := int16(0)
		if marital == 2 && r.Intn(2) == 0 {
			children = 1
		}
		wants := int16(2)
		switch r.Intn(4) {
		case 0:
			wants = 3
		case 1:
			wants = 1
		}
		bio := fmt.Sprintf("测试用户（TEST DATA），喜欢%s", []string{"徒步", "阅读", "旅行", "电影", "烹饪"}[r.Intn(5)])
		nickname := fmt.Sprintf("%s%d", nicknames[r.Intn(len(nicknames))], i+1)

		profParams := sqlcdb.CreateProfileParams{
			UserID: user.ID, Nickname: nickname, Gender: gender, BirthDate: &birth,
			ProvinceCode: loc.Province, CityCode: loc.City, HeightCm: &height,
			EducationLevel: education, OccupationCategory: int16(1 + r.Intn(11)),
			EmploymentStatus: 1, AnnualIncomeBand: income, DebtBand: int16(r.Intn(3)),
			HouseStatus: int16(1 + r.Intn(6)), VehicleStatus: int16(1 + r.Intn(4)),
			MaritalStatus: marital, ChildrenCount: children, WantsChildren: wants,
			MarriageIntent: int16(1 + r.Intn(5)), MarriageTimeline: int16(1 + r.Intn(6)),
			SmokingStatus: int16(1 + r.Intn(4)), DrinkingStatus: int16(1 + r.Intn(4)), Bio: bio,
		}
		if _, err := st.Q.CreateProfile(ctx, profParams); err != nil {
			return created, fmt.Errorf("创建测试资料失败: %w", err)
		}
		completion := profile.Compute(profile.Data{
			Nickname: profParams.Nickname, Gender: profParams.Gender, BirthDate: profParams.BirthDate,
			CityCode: profParams.CityCode, HeightCM: profParams.HeightCm,
			EducationLevel: profParams.EducationLevel, OccupationCategory: profParams.OccupationCategory,
			EmploymentStatus: profParams.EmploymentStatus, AnnualIncomeBand: profParams.AnnualIncomeBand,
			DebtBand: profParams.DebtBand, HouseStatus: profParams.HouseStatus,
			VehicleStatus: profParams.VehicleStatus, MaritalStatus: profParams.MaritalStatus,
			ChildrenCount: profParams.ChildrenCount, WantsChildren: profParams.WantsChildren,
			MarriageIntent: profParams.MarriageIntent, MarriageTimeline: profParams.MarriageTimeline,
			SmokingStatus: profParams.SmokingStatus, DrinkingStatus: profParams.DrinkingStatus,
			Bio: profParams.Bio,
		}, true) // 测试用户视为主图已通过审核，便于候选卡展示逻辑
		if err := st.Q.UpdateProfileCompletion(ctx, sqlcdb.UpdateProfileCompletionParams{
			UserID: user.ID, ProfileCompletion: int16(completion),
		}); err != nil {
			return created, err
		}

		targetGender := int16(1)
		if gender == 1 {
			targetGender = 2
		}
		prefParams := sqlcdb.CreatePreferencesParams{
			UserID: user.ID, TargetGenders: []int16{targetGender},
			MinAge: int16(max2(18, age-5)), MaxAge: int16(age + 6), AgeHard: true,
			MinHeightCm: int16(155 + r.Intn(10)), MaxHeightCm: 200, HeightHard: false,
			AllowedCityCodes: []string{loc.City}, CityHard: r.Intn(2) == 0,
			MinEducationLevel: int16(2 + r.Intn(3)), EducationHard: false,
			MinIncomeBand: int16(1 + r.Intn(4)), IncomeHard: false,
			AllowedMaritalStatuses: []int16{}, MaritalStatusHard: false,
			AllowedChildrenStates: []int16{}, ChildrenHard: false,
			AllowedWantsChildren: []int16{2, 3}, WantsChildrenHard: r.Intn(2) == 0,
			AllowedSmokingStatuses: []int16{}, SmokingHard: false,
			AllowedDrinkingStatuses: []int16{}, DrinkingHard: false,
			MinMarriageIntent: 2, MarriageIntentHard: false,
			Weights: []byte("{}"),
		}
		if _, err := st.Q.CreatePreferences(ctx, prefParams); err != nil {
			return created, fmt.Errorf("创建测试偏好失败: %w", err)
		}

		ruleCount := 6 + r.Intn(4)
		for k := 0; k < ruleCount && k < len(questions); k++ {
			q := questions[k]
			var opts []struct {
				Value string `json:"value"`
			}
			_ = json.Unmarshal(q.Options, &opts)
			if len(opts) == 0 {
				continue
			}
			own := opts[r.Intn(len(opts))].Value
			accepted := []string{own}
			if len(opts) > 1 && r.Intn(2) == 0 {
				other := opts[r.Intn(len(opts))].Value
				if other != own {
					accepted = append(accepted, other)
				}
			}
			ownJSON, _ := json.Marshal([]string{own})
			acceptedJSON, _ := json.Marshal(accepted)
			if _, err := st.Q.UpsertRule(ctx, sqlcdb.UpsertRuleParams{
				UserID: user.ID, QuestionID: q.ID, OwnAnswer: ownJSON, AcceptedAnswers: acceptedJSON,
				Importance: int16(3 + r.Intn(3)), IsHard: false,
			}); err != nil {
				return created, fmt.Errorf("创建测试价值观规则失败: %w", err)
			}
		}

		if err := st.Q.SetUserDiscoveryEnabled(ctx, sqlcdb.SetUserDiscoveryEnabledParams{
			ID: user.ID, DiscoveryEnabled: true,
		}); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ = uuid.Nil
