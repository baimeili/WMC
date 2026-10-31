// Package seed 生成开发/测试用的虚构数据。
// 所有数据必须明确标记为 TEST DATA / DEMO，生产环境禁止使用。
package seed

import (
	"fmt"
	"math/rand"

	"wmc/backend/internal/calculator"
)

// DemoDatasetName 演示数据集名称（会在页面上显著标注）。
const DemoDatasetName = "DEMO-示例数据集（非真实统计）"

// DemoDisclaimer 演示数据免责声明。
const DemoDisclaimer = "本数据集为演示用虚构数据，不代表任何真实人口统计结果。"

// DemoDataset 生成演示用的加权基线记录。
func DemoDataset(count int, seedValue int64) []calculator.Record {
	r := rand.New(rand.NewSource(seedValue))
	cities := []string{
		"310100", "110100", "440300", "440100", "330100", "320100", "510100",
		"420100", "430100", "500100", "370200", "370100", "120100", "350200", "610100",
	}
	records := make([]calculator.Record, 0, count)
	for i := 0; i < count; i++ {
		gender := int16(1)
		if r.Intn(2) == 0 {
			gender = 2
		}
		age := int16(22 + r.Intn(21)) // 22 ~ 42
		height := int16(155 + r.Intn(35))
		education := int16(1 + r.Intn(5))
		income := int16(1 + r.Intn(7))
		marital := int16(1)
		if age > 32 && r.Intn(10) < 2 {
			marital = 2
		}
		children := int16(0)
		if marital == 2 && r.Intn(2) == 0 {
			children = 1
		}
		wants := int16(2)
		switch r.Intn(10) {
		case 0, 1, 2:
			wants = 2
		case 3, 4:
			wants = 3
		case 5:
			wants = 1
		default:
			wants = 2
		}
		weight := 40 + r.Float64()*460
		var h *int16
		if r.Intn(100) < 95 {
			h = &height
		}
		records = append(records, calculator.Record{
			Gender:           gender,
			Age:              age,
			ProvinceCode:     cities[r.Intn(len(cities))][:2] + "0000",
			CityCode:         cities[r.Intn(len(cities))],
			HeightCM:         h,
			EducationLevel:   education,
			AnnualIncomeBand: income,
			MaritalStatus:    marital,
			ChildrenCount:    children,
			WantsChildren:    wants,
			Weight:           weight,
		})
	}
	return records
}

// BulkRecord 导入用的记录结构。
type BulkRecord struct {
	Gender           int16
	Age              int16
	ProvinceCode     string
	CityCode         string
	HeightCM         *int16
	EducationLevel   int16
	AnnualIncomeBand int16
	MaritalStatus    int16
	ChildrenCount    int16
	WantsChildren    int16
	Weight           float64
}

// ToBulk 转换演示记录。
func ToBulk(records []calculator.Record) []BulkRecord {
	out := make([]BulkRecord, 0, len(records))
	for _, rec := range records {
		out = append(out, BulkRecord{
			Gender: rec.Gender, Age: rec.Age, ProvinceCode: rec.ProvinceCode, CityCode: rec.CityCode,
			HeightCM: rec.HeightCM, EducationLevel: rec.EducationLevel, AnnualIncomeBand: rec.AnnualIncomeBand,
			MaritalStatus: rec.MaritalStatus, ChildrenCount: rec.ChildrenCount, WantsChildren: rec.WantsChildren,
			Weight: rec.Weight,
		})
	}
	return out
}

// Validate 校验导入记录。
func (b BulkRecord) Validate(index int) error {
	if b.Gender != 1 && b.Gender != 2 && b.Gender != 3 {
		return fmt.Errorf("第 %d 行：性别编码非法", index)
	}
	if b.Age < 18 || b.Age > 100 {
		return fmt.Errorf("第 %d 行：年龄必须在 18~100 之间", index)
	}
	if b.Weight <= 0 {
		return fmt.Errorf("第 %d 行：样本权重必须大于 0", index)
	}
	if b.HeightCM != nil && (*b.HeightCM < 100 || *b.HeightCM > 250) {
		return fmt.Errorf("第 %d 行：身高必须在 100~250 之间", index)
	}
	return nil
}
