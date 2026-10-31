package profile

import (
	"testing"
	"time"
)

func fullData() Data {
	birth := time.Date(1995, 5, 20, 0, 0, 0, 0, time.UTC)
	h := int16(170)
	return Data{
		Nickname: "小明", Gender: 1, BirthDate: &birth, CityCode: "310100", HeightCM: &h,
		EducationLevel: 3, OccupationCategory: 1, EmploymentStatus: 1, AnnualIncomeBand: 4,
		MaritalStatus: 1, WantsChildren: 2, MarriageIntent: 2, Bio: "你好",
		SmokingStatus: 1, DrinkingStatus: 1, HouseStatus: 2, VehicleStatus: 2,
	}
}

func TestComputeComplete(t *testing.T) {
	if got := Compute(fullData(), true); got != 100 {
		t.Fatalf("完整资料且有照片应为 100，实际 %d", got)
	}
	if got := Compute(fullData(), false); got != 92 {
		t.Fatalf("无照片应为 92，实际 %d", got)
	}
}

func TestComputeEmpty(t *testing.T) {
	if got := Compute(Data{}, false); got != 0 {
		t.Fatalf("空资料应为 0，实际 %d", got)
	}
}

func TestMissingFields(t *testing.T) {
	d := fullData()
	d.CityCode = ""
	d.Bio = ""
	missing := MissingFields(d, false)
	found := map[string]bool{}
	for _, m := range missing {
		found[m] = true
	}
	if !found["所在城市"] || !found["个人简介"] || !found["主照片"] {
		t.Fatalf("缺失字段识别不正确: %+v", missing)
	}
}

func TestAge(t *testing.T) {
	birth := time.Date(2000, 12, 31, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	if got := Age(&birth, now); got != 25 {
		t.Fatalf("生日未到应为 25，实际 %d", got)
	}
	now = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	if got := Age(&birth, now); got != 26 {
		t.Fatalf("生日当天应为 26，实际 %d", got)
	}
	if got := Age(nil, now); got != 0 {
		t.Fatalf("未填写生日应为 0，实际 %d", got)
	}
}
