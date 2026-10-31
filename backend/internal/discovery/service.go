// Package discovery 提供本站真实婚恋市场的候选池、漏斗与条件敏感性。
// 与理论市场严格区分：数据全部来自注册用户，且必须双向满足硬条件。
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/matching"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/platform/cursor"
	"wmc/backend/internal/profile"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/store"
)

// scanLimit 单次扫描候选上限（早期用户规模下的保护措施）。
const scanLimit = 500

// Service 候选池服务。
type Service struct {
	store    *store.Store
	pool     *pgxpool.Pool
	settings *settings.Service
	obj      storage.ObjectStorage
	now      func() time.Time
}

// NewService 构造服务。
func NewService(st *store.Store, pool *pgxpool.Pool, set *settings.Service, obj storage.ObjectStorage) *Service {
	return &Service{store: st, pool: pool, settings: set, obj: obj, now: time.Now}
}

// snapshot 当前用户的匹配快照。
type snapshot struct {
	user    sqlcdb.User
	profile sqlcdb.UserProfile
	pref    sqlcdb.PartnerPreference
	rules   []matching.ValueRule
	age     int
}

// 漏斗固定顺序（顺序会影响每一层的损失展示）。
var funnelDims = []struct{ Key, Label string }{
	{"gender", "目标性别"},
	{"age", "年龄"},
	{"city", "城市"},
	{"marital_status", "婚姻状态"},
	{"children", "子女情况"},
	{"wants_children", "生育意愿"},
	{"height", "身高"},
	{"education", "学历"},
	{"income", "年收入"},
}

func dimIndex(key string) int {
	for i, d := range funnelDims {
		if d.Key == key {
			return i
		}
	}
	return -1
}

// queryOpts 控制 SQL 条件生成。
type queryOpts struct {
	includeReverse bool   // 是否要求「我也满足对方的硬条件」
	excludeDim     string // 取消某个维度（敏感性分析）
	upToDim        string // 只累积到某个维度（漏斗）
	excludeAll     bool   // 不应用任何择偶条件（漏斗基数）
}

func (o queryOpts) active(dim string) bool {
	if o.excludeAll {
		return false
	}
	if o.upToDim != "" {
		limit := dimIndex(o.upToDim)
		idx := dimIndex(dim)
		if limit >= 0 && idx >= 0 && idx > limit {
			return false
		}
	}
	if o.excludeDim != "" && o.excludeDim == dim {
		return false
	}
	return true
}

func (s *Service) loadSnapshot(ctx context.Context, userID uuid.UUID) (*snapshot, error) {
	u, err := s.store.Q.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound.WithMessage("用户不存在")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	p, err := s.store.Q.GetProfile(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrProfileIncomplete.WithMessage("请先填写个人资料")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	pref, err := s.store.Q.GetPreferences(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrProfileIncomplete.WithMessage("请先填写择偶条件")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	rules, err := s.loadRules(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &snapshot{user: u, profile: p, pref: pref, rules: rules, age: profile.Age(p.BirthDate, s.now())}, nil
}

func (s *Service) loadRules(ctx context.Context, userID uuid.UUID) ([]matching.ValueRule, error) {
	rows, err := s.store.Q.ListRulesWithQuestions(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]matching.ValueRule, 0, len(rows))
	for _, r := range rows {
		out = append(out, matching.ValueRule{
			QuestionID: r.QuestionID, QuestionCode: r.Code, Category: r.Category,
			OwnAnswer: convert.JSONStrings(r.OwnAnswer), AcceptedAnswers: convert.JSONStrings(r.AcceptedAnswers),
			Importance: r.Importance, IsHard: r.IsHard,
		})
	}
	return out, nil
}

func (s *Service) loadRulesBatch(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID][]matching.ValueRule, error) {
	out := make(map[uuid.UUID][]matching.ValueRule, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.store.Q.ListRulesByUsers(ctx, ids)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	questions, err := s.store.Q.ListActiveQuestions(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	meta := make(map[uuid.UUID]sqlcdb.CompatibilityQuestion, len(questions))
	for _, q := range questions {
		meta[q.ID] = q
	}
	for _, r := range rows {
		if _, ok := meta[r.QuestionID]; !ok {
			continue
		}
		q := meta[r.QuestionID]
		out[r.UserID] = append(out[r.UserID], matching.ValueRule{
			QuestionID: r.QuestionID, QuestionCode: q.Code, Category: q.Category,
			OwnAnswer: convert.JSONStrings(r.OwnAnswer), AcceptedAnswers: convert.JSONStrings(r.AcceptedAnswers),
			Importance: r.Importance, IsHard: r.IsHard,
		})
	}
	return out, nil
}

// candidateSelect 候选查询的公共列。
const candidateSelect = `
SELECT u.id, p.nickname, p.gender, p.birth_date, p.city_code, p.height_cm, p.education_level,
       p.occupation_category, p.employment_status, p.annual_income_band, p.marital_status,
       p.children_count, p.wants_children, p.smoking_status, p.drinking_status, p.marriage_intent, p.bio,
       (SELECT ph.object_key FROM user_photos ph
         WHERE ph.user_id = u.id AND ph.is_primary AND ph.deleted_at IS NULL AND ph.moderation_status = 2
         ORDER BY ph.sort_order LIMIT 1) AS photo_key
FROM users u
JOIN user_profiles p ON p.user_id = u.id
JOIN partner_preferences pp ON pp.user_id = u.id
`

// buildWhere 生成筛选条件 SQL 与参数（全部通过占位符传递，禁止拼接用户输入）。
func (s *Service) buildWhere(me *snapshot, opts queryOpts) (string, []any) {
	args := []any{me.user.ID, s.settings.Int(context.Background(), settings.KeyMinProfileCompletion, 80)}
	selfPos, completionPos := 1, 2
	clauses := []string{
		"u.status = 2",
		"u.discovery_enabled",
		fmt.Sprintf("u.id <> $%d", selfPos),
		fmt.Sprintf("p.profile_completion >= $%d", completionPos),
		"p.birth_date IS NOT NULL",
		"EXTRACT(YEAR FROM age(p.birth_date)) >= 18",
		fmt.Sprintf("NOT EXISTS (SELECT 1 FROM user_blocks b WHERE (b.blocker_user_id = $%d AND b.blocked_user_id = u.id) OR (b.blocker_user_id = u.id AND b.blocked_user_id = $%d))", selfPos, selfPos),
	}
	add := func(cond string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(cond, len(args)))
	}

	p := me.pref
	if opts.active("gender") && len(p.TargetGenders) > 0 {
		add("p.gender = ANY($%d)", p.TargetGenders)
	}
	if opts.active("age") && p.AgeHard {
		add("EXTRACT(YEAR FROM age(p.birth_date)) >= $%d", p.MinAge)
		add("EXTRACT(YEAR FROM age(p.birth_date)) <= $%d", p.MaxAge)
	}
	if opts.active("city") && p.CityHard && len(p.AllowedCityCodes) > 0 {
		add("p.city_code = ANY($%d)", p.AllowedCityCodes)
	}
	if opts.active("marital_status") && p.MaritalStatusHard && len(p.AllowedMaritalStatuses) > 0 {
		add("p.marital_status = ANY($%d)", p.AllowedMaritalStatuses)
	}
	if opts.active("children") && p.ChildrenHard && len(p.AllowedChildrenStates) > 0 {
		add("(CASE WHEN p.children_count > 0 THEN 1 ELSE 0 END) = ANY($%d)", p.AllowedChildrenStates)
	}
	if opts.active("wants_children") && p.WantsChildrenHard && len(p.AllowedWantsChildren) > 0 {
		add("p.wants_children = ANY($%d)", p.AllowedWantsChildren)
	}
	if opts.active("height") && p.HeightHard {
		add("p.height_cm IS NOT NULL AND p.height_cm >= $%d", p.MinHeightCm)
		add("p.height_cm <= $%d", p.MaxHeightCm)
	}
	if opts.active("education") && p.EducationHard {
		add("p.education_level >= $%d", p.MinEducationLevel)
	}
	if opts.active("income") && p.IncomeHard {
		add("p.annual_income_band BETWEEN 1 AND 7 AND p.annual_income_band >= $%d", p.MinIncomeBand)
	}
	if opts.active("smoking") && p.SmokingHard && len(p.AllowedSmokingStatuses) > 0 {
		add("p.smoking_status = ANY($%d)", p.AllowedSmokingStatuses)
	}
	if opts.active("drinking") && p.DrinkingHard && len(p.AllowedDrinkingStatuses) > 0 {
		add("p.drinking_status = ANY($%d)", p.AllowedDrinkingStatuses)
	}
	if opts.active("marriage_intent") && p.MarriageIntentHard && p.MinMarriageIntent > 0 {
		add("(CASE p.marriage_intent WHEN 1 THEN 5 WHEN 2 THEN 4 WHEN 3 THEN 3 WHEN 4 THEN 2 WHEN 5 THEN 1 ELSE 0 END) >= "+
			"(CASE $%d::smallint WHEN 1 THEN 5 WHEN 2 THEN 4 WHEN 3 THEN 3 WHEN 4 THEN 2 WHEN 5 THEN 1 ELSE 0 END)", p.MinMarriageIntent)
	}

	if opts.includeReverse {
		add("(cardinality(pp.target_genders) = 0 OR $%d = ANY(pp.target_genders))", me.profile.Gender)
		add("(NOT pp.age_hard OR ($%d BETWEEN pp.min_age AND pp.max_age))", me.age)
		if me.profile.HeightCm != nil {
			add("(NOT pp.height_hard OR ($%d BETWEEN pp.min_height_cm AND pp.max_height_cm))", *me.profile.HeightCm)
		} else {
			clauses = append(clauses, "(NOT pp.height_hard)")
		}
		add("(NOT pp.city_hard OR $%d = ANY(pp.allowed_city_codes))", me.profile.CityCode)
		add("(NOT pp.education_hard OR $%d >= pp.min_education_level)", me.profile.EducationLevel)
		add("(NOT pp.income_hard OR $%d BETWEEN 1 AND 7)", me.profile.AnnualIncomeBand)
		add("(NOT pp.income_hard OR $%d >= pp.min_income_band)", me.profile.AnnualIncomeBand)
		add("(NOT pp.marital_status_hard OR $%d = ANY(pp.allowed_marital_statuses))", me.profile.MaritalStatus)
		add("(NOT pp.children_hard OR $%d = ANY(pp.allowed_children_states))",
			matching.ChildrenState(me.profile.ChildrenCount))
		add("(NOT pp.wants_children_hard OR $%d = ANY(pp.allowed_wants_children))", me.profile.WantsChildren)
		add("(NOT pp.smoking_hard OR $%d = ANY(pp.allowed_smoking_statuses))", me.profile.SmokingStatus)
		add("(NOT pp.drinking_hard OR $%d = ANY(pp.allowed_drinking_statuses))", me.profile.DrinkingStatus)
		add("(NOT pp.marriage_intent_hard OR (CASE $%d::smallint WHEN 1 THEN 5 WHEN 2 THEN 4 WHEN 3 THEN 3 WHEN 4 THEN 2 WHEN 5 THEN 1 ELSE 0 END) >= "+
			"(CASE pp.min_marriage_intent WHEN 1 THEN 5 WHEN 2 THEN 4 WHEN 3 THEN 3 WHEN 4 THEN 2 WHEN 5 THEN 1 ELSE 0 END))",
			me.profile.MarriageIntent)
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (s *Service) countCandidates(ctx context.Context, me *snapshot, opts queryOpts) (int, error) {
	where, args := s.buildWhere(me, opts)
	var n int
	if err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users u JOIN user_profiles p ON p.user_id = u.id JOIN partner_preferences pp ON pp.user_id = u.id"+where, args...).Scan(&n); err != nil {
		return 0, apperr.ErrInternal.Wrap(err)
	}
	return n, nil
}

type candidateRow struct {
	UserID             uuid.UUID
	Nickname           string
	Gender             int16
	BirthDate          time.Time
	CityCode           string
	HeightCM           *int16
	EducationLevel     int16
	OccupationCategory int16
	EmploymentStatus   int16
	AnnualIncomeBand   int16
	MaritalStatus      int16
	ChildrenCount      int16
	WantsChildren      int16
	SmokingStatus      int16
	DrinkingStatus     int16
	MarriageIntent     int16
	Bio                string
	PhotoKey           *string
}

func (s *Service) fetchCandidates(ctx context.Context, me *snapshot, opts queryOpts, extraWhere string, extraArgs []any, limit int) ([]candidateRow, error) {
	where, args := s.buildWhere(me, opts)
	query := candidateSelect + where
	if extraWhere != "" {
		query += " AND " + extraWhere
		args = append(args, extraArgs...)
	}
	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT %d", limit)
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	out := make([]candidateRow, 0, limit)
	for rows.Next() {
		var r candidateRow
		if err := rows.Scan(&r.UserID, &r.Nickname, &r.Gender, &r.BirthDate, &r.CityCode, &r.HeightCM,
			&r.EducationLevel, &r.OccupationCategory, &r.EmploymentStatus, &r.AnnualIncomeBand,
			&r.MaritalStatus, &r.ChildrenCount, &r.WantsChildren, &r.SmokingStatus, &r.DrinkingStatus,
			&r.MarriageIntent, &r.Bio, &r.PhotoKey); err != nil {
			return nil, apperr.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) loadPreferencesBatch(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]matching.Preference, error) {
	out := make(map[uuid.UUID]matching.Preference, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.store.Q.ListPreferencesByUsers(ctx, ids)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	for _, r := range rows {
		out[r.UserID] = toMatchingPreference(r)
	}
	return out, nil
}

func (s *Service) loadActions(ctx context.Context, actorID uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]int16, error) {
	out := make(map[uuid.UUID]int16, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.store.Q.ListInteractionsByActorTargets(ctx, sqlcdb.ListInteractionsByActorTargetsParams{
		ActorUserID: actorID, Column2: ids,
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	for _, r := range rows {
		out[r.TargetUserID] = r.Action
	}
	return out, nil
}

func (s *Service) newEngine(ctx context.Context) *matching.Engine {
	return matching.NewEngine(matching.Weights{
		Preference:    s.settings.Float(ctx, settings.KeyPreferenceWeight, 0.4),
		Compatibility: s.settings.Float(ctx, settings.KeyCompatibilityWeight, 0.6),
	})
}

func (s *Service) toMatchingProfile(me *snapshot) matching.Profile {
	p := me.profile
	return matching.Profile{
		UserID: me.user.ID, Nickname: p.Nickname, Gender: p.Gender, Age: me.age, HeightCM: p.HeightCm,
		CityCode: p.CityCode, EducationLevel: p.EducationLevel, OccupationCategory: p.OccupationCategory,
		AnnualIncomeBand: p.AnnualIncomeBand, MaritalStatus: p.MaritalStatus, ChildrenCount: p.ChildrenCount,
		WantsChildren: p.WantsChildren, MarriageIntent: p.MarriageIntent, SmokingStatus: p.SmokingStatus,
		DrinkingStatus: p.DrinkingStatus, Bio: p.Bio,
	}
}

func toMatchingPreference(p sqlcdb.PartnerPreference) matching.Preference {
	weights := map[string]float64{}
	if len(p.Weights) > 0 {
		_ = json.Unmarshal(p.Weights, &weights)
	}
	return matching.Preference{
		TargetGenders: p.TargetGenders, MinAge: p.MinAge, MaxAge: p.MaxAge, AgeHard: p.AgeHard,
		MinHeightCM: p.MinHeightCm, MaxHeightCM: p.MaxHeightCm, HeightHard: p.HeightHard,
		AllowedCityCodes: p.AllowedCityCodes, CityHard: p.CityHard,
		MinEducationLevel: p.MinEducationLevel, EducationHard: p.EducationHard,
		MinIncomeBand: p.MinIncomeBand, IncomeHard: p.IncomeHard,
		AllowedMarital: p.AllowedMaritalStatuses, MaritalHard: p.MaritalStatusHard,
		AllowedChildren: p.AllowedChildrenStates, ChildrenHard: p.ChildrenHard,
		AllowedWantsChildren: p.AllowedWantsChildren, WantsChildrenHard: p.WantsChildrenHard,
		AllowedSmoking: p.AllowedSmokingStatuses, SmokingHard: p.SmokingHard,
		AllowedDrinking: p.AllowedDrinkingStatuses, DrinkingHard: p.DrinkingHard,
		MinMarriageIntent: p.MinMarriageIntent, MarriageIntentHard: p.MarriageIntentHard,
		Weights: weights,
	}
}

func yearsBetween(birth, now time.Time) int {
	years := now.Year() - birth.Year()
	if now.Month() < birth.Month() || (now.Month() == birth.Month() && now.Day() < birth.Day()) {
		years--
	}
	if years < 0 {
		return 0
	}
	return years
}

func (s *Service) signedPhotoURL(ctx context.Context, key string) (string, error) {
	if s.obj == nil {
		return "", errors.New("未配置对象存储")
	}
	return s.obj.SignedURL(ctx, key, time.Hour)
}
// Summary 候选池概览（含 K 匿名与冷启动文案）。
type Summary struct {
	InMarket          bool         `json:"in_market"`
	OneWayCount       int          `json:"one_way_count"`
	ReciprocalCount   int          `json:"reciprocal_count"`
	EligibleCount     int          `json:"eligible_count"`
	EligibleCountText string       `json:"eligible_count_text"`
	EarlyStage        bool         `json:"early_stage"`
	EarlyStageMessage string       `json:"early_stage_message"`
	ProfileCompletion int16        `json:"profile_completion"`
	Requirements      Requirements `json:"requirements"`
}

// Requirements 入池条件状态。
type Requirements struct {
	MinProfileCompletion int  `json:"min_profile_completion"`
	AgeOK                bool `json:"age_ok"`
	ConsentOK            bool `json:"consent_ok"`
	AccountOK            bool `json:"account_ok"`
	PreferencesOK        bool `json:"preferences_ok"`
}

// Anonymize 小样本匿名处理：小于阈值统一返回「<K」。
func Anonymize(count, threshold int) string {
	if count < threshold {
		return fmt.Sprintf("<%d", threshold)
	}
	return fmt.Sprintf("%d", count)
}

// EarlyStageMessage 冷启动文案（禁止显示「0 个人」）。
func EarlyStageMessage(threshold int) string {
	return fmt.Sprintf("本站目前仍处于早期测试阶段，你当前已有 <%d 名双向基础候选人，随着目标城市用户增加，候选池会自动更新。", threshold)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetSummary 计算候选池概览。
func (s *Service) GetSummary(ctx context.Context, userID uuid.UUID) (*Summary, error) {
	me, err := s.loadSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	threshold := s.settings.Int(ctx, settings.KeyKAnonymityThreshold, 10)
	oneWay, err := s.countCandidates(ctx, me, queryOpts{includeReverse: false})
	if err != nil {
		return nil, err
	}
	reciprocal, err := s.countCandidates(ctx, me, queryOpts{includeReverse: true})
	if err != nil {
		return nil, err
	}
	early := reciprocal < threshold
	summary := &Summary{
		InMarket:          me.user.DiscoveryEnabled,
		OneWayCount:       oneWay,
		ReciprocalCount:   reciprocal,
		EligibleCount:     reciprocal,
		EligibleCountText: Anonymize(reciprocal, threshold),
		EarlyStage:        early,
		ProfileCompletion: me.profile.ProfileCompletion,
		Requirements: Requirements{
			MinProfileCompletion: s.settings.Int(ctx, settings.KeyMinProfileCompletion, 80),
			AgeOK:                me.age >= 18,
			AccountOK:            me.user.Status == 2,
			PreferencesOK:        true,
		},
	}
	if early {
		summary.EarlyStageMessage = EarlyStageMessage(threshold)
	}
	if consents, err := s.store.Q.ListConsents(ctx, userID); err == nil {
		set := map[int16]bool{}
		for _, c := range consents {
			if !c.RevokedAt.Valid {
				set[c.ConsentType] = true
			}
		}
		summary.Requirements.ConsentOK = set[1] && set[2] && set[3]
	}
	return summary, nil
}

// FunnelStep 漏斗步骤。
type FunnelStep struct {
	Key      string  `json:"key"`
	Label    string  `json:"label"`
	Count    int     `json:"count"`
	Ratio    float64 `json:"ratio"`
	DropRate float64 `json:"drop_ratio"`
	Applied  bool    `json:"applied"`
}

// Funnel 真实候选池漏斗。
type Funnel struct {
	Steps           []FunnelStep `json:"steps"`
	Total           int          `json:"total"`
	FinalCount      int          `json:"final_count"`
	FinalRatio      float64      `json:"final_ratio"`
	ReciprocalCount int          `json:"reciprocal_count"`
	ReciprocalRatio float64      `json:"reciprocal_ratio"`
	ReciprocalText  string       `json:"reciprocal_count_text"`
	Note            string       `json:"note"`
}

// FunnelNote 说明文案。
const FunnelNote = "真实候选池按当前条件顺序计算；最后一层要求双方同时满足硬条件。"

// GetFunnel 计算真实候选池漏斗。
func (s *Service) GetFunnel(ctx context.Context, userID uuid.UUID) (*Funnel, error) {
	me, err := s.loadSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	threshold := s.settings.Int(ctx, settings.KeyKAnonymityThreshold, 10)
	total, err := s.countCandidates(ctx, me, queryOpts{includeReverse: false, excludeAll: true})
	if err != nil {
		return nil, err
	}
	out := &Funnel{Total: total, Note: FunnelNote, Steps: make([]FunnelStep, 0, len(funnelDims)+1)}
	prev := total
	for _, d := range funnelDims {
		count, err := s.countCandidates(ctx, me, queryOpts{includeReverse: false, upToDim: d.Key})
		if err != nil {
			return nil, err
		}
		ratio, drop := 0.0, 0.0
		if total > 0 {
			ratio = float64(count) / float64(total)
		}
		if prev > 0 {
			drop = float64(prev-count) / float64(prev)
		}
		out.Steps = append(out.Steps, FunnelStep{
			Key: d.Key, Label: d.Label, Count: count, Ratio: ratio, DropRate: drop, Applied: true,
		})
		prev = count
	}
	reciprocal, err := s.countCandidates(ctx, me, queryOpts{includeReverse: true})
	if err != nil {
		return nil, err
	}
	finalRatio := 0.0
	if total > 0 {
		finalRatio = float64(reciprocal) / float64(total)
	}
	out.Steps = append(out.Steps, FunnelStep{
		Key: "reciprocal", Label: "同时满足双方硬条件", Count: reciprocal,
		Ratio: finalRatio, Applied: true,
	})
	out.FinalCount = prev
	out.ReciprocalCount = reciprocal
	out.ReciprocalRatio = finalRatio
	out.ReciprocalText = Anonymize(reciprocal, threshold)
	return out, nil
}

// SensitivityItem 单条件敏感性。
type SensitivityItem struct {
	Key          string  `json:"key"`
	Label        string  `json:"label"`
	BaseCount    int     `json:"base_count"`
	WithoutCount int     `json:"without_count"`
	DeltaRatio   float64 `json:"delta_ratio"`
	AddedCount   int     `json:"added_count"`
}

// SensitivityResult 敏感性分析结果。
type SensitivityResult struct {
	Items []SensitivityItem `json:"items"`
	Note  string            `json:"note"`
}

// SensitivityNote 说明。
const SensitivityNote = "逐条取消你的硬条件后重新计算双向候选数；取消条件后候选数不会减少。"

// GetSensitivity 计算条件敏感性（Leave-One-Out）。
func (s *Service) GetSensitivity(ctx context.Context, userID uuid.UUID) (*SensitivityResult, error) {
	me, err := s.loadSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	base, err := s.countCandidates(ctx, me, queryOpts{includeReverse: true})
	if err != nil {
		return nil, err
	}
	items := make([]SensitivityItem, 0, len(funnelDims))
	for _, d := range funnelDims {
		count, err := s.countCandidates(ctx, me, queryOpts{includeReverse: true, excludeDim: d.Key})
		if err != nil {
			return nil, err
		}
		delta := 0.0
		if base > 0 {
			delta = float64(count-base) / float64(base)
		}
		items = append(items, SensitivityItem{
			Key: d.Key, Label: d.Label, BaseCount: base, WithoutCount: count,
			DeltaRatio: delta, AddedCount: count - base,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].DeltaRatio != items[j].DeltaRatio {
			return items[i].DeltaRatio > items[j].DeltaRatio
		}
		return items[i].Key < items[j].Key
	})
	return &SensitivityResult{Items: items, Note: SensitivityNote}, nil
}

// CandidateCard 候选卡。
type CandidateCard struct {
	UserID                    string                     `json:"user_id"`
	Nickname                  string                     `json:"nickname"`
	Age                       int                        `json:"age"`
	CityCode                  string                     `json:"city_code"`
	CityName                  string                     `json:"city_name"`
	HeightCM                  *int16                     `json:"height_cm"`
	EducationLevel            int16                      `json:"education_level"`
	OccupationCategory        int16                      `json:"occupation_category"`
	MaritalStatus             int16                      `json:"marital_status"`
	WantsChildren             int16                      `json:"wants_children"`
	Bio                       string                     `json:"bio"`
	PrimaryPhotoURL           string                     `json:"primary_photo_url"`
	PreferenceScore           float64                    `json:"preference_score"`
	CompatibilityScore        float64                    `json:"compatibility_score"`
	OverallScore              float64                    `json:"overall_score"`
	CompatibilityInsufficient bool                       `json:"compatibility_insufficient"`
	MyAction                  *int16                     `json:"my_action"`
	Reasons                   []string                   `json:"reasons,omitempty"`
	Conflicts                 []string                   `json:"conflicts,omitempty"`
	Dimensions                []matching.DimensionScore `json:"dimensions,omitempty"`
}

// CandidatePage 分页结果。
type CandidatePage struct {
	Items      []CandidateCard `json:"items"`
	NextCursor string          `json:"next_cursor"`
	HasMore    bool            `json:"has_more"`
}

// Candidates 返回双向候选人列表（keyset 游标分页）。
func (s *Service) Candidates(ctx context.Context, userID uuid.UUID, cursorStr string, limit int) (*CandidatePage, error) {
	me, err := s.loadSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !me.user.DiscoveryEnabled {
		return nil, apperr.ErrForbidden.WithMessage("你还没有加入真实婚恋市场")
	}
	pageSize := s.settings.Int(ctx, settings.KeyCandidatePageSize, 20)
	if limit > 0 {
		pageSize = limit
	}
	if maxSize := s.settings.Int(ctx, settings.KeyMaxCandidatePageSize, 50); pageSize > maxSize {
		pageSize = maxSize
	}

	rows, err := s.fetchCandidates(ctx, me, queryOpts{includeReverse: true}, "", nil, scanLimit)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return &CandidatePage{Items: []CandidateCard{}}, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.UserID)
	}
	rulesByUser, err := s.loadRulesBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	prefs, err := s.loadPreferencesBatch(ctx, ids)
	if err != nil {
		return nil, err
	}
	actions, err := s.loadActions(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	cityNames, err := s.loadCityNames(ctx)
	if err != nil {
		return nil, err
	}

	eng := s.newEngine(ctx)
	meProfile := s.toMatchingProfile(me)
	mePref := toMatchingPreference(me.pref)
	cards := make([]CandidateCard, 0, len(rows))
	for _, r := range rows {
		cand := matching.Profile{
			UserID: r.UserID, Nickname: r.Nickname, Gender: r.Gender, Age: yearsBetween(r.BirthDate, s.now()),
			HeightCM: r.HeightCM, CityCode: r.CityCode, EducationLevel: r.EducationLevel,
			OccupationCategory: r.OccupationCategory, AnnualIncomeBand: r.AnnualIncomeBand,
			MaritalStatus: r.MaritalStatus, ChildrenCount: r.ChildrenCount,
			WantsChildren: r.WantsChildren, MarriageIntent: r.MarriageIntent,
			SmokingStatus: r.SmokingStatus, DrinkingStatus: r.DrinkingStatus, Bio: r.Bio,
		}
		res := eng.Evaluate(matching.Input{
			Me: meProfile, MyPreference: mePref, MyRules: me.rules,
			Candidate: cand, PeerPreference: prefs[r.UserID], PeerRules: rulesByUser[r.UserID],
		})
		if !res.Eligible {
			continue
		}
		card := CandidateCard{
			UserID: r.UserID.String(), Nickname: r.Nickname, Age: cand.Age, CityCode: r.CityCode,
			CityName: cityNames[r.CityCode],
			HeightCM: r.HeightCM, EducationLevel: r.EducationLevel, OccupationCategory: r.OccupationCategory,
			MaritalStatus: r.MaritalStatus, WantsChildren: r.WantsChildren, Bio: r.Bio,
			PreferenceScore: res.PreferenceScore, CompatibilityScore: res.CompatibilityScore,
			OverallScore: res.OverallScore, CompatibilityInsufficient: res.Insufficient,
			Dimensions: res.CategoryScores,
		}
		if r.PhotoKey != nil {
			if url, err := s.signedPhotoURL(ctx, *r.PhotoKey); err == nil {
				card.PrimaryPhotoURL = url
			}
		}
		if a, ok := actions[r.UserID]; ok {
			v := a
			card.MyAction = &v
		}
		cards = append(cards, card)
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].OverallScore != cards[j].OverallScore {
			return cards[i].OverallScore > cards[j].OverallScore
		}
		return cards[i].UserID < cards[j].UserID
	})

	cur, err := cursor.Decode(cursorStr)
	if err != nil {
		return nil, apperr.Invalid("分页游标非法")
	}
	start := 0
	if cur.Valid() {
		for i, c := range cards {
			if c.UserID == cur.UserID {
				start = i + 1
				break
			}
		}
	}
	if start > len(cards) {
		start = len(cards)
	}
	end := start + pageSize
	if end > len(cards) {
		end = len(cards)
	}
	page := cards[start:end]
	next := ""
	hasMore := end < len(cards)
	if hasMore && len(page) > 0 {
		last := page[len(page)-1]
		next = cursor.Encode(cursor.Cursor{Score: last.OverallScore, UserID: last.UserID})
	}
	return &CandidatePage{Items: page, NextCursor: next, HasMore: hasMore}, nil
}

// CandidateDetail 候选人详情（含「为什么匹配」解释）。
func (s *Service) CandidateDetail(ctx context.Context, userID, candidateID uuid.UUID) (*CandidateCard, error) {
	me, err := s.loadSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !me.user.DiscoveryEnabled {
		return nil, apperr.ErrForbidden.WithMessage("你还没有加入真实婚恋市场")
	}
	detailRows, err := s.fetchCandidateByID(ctx, me, candidateID)
	if err != nil {
		return nil, err
	}
	if len(detailRows) == 0 {
		return nil, apperr.ErrNotFound.WithMessage("该用户不可见或与你的条件不匹配")
	}
	r := detailRows[0]
	rulesByUser, err := s.loadRulesBatch(ctx, []uuid.UUID{candidateID})
	if err != nil {
		return nil, err
	}
	prefs, err := s.loadPreferencesBatch(ctx, []uuid.UUID{candidateID})
	if err != nil {
		return nil, err
	}
	actions, err := s.loadActions(ctx, userID, []uuid.UUID{candidateID})
	if err != nil {
		return nil, err
	}
	cityNames, _ := s.loadCityNames(ctx)
	eng := s.newEngine(ctx)
	cand := matching.Profile{
		UserID: r.UserID, Nickname: r.Nickname, Gender: r.Gender, Age: yearsBetween(r.BirthDate, s.now()),
		HeightCM: r.HeightCM, CityCode: r.CityCode, EducationLevel: r.EducationLevel,
		OccupationCategory: r.OccupationCategory, AnnualIncomeBand: r.AnnualIncomeBand,
		MaritalStatus: r.MaritalStatus, ChildrenCount: r.ChildrenCount,
		WantsChildren: r.WantsChildren, MarriageIntent: r.MarriageIntent,
		SmokingStatus: r.SmokingStatus, DrinkingStatus: r.DrinkingStatus, Bio: r.Bio,
	}
	res := eng.Evaluate(matching.Input{
		Me: s.toMatchingProfile(me), MyPreference: toMatchingPreference(me.pref), MyRules: me.rules,
		Candidate: cand, PeerPreference: prefs[r.UserID], PeerRules: rulesByUser[r.UserID],
	})
	if !res.Eligible {
		return nil, apperr.ErrNotFound.WithMessage("你们之间存在硬条件冲突")
	}
	card := CandidateCard{
		UserID: r.UserID.String(), Nickname: r.Nickname, Age: cand.Age, CityCode: r.CityCode,
		CityName: cityNames[r.CityCode],
		HeightCM: r.HeightCM, EducationLevel: r.EducationLevel, OccupationCategory: r.OccupationCategory,
		MaritalStatus: r.MaritalStatus, WantsChildren: r.WantsChildren, Bio: r.Bio,
		PreferenceScore: res.PreferenceScore, CompatibilityScore: res.CompatibilityScore,
		OverallScore: res.OverallScore, CompatibilityInsufficient: res.Insufficient,
		Reasons: res.Reasons, Dimensions: res.CategoryScores,
	}
	for _, c := range res.Conflicts {
		card.Conflicts = append(card.Conflicts, c.Label+"："+c.Detail)
	}
	if r.PhotoKey != nil {
		if url, err := s.signedPhotoURL(ctx, *r.PhotoKey); err == nil {
			card.PrimaryPhotoURL = url
		}
	}
	if a, ok := actions[r.UserID]; ok {
		v := a
		card.MyAction = &v
	}
	return &card, nil
}

func (s *Service) fetchCandidateByID(ctx context.Context, me *snapshot, candidateID uuid.UUID) ([]candidateRow, error) {
	where, args := s.buildWhere(me, queryOpts{includeReverse: true})
	args = append(args, candidateID)
	query := candidateSelect + where + fmt.Sprintf(" AND u.id = $%d LIMIT 1", len(args))
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	defer rows.Close()
	out := make([]candidateRow, 0, 1)
	for rows.Next() {
		var r candidateRow
		if err := rows.Scan(&r.UserID, &r.Nickname, &r.Gender, &r.BirthDate, &r.CityCode, &r.HeightCM,
			&r.EducationLevel, &r.OccupationCategory, &r.EmploymentStatus, &r.AnnualIncomeBand,
			&r.MaritalStatus, &r.ChildrenCount, &r.WantsChildren, &r.SmokingStatus, &r.DrinkingStatus,
			&r.MarriageIntent, &r.Bio, &r.PhotoKey); err != nil {
			return nil, apperr.ErrInternal.Wrap(err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) loadCityNames(ctx context.Context) (map[string]string, error) {
	regions, err := s.store.Q.ListRegions(ctx)
	if err != nil {
		return map[string]string{}, nil
	}
	out := make(map[string]string, len(regions))
	for _, r := range regions {
		out[r.Code] = r.Name
	}
	return out, nil
}