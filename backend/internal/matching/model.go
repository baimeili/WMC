// Package matching 实现双向婚恋匹配引擎。
// 引擎为纯函数：不访问数据库，可完整单元测试。
// 规则：硬条件双向全部满足才 Eligible，绝不允许软条件补偿。
package matching

import (
	"fmt"

	"github.com/google/uuid"
)

// Profile 参与匹配的用户画像（由资料表推导）。
type Profile struct {
	UserID             uuid.UUID
	Nickname           string
	Gender             int16
	Age                int
	HeightCM           *int16
	CityCode           string
	EducationLevel     int16
	OccupationCategory int16
	AnnualIncomeBand   int16
	MaritalStatus      int16
	ChildrenCount      int16
	WantsChildren      int16
	MarriageIntent     int16
	SmokingStatus      int16
	DrinkingStatus     int16
	Bio                string
}

// Preference 择偶条件。Hard=true 表示硬条件（不满足直接淘汰）。
type Preference struct {
	TargetGenders       []int16
	MinAge              int16
	MaxAge              int16
	AgeHard             bool
	MinHeightCM         int16
	MaxHeightCM         int16
	HeightHard          bool
	AllowedCityCodes    []string
	CityHard            bool
	MinEducationLevel   int16
	EducationHard       bool
	MinIncomeBand       int16
	IncomeHard          bool
	AllowedMarital      []int16
	MaritalHard         bool
	AllowedChildren     []int16
	ChildrenHard        bool
	AllowedWantsChildren []int16
	WantsChildrenHard   bool
	AllowedSmoking      []int16
	SmokingHard         bool
	AllowedDrinking     []int16
	DrinkingHard        bool
	MinMarriageIntent   int16
	MarriageIntentHard  bool
	Weights             map[string]float64
}

// ValueRule 价值观规则。
type ValueRule struct {
	QuestionID      uuid.UUID
	QuestionCode    string
	Category        string
	OwnAnswer       []string
	AcceptedAnswers []string
	Importance      int16 // 0 不在意 ~ 5 非常重要
	IsHard          bool
}

// Conflict 硬条件冲突。
type Conflict struct {
	Dimension string `json:"dimension"`
	Label     string `json:"label"`
	Detail    string `json:"detail"`
	Direction string `json:"direction"` // a_to_b / b_to_a
}

// DimensionScore 单维度得分。
type DimensionScore struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Score float64 `json:"score"`
}

// Result 匹配结果。
type Result struct {
	Eligible          bool             `json:"eligible"`
	Conflicts         []Conflict       `json:"conflicts"`
	PreferenceScore   float64          `json:"preference_score"`
	CompatibilityScore float64         `json:"compatibility_score"`
	OverallScore      float64          `json:"overall_score"`
	PreferenceDetail  []DimensionScore `json:"preference_detail"`
	CategoryScores    []DimensionScore `json:"category_scores"`
	ValidQuestions    int              `json:"valid_questions"`
	Insufficient      bool             `json:"compatibility_insufficient"`
	Reasons           []string         `json:"reasons"`
}

// Weights 综合匹配度权重。
type Weights struct {
	Preference    float64
	Compatibility float64
}

// DefaultWeights 默认权重（可由系统参数覆盖）。
func DefaultWeights() Weights { return Weights{Preference: 0.4, Compatibility: 0.6} }

// Normalize 归一化权重，保证和为 1。
func (w Weights) Normalize() Weights {
	sum := w.Preference + w.Compatibility
	if sum <= 0 {
		return DefaultWeights()
	}
	return Weights{Preference: w.Preference / sum, Compatibility: w.Compatibility / sum}
}

// MinValidQuestions 有效价值观题最少数量，低于该值不输出精确分数。
const MinValidQuestions = 6

// InsufficientText 数据不足文案。
const InsufficientText = "关系观念数据不足"

// Disclaimer 结果免责声明。
const Disclaimer = "以上分数仅根据双方填写条件计算，不代表婚姻成功概率。"

// FormatScore 输出两位小数的百分比文案（后端内部使用）。
func FormatScore(v float64) string { return fmt.Sprintf("%.1f", v*100) }
