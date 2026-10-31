// Package user 处理当前用户的资料、择偶条件、价值观规则、入池与注销。
package user

import "time"

// Profile 资料 DTO。
type Profile struct {
	UserID             string  `json:"user_id"`
	Nickname           string  `json:"nickname"`
	Gender             int16   `json:"gender"`
	BirthDate          *string `json:"birth_date"`
	Age                int     `json:"age"`
	ProvinceCode       string  `json:"province_code"`
	CityCode           string  `json:"city_code"`
	DistrictCode       *string `json:"district_code"`
	HeightCM           *int16  `json:"height_cm"`
	EducationLevel     int16   `json:"education_level"`
	OccupationCategory int16   `json:"occupation_category"`
	EmploymentStatus   int16   `json:"employment_status"`
	AnnualIncomeBand   int16   `json:"annual_income_band"`
	FamilyAssetBand    *int16  `json:"family_asset_band"`
	DebtBand           int16   `json:"debt_band"`
	HouseStatus        int16   `json:"house_status"`
	VehicleStatus      int16   `json:"vehicle_status"`
	MaritalStatus      int16   `json:"marital_status"`
	ChildrenCount      int16   `json:"children_count"`
	WantsChildren      int16   `json:"wants_children"`
	MarriageIntent     int16   `json:"marriage_intent"`
	MarriageTimeline   int16   `json:"marriage_timeline"`
	SmokingStatus      int16   `json:"smoking_status"`
	DrinkingStatus     int16   `json:"drinking_status"`
	Bio                string  `json:"bio"`
	ProfileCompletion  int16   `json:"profile_completion"`
	UpdatedAt          string  `json:"updated_at"`
}

// ProfileInput 保存资料输入。
type ProfileInput struct {
	Nickname           string  `json:"nickname" validate:"required,max=32"`
	Gender             int16   `json:"gender" validate:"required,oneof=1 2 3"`
	BirthDate          *string `json:"birth_date"`
	ProvinceCode       string  `json:"province_code"`
	CityCode           string  `json:"city_code" validate:"required,max=12"`
	DistrictCode       *string `json:"district_code"`
	HeightCM           *int16  `json:"height_cm" validate:"omitempty,gte=100,lte=250"`
	EducationLevel     int16   `json:"education_level" validate:"gte=0,lte=5"`
	OccupationCategory int16   `json:"occupation_category" validate:"gte=0,lte=11"`
	EmploymentStatus   int16   `json:"employment_status" validate:"gte=0,lte=6"`
	AnnualIncomeBand   int16   `json:"annual_income_band" validate:"gte=0,lte=8"`
	FamilyAssetBand    *int16  `json:"family_asset_band" validate:"omitempty,gte=0,lte=6"`
	DebtBand           int16   `json:"debt_band" validate:"gte=0,lte=6"`
	HouseStatus        int16   `json:"house_status" validate:"gte=0,lte=6"`
	VehicleStatus      int16   `json:"vehicle_status" validate:"gte=0,lte=4"`
	MaritalStatus      int16   `json:"marital_status" validate:"gte=0,lte=4"`
	ChildrenCount      int16   `json:"children_count" validate:"gte=0,lte=10"`
	WantsChildren      int16   `json:"wants_children" validate:"gte=0,lte=4"`
	MarriageIntent     int16   `json:"marriage_intent" validate:"gte=0,lte=5"`
	MarriageTimeline   int16   `json:"marriage_timeline" validate:"gte=0,lte=6"`
	SmokingStatus      int16   `json:"smoking_status" validate:"gte=0,lte=4"`
	DrinkingStatus     int16   `json:"drinking_status" validate:"gte=0,lte=4"`
	Bio                string  `json:"bio" validate:"max=500"`
}

// Preferences 择偶条件 DTO。
type Preferences struct {
	TargetGenders          []int16          `json:"target_genders"`
	MinAge                 int16            `json:"min_age"`
	MaxAge                 int16            `json:"max_age"`
	AgeHard                bool             `json:"age_hard"`
	MinHeightCM            int16            `json:"min_height_cm"`
	MaxHeightCM            int16            `json:"max_height_cm"`
	HeightHard             bool             `json:"height_hard"`
	AllowedCityCodes       []string         `json:"allowed_city_codes"`
	CityHard               bool             `json:"city_hard"`
	MinEducationLevel      int16            `json:"min_education_level"`
	EducationHard          bool             `json:"education_hard"`
	MinIncomeBand          int16            `json:"min_income_band"`
	IncomeHard             bool             `json:"income_hard"`
	AllowedMaritalStatuses []int16          `json:"allowed_marital_statuses"`
	MaritalStatusHard      bool             `json:"marital_status_hard"`
	AllowedChildrenStates  []int16          `json:"allowed_children_states"`
	ChildrenHard           bool             `json:"children_hard"`
	AllowedWantsChildren   []int16          `json:"allowed_wants_children"`
	WantsChildrenHard      bool             `json:"wants_children_hard"`
	AllowedSmokingStatuses []int16          `json:"allowed_smoking_statuses"`
	SmokingHard            bool             `json:"smoking_hard"`
	AllowedDrinkingStatuses []int16         `json:"allowed_drinking_statuses"`
	DrinkingHard           bool             `json:"drinking_hard"`
	MinMarriageIntent      int16            `json:"min_marriage_intent"`
	MarriageIntentHard     bool             `json:"marriage_intent_hard"`
	Weights                map[string]float64 `json:"weights"`
	UpdatedAt              string           `json:"updated_at"`
}

// CompatibilityRule 价值观规则 DTO。
type CompatibilityRule struct {
	QuestionID      string   `json:"question_id"`
	QuestionCode    string   `json:"question_code"`
	Category        string   `json:"category"`
	Title           string   `json:"title"`
	OwnAnswer       []string `json:"own_answer"`
	AcceptedAnswers []string `json:"accepted_answers"`
	Importance      int16    `json:"importance"`
	IsHard          bool     `json:"is_hard"`
}

// CompatibilityRuleInput 保存规则输入。
type CompatibilityRuleInput struct {
	QuestionID      string   `json:"question_id" validate:"required,uuid"`
	OwnAnswer       []string `json:"own_answer" validate:"required,min=1"`
	AcceptedAnswers []string `json:"accepted_answers" validate:"required,min=1"`
	Importance      int16    `json:"importance" validate:"gte=0,lte=5"`
	IsHard          bool     `json:"is_hard"`
}

// MarketStatus 入池状态。
type MarketStatus struct {
	InMarket            bool     `json:"in_market"`
	DiscoveryEnabled    bool     `json:"discovery_enabled"`
	ProfileCompletion   int16    `json:"profile_completion"`
	RequiredCompletion  int      `json:"required_completion"`
	Age                 int      `json:"age"`
	AgeOK               bool     `json:"age_ok"`
	CompletionOK        bool     `json:"completion_ok"`
	ConsentOK           bool     `json:"consent_ok"`
	AccountOK           bool     `json:"account_ok"`
	MissingRequirements []string `json:"missing_requirements"`
	RuleCount           int      `json:"rule_count"`
}

// ConsentInput 同意记录输入。
type ConsentInput struct {
	ConsentType   int16  `json:"consent_type" validate:"required,oneof=1 2 3"`
	PolicyVersion string `json:"policy_version" validate:"max=32"`
}

// nowFunc 便于测试注入时间。
var nowFunc = time.Now
