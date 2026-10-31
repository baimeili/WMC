package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/platform/enums"
)

// ReportItem 举报列表项。
type ReportItem struct {
	ID               string `json:"id"`
	ReporterUserID   string `json:"reporter_user_id"`
	ReportedUserID   string `json:"reported_user_id"`
	ReportedNickname string `json:"reported_nickname"`
	Reason           int16  `json:"reason"`
	ReasonLabel      string `json:"reason_label"`
	Description      string `json:"description"`
	Status           int16  `json:"status"`
	StatusLabel      string `json:"status_label"`
	CreatedAt        string `json:"created_at"`
}

// ListReports 举报列表。
func (s *Service) ListReports(ctx context.Context, status int16, limit, offset int32) ([]ReportItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := s.store.Q.AdminListReports(ctx, sqlcdb.AdminListReportsParams{
		Column1: int32(status), Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]ReportItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, ReportItem{
			ID: r.ID.String(), ReporterUserID: r.ReporterUserID.String(), ReportedUserID: r.ReportedUserID.String(),
			ReportedNickname: r.ReportedNickname, Reason: r.Reason,
			ReasonLabel: enums.Labels(enums.ReportReasonLabels, r.Reason),
			Description: r.Description, Status: r.Status,
			StatusLabel: enums.Labels(enums.ReportStatusLabels, r.Status),
			CreatedAt:   r.CreatedAt.Time.Format(time.RFC3339),
		})
	}
	return out, nil
}

// ResolveReport 处理举报并写审计日志。
func (s *Service) ResolveReport(ctx context.Context, actorID, reportID uuid.UUID, status int16, note, ip string) error {
	if status < enums.ReportStatusReviewing || status > enums.ReportStatusRejected {
		return apperr.Invalid("处理状态非法")
	}
	if err := s.store.Q.AdminResolveReport(ctx, sqlcdb.AdminResolveReportParams{
		ID: reportID, Column2: status, ReviewerID: &actorID, ResolutionNote: note,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	detail, _ := json.Marshal(map[string]any{"status": status, "note": note})
	return s.writeAudit(ctx, actorID, "report.resolve", "report", reportID.String(), detail, ip)
}

// PhotoItem 待审核照片。
type PhotoItem struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Nickname  string `json:"nickname"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

// ListPendingPhotos 待审核照片列表。
func (s *Service) ListPendingPhotos(ctx context.Context, limit, offset int32) ([]PhotoItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.store.Q.AdminListPendingPhotos(ctx, sqlcdb.AdminListPendingPhotosParams{
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]PhotoItem, 0, len(rows))
	for _, r := range rows {
		item := PhotoItem{
			ID: r.ID.String(), UserID: r.UserID.String(), Nickname: r.Nickname,
			CreatedAt: r.CreatedAt.Time.Format(time.RFC3339),
		}
		if s.obj != nil {
			if url, err := s.obj.SignedURL(ctx, r.ObjectKey, time.Hour); err == nil {
				item.URL = url
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// ModeratePhoto 审核照片（2 通过 / 3 拒绝）。
func (s *Service) ModeratePhoto(ctx context.Context, actorID, photoID uuid.UUID, status int16, reason, ip string) error {
	if status != enums.ModerationApproved && status != enums.ModerationRejected {
		return apperr.Invalid("审核状态非法")
	}
	photo, err := s.store.Q.GetPhotoByID(ctx, photoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.ErrNotFound.WithMessage("照片不存在")
	}
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if err := s.store.Q.ModeratePhoto(ctx, sqlcdb.ModeratePhotoParams{ID: photoID, ModerationStatus: status}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	detail, _ := json.Marshal(map[string]any{"status": status, "reason": reason, "user_id": photo.UserID.String()})
	return s.writeAudit(ctx, actorID, "photo.moderate", "photo", photoID.String(), detail, ip)
}

// SettingsBundle 系统参数集合。
type SettingsBundle struct {
	Features map[string]bool `json:"features"`
	Settings []SettingItem   `json:"settings"`
}

// SettingItem 单条参数。
type SettingItem struct {
	Key         string `json:"key"`
	Value       any    `json:"value"`
	Description string `json:"description"`
	UpdatedAt   string `json:"updated_at"`
}

// GetSettings 读取全部系统参数。
func (s *Service) GetSettings(ctx context.Context) (*SettingsBundle, error) {
	rows, err := s.store.Q.ListSettingsRows(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	bundle := &SettingsBundle{Features: s.settings.Features(ctx), Settings: make([]SettingItem, 0, len(rows))}
	for _, r := range rows {
		item := SettingItem{Key: r.Key, Description: r.Description}
		var v any
		if err := json.Unmarshal(r.Value, &v); err == nil {
			item.Value = v
		} else {
			item.Value = string(r.Value)
		}
		if r.UpdatedAt.Valid {
			item.UpdatedAt = r.UpdatedAt.Time.Format(time.RFC3339)
		}
		bundle.Settings = append(bundle.Settings, item)
	}
	return bundle, nil
}

// UpdateSetting 修改系统参数并写审计日志。
func (s *Service) UpdateSetting(ctx context.Context, actorID uuid.UUID, key string, value json.RawMessage, ip string) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return apperr.Invalid("参数 key 不能为空")
	}
	if len(value) == 0 {
		return apperr.Invalid("参数 value 不能为空")
	}
	if !json.Valid(value) {
		return apperr.Invalid("参数 value 必须是合法 JSON")
	}
	if _, err := s.store.Q.UpsertSetting(ctx, sqlcdb.UpsertSettingParams{
		Key: key, Value: value, UpdatedBy: &actorID,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	s.settings.Invalidate()
	detail, _ := json.Marshal(map[string]any{"key": key, "value": json.RawMessage(value)})
	return s.writeAudit(ctx, actorID, "setting.update", "setting", key, detail, ip)
}

// UserReports 查询某个用户被举报记录。
func (s *Service) UserReports(ctx context.Context, userID uuid.UUID) ([]ReportItem, error) {
	rows, err := s.store.Q.AdminListUserReports(ctx, userID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]ReportItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, ReportItem{
			ID: r.ID.String(), ReporterUserID: r.ReporterUserID.String(), ReportedUserID: r.ReportedUserID.String(),
			Reason: r.Reason, ReasonLabel: enums.Labels(enums.ReportReasonLabels, r.Reason),
			Description: r.Description, Status: r.Status,
			StatusLabel: enums.Labels(enums.ReportStatusLabels, r.Status),
			CreatedAt:   convert.Time(r.CreatedAt).Format(time.RFC3339),
		})
	}
	return out, nil
}
