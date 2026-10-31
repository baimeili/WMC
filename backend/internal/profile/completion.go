// Package profile 处理用户资料与完成度计算。
package profile

import "time"

// Data 用户资料领域数据。
type Data struct {
	Nickname           string
	Gender             int16
	BirthDate          *time.Time
	ProvinceCode       string
	CityCode           string
	DistrictCode       *string
	HeightCM           *int16
	EducationLevel     int16
	OccupationCategory int16
	EmploymentStatus   int16
	AnnualIncomeBand   int16
	FamilyAssetBand    *int16
	DebtBand           int16
	HouseStatus        int16
	VehicleStatus      int16
	MaritalStatus      int16
	ChildrenCount      int16
	WantsChildren      int16
	MarriageIntent     int16
	MarriageTimeline   int16
	SmokingStatus      int16
	DrinkingStatus     int16
	Bio                string
}

// Age 计算年龄（按自然年周岁）。
func Age(birth *time.Time, now time.Time) int {
	if birth == nil {
		return 0
	}
	years := now.Year() - birth.Year()
	if now.Month() < birth.Month() || (now.Month() == birth.Month() && now.Day() < birth.Day()) {
		years--
	}
	if years < 0 {
		return 0
	}
	return years
}

type weightedField struct {
	label  string
	weight int
	filled func(Data) bool
}

// 资料完成度权重，合计 100 分。照片单独计 8 分，由 hasPhoto 决定。
var coreFields = []weightedField{
	{"昵称", 5, func(d Data) bool { return d.Nickname != "" }},
	{"性别", 5, func(d Data) bool { return d.Gender > 0 }},
	{"出生日期", 10, func(d Data) bool { return d.BirthDate != nil }},
	{"所在城市", 10, func(d Data) bool { return d.CityCode != "" }},
	{"身高", 5, func(d Data) bool { return d.HeightCM != nil }},
	{"学历", 5, func(d Data) bool { return d.EducationLevel > 0 }},
	{"职业类别", 5, func(d Data) bool { return d.OccupationCategory > 0 }},
	{"就业状态", 5, func(d Data) bool { return d.EmploymentStatus > 0 }},
	{"年收入区间", 10, func(d Data) bool { return d.AnnualIncomeBand > 0 }},
	{"婚姻状态", 5, func(d Data) bool { return d.MaritalStatus > 0 }},
	{"生育意愿", 5, func(d Data) bool { return d.WantsChildren > 0 }},
	{"结婚意愿", 5, func(d Data) bool { return d.MarriageIntent > 0 }},
	{"个人简介", 5, func(d Data) bool { return d.Bio != "" }},
	{"吸烟习惯", 3, func(d Data) bool { return d.SmokingStatus > 0 }},
	{"饮酒习惯", 3, func(d Data) bool { return d.DrinkingStatus > 0 }},
	{"房产状态", 3, func(d Data) bool { return d.HouseStatus > 0 }},
	{"车辆状态", 3, func(d Data) bool { return d.VehicleStatus > 0 }},
}

// PhotoWeight 照片权重。
const PhotoWeight = 8

// MaxCompletion 满分。
const MaxCompletion = 100

// Compute 计算资料完成度（0~100）。
func Compute(d Data, hasPhoto bool) int {
	score := 0
	for _, f := range coreFields {
		if f.filled(d) {
			score += f.weight
		}
	}
	if hasPhoto {
		score += PhotoWeight
	}
	if score > MaxCompletion {
		score = MaxCompletion
	}
	return score
}

// MissingFields 返回未填写字段的中文名称，用于引导用户完善。
func MissingFields(d Data, hasPhoto bool) []string {
	out := make([]string, 0, len(coreFields))
	for _, f := range coreFields {
		if !f.filled(d) {
			out = append(out, f.label)
		}
	}
	if !hasPhoto {
		out = append(out, "主照片")
	}
	return out
}

// DisplayAge 返回用于展示的年龄（未填写返回 0）。
func DisplayAge(d Data, now time.Time) int {
	return Age(d.BirthDate, now)
}
