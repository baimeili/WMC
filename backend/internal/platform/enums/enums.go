// Package enums 集中定义所有 smallint 编码与中文标签。
// 前后端只交换编码，标签统一由 /meta/options 下发，避免两端文案分裂。
package enums

import "sort"

// Option 可选项。
type Option struct {
	Value int16  `json:"value"`
	Label string `json:"label"`
}

// ---- 角色 ----
const (
	RoleUser       int16 = 1
	RoleModerator  int16 = 2
	RoleAdmin      int16 = 3
	RoleSuperAdmin int16 = 4
)

var RoleLabels = map[int16]string{
	RoleUser: "普通用户", RoleModerator: "内容审核员", RoleAdmin: "管理员", RoleSuperAdmin: "超级管理员",
}

// ---- 用户状态 ----
const (
	UserStatusPending         int16 = 1
	UserStatusActive          int16 = 2
	UserStatusSuspended       int16 = 3
	UserStatusDeleted         int16 = 4
	UserStatusDeletionPending int16 = 5
)

var UserStatusLabels = map[int16]string{
	UserStatusPending: "待激活", UserStatusActive: "正常", UserStatusSuspended: "已冻结",
	UserStatusDeleted: "已注销", UserStatusDeletionPending: "注销保留期",
}

// ---- 性别 ----
const (
	GenderUnset       int16 = 0
	GenderMale        int16 = 1
	GenderFemale      int16 = 2
	GenderUndisclosed int16 = 3
)

var GenderLabels = map[int16]string{
	GenderMale: "男", GenderFemale: "女", GenderUndisclosed: "不愿透露",
}

// ---- 学历 ----
const (
	EducationUnknown int16 = 0
	EducationHighSchool int16 = 1
	EducationCollege    int16 = 2
	EducationBachelor   int16 = 3
	EducationMaster     int16 = 4
	EducationDoctor     int16 = 5
)

var EducationLabels = map[int16]string{
	EducationUnknown: "未填写", EducationHighSchool: "高中及以下", EducationCollege: "中专/大专",
	EducationBachelor: "本科", EducationMaster: "硕士", EducationDoctor: "博士",
}

// ---- 年收入区间（禁止精确工资）----
const (
	IncomeUnknown  int16 = 0
	IncomeBelow5   int16 = 1
	Income5to10    int16 = 2
	Income10to20   int16 = 3
	Income20to30   int16 = 4
	Income30to50   int16 = 5
	Income50to100  int16 = 6
	IncomeAbove100 int16 = 7
	IncomeUndisclosed int16 = 8
)

var IncomeBandLabels = map[int16]string{
	IncomeUnknown: "未填写", IncomeBelow5: "5万以下", Income5to10: "5-10万", Income10to20: "10-20万",
	Income20to30: "20-30万", Income30to50: "30-50万", Income50to100: "50-100万",
	IncomeAbove100: "100万以上", IncomeUndisclosed: "不愿透露",
}

// ---- 家庭资产区间 ----
const (
	AssetUnknown int16 = 0
	AssetBelow50 int16 = 1
	Asset50to100 int16 = 2
	Asset100to300 int16 = 3
	Asset300to1000 int16 = 4
	AssetAbove1000 int16 = 5
	AssetUndisclosed int16 = 6
)

var AssetBandLabels = map[int16]string{
	AssetUnknown: "未填写", AssetBelow50: "50万以下", Asset50to100: "50-100万",
	Asset100to300: "100-300万", Asset300to1000: "300-1000万", AssetAbove1000: "1000万以上",
	AssetUndisclosed: "不愿透露",
}

// ---- 负债区间 ----
const (
	DebtNone int16 = 0
	DebtBelow5 int16 = 1
	Debt5to20 int16 = 2
	Debt20to50 int16 = 3
	Debt50to100 int16 = 4
	DebtAbove100 int16 = 5
	DebtUndisclosed int16 = 6
)

var DebtBandLabels = map[int16]string{
	DebtNone: "无负债", DebtBelow5: "5万以下", Debt5to20: "5-20万", Debt20to50: "20-50万",
	Debt50to100: "50-100万", DebtAbove100: "100万以上", DebtUndisclosed: "不愿透露",
}

// ---- 房产 ----
const (
	HouseUnknown int16 = 0
	HouseNone int16 = 1
	HouseWithLoan int16 = 2
	HouseNoLoan int16 = 3
	HouseMultiple int16 = 4
	HouseWithParents int16 = 5
	HouseUndisclosed int16 = 6
)

var HouseStatusLabels = map[int16]string{
	HouseUnknown: "未填写", HouseNone: "暂无房产", HouseWithLoan: "有房有贷款",
	HouseNoLoan: "有房无贷款", HouseMultiple: "多套房产", HouseWithParents: "与父母同住",
	HouseUndisclosed: "不愿透露",
}

// ---- 车辆 ----
const (
	VehicleUnknown int16 = 0
	VehicleNone int16 = 1
	VehicleOwned int16 = 2
	VehicleMultiple int16 = 3
	VehicleUndisclosed int16 = 4
)

var VehicleStatusLabels = map[int16]string{
	VehicleUnknown: "未填写", VehicleNone: "暂无车辆", VehicleOwned: "有车",
	VehicleMultiple: "多辆车", VehicleUndisclosed: "不愿透露",
}

// ---- 婚姻状态 ----
const (
	MaritalUnknown int16 = 0
	MaritalSingle int16 = 1
	MaritalDivorced int16 = 2
	MaritalWidowed int16 = 3
	MaritalUndisclosed int16 = 4
)

var MaritalStatusLabels = map[int16]string{
	MaritalUnknown: "未填写", MaritalSingle: "未婚", MaritalDivorced: "离异",
	MaritalWidowed: "丧偶", MaritalUndisclosed: "不愿透露",
}

// ---- 子女状态（用于偏好匹配）----
const (
	ChildrenStateNone int16 = 0
	ChildrenStateHas  int16 = 1
)

var ChildrenStateLabels = map[int16]string{
	ChildrenStateNone: "无子女", ChildrenStateHas: "有子女",
}

// ChildrenStateOf 由子女数量推导子女状态。
func ChildrenStateOf(count int16) int16 {
	if count > 0 {
		return ChildrenStateHas
	}
	return ChildrenStateNone
}

// ---- 生育意愿 ----
const (
	WantsChildrenUnknown int16 = 0
	WantsChildrenNo      int16 = 1
	WantsChildrenYes     int16 = 2
	WantsChildrenDepends int16 = 3
	WantsChildrenDone    int16 = 4
)

var WantsChildrenLabels = map[int16]string{
	WantsChildrenUnknown: "未填写", WantsChildrenNo: "不想要孩子", WantsChildrenYes: "想要孩子",
	WantsChildrenDepends: "看情况", WantsChildrenDone: "已有子女，不再生育",
}

// ---- 结婚意愿 ----
const (
	MarriageIntentUnknown int16 = 0
	MarriageIntentAsap    int16 = 1
	MarriageIntent1Year   int16 = 2
	MarriageIntent2Years  int16 = 3
	MarriageIntentNoPlan  int16 = 4
	MarriageIntentDateOnly int16 = 5
)

var MarriageIntentLabels = map[int16]string{
	MarriageIntentUnknown: "未填写", MarriageIntentAsap: "尽快结婚", MarriageIntent1Year: "1年内结婚",
	MarriageIntent2Years: "2年内结婚", MarriageIntentNoPlan: "暂无明确计划", MarriageIntentDateOnly: "只恋爱不结婚",
}

// marriageIntentRank 将编码映射为「结婚意愿强烈程度」，用于 min_marriage_intent 比较。
var marriageIntentRank = map[int16]int16{
	MarriageIntentUnknown: 0, MarriageIntentAsap: 5, MarriageIntent1Year: 4,
	MarriageIntent2Years: 3, MarriageIntentNoPlan: 2, MarriageIntentDateOnly: 1,
}

// IntentRank 返回结婚意愿等级（越大越强烈）。
func IntentRank(code int16) int16 { return marriageIntentRank[code] }

// MeetsMinIntent 判断用户结婚意愿是否不低于要求。
func MeetsMinIntent(actual, minRequired int16) bool {
	if minRequired <= MarriageIntentUnknown {
		return true
	}
	return IntentRank(actual) >= IntentRank(minRequired)
}

// ---- 结婚时间规划 ----
const (
	TimelineUnknown int16 = 0
	Timeline3Months int16 = 1
	Timeline6Months int16 = 2
	Timeline1Year int16 = 3
	Timeline2Years int16 = 4
	Timeline3YearsPlus int16 = 5
	TimelineUnsure int16 = 6
)

var MarriageTimelineLabels = map[int16]string{
	TimelineUnknown: "未填写", Timeline3Months: "3个月内", Timeline6Months: "6个月内",
	Timeline1Year: "1年内", Timeline2Years: "2年内", Timeline3YearsPlus: "3年以上", TimelineUnsure: "不确定",
}

// ---- 吸烟 ----
const (
	SmokingUnknown int16 = 0
	SmokingNever int16 = 1
	SmokingSometimes int16 = 2
	SmokingOften int16 = 3
	SmokingQuit int16 = 4
)

var SmokingStatusLabels = map[int16]string{
	SmokingUnknown: "未填写", SmokingNever: "不吸烟", SmokingSometimes: "偶尔吸烟",
	SmokingOften: "经常吸烟", SmokingQuit: "已戒烟",
}

// ---- 饮酒 ----
const (
	DrinkingUnknown int16 = 0
	DrinkingNever int16 = 1
	DrinkingSometimes int16 = 2
	DrinkingOften int16 = 3
	DrinkingSocial int16 = 4
)

var DrinkingStatusLabels = map[int16]string{
	DrinkingUnknown: "未填写", DrinkingNever: "不饮酒", DrinkingSometimes: "偶尔饮酒",
	DrinkingOften: "经常饮酒", DrinkingSocial: "应酬饮酒",
}

// ---- 就业状态 ----
const (
	EmploymentUnknown int16 = 0
	EmploymentFullTime int16 = 1
	EmploymentPartTime int16 = 2
	EmploymentFreelance int16 = 3
	EmploymentFounder int16 = 4
	EmploymentStudent int16 = 5
	EmploymentUnemployed int16 = 6
)

var EmploymentStatusLabels = map[int16]string{
	EmploymentUnknown: "未填写", EmploymentFullTime: "全职", EmploymentPartTime: "兼职",
	EmploymentFreelance: "自由职业", EmploymentFounder: "创业", EmploymentStudent: "学生",
	EmploymentUnemployed: "待业",
}

// ---- 职业分类 ----
const (
	OccupationUnknown int16 = 0
	OccupationIT int16 = 1
	OccupationFinance int16 = 2
	OccupationEducation int16 = 3
	OccupationHealthcare int16 = 4
	OccupationPublicSector int16 = 5
	OccupationManufacturing int16 = 6
	OccupationSalesMarketing int16 = 7
	OccupationDesignCreative int16 = 8
	OccupationLegalConsulting int16 = 9
	OccupationFreelance int16 = 10
	OccupationOther int16 = 11
)

var OccupationLabels = map[int16]string{
	OccupationUnknown: "未填写", OccupationIT: "IT/互联网", OccupationFinance: "金融",
	OccupationEducation: "教育", OccupationHealthcare: "医疗健康", OccupationPublicSector: "公务员/事业单位",
	OccupationManufacturing: "制造业", OccupationSalesMarketing: "销售/市场", OccupationDesignCreative: "设计/创意",
	OccupationLegalConsulting: "律师/咨询", OccupationFreelance: "自由职业", OccupationOther: "其他",
}

// ---- 互动 ----
const (
	InteractionLike int16 = 1
	InteractionPass int16 = 2
)

var InteractionActionLabels = map[int16]string{InteractionLike: "喜欢", InteractionPass: "跳过"}

// 互动来源
const (
	InteractionSourceCard int16 = 1
	InteractionSourceDetail int16 = 2
	InteractionSourceMatch int16 = 3
)

// ---- Match 状态 ----
const (
	MatchStatusActive int16 = 1
	MatchStatusCancelled int16 = 2
	MatchStatusBlocked int16 = 3
)

var MatchStatusLabels = map[int16]string{
	MatchStatusActive: "进行中", MatchStatusCancelled: "已取消", MatchStatusBlocked: "已屏蔽",
}

// ---- 结果反馈 ----
const (
	OutcomeContactExchanged int16 = 1
	OutcomeChatted int16 = 2
	OutcomeMetOffline int16 = 3
	OutcomeDating int16 = 4
	OutcomeRelationship int16 = 5
	OutcomeEngaged int16 = 6
	OutcomeMarried int16 = 7
	OutcomeEnded int16 = 8
)

var OutcomeTypeLabels = map[int16]string{
	OutcomeContactExchanged: "已交换联系方式", OutcomeChatted: "已经聊过", OutcomeMetOffline: "已经线下见面",
	OutcomeDating: "正在约会", OutcomeRelationship: "建立稳定关系", OutcomeEngaged: "已订婚",
	OutcomeMarried: "已结婚", OutcomeEnded: "没有继续",
}

// ---- 审核 ----
const (
	ModerationPending int16 = 1
	ModerationApproved int16 = 2
	ModerationRejected int16 = 3
)

var ModerationStatusLabels = map[int16]string{
	ModerationPending: "待审核", ModerationApproved: "已通过", ModerationRejected: "已拒绝",
}

// ---- 举报 ----
const (
	ReportReasonFraud int16 = 1
	ReportReasonHarassment int16 = 2
	ReportReasonFakeProfile int16 = 3
	ReportReasonStolenPhoto int16 = 4
	ReportReasonPorn int16 = 5
	ReportReasonAd int16 = 6
	ReportReasonMarried int16 = 7
	ReportReasonOther int16 = 8
)

var ReportReasonLabels = map[int16]string{
	ReportReasonFraud: "诈骗", ReportReasonHarassment: "骚扰", ReportReasonFakeProfile: "虚假资料",
	ReportReasonStolenPhoto: "盗图", ReportReasonPorn: "色情", ReportReasonAd: "广告",
	ReportReasonMarried: "已婚冒充单身", ReportReasonOther: "其他",
}

const (
	ReportStatusPending int16 = 1
	ReportStatusReviewing int16 = 2
	ReportStatusResolved int16 = 3
	ReportStatusRejected int16 = 4
)

var ReportStatusLabels = map[int16]string{
	ReportStatusPending: "待处理", ReportStatusReviewing: "处理中",
	ReportStatusResolved: "已处理", ReportStatusRejected: "已驳回",
}

// ---- 验证 ----
const (
	VerificationPhone int16 = 1
	VerificationIdentity int16 = 2
	VerificationEducation int16 = 3
	VerificationIncome int16 = 4
	VerificationMarriageStatus int16 = 5
)

var VerificationTypeLabels = map[int16]string{
	VerificationPhone: "手机号", VerificationIdentity: "身份认证", VerificationEducation: "学历认证",
	VerificationIncome: "收入认证", VerificationMarriageStatus: "婚姻状态认证",
}

const (
	VerifyUnverified int16 = 1
	VerifyPending int16 = 2
	VerifyVerified int16 = 3
	VerifyRejected int16 = 4
	VerifyExpired int16 = 5
)

var VerificationStatusLabels = map[int16]string{
	VerifyUnverified: "未验证", VerifyPending: "审核中", VerifyVerified: "已验证",
	VerifyRejected: "已拒绝", VerifyExpired: "已过期",
}

// ---- 同意类型 ----
const (
	ConsentTerms int16 = 1
	ConsentPrivacy int16 = 2
	ConsentMarket int16 = 3
)

var ConsentTypeLabels = map[int16]string{
	ConsentTerms: "用户协议", ConsentPrivacy: "隐私政策", ConsentMarket: "婚恋市场参与声明",
}

// ---- 问题类型 ----
const (
	QuestionSingle int16 = 1
	QuestionMulti int16 = 2
)

// ---- 数据集状态 ----
const (
	DatasetDraft int16 = 1
	DatasetReady int16 = 2
	DatasetImporting int16 = 3
	DatasetArchived int16 = 4
)

var DatasetStatusLabels = map[int16]string{
	DatasetDraft: "草稿", DatasetReady: "可用", DatasetImporting: "导入中", DatasetArchived: "已归档",
}

// Labels 返回 code→label 映射的安全取值。
func Labels(m map[int16]string, code int16) string {
	if v, ok := m[code]; ok {
		return v
	}
	return "未知"
}

func toOptions(m map[int16]string, skipZero bool) []Option {
	out := make([]Option, 0, len(m))
	for k, v := range m {
		if skipZero && k == 0 {
			continue
		}
		out = append(out, Option{Value: k, Label: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value < out[j].Value })
	return out
}

// Options 返回全部选项字典，供 GET /api/v1/meta/options 使用。
func Options() map[string][]Option {
	return map[string][]Option{
		"genders":             toOptions(GenderLabels, false),
		"education_levels":    toOptions(EducationLabels, false),
		"income_bands":        toOptions(IncomeBandLabels, false),
		"asset_bands":         toOptions(AssetBandLabels, false),
		"debt_bands":          toOptions(DebtBandLabels, false),
		"house_statuses":      toOptions(HouseStatusLabels, false),
		"vehicle_statuses":    toOptions(VehicleStatusLabels, false),
		"marital_statuses":    toOptions(MaritalStatusLabels, false),
		"children_states":     toOptions(ChildrenStateLabels, false),
		"wants_children":      toOptions(WantsChildrenLabels, false),
		"marriage_intents":    toOptions(MarriageIntentLabels, false),
		"marriage_timelines":  toOptions(MarriageTimelineLabels, false),
		"smoking_statuses":    toOptions(SmokingStatusLabels, false),
		"drinking_statuses":   toOptions(DrinkingStatusLabels, false),
		"employment_statuses": toOptions(EmploymentStatusLabels, false),
		"occupations":         toOptions(OccupationLabels, false),
		"report_reasons":      toOptions(ReportReasonLabels, false),
		"outcome_types":       toOptions(OutcomeTypeLabels, false),
		"verification_types":  toOptions(VerificationTypeLabels, false),
		"consent_types":       toOptions(ConsentTypeLabels, false),
		"roles":               toOptions(RoleLabels, false),
		"user_statuses":       toOptions(UserStatusLabels, false),
	}
}
