package matching

import (
	"math"
	"sort"

	"github.com/google/uuid"
)

// Engine 匹配引擎。
type Engine struct {
	weights Weights
}

// NewEngine 构造引擎。
func NewEngine(w Weights) *Engine { return &Engine{weights: w.Normalize()} }

// Weights 返回归一化后的权重。
func (e *Engine) Weights() Weights { return e.weights }

// Dimension 与择偶条件对应的展示维度。
type Dimension struct {
	Key   string
	Label string
}

// PreferenceDimensions 软匹配维度（顺序固定，便于展示）。
var PreferenceDimensions = []Dimension{
	{"age", "年龄"},
	{"height", "身高"},
	{"city", "城市"},
	{"education", "学历"},
	{"income", "收入"},
	{"marital_status", "婚姻状态"},
	{"children", "子女情况"},
	{"wants_children", "生育意愿"},
	{"smoking", "吸烟习惯"},
	{"drinking", "饮酒习惯"},
	{"marriage_intent", "结婚意愿"},
}

// CheckHard 检查 candidate 是否满足 owner 的全部硬条件。
// direction 用于标记是「我要求对方」还是「对方要求我」。
func CheckHard(candidate Profile, owner Preference, direction string) []Conflict {
	out := make([]Conflict, 0, 4)
	add := func(dim, label, detail string) {
		out = append(out, Conflict{Dimension: dim, Label: label, Detail: detail, Direction: direction})
	}

	// 目标性别始终视为硬条件
	if len(owner.TargetGenders) > 0 && !containsInt16(owner.TargetGenders, candidate.Gender) {
		add("gender", "目标性别", "性别不在对方期望范围内")
	}
	if owner.AgeHard && (candidate.Age < int(owner.MinAge) || candidate.Age > int(owner.MaxAge)) {
		add("age", "年龄", "年龄不在对方要求区间内")
	}
	if owner.HeightHard {
		if candidate.HeightCM == nil {
			add("height", "身高", "对方要求身高条件，但未填写身高")
		} else if int(*candidate.HeightCM) < int(owner.MinHeightCM) || int(*candidate.HeightCM) > int(owner.MaxHeightCM) {
			add("height", "身高", "身高不在对方要求区间内")
		}
	}
	if owner.CityHard && len(owner.AllowedCityCodes) > 0 && !containsString(owner.AllowedCityCodes, candidate.CityCode) {
		add("city", "城市", "所在城市不在对方期望范围")
	}
	if owner.EducationHard && candidate.EducationLevel < owner.MinEducationLevel {
		add("education", "学历", "学历低于对方要求")
	}
	if owner.IncomeHard && !incomeMeets(candidate.AnnualIncomeBand, owner.MinIncomeBand) {
		add("income", "收入", "收入区间低于对方要求，或未填写")
	}
	if owner.MaritalHard && len(owner.AllowedMarital) > 0 && !containsInt16(owner.AllowedMarital, candidate.MaritalStatus) {
		add("marital_status", "婚姻状态", "婚姻状态不在对方期望范围")
	}
	if owner.ChildrenHard && len(owner.AllowedChildren) > 0 && !containsInt16(owner.AllowedChildren, ChildrenState(candidate.ChildrenCount)) {
		add("children", "子女情况", "子女情况不在对方期望范围")
	}
	if owner.WantsChildrenHard && len(owner.AllowedWantsChildren) > 0 && !containsInt16(owner.AllowedWantsChildren, candidate.WantsChildren) {
		add("wants_children", "生育意愿", "生育意愿与对方硬条件冲突")
	}
	if owner.SmokingHard && len(owner.AllowedSmoking) > 0 && !containsInt16(owner.AllowedSmoking, candidate.SmokingStatus) {
		add("smoking", "吸烟习惯", "吸烟习惯不在对方期望范围")
	}
	if owner.DrinkingHard && len(owner.AllowedDrinking) > 0 && !containsInt16(owner.AllowedDrinking, candidate.DrinkingStatus) {
		add("drinking", "饮酒习惯", "饮酒习惯不在对方期望范围")
	}
	if owner.MarriageIntentHard && owner.MinMarriageIntent > 0 && IntentRank(candidate.MarriageIntent) < IntentRank(owner.MinMarriageIntent) {
		add("marriage_intent", "结婚意愿", "结婚意愿低于对方要求")
	}
	return out
}

// Eligible 双向硬筛：A 满足 B 的全部硬条件 AND B 满足 A 的全部硬条件。
// 任何一项不满足即返回 false，软条件不得补偿。
func (e *Engine) Eligible(a Profile, aPref Preference, aRules []ValueRule, b Profile, bPref Preference, bRules []ValueRule) (bool, []Conflict) {
	conflicts := CheckHard(b, aPref, "b_to_a")   // B 是否满足 A 的要求
	conflicts = append(conflicts, CheckHard(a, bPref, "a_to_b")...) // A 是否满足 B 的要求
	conflicts = append(conflicts, checkValueHard(aRules, answersOf(bRules), "b_to_a")...)
	conflicts = append(conflicts, checkValueHard(bRules, answersOf(aRules), "a_to_b")...)
	return len(conflicts) == 0, conflicts
}

// checkValueHard 检查 owner 的硬价值观规则是否被对方答案满足。
func checkValueHard(ownerRules []ValueRule, peerAnswers map[uuid.UUID][]string, direction string) []Conflict {
	out := make([]Conflict, 0, 2)
	for _, r := range ownerRules {
		if !r.IsHard || r.Importance == 0 || len(r.AcceptedAnswers) == 0 {
			continue
		}
		peerAns, ok := peerAnswers[r.QuestionID]
		if !ok || len(peerAns) == 0 {
			continue // 对方未填写 -> 不计入（禁止把未填写当作 0 分或冲突）
		}
		if !anyAccepted(peerAns, r.AcceptedAnswers) {
			out = append(out, Conflict{
				Dimension: "values:" + r.QuestionCode,
				Label:     "价值观",
				Detail:    "价值观硬条件未满足：" + categoryOf(r),
				Direction: direction,
			})
		}
	}
	return out
}

// Input 评估输入。
type Input struct {
	Me             Profile
	MyPreference   Preference
	MyRules        []ValueRule
	Candidate      Profile
	PeerPreference Preference
	PeerRules      []ValueRule
}

// Evaluate 计算完整匹配结果（硬条件、软分、解释）。
func (e *Engine) Evaluate(in Input) Result {
	var res Result

	conflicts := CheckHard(in.Candidate, in.MyPreference, "b_to_a")
	conflicts = append(conflicts, CheckHard(in.Me, in.PeerPreference, "a_to_b")...)

	myAnswers := answersOf(in.MyRules)
	peerAnswers := answersOf(in.PeerRules)
	conflicts = append(conflicts, checkValueHard(in.MyRules, peerAnswers, "b_to_a")...)
	conflicts = append(conflicts, checkValueHard(in.PeerRules, myAnswers, "a_to_b")...)

	res.Conflicts = conflicts
	res.Eligible = len(conflicts) == 0

	res.PreferenceScore, res.PreferenceDetail = scorePreference(in.Candidate, in.MyPreference)
	res.CompatibilityScore, res.CategoryScores, res.ValidQuestions = scoreCompatibility(
		myAnswers, peerAnswers, in.MyRules, in.PeerRules)
	res.Insufficient = res.ValidQuestions < MinValidQuestions

	res.OverallScore = e.weights.Preference*res.PreferenceScore + e.weights.Compatibility*res.CompatibilityScore
	res.Reasons = buildReasons(in, res)
	return res
}

func scorePreference(candidate Profile, pref Preference) (float64, []DimensionScore) {
	var (
		totalWeight float64
		totalScore  float64
		detail      = make([]DimensionScore, 0, len(PreferenceDimensions))
	)
	consider := func(key, label string, hard bool, ok bool, applicable bool) {
		if hard || !applicable {
			return
		}
		w := pref.Weights[key]
		if w <= 0 {
			w = 1
		}
		score := 0.0
		if ok {
			score = 1
		}
		detail = append(detail, DimensionScore{Key: key, Label: label, Score: score})
		totalWeight += w
		totalScore += w * score
	}

	consider("age", "年龄", pref.AgeHard,
		candidate.Age >= int(pref.MinAge) && candidate.Age <= int(pref.MaxAge),
		pref.MinAge > 0 || pref.MaxAge > 0)
	consider("height", "身高", pref.HeightHard,
		candidate.HeightCM != nil && int(*candidate.HeightCM) >= int(pref.MinHeightCM) && int(*candidate.HeightCM) <= int(pref.MaxHeightCM),
		pref.MinHeightCM > 0 || (pref.MaxHeightCM > 0 && pref.MaxHeightCM < 250))
	consider("city", "城市", pref.CityHard,
		len(pref.AllowedCityCodes) == 0 || containsString(pref.AllowedCityCodes, candidate.CityCode),
		len(pref.AllowedCityCodes) > 0)
	consider("education", "学历", pref.EducationHard,
		candidate.EducationLevel >= pref.MinEducationLevel, pref.MinEducationLevel > 0)
	consider("income", "收入", pref.IncomeHard,
		incomeMeets(candidate.AnnualIncomeBand, pref.MinIncomeBand), pref.MinIncomeBand > 0)
	consider("marital_status", "婚姻状态", pref.MaritalHard,
		len(pref.AllowedMarital) == 0 || containsInt16(pref.AllowedMarital, candidate.MaritalStatus),
		len(pref.AllowedMarital) > 0)
	consider("children", "子女情况", pref.ChildrenHard,
		len(pref.AllowedChildren) == 0 || containsInt16(pref.AllowedChildren, ChildrenState(candidate.ChildrenCount)),
		len(pref.AllowedChildren) > 0)
	consider("wants_children", "生育意愿", pref.WantsChildrenHard,
		len(pref.AllowedWantsChildren) == 0 || containsInt16(pref.AllowedWantsChildren, candidate.WantsChildren),
		len(pref.AllowedWantsChildren) > 0)
	consider("smoking", "吸烟习惯", pref.SmokingHard,
		len(pref.AllowedSmoking) == 0 || containsInt16(pref.AllowedSmoking, candidate.SmokingStatus),
		len(pref.AllowedSmoking) > 0)
	consider("drinking", "饮酒习惯", pref.DrinkingHard,
		len(pref.AllowedDrinking) == 0 || containsInt16(pref.AllowedDrinking, candidate.DrinkingStatus),
		len(pref.AllowedDrinking) > 0)
	consider("marriage_intent", "结婚意愿", pref.MarriageIntentHard,
		IntentRank(candidate.MarriageIntent) >= IntentRank(pref.MinMarriageIntent), pref.MinMarriageIntent > 0)

	if totalWeight == 0 {
		return 1, detail
	}
	return totalScore / totalWeight, detail
}

// scoreCompatibility 计算关系观念兼容度：只有双方都填写的问题才计入分母。
func scoreCompatibility(
	myAnswers, peerAnswers map[uuid.UUID][]string,
	myRules, peerRules []ValueRule,
) (float64, []DimensionScore, int) {
	peerByID := make(map[uuid.UUID]ValueRule, len(peerRules))
	for _, r := range peerRules {
		peerByID[r.QuestionID] = r
	}

	var (
		totalWeight float64
		totalScore  float64
		valid       int
	)
	type catAgg struct{ sum, weight float64 }
	catScore := map[string]catAgg{}

	for _, mine := range myRules {
		if mine.Importance == 0 || len(mine.AcceptedAnswers) == 0 {
			continue
		}
		peerAns, ok := peerAnswers[mine.QuestionID]
		if !ok || len(peerAns) == 0 {
			continue // 对方未填写 -> 不计入分母
		}
		peerRule, hasPeerRule := peerByID[mine.QuestionID]

		// 双向打分：我接受对方答案 + 对方接受我的答案
		score := 0.0
		if anyAccepted(peerAns, mine.AcceptedAnswers) {
			score += 0.5
		}
		if !hasPeerRule || peerRule.Importance == 0 || len(peerRule.AcceptedAnswers) == 0 {
			score += 0.5 // 对方未设接受范围，视为中性
		} else if myAns, ok := myAnswers[mine.QuestionID]; ok && anyAccepted(myAns, peerRule.AcceptedAnswers) {
			score += 0.5
		}

		weight := float64(mine.Importance)
		if hasPeerRule && peerRule.Importance > 0 {
			weight = (float64(mine.Importance) + float64(peerRule.Importance)) / 2
		}
		if weight <= 0 {
			weight = 1
		}
		totalWeight += weight
		totalScore += weight * score
		valid++

		cat := categoryOf(mine)
		agg := catScore[cat]
		agg.sum += weight * score
		agg.weight += weight
		catScore[cat] = agg
	}

	categories := make([]DimensionScore, 0, len(catScore))
	for name, agg := range catScore {
		score := 0.0
		if agg.weight > 0 {
			score = agg.sum / agg.weight
		}
		categories = append(categories, DimensionScore{Key: name, Label: name, Score: score})
	}
	sort.Slice(categories, func(i, j int) bool { return categories[i].Label < categories[j].Label })

	if totalWeight == 0 {
		return 0, categories, valid
	}
	return totalScore / totalWeight, categories, valid
}

func buildReasons(in Input, res Result) []string {
	reasons := make([]string, 0, 8)
	if res.Eligible {
		reasons = append(reasons, "双方硬条件全部满足")
	}
	if len(in.MyPreference.TargetGenders) > 0 && containsInt16(in.MyPreference.TargetGenders, in.Candidate.Gender) {
		reasons = append(reasons, "性别符合你的期望")
	}
	if in.Candidate.Age >= int(in.MyPreference.MinAge) && in.Candidate.Age <= int(in.MyPreference.MaxAge) {
		reasons = append(reasons, "年龄符合你的要求")
	}
	if in.Me.Age >= int(in.PeerPreference.MinAge) && in.Me.Age <= int(in.PeerPreference.MaxAge) {
		reasons = append(reasons, "年龄符合对方要求")
	}
	if in.Candidate.WantsChildren == in.Me.WantsChildren && in.Candidate.WantsChildren > 0 {
		reasons = append(reasons, "婚育意愿一致")
	}
	if in.Candidate.CityCode != "" && in.Candidate.CityCode == in.Me.CityCode {
		reasons = append(reasons, "同城")
	}
	if in.Candidate.EducationLevel >= in.MyPreference.MinEducationLevel && in.Me.EducationLevel >= in.PeerPreference.MinEducationLevel {
		reasons = append(reasons, "双方学历条件满足")
	}
	return reasons
}

func categoryOf(r ValueRule) string {
	if r.Category != "" {
		return r.Category
	}
	if r.QuestionCode != "" {
		return r.QuestionCode
	}
	return "其他"
}

func answersOf(rules []ValueRule) map[uuid.UUID][]string {
	out := make(map[uuid.UUID][]string, len(rules))
	for _, r := range rules {
		if len(r.OwnAnswer) > 0 {
			out[r.QuestionID] = r.OwnAnswer
		}
	}
	return out
}

func anyAccepted(answer, accepted []string) bool {
	for _, a := range answer {
		for _, b := range accepted {
			if a == b {
				return true
			}
		}
	}
	return false
}

func containsInt16(list []int16, v int16) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func containsString(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// ChildrenState 由子女数量推导状态（0 无 1 有）。
func ChildrenState(count int16) int16 {
	if count > 0 {
		return 1
	}
	return 0
}

// IntentRank 结婚意愿等级（越大越强），与 enums 保持一致。
func IntentRank(code int16) int16 {
	switch code {
	case 1:
		return 5
	case 2:
		return 4
	case 3:
		return 3
	case 4:
		return 2
	case 5:
		return 1
	}
	return 0
}

// incomeMeets 收入区间比较：8=不愿透露 不视为高收入。
func incomeMeets(actual, required int16) bool {
	if required <= 0 {
		return true
	}
	if actual < 1 || actual > 7 {
		return false
	}
	return actual >= required
}

// Round 保留小数。
func Round(v float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(v*pow) / pow
}
