// Package baseline 管理理论婚恋市场的基准人口数据集。
package baseline

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/internal/db/sqlcdb"
	"wmc/backend/internal/platform/apperr"
	"wmc/backend/internal/platform/enums"
	"wmc/backend/internal/seed"
	"wmc/backend/internal/settings"
	"wmc/backend/internal/store"
)

// CSVHeaders 导入 CSV 的必需列。
var CSVHeaders = []string{
	"gender", "age", "province_code", "city_code", "height_cm",
	"education_level", "annual_income_band", "marital_status",
	"children_count", "wants_children", "sample_weight",
}

// Service 数据集服务。
type Service struct {
	store     *store.Store
	pool      *pgxpool.Pool
	settings  *settings.Service
	appEnv    string
}

// NewService 构造数据集服务。
func NewService(st *store.Store, pool *pgxpool.Pool, set *settings.Service, appEnv string) *Service {
	return &Service{store: st, pool: pool, settings: set, appEnv: appEnv}
}

// Dataset DTO。
type Dataset struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	SourceName  string    `json:"source_name"`
	SourceURL   string    `json:"source_url"`
	SourceYear  *int16    `json:"source_year"`
	Description string    `json:"description"`
	Methodology string    `json:"methodology"`
	Disclaimer  string    `json:"disclaimer"`
	Status      int16     `json:"status"`
	Active      bool      `json:"active"`
	IsDemo      bool      `json:"is_demo"`
	RecordCount int64     `json:"record_count"`
	CreatedAt   string    `json:"created_at"`
}

func toDataset(d sqlcdb.BaselineDataset) Dataset {
	return Dataset{
		ID: d.ID, Name: d.Name, Version: d.Version, SourceName: d.SourceName, SourceURL: d.SourceUrl,
		SourceYear: d.SourceYear, Description: d.Description, Methodology: d.Methodology,
		Disclaimer: d.Disclaimer, Status: d.Status, Active: d.Active, IsDemo: d.IsDemo,
		RecordCount: d.RecordCount, CreatedAt: d.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// List 返回全部数据集。
func (s *Service) List(ctx context.Context) ([]Dataset, error) {
	rows, err := s.store.Q.ListDatasets(ctx)
	if err != nil {
		return nil, apperr.ErrInternal.Wrap(err)
	}
	out := make([]Dataset, 0, len(rows))
	for _, d := range rows {
		out = append(out, toDataset(d))
	}
	return out, nil
}

// Get 返回单个数据集。
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Dataset, error) {
	d, err := s.store.Q.GetDatasetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Dataset{}, apperr.ErrNotFound.WithMessage("数据集不存在")
	}
	if err != nil {
		return Dataset{}, apperr.ErrInternal.Wrap(err)
	}
	return toDataset(d), nil
}

// CreateInput 创建数据集输入。
type CreateInput struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	SourceName  string `json:"source_name"`
	SourceURL   string `json:"source_url"`
	SourceYear  *int16 `json:"source_year"`
	Description string `json:"description"`
	Methodology string `json:"methodology"`
	Disclaimer  string `json:"disclaimer"`
	IsDemo      bool   `json:"is_demo"`
}

// Create 创建草稿数据集。
func (s *Service) Create(ctx context.Context, in CreateInput, createdBy *uuid.UUID) (Dataset, error) {
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Version) == "" {
		return Dataset{}, apperr.Invalid("数据集名称与版本不能为空")
	}
	if in.IsDemo && s.appEnv == "production" {
		return Dataset{}, apperr.ErrForbidden.WithMessage("生产环境禁止创建演示数据集")
	}
	d, err := s.store.Q.CreateDataset(ctx, sqlcdb.CreateDatasetParams{
		Name: in.Name, Version: in.Version, SourceName: in.SourceName, SourceUrl: in.SourceURL,
		SourceYear: in.SourceYear, Description: in.Description, Methodology: in.Methodology,
		Disclaimer: in.Disclaimer, Status: 1, IsDemo: in.IsDemo, CreatedBy: createdBy,
	})
	if err != nil {
		return Dataset{}, apperr.ErrInternal.Wrap(err)
	}
	return toDataset(d), nil
}

// ImportReport 导入结果。
type ImportReport struct {
	TotalRows   int              `json:"total_rows"`
	ValidRows   int              `json:"valid_rows"`
	InvalidRows int              `json:"invalid_rows"`
	Errors      []string         `json:"errors"`
	Imported    bool             `json:"imported"`
	Preview     []map[string]any `json:"preview"`
}

// ImportCSV 导入 CSV（dryRun 只校验不写库）。
func (s *Service) ImportCSV(ctx context.Context, datasetID uuid.UUID, r io.Reader, dryRun bool) (ImportReport, error) {
	ds, err := s.store.Q.GetDatasetByID(ctx, datasetID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ImportReport{}, apperr.ErrNotFound.WithMessage("数据集不存在")
	}
	if err != nil {
		return ImportReport{}, apperr.ErrInternal.Wrap(err)
	}

	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return ImportReport{}, apperr.Invalid("CSV 表头读取失败：文件为空或格式错误")
	}
	index := make(map[string]int, len(header))
	for i, h := range header {
		// 兼容 Excel/PowerShell 导出的 UTF-8 BOM
		key := strings.TrimPrefix(strings.TrimSpace(strings.ToLower(h)), "\ufeff")
		index[key] = i
	}
	for _, required := range CSVHeaders {
		if _, ok := index[required]; !ok {
			return ImportReport{}, apperr.Invalid("CSV 缺少必需列：" + required)
		}
	}

	report := ImportReport{Errors: []string{}, Preview: []map[string]any{}}
	rows := make([][]any, 0, 1024)
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return ImportReport{}, apperr.Invalid("CSV 解析失败：" + err.Error())
		}
		report.TotalRows++
		parsed, parseErr := parseRow(record, index)
		if parseErr != nil {
			report.InvalidRows++
			if len(report.Errors) < 20 {
				report.Errors = append(report.Errors, fmt.Sprintf("第 %d 行：%s", report.TotalRows+1, parseErr.Error()))
			}
			continue
		}
		if validErr := parsed.Validate(report.TotalRows + 1); validErr != nil {
			report.InvalidRows++
			if len(report.Errors) < 20 {
				report.Errors = append(report.Errors, validErr.Error())
			}
			continue
		}
		report.ValidRows++
		if len(report.Preview) < 10 {
			report.Preview = append(report.Preview, map[string]any{
				"gender": parsed.Gender, "age": parsed.Age, "city_code": parsed.CityCode,
				"height_cm": parsed.HeightCM, "education_level": parsed.EducationLevel,
				"annual_income_band": parsed.AnnualIncomeBand, "marital_status": parsed.MaritalStatus,
				"children_count": parsed.ChildrenCount, "wants_children": parsed.WantsChildren,
				"sample_weight": parsed.Weight,
			})
		}
		rows = append(rows, []any{
			ds.ID, parsed.Gender, parsed.Age, parsed.ProvinceCode, parsed.CityCode, parsed.HeightCM,
			parsed.EducationLevel, parsed.AnnualIncomeBand, parsed.MaritalStatus,
			parsed.ChildrenCount, parsed.WantsChildren, parsed.Weight,
		})
	}

	if report.ValidRows == 0 {
		return report, apperr.Invalid("CSV 中没有有效数据行")
	}
	if dryRun {
		return report, nil
	}

	err = s.store.InTx(ctx, func(tx *store.Store) error {
		if err := tx.Q.DeleteBaselineRecords(ctx, ds.ID); err != nil {
			return err
		}
		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{"baseline_records"},
			[]string{"dataset_id", "gender", "age", "province_code", "city_code", "height_cm",
				"education_level", "annual_income_band", "marital_status", "children_count",
				"wants_children", "sample_weight"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return err
		}
		return tx.Q.UpdateDatasetStats(ctx, sqlcdb.UpdateDatasetStatsParams{
			ID: ds.ID, RecordCount: int64(report.ValidRows), Status: enums.DatasetReady,
		})
	})
	if err != nil {
		return report, apperr.ErrInternal.Wrap(err)
	}
	report.Imported = true
	return report, nil
}

func parseRow(record []string, index map[string]int) (seed.BulkRecord, error) {
	get := func(key string) string {
		i, ok := index[key]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}
	parseInt := func(key string) (int, error) {
		v := get(key)
		if v == "" {
			return 0, nil
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("%s 不是整数：%s", key, v)
		}
		return n, nil
	}
	gender, err := parseInt("gender")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	age, err := parseInt("age")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	height, err := parseInt("height_cm")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	education, err := parseInt("education_level")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	income, err := parseInt("annual_income_band")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	marital, err := parseInt("marital_status")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	children, err := parseInt("children_count")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	wants, err := parseInt("wants_children")
	if err != nil {
		return seed.BulkRecord{}, err
	}
	weightStr := get("sample_weight")
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return seed.BulkRecord{}, fmt.Errorf("sample_weight 不是数字：%s", weightStr)
	}
	var heightPtr *int16
	if height > 0 {
		h := int16(height)
		heightPtr = &h
	}
	return seed.BulkRecord{
		Gender: int16(gender), Age: int16(age), ProvinceCode: get("province_code"),
		CityCode: get("city_code"), HeightCM: heightPtr, EducationLevel: int16(education),
		AnnualIncomeBand: int16(income), MaritalStatus: int16(marital),
		ChildrenCount: int16(children), WantsChildren: int16(wants), Weight: weight,
	}, nil
}

// Activate 激活数据集（同一时刻仅一个）。
func (s *Service) Activate(ctx context.Context, id uuid.UUID) error {
	ds, err := s.store.Q.GetDatasetByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.ErrNotFound.WithMessage("数据集不存在")
	}
	if err != nil {
		return apperr.ErrInternal.Wrap(err)
	}
	if ds.IsDemo && s.appEnv == "production" {
		return apperr.ErrForbidden.WithMessage("生产环境禁止激活演示数据集")
	}
	if ds.RecordCount == 0 {
		return apperr.ErrConflict.WithMessage("数据集没有记录，无法激活")
	}
	return s.store.InTx(ctx, func(tx *store.Store) error {
		if err := tx.Q.DeactivateAllDatasets(ctx); err != nil {
			return err
		}
		return tx.Q.SetDatasetActive(ctx, id)
	})
}

// Rollback 回滚数据集：删除全部记录并归档。
func (s *Service) Rollback(ctx context.Context, id uuid.UUID) error {
	return s.store.InTx(ctx, func(tx *store.Store) error {
		if err := tx.Q.DeleteBaselineRecords(ctx, id); err != nil {
			return err
		}
		if err := tx.Q.UpdateDatasetStats(ctx, sqlcdb.UpdateDatasetStatsParams{
			ID: id, RecordCount: 0, Status: enums.DatasetArchived,
		}); err != nil {
			return err
		}
		return tx.Q.DeactivateAllDatasets(ctx)
	})
}

// CreateDemoDataset 生成演示数据集（仅开发/测试环境）。
func (s *Service) CreateDemoDataset(ctx context.Context, count int, createdBy *uuid.UUID) (Dataset, error) {
	if s.appEnv == "production" {
		return Dataset{}, apperr.ErrForbidden.WithMessage("生产环境禁止生成演示数据")
	}
	if !s.settings.Bool(ctx, settings.KeyDemoDatasetEnabled, true) {
		return Dataset{}, apperr.ErrFeatureDisabled.WithMessage("演示数据集功能已关闭")
	}
	// 幂等：若已存在同名演示数据集且已有记录，直接复用，避免重复生成
	existing, err := s.store.Q.ListDatasets(ctx)
	if err == nil {
		for _, d := range existing {
			if d.Name == seed.DemoDatasetName && d.RecordCount > 0 {
				return toDataset(d), nil
			}
		}
	}
	ds, err := s.Create(ctx, CreateInput{
		Name: seed.DemoDatasetName, Version: "demo-1",
		SourceName: "虚构演示数据", SourceURL: "", SourceYear: nil,
		Description: "用于开发与演示的虚构加权样本数据集",
		Methodology: "随机生成，非任何真实统计调查",
		Disclaimer:  seed.DemoDisclaimer, IsDemo: true,
	}, createdBy)
	if err != nil {
		return Dataset{}, err
	}
	records := seed.ToBulk(seed.DemoDataset(count, 20260101))
	rows := make([][]any, 0, len(records))
	for _, r := range records {
		rows = append(rows, []any{
			ds.ID, r.Gender, r.Age, r.ProvinceCode, r.CityCode, r.HeightCM, r.EducationLevel,
			r.AnnualIncomeBand, r.MaritalStatus, r.ChildrenCount, r.WantsChildren, r.Weight,
		})
	}
	err = s.store.InTx(ctx, func(tx *store.Store) error {
		_, err := tx.CopyFrom(ctx,
			pgx.Identifier{"baseline_records"},
			[]string{"dataset_id", "gender", "age", "province_code", "city_code", "height_cm",
				"education_level", "annual_income_band", "marital_status", "children_count",
				"wants_children", "sample_weight"},
			pgx.CopyFromRows(rows),
		)
		if err != nil {
			return err
		}
		return tx.Q.UpdateDatasetStats(ctx, sqlcdb.UpdateDatasetStatsParams{
			ID: ds.ID, RecordCount: int64(len(rows)), Status: enums.DatasetReady,
		})
	})
	if err != nil {
		return Dataset{}, apperr.ErrInternal.Wrap(err)
	}
	return s.Get(ctx, ds.ID)
}

// 导入规则：CSV 必须包含全部必需列，逐行校验后整体替换数据集记录。