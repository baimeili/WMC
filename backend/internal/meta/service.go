// Package meta 提供元数据接口：选项字典、行政区划、价值观问题、应用配置。
package meta

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/enums"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

// Service 元数据服务。
type Service struct {
	store    *store.Store
	settings *settings.Service
}

// NewService 构造元数据服务。
func NewService(st *store.Store, set *settings.Service) *Service {
	return &Service{store: st, settings: set}
}

// Options 返回全部选项字典。
func (s *Service) Options() map[string][]enums.Option { return enums.Options() }

// Question 问题 DTO。
type Question struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Category     string    `json:"category"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	QuestionType int16     `json:"question_type"`
	Options      []Option  `json:"options"`
	Active       bool      `json:"active"`
	SortOrder    int16     `json:"sort_order"`
	Version      int16     `json:"version"`
}

// Option 问题选项。
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

func toQuestion(q sqlcdb.CompatibilityQuestion) Question {
	opts := make([]Option, 0, 4)
	if len(q.Options) > 0 {
		_ = json.Unmarshal(q.Options, &opts)
	}
	return Question{
		ID: q.ID, Code: q.Code, Category: q.Category, Title: q.Title, Description: q.Description,
		QuestionType: q.QuestionType, Options: opts, Active: q.Active, SortOrder: q.SortOrder, Version: q.Version,
	}
}

// ActiveQuestions 返回启用的价值观问题。
func (s *Service) ActiveQuestions(ctx context.Context) ([]Question, error) {
	rows, err := s.store.Q.ListActiveQuestions(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]Question, 0, len(rows))
	for _, q := range rows {
		out = append(out, toQuestion(q))
	}
	return out, nil
}

// AllQuestions 返回全部问题（后台使用）。
func (s *Service) AllQuestions(ctx context.Context) ([]Question, error) {
	rows, err := s.store.Q.ListAllQuestions(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]Question, 0, len(rows))
	for _, q := range rows {
		out = append(out, toQuestion(q))
	}
	return out, nil
}

// Region DTO。
type Region struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Level      int16   `json:"level"`
	ParentCode *string `json:"parent_code"`
}

// Regions 返回行政区划。
func (s *Service) Regions(ctx context.Context) ([]Region, error) {
	rows, err := s.store.Q.ListRegions(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]Region, 0, len(rows))
	for _, r := range rows {
		out = append(out, Region{Code: r.Code, Name: r.Name, Level: r.Level, ParentCode: r.ParentCode})
	}
	return out, nil
}

// AppConfig 返回前端配置。
func (s *Service) AppConfig(ctx context.Context) map[string]any {
	return s.settings.PublicConfig(ctx)
}

// CreateQuestion 新增问题。
type QuestionInput struct {
	Code         string   `json:"code"`
	Category     string   `json:"category"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	QuestionType int16    `json:"question_type"`
	Options      []Option `json:"options"`
	Active       *bool    `json:"active"`
	SortOrder    int16    `json:"sort_order"`
}

// CreateQuestion 新增价值观问题。
func (s *Service) CreateQuestion(ctx context.Context, in QuestionInput) (Question, error) {
	if in.Code == "" || in.Title == "" {
		return Question{}, apperr.Invalid("问题编码与标题不能为空")
	}
	if len(in.Options) == 0 {
		return Question{}, apperr.Invalid("问题至少需要一个选项")
	}
	qtype := in.QuestionType
	if qtype == 0 {
		qtype = 1
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	opts, _ := json.Marshal(in.Options)
	q, err := s.store.Q.CreateQuestion(ctx, sqlcdb.CreateQuestionParams{
		Code: in.Code, Category: in.Category, Title: in.Title, Description: in.Description,
		QuestionType: qtype, Options: opts, Active: active, SortOrder: in.SortOrder, Version: 1,
	})
	if err != nil {
		return Question{}, apperr.ErrInternal.Wrap(err)
	}
	return toQuestion(q), nil
}

// UpdateQuestion 修改问题。
func (s *Service) UpdateQuestion(ctx context.Context, id uuid.UUID, in QuestionInput) (Question, error) {
	qtype := in.QuestionType
	if qtype == 0 {
		qtype = 1
	}
	active := true
	if in.Active != nil {
		active = *in.Active
	}
	opts, _ := json.Marshal(in.Options)
	q, err := s.store.Q.UpdateQuestion(ctx, sqlcdb.UpdateQuestionParams{
		ID: id, Category: in.Category, Title: in.Title, Description: in.Description,
		QuestionType: qtype, Options: opts, Active: active, SortOrder: in.SortOrder,
	})
	if err != nil {
		return Question{}, apperr.ErrInternal.Wrap(err)
	}
	return toQuestion(q), nil
}
