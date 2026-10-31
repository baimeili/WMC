package admin

import (
	"context"
	"math"
	"time"

	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/convert"
	"wmc/backend/internal/platform/enums"
	"wmc/backend/internal/settings"
)

// DistributionItem 分布项。
type DistributionItem struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

// FunnelItem 经营漏斗项。
type FunnelItem struct {
	Key        string  `json:"key"`
	Label      string  `json:"label"`
	Count      int64   `json:"count"`
	Conversion float64 `json:"conversion"`
}

// Dashboard 运营看板。
type Dashboard struct {
	TotalUsers        int64              `json:"total_users"`
	NewUsers7d        int64              `json:"new_users_7d"`
	DAU               int64              `json:"dau"`
	WAU               int64              `json:"wau"`
	MAU               int64              `json:"mau"`
	ProfileCompleted  int64              `json:"profile_completed_users"`
	MarketUsers       int64              `json:"market_users"`
	Gender            []DistributionItem `json:"gender_distribution"`
	Cities            []DistributionItem `json:"city_distribution"`
	Ages              []DistributionItem `json:"age_distribution"`
	InteractionsLike  int64              `json:"interactions_like"`
	MatchesTotal      int64              `json:"matches_total"`
	MatchRate         float64            `json:"match_rate"`
	OfflineReports    int64              `json:"offline_meeting_reports"`
	ConfirmedMeetings int64              `json:"confirmed_meetings"`
	RelationshipCount int64              `json:"relationship_reports"`
	NorthStar         float64            `json:"north_star"`
	Funnel            []FunnelItem       `json:"funnel"`
	Note              string             `json:"note"`
}

// DashboardNote 指标口径说明。
const DashboardNote = "North Star：每 1000 个月活用户产生的双方确认线下见面数。漏斗转化率按相对上一步计算；" +
	"上线初期历史用户可能缺少前置事件，个别步骤转化率可能大于 100%，属于数据完整度问题，不代表产品异常。"

// GetDashboard 汇总运营指标。
func (s *Service) GetDashboard(ctx context.Context) (*Dashboard, error) {
	now := time.Now()
	minCompletion := int16(s.settings.Int(ctx, settings.KeyMinProfileCompletion, 80))
	d := &Dashboard{Note: DashboardNote}

	total, err := s.store.Q.CountUsers(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	d.TotalUsers = total
	if v, err := s.store.Q.StatsCountRegisteredSince(ctx, convert.ToTime(now.AddDate(0, 0, -7))); err == nil {
		d.NewUsers7d = v
	}
	if v, err := s.store.Q.StatsActiveUsers(ctx, convert.ToTime(now.AddDate(0, 0, -1))); err == nil {
		d.DAU = v
	}
	if v, err := s.store.Q.StatsActiveUsers(ctx, convert.ToTime(now.AddDate(0, 0, -7))); err == nil {
		d.WAU = v
	}
	if v, err := s.store.Q.StatsActiveUsers(ctx, convert.ToTime(now.AddDate(0, 0, -30))); err == nil {
		d.MAU = v
	}
	if v, err := s.store.Q.StatsCountCompletedProfiles(ctx, minCompletion); err == nil {
		d.ProfileCompleted = v
	}
	if v, err := s.store.Q.StatsCountMarketUsers(ctx, minCompletion); err == nil {
		d.MarketUsers = v
	}

	if rows, err := s.store.Q.AdminGenderDistribution(ctx); err == nil {
		for _, r := range rows {
			d.Gender = append(d.Gender, DistributionItem{
				Label: enums.Labels(enums.GenderLabels, r.Gender), Value: r.Count,
			})
		}
	}
	cityNames := s.cityNames(ctx)
	if rows, err := s.store.Q.AdminCityDistribution(ctx); err == nil {
		for _, r := range rows {
			label := cityNames[r.CityCode]
			if label == "" {
				label = r.CityCode
			}
			d.Cities = append(d.Cities, DistributionItem{Label: label, Value: r.Count})
		}
	}
	if rows, err := s.store.Q.AdminAgeDistribution(ctx); err == nil {
		for _, r := range rows {
			d.Ages = append(d.Ages, DistributionItem{Label: ageBucket(r.Bucket), Value: r.Count})
		}
	}

	if v, err := s.store.Q.StatsCountInteractions(ctx, enums.InteractionLike); err == nil {
		d.InteractionsLike = v
	}
	if v, err := s.store.Q.StatsCountMatches(ctx); err == nil {
		d.MatchesTotal = v
	}
	if d.MarketUsers > 0 {
		d.MatchRate = math.Round(float64(d.MatchesTotal)/float64(d.MarketUsers)*10000) / 10000
	}
	if v, err := s.store.Q.StatsCountOutcomes(ctx, enums.OutcomeMetOffline); err == nil {
		d.OfflineReports = v
	}
	if v, err := s.store.Q.StatsConfirmedMeetings(ctx); err == nil {
		d.ConfirmedMeetings = v
	}
	if v, err := s.store.Q.StatsCountOutcomes(ctx, enums.OutcomeRelationship); err == nil {
		d.RelationshipCount = v
	}
	if d.MAU > 0 {
		d.NorthStar = math.Round(float64(d.ConfirmedMeetings)*1000/float64(d.MAU)*100) / 100
	}

	// 核心经营漏斗
	steps := []struct {
		Key    string
		Label  string
		Loader func() int64
	}{
		{"calculator_started", "访问计算器", func() int64 { v, _ := s.store.Q.StatsCountEventUsers(ctx, "calculator_started"); return v }},
		{"calculator_completed", "完成计算器", func() int64 { v, _ := s.store.Q.StatsCountCalculatorRuns(ctx, convert.ToTime(now.AddDate(-1, 0, 0))); return v }},
		{"registered", "注册", func() int64 { return d.TotalUsers }},
		{"profile_completed", "完成资料", func() int64 { return d.ProfileCompleted }},
		{"market_joined", "加入婚恋池", func() int64 { return d.MarketUsers }},
		{"discovery_viewed", "浏览候选池", func() int64 { v, _ := s.store.Q.StatsCountEventUsers(ctx, "discovery_viewed"); return v }},
		{"liked", "Like", func() int64 { return d.InteractionsLike }},
		{"matched", "Match", func() int64 { return d.MatchesTotal }},
		{"met_offline", "线下见面", func() int64 { return d.OfflineReports }},
		{"relationship", "建立关系", func() int64 { return d.RelationshipCount }},
	}
	// 转化率按「相对上一步」计算；首步为 100%。
	prev := int64(0)
	for i, st := range steps {
		count := st.Loader()
		conv := 100.0
		if i > 0 {
			conv = 0
			if prev > 0 {
				conv = math.Round(float64(count)/float64(prev)*10000) / 100
			}
		}
		d.Funnel = append(d.Funnel, FunnelItem{Key: st.Key, Label: st.Label, Count: count, Conversion: conv})
		prev = count
	}
	return d, nil
}

func ageBucket(bucket int32) string {
	// width_bucket 以 18 岁为起点、13 个桶、区间到 60 岁
	age := 18 + (int(bucket)-1)*3
	if bucket <= 0 || age > 60 {
		return "其他"
	}
	return timeLabel(age)
}

func timeLabel(age int) string {
	return itoa(age) + "-" + itoa(age+2) + "岁"
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [8]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}

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
