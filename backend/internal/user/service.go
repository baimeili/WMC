package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/profile"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

// Service 用户资料与价值观服务。
type Service struct {
	store    *store.Store
	settings *settings.Service
	validate *validator.Validate
}

// NewService 构造服务。
func NewService(st *store.Store, set *settings.Service) *Service {
	return &Service{store: st, settings: set, validate: validator.New()}
}

func (s *Service) validateInput(in any) error {
	if err := s.validate.Struct(in); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) && len(ve) > 0 {
			fe := ve[0]
			return apperr.Invalid(fmt.Sprintf("字段 %s 校验失败（%s）", fe.Field(), fe.Tag()))
		}
		return apperr.Invalid("参数校验失败")
	}
	return nil
}

// DateLayout 生日格式。
const DateLayout = "2006-01-02"

// GetProfile 读取资料；不存在时返回空资料（完成度 0）。
func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (Profile, error) {
	row, err := s.store.Q.GetProfile(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{UserID: userID.String()}, nil
	}
	if err != nil {
		return Profile{}, apperr.ErrInternal.Wrap(err)
	}
	return toProfile(row), nil
}

// SaveProfile 保存资料并重算完成度。
func (s *Service) SaveProfile(ctx context.Context, userID uuid.UUID, in ProfileInput) (Profile, error) {
	if err := s.validateInput(in); err != nil {
		return Profile{}, err
	}
	var birth *time.Time
	if in.BirthDate != nil && *in.BirthDate != "" {
		t, err := time.Parse(DateLayout, *in.BirthDate)
		if err != nil {
			return Profile{}, apperr.Invalid("出生日期格式应为 YYYY-MM-DD")
		}
		if t.After(nowFunc()) {
			return Profile{}, apperr.Invalid("出生日期不能晚于今天")
		}
		birth = &t
	}

	params := sqlcdb.CreateProfileParams{
		UserID: userID, Nickname: in.Nickname, Gender: in.Gender, BirthDate: birth,
		ProvinceCode: in.ProvinceCode, CityCode: in.CityCode, DistrictCode: in.DistrictCode,
		HeightCm: in.HeightCM, EducationLevel: in.EducationLevel, OccupationCategory: in.OccupationCategory,
		EmploymentStatus: in.EmploymentStatus, AnnualIncomeBand: in.AnnualIncomeBand,
		FamilyAssetBand: in.FamilyAssetBand, DebtBand: in.DebtBand, HouseStatus: in.HouseStatus,
		VehicleStatus: in.VehicleStatus, MaritalStatus: in.MaritalStatus, ChildrenCount: in.ChildrenCount,
		WantsChildren: in.WantsChildren, MarriageIntent: in.MarriageIntent,
		MarriageTimeline: in.MarriageTimeline, SmokingStatus: in.SmokingStatus,
		DrinkingStatus: in.DrinkingStatus, Bio: in.Bio, ProfileCompletion: 0,
	}

	exists := true
	if _, err := s.store.Q.GetProfile(ctx, userID); errors.Is(err, pgx.ErrNoRows) {
		exists = false
	}
	var completion int16
	err := s.store.InTx(ctx, func(tx *store.Store) error {
		if exists {
			row, err := tx.Q.UpdateProfile(ctx, sqlcdb.UpdateProfileParams{
				UserID: userID, Nickname: params.Nickname, Gender: params.Gender, BirthDate: params.BirthDate,
				ProvinceCode: params.ProvinceCode, CityCode: params.CityCode, DistrictCode: params.DistrictCode,
				HeightCm: params.HeightCm, EducationLevel: params.EducationLevel,
				OccupationCategory: params.OccupationCategory, EmploymentStatus: params.EmploymentStatus,
				AnnualIncomeBand: params.AnnualIncomeBand, FamilyAssetBand: params.FamilyAssetBand,
				DebtBand: params.DebtBand, HouseStatus: params.HouseStatus, VehicleStatus: params.VehicleStatus,
				MaritalStatus: params.MaritalStatus, ChildrenCount: params.ChildrenCount,
				WantsChildren: params.WantsChildren, MarriageIntent: params.MarriageIntent,
				MarriageTimeline: params.MarriageTimeline, SmokingStatus: params.SmokingStatus,
				DrinkingStatus: params.DrinkingStatus, Bio: params.Bio, ProfileCompletion: 0,
			})
			if err != nil {
				return err
			}
			_ = row
		} else {
			if _, err := tx.Q.CreateProfile(ctx, params); err != nil {
				return err
			}
		}
		photos, err := tx.Q.CountPhotosByUser(ctx, userID)
		if err != nil {
			return err
		}
		completion = int16(profile.Compute(dataOf(params), photos > 0))
		return tx.Q.UpdateProfileCompletion(ctx, sqlcdb.UpdateProfileCompletionParams{
			UserID: userID, ProfileCompletion: completion,
		})
	})
	if err != nil {
		return Profile{}, apperr.ErrInternal.Wrap(err)
	}

	// 资料变化后若不再满足入池门槛，自动退出婚恋池
	if completion < int16(s.settings.Int(ctx, settings.KeyMinProfileCompletion, 80)) {
		_ = s.store.Q.SetUserDiscoveryEnabled(ctx, sqlcdb.SetUserDiscoveryEnabledParams{
			ID: userID, DiscoveryEnabled: false,
		})
	}
	return s.GetProfile(ctx, userID)
}

func dataOf(p sqlcdb.CreateProfileParams) profile.Data {
	now := nowFunc()
	_ = now
	return profile.Data{
		Nickname: p.Nickname, Gender: p.Gender, BirthDate: p.BirthDate,
		ProvinceCode: p.ProvinceCode, CityCode: p.CityCode, DistrictCode: p.DistrictCode,
		HeightCM: p.HeightCm, EducationLevel: p.EducationLevel, OccupationCategory: p.OccupationCategory,
		EmploymentStatus: p.EmploymentStatus, AnnualIncomeBand: p.AnnualIncomeBand,
		FamilyAssetBand: p.FamilyAssetBand, DebtBand: p.DebtBand, HouseStatus: p.HouseStatus,
		VehicleStatus: p.VehicleStatus, MaritalStatus: p.MaritalStatus, ChildrenCount: p.ChildrenCount,
		WantsChildren: p.WantsChildren, MarriageIntent: p.MarriageIntent,
		MarriageTimeline: p.MarriageTimeline, SmokingStatus: p.SmokingStatus,
		DrinkingStatus: p.DrinkingStatus, Bio: p.Bio,
	}
}

func toProfile(row sqlcdb.UserProfile) Profile {
	age := 0
	if row.BirthDate != nil {
		age = profile.Age(row.BirthDate, nowFunc())
	}
	var birth *string
	if row.BirthDate != nil {
		v := row.BirthDate.Format(DateLayout)
		birth = &v
	}
	return Profile{
		UserID: row.UserID.String(), Nickname: row.Nickname, Gender: row.Gender, BirthDate: birth, Age: age,
		ProvinceCode: row.ProvinceCode, CityCode: row.CityCode, DistrictCode: row.DistrictCode,
		HeightCM: row.HeightCm, EducationLevel: row.EducationLevel, OccupationCategory: row.OccupationCategory,
		EmploymentStatus: row.EmploymentStatus, AnnualIncomeBand: row.AnnualIncomeBand,
		FamilyAssetBand: row.FamilyAssetBand, DebtBand: row.DebtBand, HouseStatus: row.HouseStatus,
		VehicleStatus: row.VehicleStatus, MaritalStatus: row.MaritalStatus, ChildrenCount: row.ChildrenCount,
		WantsChildren: row.WantsChildren, MarriageIntent: row.MarriageIntent,
		MarriageTimeline: row.MarriageTimeline, SmokingStatus: row.SmokingStatus,
		DrinkingStatus: row.DrinkingStatus, Bio: row.Bio, ProfileCompletion: row.ProfileCompletion,
		UpdatedAt: row.UpdatedAt.Time.Format(time.RFC3339),
	}
}

// GetPreferences 读取择偶条件（未设置时返回安全默认值）。
func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (Preferences, error) {
	row, err := s.store.Q.GetPreferences(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Preferences{
			TargetGenders: []int16{}, AllowedCityCodes: []string{}, AllowedMaritalStatuses: []int16{},
			AllowedChildrenStates: []int16{}, AllowedWantsChildren: []int16{},
			AllowedSmokingStatuses: []int16{}, AllowedDrinkingStatuses: []int16{},
			MinAge: 18, MaxAge: 99, AgeHard: true, MinHeightCM: 0, MaxHeightCM: 250,
			Weights: map[string]float64{},
		}, nil
	}
	if err != nil {
		return Preferences{}, apperr.ErrInternal.Wrap(err)
	}
	return toPreferences(row), nil
}

func toPreferences(row sqlcdb.PartnerPreference) Preferences {
	weights := map[string]float64{}
	if len(row.Weights) > 0 {
		_ = json.Unmarshal(row.Weights, &weights)
	}
	return Preferences{
		TargetGenders: row.TargetGenders, MinAge: row.MinAge, MaxAge: row.MaxAge, AgeHard: row.AgeHard,
		MinHeightCM: row.MinHeightCm, MaxHeightCM: row.MaxHeightCm, HeightHard: row.HeightHard,
		AllowedCityCodes: row.AllowedCityCodes, CityHard: row.CityHard,
		MinEducationLevel: row.MinEducationLevel, EducationHard: row.EducationHard,
		MinIncomeBand: row.MinIncomeBand, IncomeHard: row.IncomeHard,
		AllowedMaritalStatuses: row.AllowedMaritalStatuses, MaritalStatusHard: row.MaritalStatusHard,
		AllowedChildrenStates: row.AllowedChildrenStates, ChildrenHard: row.ChildrenHard,
		AllowedWantsChildren: row.AllowedWantsChildren, WantsChildrenHard: row.WantsChildrenHard,
		AllowedSmokingStatuses: row.AllowedSmokingStatuses, SmokingHard: row.SmokingHard,
		AllowedDrinkingStatuses: row.AllowedDrinkingStatuses, DrinkingHard: row.DrinkingHard,
		MinMarriageIntent: row.MinMarriageIntent, MarriageIntentHard: row.MarriageIntentHard,
		Weights: weights, UpdatedAt: row.UpdatedAt.Time.Format(time.RFC3339),
	}
}

func (p Preferences) validate() error {
	if p.MinAge < 18 || p.MaxAge > 100 {
		return apperr.Invalid("年龄要求应在 18~100 之间")
	}
	if p.MinAge > p.MaxAge {
		return apperr.Invalid("最小年龄不能大于最大年龄")
	}
	if p.MinHeightCM < 0 || p.MaxHeightCM > 250 || p.MinHeightCM > p.MaxHeightCM {
		return apperr.Invalid("身高要求不合理")
	}
	for _, g := range p.TargetGenders {
		if g < 1 || g > 3 {
			return apperr.Invalid("目标性别编码非法")
		}
	}
	for _, b := range p.AllowedMaritalStatuses {
		if b < 1 || b > 4 {
			return apperr.Invalid("婚姻状态编码非法")
		}
	}
	for _, b := range p.AllowedWantsChildren {
		if b < 1 || b > 4 {
			return apperr.Invalid("生育意愿编码非法")
		}
	}
	return nil
}

// SavePreferences 保存择偶条件。
func (s *Service) SavePreferences(ctx context.Context, userID uuid.UUID, in Preferences) (Preferences, error) {
	if err := in.validate(); err != nil {
		return Preferences{}, err
	}
	weights := in.Weights
	if weights == nil {
		weights = map[string]float64{}
	}
	weightsJSON, _ := json.Marshal(weights)
	params := sqlcdb.CreatePreferencesParams{
		UserID: userID, TargetGenders: in.TargetGenders, MinAge: in.MinAge, MaxAge: in.MaxAge,
		AgeHard: in.AgeHard, MinHeightCm: in.MinHeightCM, MaxHeightCm: in.MaxHeightCM,
		HeightHard: in.HeightHard, AllowedCityCodes: in.AllowedCityCodes, CityHard: in.CityHard,
		MinEducationLevel: in.MinEducationLevel, EducationHard: in.EducationHard,
		MinIncomeBand: in.MinIncomeBand, IncomeHard: in.IncomeHard,
		AllowedMaritalStatuses: in.AllowedMaritalStatuses, MaritalStatusHard: in.MaritalStatusHard,
		AllowedChildrenStates: in.AllowedChildrenStates, ChildrenHard: in.ChildrenHard,
		AllowedWantsChildren: in.AllowedWantsChildren, WantsChildrenHard: in.WantsChildrenHard,
		AllowedSmokingStatuses: in.AllowedSmokingStatuses, SmokingHard: in.SmokingHard,
		AllowedDrinkingStatuses: in.AllowedDrinkingStatuses, DrinkingHard: in.DrinkingHard,
		MinMarriageIntent: in.MinMarriageIntent, MarriageIntentHard: in.MarriageIntentHard,
		Weights: weightsJSON,
	}
	_, err := s.store.Q.GetPreferences(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err := s.store.Q.CreatePreferences(ctx, params); err != nil {
			return Preferences{}, apperr.ErrInternal.Wrap(err)
		}
	} else {
		if _, err := s.store.Q.UpdatePreferences(ctx, sqlcdb.UpdatePreferencesParams{
			UserID: userID, TargetGenders: params.TargetGenders, MinAge: params.MinAge, MaxAge: params.MaxAge,
			AgeHard: params.AgeHard, MinHeightCm: params.MinHeightCm, MaxHeightCm: params.MaxHeightCm,
			HeightHard: params.HeightHard, AllowedCityCodes: params.AllowedCityCodes, CityHard: params.CityHard,
			MinEducationLevel: params.MinEducationLevel, EducationHard: params.EducationHard,
			MinIncomeBand: params.MinIncomeBand, IncomeHard: params.IncomeHard,
			AllowedMaritalStatuses: params.AllowedMaritalStatuses, MaritalStatusHard: params.MaritalStatusHard,
			AllowedChildrenStates: params.AllowedChildrenStates, ChildrenHard: params.ChildrenHard,
			AllowedWantsChildren: params.AllowedWantsChildren, WantsChildrenHard: params.WantsChildrenHard,
			AllowedSmokingStatuses: params.AllowedSmokingStatuses, SmokingHard: params.SmokingHard,
			AllowedDrinkingStatuses: params.AllowedDrinkingStatuses, DrinkingHard: params.DrinkingHard,
			MinMarriageIntent: params.MinMarriageIntent, MarriageIntentHard: params.MarriageIntentHard,
			Weights: params.Weights,
		}); err != nil {
			return Preferences{}, apperr.ErrInternal.Wrap(err)
		}
	}
	return s.GetPreferences(ctx, userID)
}

// GetRules 读取价值观规则（含问题信息）。
func (s *Service) GetRules(ctx context.Context, userID uuid.UUID) ([]CompatibilityRule, error) {
	rows, err := s.store.Q.ListRulesWithQuestions(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]CompatibilityRule, 0, len(rows))
	for _, r := range rows {
		out = append(out, CompatibilityRule{
			QuestionID: r.QuestionID.String(), QuestionCode: r.Code, Category: r.Category, Title: r.Title,
			OwnAnswer: convert.JSONStrings(r.OwnAnswer), AcceptedAnswers: convert.JSONStrings(r.AcceptedAnswers),
			Importance: r.Importance, IsHard: r.IsHard,
		})
	}
	return out, nil
}

// SaveRules 全量替换价值观规则。
func (s *Service) SaveRules(ctx context.Context, userID uuid.UUID, rules []CompatibilityRuleInput) error {
	questions, err := s.store.Q.ListActiveQuestions(ctx)
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	optionsByQuestion := make(map[uuid.UUID]map[string]bool, len(questions))
	for _, q := range questions {
		var opts []struct {
			Value string `json:"value"`
		}
		_ = json.Unmarshal(q.Options, &opts)
		set := make(map[string]bool, len(opts))
		for _, o := range opts {
			set[o.Value] = true
		}
		optionsByQuestion[q.ID] = set
	}

	for i, r := range rules {
		if err := s.validateInput(r); err != nil {
			return err
		}
		qid, err := uuid.Parse(r.QuestionID)
		if err != nil {
			return apperr.Invalid(fmt.Sprintf("第 %d 条规则问题 ID 非法", i+1))
		}
		opts, ok := optionsByQuestion[qid]
		if !ok {
			return apperr.Invalid(fmt.Sprintf("第 %d 条规则引用了不存在或已停用的问题", i+1))
		}
		for _, ans := range r.OwnAnswer {
			if !opts[ans] {
				return apperr.Invalid(fmt.Sprintf("第 %d 条规则答案不在选项范围内", i+1))
			}
		}
		for _, ans := range r.AcceptedAnswers {
			if !opts[ans] {
				return apperr.Invalid(fmt.Sprintf("第 %d 条规则接受答案不在选项范围内", i+1))
			}
		}
		if r.IsHard && r.Importance == 0 {
			return apperr.Invalid(fmt.Sprintf("第 %d 条规则设为硬条件时重要度不能为 0", i+1))
		}
	}

	return s.store.InTx(ctx, func(tx *store.Store) error {
		if err := tx.Q.DeleteRulesByUser(ctx, userID); err != nil {
			return err
		}
		for i, r := range rules {
			qid, _ := uuid.Parse(r.QuestionID)
			own, _ := json.Marshal(r.OwnAnswer)
			accepted, _ := json.Marshal(r.AcceptedAnswers)
			importance := r.Importance
			if importance == 0 && !r.IsHard {
				importance = 0
			}
			if _, err := tx.Q.UpsertRule(ctx, sqlcdb.UpsertRuleParams{
				UserID: userID, QuestionID: qid, OwnAnswer: own, AcceptedAnswers: accepted,
				Importance: importance, IsHard: r.IsHard,
			}); err != nil {
				return fmt.Errorf("写入第 %d 条规则失败: %w", i+1, err)
			}
		}
		return nil
	})
}

// MarketStatus 计算当前用户的入池状态。
func (s *Service) MarketStatus(ctx context.Context, userID uuid.UUID) (MarketStatus, error) {
	u, err := s.store.Q.GetUserByID(ctx, userID)
	if err != nil {
		return MarketStatus{}, apperr.ErrInternal.Wrap(err)
	}
	prof, err := s.GetProfile(ctx, userID)
	if err != nil {
		return MarketStatus{}, err
	}
	required := s.settings.Int(ctx, settings.KeyMinProfileCompletion, 80)
	consents, err := s.store.Q.ListConsents(ctx, userID)
	if err != nil {
		return MarketStatus{}, apperr.ErrInternal.Wrap(err)
	}
	consentSet := map[int16]bool{}
	for _, c := range consents {
		if !c.RevokedAt.Valid {
			consentSet[c.ConsentType] = true
		}
	}
	ruleCount, err := s.store.Q.CountRulesByUser(ctx, userID)
	if err != nil {
		return MarketStatus{}, apperr.ErrInternal.Wrap(err)
	}

	status := MarketStatus{
		InMarket:           u.DiscoveryEnabled,
		DiscoveryEnabled:   u.DiscoveryEnabled,
		ProfileCompletion:  prof.ProfileCompletion,
		RequiredCompletion: required,
		Age:                prof.Age,
		AgeOK:              prof.Age >= 18,
		CompletionOK:       int(prof.ProfileCompletion) >= required,
		ConsentOK:          consentSet[1] && consentSet[2] && consentSet[3],
		AccountOK:          u.Status == 2,
		RuleCount:          int(ruleCount),
	}
	if !status.AgeOK {
		status.MissingRequirements = append(status.MissingRequirements, "需年满 18 岁并填写出生日期")
	}
	if !status.CompletionOK {
		status.MissingRequirements = append(status.MissingRequirements,
			fmt.Sprintf("资料完成度需达到 %d%%，当前 %d%%", required, prof.ProfileCompletion))
	}
	if !status.ConsentOK {
		status.MissingRequirements = append(status.MissingRequirements, "需同意用户协议、隐私政策与婚恋市场参与声明")
	}
	if !status.AccountOK {
		status.MissingRequirements = append(status.MissingRequirements, "账号状态异常")
	}
	return status, nil
}

// JoinMarket 加入真实婚恋市场。
func (s *Service) JoinMarket(ctx context.Context, userID uuid.UUID) error {
	status, err := s.MarketStatus(ctx, userID)
	if err != nil {
		return err
	}
	if len(status.MissingRequirements) > 0 {
		return apperr.ErrProfileIncomplete.WithMessage("还不满足加入婚恋市场的条件：" +
			joinStrings(status.MissingRequirements, "；"))
	}
	return s.store.Q.SetUserDiscoveryEnabled(ctx, sqlcdb.SetUserDiscoveryEnabledParams{
		ID: userID, DiscoveryEnabled: true,
	})
}

// LeaveMarket 退出真实婚恋市场。
func (s *Service) LeaveMarket(ctx context.Context, userID uuid.UUID) error {
	return s.store.Q.SetUserDiscoveryEnabled(ctx, sqlcdb.SetUserDiscoveryEnabledParams{
		ID: userID, DiscoveryEnabled: false,
	})
}

// RecordConsent 记录协议同意（幂等）。
func (s *Service) RecordConsent(ctx context.Context, userID uuid.UUID, in ConsentInput) error {
	if err := s.validateInput(in); err != nil {
		return err
	}
	version := in.PolicyVersion
	if version == "" {
		switch in.ConsentType {
		case 1:
			version = s.settings.String(ctx, settings.KeyPolicyTerms, "2026.01")
		case 2:
			version = s.settings.String(ctx, settings.KeyPolicyPrivacy, "2026.01")
		case 3:
			version = s.settings.String(ctx, settings.KeyPolicyMarket, "2026.01")
		}
	}
	if _, err := s.store.Q.UpsertConsent(ctx, sqlcdb.UpsertConsentParams{
		UserID: userID, ConsentType: in.ConsentType, PolicyVersion: version,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// ListConsents 返回同意记录。
func (s *Service) ListConsents(ctx context.Context, userID uuid.UUID) ([]map[string]any, error) {
	rows, err := s.store.Q.ListConsents(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]map[string]any, 0, len(rows))
	for _, c := range rows {
		item := map[string]any{
			"consent_type":   c.ConsentType,
			"policy_version": c.PolicyVersion,
			"consented_at":   c.ConsentedAt.Time.Format(time.RFC3339),
		}
		if c.RevokedAt.Valid {
			item["revoked_at"] = c.RevokedAt.Time.Format(time.RFC3339)
		}
		out = append(out, item)
	}
	return out, nil
}

// RequestDeletion 申请注销：立即退出婚恋池并吊销全部 Token。
func (s *Service) RequestDeletion(ctx context.Context, userID uuid.UUID) error {
	if err := s.store.Q.RequestUserDeletion(ctx, userID); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if err := s.store.Q.RevokeAllUserTokens(ctx, userID); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// Seller 辅助：把多个缺失条件拼接。
func joinStrings(items []string, sep string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

// FormatInt 供日志使用。
func FormatInt(v int) string { return strconv.Itoa(v) }
