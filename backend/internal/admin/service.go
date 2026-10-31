// Package admin 提供运营后台能力：用户管理、审计日志、运营指标。
package admin

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/platform/enums"
	"wmc/backend/internal/platform/secure"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/store"
)

// Service 后台服务。
type Service struct {
	store    *store.Store
	settings *settings.Service
	obj      storage.ObjectStorage
	phoneKey []byte
}

// NewService 构造后台服务。
func NewService(st *store.Store, set *settings.Service, obj storage.ObjectStorage, phoneKey []byte) *Service {
	return &Service{store: st, settings: set, obj: obj, phoneKey: phoneKey}
}

// UserItem 用户列表项。
type UserItem struct {
	ID                string `json:"id"`
	Nickname          string `json:"nickname"`
	PhoneMasked       string `json:"phone_masked"`
	Status            int16  `json:"status"`
	Role              int16  `json:"role"`
	DiscoveryEnabled  bool   `json:"discovery_enabled"`
	ProfileCompletion int16  `json:"profile_completion"`
	IsSeed            bool   `json:"is_seed"`
	Gender            int16  `json:"gender"`
	CityCode          string `json:"city_code"`
	CreatedAt         string `json:"created_at"`
	LastLoginAt       string `json:"last_login_at,omitempty"`
}

// UserPage 分页结果。
type UserPage struct {
	Items    []UserItem `json:"items"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

// ListUsers 查询用户。
func (s *Service) ListUsers(ctx context.Context, keyword string, status int16, page, pageSize int) (UserPage, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	rows, err := s.store.Q.AdminListUsers(ctx, sqlcdb.AdminListUsersParams{
		Column1: keyword, Column2: int32(status), Limit: int32(pageSize), Offset: int32(offset),
	})
	if err != nil {
		return UserPage{}, apperr.ErrInternal.Wrap(err)
	}
	total, err := s.store.Q.AdminCountUsers(ctx, sqlcdb.AdminCountUsersParams{Column1: keyword, Column2: int32(status)})
	if err != nil {
		return UserPage{}, apperr.ErrInternal.Wrap(err)
	}
	items := make([]UserItem, 0, len(rows))
	for _, r := range rows {
		item := UserItem{
			ID: r.ID.String(), Nickname: r.Nickname, Status: r.Status, Role: r.Role,
			DiscoveryEnabled: r.DiscoveryEnabled, ProfileCompletion: r.ProfileCompletion,
			IsSeed: r.IsSeed, Gender: r.Gender, CityCode: r.CityCode,
			CreatedAt: r.CreatedAt.Time.Format(time.RFC3339),
		}
		if r.LastLoginAt.Valid {
			item.LastLoginAt = r.LastLoginAt.Time.Format(time.RFC3339)
		}
		// 默认不显示完整手机号，只显示脱敏信息
		if phone, err := s.decryptPhone(ctx, r.ID); err == nil && phone != "" {
			item.PhoneMasked = secure.MaskPhone(phone)
		}
		items = append(items, item)
	}
	return UserPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) decryptPhone(ctx context.Context, userID uuid.UUID) (string, error) {
	row, err := s.store.Q.AdminGetUserDetail(ctx, userID)
	if err != nil {
		return "", err
	}
	if len(row.PhoneEncrypted) == 0 {
		return "", nil
	}
	return secure.DecryptPhone(s.phoneKey, row.PhoneEncrypted)
}

// UserDetail 用户详情。
type UserDetail struct {
	UserItem
	Age            int    `json:"age"`
	EducationLevel int16  `json:"education_level"`
	MaritalStatus  int16  `json:"marital_status"`
	RoleLabel      string `json:"role_label"`
	StatusLabel    string `json:"status_label"`
}

// GetUser 读取用户详情。
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (UserDetail, error) {
	row, err := s.store.Q.AdminGetUserDetail(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserDetail{}, apperr.ErrNotFound.WithMessage("用户不存在")
	}
	if err != nil {
		return UserDetail{}, apperr.ErrInternal.Wrap(err)
	}
	detail := UserDetail{
		UserItem: UserItem{
			ID: row.ID.String(), Nickname: row.Nickname, Status: row.Status, Role: row.Role,
			DiscoveryEnabled: row.DiscoveryEnabled, ProfileCompletion: row.ProfileCompletion,
			IsSeed: row.IsSeed, Gender: row.Gender, CityCode: row.CityCode,
			CreatedAt: row.CreatedAt.Time.Format(time.RFC3339),
		},
		EducationLevel: row.EducationLevel,
		MaritalStatus:  row.MaritalStatus,
		RoleLabel:      enums.Labels(enums.RoleLabels, row.Role),
		StatusLabel:    enums.Labels(enums.UserStatusLabels, row.Status),
	}
	if row.LastLoginAt.Valid {
		detail.LastLoginAt = row.LastLoginAt.Time.Format(time.RFC3339)
	}
	if phone, err := secure.DecryptPhone(s.phoneKey, row.PhoneEncrypted); err == nil {
		detail.PhoneMasked = secure.MaskPhone(phone)
	}
	if row.BirthDate != nil && row.BirthDate.Year() > 1900 {
		detail.Age = int(time.Since(*row.BirthDate).Hours() / 24 / 365)
	}
	return detail, nil
}

// Freeze 冻结用户并写审计日志。
func (s *Service) Freeze(ctx context.Context, actorID, targetID uuid.UUID, reason, ip string) error {
	if targetID == actorID {
		return apperr.Invalid("不能冻结自己")
	}
	if err := s.store.Q.AdminSetUserStatus(ctx, sqlcdb.AdminSetUserStatusParams{
		ID: targetID, Column2: enums.UserStatusSuspended,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	detail, _ := json.Marshal(map[string]string{"reason": reason})
	return s.writeAudit(ctx, actorID, "user.freeze", "user", targetID.String(), detail, ip)
}

// Unfreeze 解冻用户并写审计日志。
func (s *Service) Unfreeze(ctx context.Context, actorID, targetID uuid.UUID, ip string) error {
	if err := s.store.Q.AdminSetUserStatus(ctx, sqlcdb.AdminSetUserStatusParams{
		ID: targetID, Column2: enums.UserStatusActive,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return s.writeAudit(ctx, actorID, "user.unfreeze", "user", targetID.String(), []byte("{}"), ip)
}

func (s *Service) writeAudit(ctx context.Context, actorID uuid.UUID, action, targetType, targetID string, detail []byte, ip string) error {
	var ipHash []byte
	if ip != "" {
		ipHash = secure.TokenHash(ip)
	}
	err := s.store.Q.InsertAuditLog(ctx, sqlcdb.InsertAuditLogParams{
		ActorUserID: &actorID, Action: action, TargetType: targetType, TargetID: targetID,
		Detail: detail, IpHash: ipHash,
	})
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// AuditItem 审计日志项。
type AuditItem struct {
	ID          int64  `json:"id"`
	ActorUserID string `json:"actor_user_id,omitempty"`
	Action      string `json:"action"`
	TargetType  string `json:"target_type"`
	TargetID    string `json:"target_id"`
	Detail      any    `json:"detail"`
	CreatedAt   string `json:"created_at"`
}

// ListAuditLogs 审计日志列表。
func (s *Service) ListAuditLogs(ctx context.Context, limit, offset int32) ([]AuditItem, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.store.Q.ListAuditLogs(ctx, sqlcdb.ListAuditLogsParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]AuditItem, 0, len(rows))
	for _, r := range rows {
		item := AuditItem{
			ID: r.ID, Action: r.Action, TargetType: r.TargetType, TargetID: r.TargetID,
			Detail: convert.JSONMap(r.Detail), CreatedAt: r.CreatedAt.Time.Format(time.RFC3339),
		}
		if r.ActorUserID != nil {
			item.ActorUserID = r.ActorUserID.String()
		}
		out = append(out, item)
	}
	return out, nil
}