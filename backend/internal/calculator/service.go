package calculator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/events"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/platform/redisx"
	"wmc/backend/internal/platform/secure"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

// DatasetInfo 数据集对外信息。
type DatasetInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	SourceName  string    `json:"source_name"`
	SourceURL   string    `json:"source_url"`
	SourceYear  *int16    `json:"source_year"`
	Description string    `json:"description"`
	Methodology string    `json:"methodology"`
	Disclaimer  string    `json:"disclaimer"`
	IsDemo      bool      `json:"is_demo"`
}

// ConditionInput 计算器条件输入。
type ConditionInput struct {
	TargetGenders          []int16  `json:"target_genders"`
	MinAge                 *int     `json:"min_age"`
	MaxAge                 *int     `json:"max_age"`
	AllowedCityCodes       []string `json:"allowed_city_codes"`
	MinHeightCM            *int     `json:"min_height_cm"`
	MaxHeightCM            *int     `json:"max_height_cm"`
	MinEducationLevel      *int16   `json:"min_education_level"`
	MinIncomeBand          *int16   `json:"min_income_band"`
	AllowedMaritalStatuses []int16  `json:"allowed_marital_statuses"`
	AllowedChildrenStates  []int16  `json:"allowed_children_states"`
	AllowedWantsChildren   []int16  `json:"allowed_wants_children"`
}

// SelfInput 用户自述信息（可选，仅用于快照）。
type SelfInput struct {
	Gender           int16   `json:"gender"`
	Age              *int    `json:"age"`
	CityCode         string  `json:"city_code"`
	HeightCM         *int    `json:"height_cm"`
	EducationLevel   int16   `json:"education_level"`
	AnnualIncomeBand int16   `json:"annual_income_band"`
}

// EstimateRequest 估算请求。
type EstimateRequest struct {
	AnonymousSessionID string         `json:"anonymous_session_id"`
	Self               *SelfInput     `json:"self"`
	Conditions         ConditionInput `json:"conditions"`
}

// EstimateResponse 估算响应。
type EstimateResponse struct {
	RunID           uuid.UUID     `json:"run_id"`
	AnonymousID     string        `json:"anonymous_session_id,omitempty"`
	Dataset         DatasetInfo   `json:"dataset"`
	Estimate        Estimate      `json:"estimate"`
	ShareToken      *string       `json:"share_token"`
	MarketNotice    string        `json:"market_notice"`
	RegisterCTAText string        `json:"register_cta"`
}

// MarketNotice 理论市场提示文案（禁止说成精确人数或双向匹配率）。
const MarketNotice = "以下结果为基于基准人口数据集的统计估算，用于判断条件筛选的松紧程度，并非精确人数，也不代表双向匹配概率。"

// RegisterCTA 注册引导文案。
const RegisterCTAText = "注册后可以查看本站真实用户中，有多少人与你双向满足硬条件。"

// Service 计算器服务。
type Service struct {
	store    *store.Store
	pool     *pgxpool.Pool
	settings *settings.Service
	cache    redisx.Store
	now      func() time.Time
}

// NewService 构造计算器服务。
func NewService(st *store.Store, pool *pgxpool.Pool, set *settings.Service, cache redisx.Store) *Service {
	return &Service{store: st, pool: pool, settings: set, cache: cache, now: time.Now}
}

// ToFilter 将接口输入转换为引擎条件。
func (in ConditionInput) ToFilter() Filter {
	return Filter{
		Genders:         in.TargetGenders,
		MinAge:          in.MinAge,
		MaxAge:          in.MaxAge,
		Cities:          in.AllowedCityCodes,
		MaritalStatuses: in.AllowedMaritalStatuses,
		ChildrenStates:  in.AllowedChildrenStates,
		WantsChildren:   in.AllowedWantsChildren,
		MinHeight:       in.MinHeightCM,
		MaxHeight:       in.MaxHeightCM,
		MinEducation:    in.MinEducationLevel,
		MinIncome:       in.MinIncomeBand,
	}
}

type cachedEstimate struct {
	Dataset  DatasetInfo `json:"dataset"`
	Estimate Estimate    `json:"estimate"`
}

// Estimate 计算理论市场估算并落库。
func (s *Service) Estimate(ctx context.Context, req EstimateRequest, userID *uuid.UUID) (*EstimateResponse, error) {
	if !s.settings.Bool(ctx, settings.KeyCalculatorEnabled, true) {
		return nil, apperr.ErrFeatureDisabled.WithMessage("计算器功能暂未开放")
	}
	// 记录产品事件，保证运营漏斗「访问计算器 → 完成计算器」有数据
	var anonEventID *string
	if userID == nil && req.AnonymousSessionID != "" {
		anonEventID = &req.AnonymousSessionID
	}
	events.Record(ctx, s.store, userID, anonEventID, events.NameCalculatorStarted, nil)

	filter := req.Conditions.ToFilter()
	if err := filter.Validate(); err != nil {
		return nil, apperr.Invalid(err.Error())
	}

	ds, err := s.store.Q.GetActiveDataset(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrDatasetUnavailable
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}

	cacheKey := s.cacheKey(ds.ID, req.Conditions)
	payload, err := s.computeEstimate(ctx, ds, filter, cacheKey)
	if err != nil {
		return nil, err
	}

	anonID := req.AnonymousSessionID
	if userID == nil && anonID == "" {
		anonID = uuid.NewString()
	}
	if userID != nil {
		anonID = ""
	}

	response := &EstimateResponse{
		AnonymousID:     anonID,
		Dataset:         payload.Dataset,
		Estimate:        payload.Estimate,
		MarketNotice:    MarketNotice,
		RegisterCTAText: RegisterCTAText,
	}

	resultJSON := convert.MustJSON(map[string]any{
		"dataset":  payload.Dataset,
		"estimate": payload.Estimate,
	})
	inputJSON := convert.MustJSON(req)

	var userPtr *uuid.UUID
	if userID != nil {
		userPtr = userID
	}
	var anonPtr *string
	if anonID != "" {
		anonPtr = &anonID
	}
	run, err := s.store.Q.CreateCalculatorRun(ctx, sqlcdb.CreateCalculatorRunParams{
		UserID:             userPtr,
		AnonymousSessionID: anonPtr,
		DatasetID:          &ds.ID,
		InputSnapshot:      inputJSON,
		ResultSnapshot:     resultJSON,
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	response.RunID = run.ID
	events.Record(ctx, s.store, userID, anonEventID, events.NameCalculatorCompleted, map[string]any{
		"run_id": run.ID.String(), "matched_ratio": payload.Estimate.MatchedRatio,
	})
	return response, nil
}

func (s *Service) computeEstimate(ctx context.Context, ds sqlcdb.BaselineDataset, filter Filter, cacheKey string) (*cachedEstimate, error) {
	if s.cache != nil && cacheKey != "" {
		if raw, ok, _ := s.cache.Get(ctx, cacheKey); ok {
			var cached cachedEstimate
			if err := json.Unmarshal([]byte(raw), &cached); err == nil {
				return &cached, nil
			}
		}
	}

	counter := NewSQLCounter(s.pool, ds.ID)
	engine := NewEngine(counter)
	est, err := engine.Estimate(ctx, filter)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	est.DatasetID = ds.ID
	est.DatasetName = ds.Name
	est.DatasetVersion = ds.Version
	est.SourceName = ds.SourceName
	est.Disclaimer = ds.Disclaimer

	payload := &cachedEstimate{
		Dataset: DatasetInfo{
			ID: ds.ID, Name: ds.Name, Version: ds.Version, SourceName: ds.SourceName,
			SourceURL: ds.SourceUrl, SourceYear: ds.SourceYear, Description: ds.Description,
			Methodology: ds.Methodology, Disclaimer: ds.Disclaimer, IsDemo: ds.IsDemo,
		},
		Estimate: *est,
	}
	if s.cache != nil && cacheKey != "" {
		if raw, err := json.Marshal(payload); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(raw), 10*time.Minute)
		}
	}
	return payload, nil
}

func (s *Service) cacheKey(datasetID uuid.UUID, cond ConditionInput) string {
	normalized, err := json.Marshal(cond)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(normalized)
	return "calc:" + datasetID.String() + ":" + hex.EncodeToString(sum[:])
}

// CreateShare 为一次计算生成分享链接。
func (s *Service) CreateShare(ctx context.Context, runID uuid.UUID, anonID string, userID *uuid.UUID) (string, error) {
	run, err := s.getRun(ctx, runID)
	if err != nil {
		return "", err
	}
	if !s.ownsRun(run, anonID, userID) {
		return "", apperr.ErrForbidden.WithMessage("无权分享该结果")
	}
	if run.ShareToken != nil && *run.ShareToken != "" {
		return *run.ShareToken, nil
	}
	token, err := secure.RandomToken(16)
	if err != nil {
		return "", apperr.ErrInternal.Wrap(err)
	}
	if err := s.store.Q.UpdateCalculatorRunShareToken(ctx, sqlcdb.UpdateCalculatorRunShareTokenParams{
		ID: run.ID, ShareToken: &token,
	}); err != nil {
		return "", apperr.ErrInternal.Wrap(err)
	}
	return token, nil
}

// GetShare 读取分享结果。
func (s *Service) GetShare(ctx context.Context, token string) (*EstimateResponse, error) {
	run, err := s.store.Q.GetCalculatorRunByShareToken(ctx, &token)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound.WithMessage("分享链接不存在或已失效")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	var payload cachedEstimate
	if err := json.Unmarshal(run.ResultSnapshot, &payload); err != nil {
		return nil, apperr.ErrInternal.Wrap(fmt.Errorf("解析快照失败: %w", err))
	}
	return &EstimateResponse{
		RunID:           run.ID,
		Dataset:         payload.Dataset,
		Estimate:        payload.Estimate,
		ShareToken:      run.ShareToken,
		MarketNotice:    MarketNotice,
		RegisterCTAText: RegisterCTAText,
	}, nil
}

func (s *Service) getRun(ctx context.Context, runID uuid.UUID) (sqlcdb.CalculatorRun, error) {
	var run sqlcdb.CalculatorRun
	err := s.pool.QueryRow(ctx, `SELECT id, user_id, anonymous_session_id, dataset_id, input_snapshot, result_snapshot, share_token, created_at FROM calculator_runs WHERE id = $1`, runID).
		Scan(&run.ID, &run.UserID, &run.AnonymousSessionID, &run.DatasetID, &run.InputSnapshot,
			&run.ResultSnapshot, &run.ShareToken, &run.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return run, apperr.ErrNotFound.WithMessage("计算记录不存在")
	}
	if err != nil {
		return run, apperr.ErrInternal.Wrap(err)
	}
	return run, nil
}

func (s *Service) ownsRun(run sqlcdb.CalculatorRun, anonID string, userID *uuid.UUID) bool {
	if userID != nil && run.UserID != nil && *run.UserID == *userID {
		return true
	}
	if anonID != "" && run.AnonymousSessionID != nil && *run.AnonymousSessionID == anonID {
		return true
	}
	return false
}