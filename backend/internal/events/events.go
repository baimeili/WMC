// Package events 记录产品事件，用于运营漏斗与北极星指标。
package events

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/store"
)

// 事件名称常量（与规格 §47 对齐）。
const (
	NameAppOpen               = "app_open"
	NameCalculatorStarted     = "calculator_started"
	NameCalculatorCompleted   = "calculator_completed"
	NameRegistrationStarted   = "registration_started"
	NameRegistrationCompleted = "registration_completed"
	NameProfileCompleted      = "profile_completed"
	NameMarketJoined          = "market_joined"
	NameDiscoveryViewed       = "discovery_viewed"
	NameCandidateImpression   = "candidate_impression"
	NameCandidateLiked        = "candidate_liked"
	NameCandidatePassed       = "candidate_passed"
	NameMatchCreated          = "match_created"
	NameOutcomeContact        = "outcome_contact_exchanged"
	NameOutcomeMeeting        = "outcome_offline_meeting"
	NameOutcomeRelationship   = "outcome_relationship"
)

// Record 写入一条产品事件（失败不影响主流程）。
func Record(ctx context.Context, st *store.Store, userID *uuid.UUID, anonymousID *string, name string, props map[string]any) {
	if st == nil || name == "" {
		return
	}
	payload := []byte("{}")
	if len(props) > 0 {
		if b, err := json.Marshal(props); err == nil {
			payload = b
		}
	}
	_ = st.Q.InsertEvent(ctx, sqlcdb.InsertEventParams{
		UserID: userID, AnonymousID: anonymousID, EventName: name, Properties: payload,
	})
}
