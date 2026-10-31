package calculator

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQLCounter 基于 PostgreSQL 加权聚合的条件计数器。
// 条件为动态拼接，参数全部通过占位符传递，避免 SQL 注入。
type SQLCounter struct {
	pool      *pgxpool.Pool
	datasetID uuid.UUID
	total     float64
	hasTotal  bool
}

// NewSQLCounter 构造 SQL 计数器。
func NewSQLCounter(pool *pgxpool.Pool, datasetID uuid.UUID) *SQLCounter {
	return &SQLCounter{pool: pool, datasetID: datasetID}
}

// Total 返回数据集总权重。
func (c *SQLCounter) Total(ctx context.Context) (float64, error) {
	if c.hasTotal {
		return c.total, nil
	}
	var total float64
	err := c.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(sample_weight), 0)::float8 FROM baseline_records WHERE dataset_id = $1`,
		c.datasetID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("统计基线总权重失败: %w", err)
	}
	c.total, c.hasTotal = total, true
	return total, nil
}

// Count 返回满足条件的样本权重之和。
func (c *SQLCounter) Count(ctx context.Context, f Filter) (float64, error) {
	var (
		sb   strings.Builder
		args []any
	)
	sb.WriteString(`SELECT COALESCE(SUM(sample_weight), 0)::float8 FROM baseline_records WHERE dataset_id = $1`)
	args = append(args, c.datasetID)

	add := func(cond string, value any) {
		args = append(args, value)
		fmt.Fprintf(&sb, " AND %s($%d)", cond, len(args))
	}

	if len(f.Genders) > 0 {
		add("gender = ANY", f.Genders)
	}
	if f.MinAge != nil {
		add("age >=", *f.MinAge)
	}
	if f.MaxAge != nil {
		add("age <=", *f.MaxAge)
	}
	if len(f.Cities) > 0 {
		add("city_code = ANY", f.Cities)
	}
	if len(f.MaritalStatuses) > 0 {
		add("marital_status = ANY", f.MaritalStatuses)
	}
	if len(f.ChildrenStates) > 0 {
		add("(CASE WHEN children_count > 0 THEN 1 ELSE 0 END) = ANY", f.ChildrenStates)
	}
	if len(f.WantsChildren) > 0 {
		add("wants_children = ANY", f.WantsChildren)
	}
	if f.MinHeight != nil {
		add("height_cm IS NOT NULL AND height_cm >=", *f.MinHeight)
	}
	if f.MaxHeight != nil {
		add("height_cm IS NOT NULL AND height_cm <=", *f.MaxHeight)
	}
	if f.MinEducation != nil {
		add("education_level >=", *f.MinEducation)
	}
	if f.MinIncome != nil {
		// 收入区间 8 = 不愿透露，不能当作高收入参与比较
		add("annual_income_band BETWEEN 1 AND 7 AND annual_income_band >=", *f.MinIncome)
	}

	var total float64
	if err := c.pool.QueryRow(ctx, sb.String(), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("统计基线候选权重失败: %w", err)
	}
	return total, nil
}
