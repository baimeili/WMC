// Package store 聚合数据库访问：sqlc 生成查询 + 少量动态 SQL。
package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wmc/backend/internal/db/sqlcdb"
)

// Store 数据库访问入口。
type Store struct {
	Pool *pgxpool.Pool
	Tx   pgx.Tx
	Q    *sqlcdb.Queries
}

// New 构造 Store。
func New(pool *pgxpool.Pool) *Store {
	return &Store{Pool: pool, Q: sqlcdb.New(pool)}
}

// InTx 在事务中执行。
func (s *Store) InTx(ctx context.Context, fn func(*Store) error) error {
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("开启事务失败: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(NewTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// NewTx 基于事务构造 Store。
func NewTx(tx pgx.Tx) *Store {
	return &Store{Tx: tx, Q: sqlcdb.New(tx)}
}

// CopyFrom 批量写入，自动区分连接池与事务。
func (s *Store) CopyFrom(ctx context.Context, table pgx.Identifier, columns []string, rows pgx.CopyFromSource) (int64, error) {
	if s.Tx != nil {
		return s.Tx.CopyFrom(ctx, table, columns, rows)
	}
	return s.Pool.CopyFrom(ctx, table, columns, rows)
}

// LoadSettings 实现 settings.Loader：返回全部系统参数原始 JSON。
func (s *Store) LoadSettings(ctx context.Context) (map[string]json.RawMessage, error) {
	rows, err := s.Q.ListSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("查询系统参数失败: %w", err)
	}
	out := make(map[string]json.RawMessage, len(rows))
	for _, r := range rows {
		out[r.Key] = json.RawMessage(r.Value)
	}
	return out, nil
}