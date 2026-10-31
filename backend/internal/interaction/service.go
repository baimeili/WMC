// Package interaction 处理 Like / Pass、Match、结果反馈、Block 与举报。
package interaction

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/events"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/platform/enums"
	"wmc/backend/internal/profile"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/storage"
	"wmc/backend/internal/store"
)

// Service 互动服务。
type Service struct {
	store    *store.Store
	settings *settings.Service
	obj      storage.ObjectStorage
	now      func() time.Time
}

// NewService 构造服务。
func NewService(st *store.Store, set *settings.Service, obj storage.ObjectStorage) *Service {
	return &Service{store: st, settings: set, obj: obj, now: time.Now}
}

// ActResult Like/Pass 结果。
type ActResult struct {
	Matched bool       `json:"matched"`
	MatchID *uuid.UUID `json:"match_id"`
}

// Act 记录 Like / Pass，并在双向 Like 时创建 Match。
// 唯一约束保证同一对用户永远只有一条 Match。
func (s *Service) Act(ctx context.Context, actorID, targetID uuid.UUID, action, source int16) (*ActResult, error) {
	if !s.settings.Bool(ctx, settings.KeyLikesEnabled, true) {
		return nil, apperr.ErrFeatureDisabled.WithMessage("喜欢/跳过功能暂未开放")
	}
	if actorID == targetID {
		return nil, apperr.Invalid("不能对自己操作")
	}
	if action != enums.InteractionLike && action != enums.InteractionPass {
		return nil, apperr.Invalid("action 取值非法")
	}
	if source < 1 || source > 3 {
		source = enums.InteractionSourceCard
	}

	blocked, err := s.store.Q.IsBlocked(ctx, sqlcdb.IsBlockedParams{BlockerUserID: actorID, BlockedUserID: targetID})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if blocked {
		return nil, apperr.ErrForbidden.WithMessage("你们之间已存在屏蔽关系")
	}
	target, err := s.store.Q.GetUserByID(ctx, targetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound.WithMessage("该用户不存在")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if target.Status != enums.UserStatusActive {
		return nil, apperr.ErrForbidden.WithMessage("该用户当前不可互动")
	}

	result := &ActResult{}
	err = s.store.InTx(ctx, func(tx *store.Store) error {
		if _, err := tx.Q.UpsertInteraction(ctx, sqlcdb.UpsertInteractionParams{
			ActorUserID: actorID, TargetUserID: targetID, Action: action, Source: source,
		}); err != nil {
			return err
		}
		eventName := events.NameCandidatePassed
		if action == enums.InteractionLike {
			eventName = events.NameCandidateLiked
		}
		events.Record(ctx, tx, &actorID, nil, eventName, map[string]any{"target": targetID.String()})

		if action != enums.InteractionLike {
			return nil
		}
		if !s.settings.Bool(ctx, settings.KeyMatchesEnabled, true) {
			return nil
		}
		reciprocal, err := tx.Q.ListLikesBetween(ctx, sqlcdb.ListLikesBetweenParams{
			ActorUserID: targetID, TargetUserID: actorID,
		})
		if err != nil {
			return err
		}
		if reciprocal == 0 {
			return nil
		}
		match, err := tx.Q.CreateMatch(ctx, sqlcdb.CreateMatchParams{Column1: actorID, Column2: targetID})
		if errors.Is(err, pgx.ErrNoRows) {
			// 已存在（并发或重复操作），读取已有 Match
			existing, err := tx.Q.GetMatchPair(ctx, sqlcdb.GetMatchPairParams{Column1: actorID, Column2: targetID})
			if err != nil {
				return err
			}
			result.Matched = true
			result.MatchID = &existing.ID
			return nil
		}
		if err != nil {
			return err
		}
		result.Matched = true
		result.MatchID = &match.ID
		events.Record(ctx, tx, &actorID, nil, events.NameMatchCreated, map[string]any{"match_id": match.ID.String()})
		return nil
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	return result, nil
}

// MatchSummary Match 列表项。
type MatchSummary struct {
	ID               string  `json:"id"`
	PeerUserID       string  `json:"peer_user_id"`
	PeerNickname     string  `json:"peer_nickname"`
	PeerAge          int     `json:"peer_age"`
	PeerCityCode     string  `json:"peer_city_code"`
	PeerCityName     string  `json:"peer_city_name"`
	PeerPhotoURL     string  `json:"peer_photo_url"`
	MatchedAt        string  `json:"matched_at"`
	Status           int16   `json:"status"`
	ConfirmedMeeting bool    `json:"confirmed_meeting"`
}

// MatchDetail Match 详情。
type MatchDetail struct {
	MatchSummary
	MyOutcomes   []int16 `json:"my_outcomes"`
	PeerOutcomes []int16 `json:"peer_outcomes"`
	ChatEnabled  bool    `json:"chat_enabled"`
	Disclaimer   string  `json:"disclaimer"`
}

// MatchDisclaimer 说明文案。
const MatchDisclaimer = "你们已经互相感兴趣。匹配结果仅代表双方填写条件的契合，不代表婚姻成功概率。"

// ListMatches 我的 Match 列表。
func (s *Service) ListMatches(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]MatchSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.store.Q.ListMatchesByUser(ctx, sqlcdb.ListMatchesByUserParams{
		UserAID: userID, Limit: limit, Offset: offset,
	})
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	cityNames := s.cityNames(ctx)
	out := make([]MatchSummary, 0, len(rows))
	for _, r := range rows {
		item := MatchSummary{
			ID: r.ID.String(), PeerUserID: r.PeerID.String(), PeerNickname: r.PeerNickname,
			PeerCityCode: r.PeerCityCode, PeerCityName: cityNames[r.PeerCityCode],
			MatchedAt: r.MatchedAt.Time.Format(time.RFC3339),
			Status: r.Status, ConfirmedMeeting: r.ConfirmedMeeting,
		}
		if r.PeerBirthDate != nil {
			item.PeerAge = profile.Age(r.PeerBirthDate, s.now())
		}
		if r.PeerPhotoKey != "" && s.obj != nil {
			if url, err := s.obj.SignedURL(ctx, r.PeerPhotoKey, time.Hour); err == nil {
				item.PeerPhotoURL = url
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// GetMatch 读取 Match 详情（仅限双方）。
func (s *Service) GetMatch(ctx context.Context, userID, matchID uuid.UUID) (*MatchDetail, error) {
	m, err := s.store.Q.GetMatchByID(ctx, matchID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrNotFound.WithMessage("Match 不存在")
	}
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	if m.UserAID != userID && m.UserBID != userID {
		return nil, apperr.ErrForbidden.WithMessage("无权查看该 Match")
	}
	peerID := m.UserBID
	if m.UserAID != userID {
		peerID = m.UserAID
	}
	peer, err := s.store.Q.GetUserByID(ctx, peerID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	peerProfile, err := s.store.Q.GetProfile(ctx, peerID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	outcomes, err := s.store.Q.ListOutcomesByMatch(ctx, matchID)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	detail := &MatchDetail{
		MatchSummary: MatchSummary{
			ID: m.ID.String(), PeerUserID: peer.ID.String(), PeerNickname: peerProfile.Nickname,
			PeerCityCode: peerProfile.CityCode, MatchedAt: m.MatchedAt.Time.Format(time.RFC3339),
			Status: m.Status, ConfirmedMeeting: m.ConfirmedMeeting,
			PeerAge: profile.Age(peerProfile.BirthDate, s.now()),
		},
		ChatEnabled: s.settings.Bool(ctx, settings.KeyChatEnabled, false),
		Disclaimer:  MatchDisclaimer,
	}
	for _, o := range outcomes {
		if o.ReporterUserID == userID {
			detail.MyOutcomes = append(detail.MyOutcomes, o.OutcomeType)
		} else {
			detail.PeerOutcomes = append(detail.PeerOutcomes, o.OutcomeType)
		}
	}
	if detail.MyOutcomes == nil {
		detail.MyOutcomes = []int16{}
	}
	if detail.PeerOutcomes == nil {
		detail.PeerOutcomes = []int16{}
	}
	return detail, nil
}

// CancelMatch 取消 Match。
func (s *Service) CancelMatch(ctx context.Context, userID, matchID uuid.UUID) error {
	m, err := s.store.Q.GetMatchByID(ctx, matchID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.ErrNotFound.WithMessage("Match 不存在")
	}
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if m.UserAID != userID && m.UserBID != userID {
		return apperr.ErrForbidden.WithMessage("无权操作该 Match")
	}
	if err := s.store.Q.CancelMatch(ctx, matchID); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// AddOutcome 反馈真实结果（幂等）。双方都报告线下见面时标记 confirmed_meeting。
func (s *Service) AddOutcome(ctx context.Context, userID, matchID uuid.UUID, outcomeType int16, occurredAt *time.Time, note string) error {
	if outcomeType < 1 || outcomeType > 8 {
		return apperr.Invalid("反馈类型非法")
	}
	m, err := s.store.Q.GetMatchByID(ctx, matchID)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.ErrNotFound.WithMessage("Match 不存在")
	}
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if m.UserAID != userID && m.UserBID != userID {
		return apperr.ErrForbidden.WithMessage("无权反馈该 Match")
	}
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	err = s.store.InTx(ctx, func(tx *store.Store) error {
		if _, err := tx.Q.UpsertOutcome(ctx, sqlcdb.UpsertOutcomeParams{
			MatchID: matchID, ReporterUserID: userID, OutcomeType: outcomeType,
			Column4: convert.ToTimePtr(occurredAt), Note: notePtr,
		}); err != nil {
			return err
		}
		// 双方都反馈线下见面 -> 确认见面
		reporters, err := tx.Q.CountOutcomeReporters(ctx, sqlcdb.CountOutcomeReportersParams{
			MatchID: matchID, OutcomeType: enums.OutcomeMetOffline,
		})
		if err != nil {
			return err
		}
		if reporters >= 2 {
			if err := tx.Q.SetConfirmedMeeting(ctx, matchID); err != nil {
				return err
			}
		}
		eventName := events.NameOutcomeContact
		switch outcomeType {
		case enums.OutcomeMetOffline:
			eventName = events.NameOutcomeMeeting
		case enums.OutcomeRelationship:
			eventName = events.NameOutcomeRelationship
		}
		events.Record(ctx, tx, &userID, nil, eventName, map[string]any{"match_id": matchID.String()})
		return nil
	})
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// Block 屏蔽用户。
func (s *Service) Block(ctx context.Context, userID, targetID uuid.UUID) error {
	if userID == targetID {
		return apperr.Invalid("不能屏蔽自己")
	}
	if err := s.store.Q.UpsertBlock(ctx, sqlcdb.UpsertBlockParams{
		BlockerUserID: userID, BlockedUserID: targetID,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// Unblock 取消屏蔽。
func (s *Service) Unblock(ctx context.Context, userID, targetID uuid.UUID) error {
	if err := s.store.Q.DeleteBlock(ctx, sqlcdb.DeleteBlockParams{
		BlockerUserID: userID, BlockedUserID: targetID,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}

// Report 举报用户。
func (s *Service) cityNames(ctx context.Context) map[string]string {
	regions, err := s.store.Q.ListRegions(ctx)
	if err != nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(regions))
	for _, r := range regions {
		out[r.Code] = r.Name
	}
	return out
}

// Report 举报用户。
func (s *Service) Report(ctx context.Context, reporterID, targetID uuid.UUID, reason int16, description string, evidence []string) error {
	if reporterID == targetID {
		return apperr.Invalid("不能举报自己")
	}
	if reason < 1 || reason > 8 {
		return apperr.Invalid("举报原因非法")
	}
	evidenceJSON := convert.MustJSON(evidence)
	if _, err := s.store.Q.CreateReport(ctx, sqlcdb.CreateReportParams{
		ReporterUserID: reporterID, ReportedUserID: targetID, Reason: reason,
		Description: description, EvidenceKeys: evidenceJSON,
	}); err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	return nil
}