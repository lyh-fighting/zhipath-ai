package platform

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
)

// RelationalStore 抽象 MySQL 访问。MySQL 是业务事实源。
type RelationalStore interface {
	Ping(ctx context.Context) error
	DB() *sql.DB
}

// CacheStore 抽象 Redis 访问（短期记忆、限流、幂等）。
type CacheStore interface {
	Ping(ctx context.Context) error
}

type mysqlStore struct {
	db *sql.DB
}

// NewMySQLStore 按 DSN 建立 MySQL 连接，连接失败立即返回错误。
func NewMySQLStore(dsn string) (RelationalStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("platform: 打开 MySQL 失败: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("platform: 连接 MySQL 失败: %w", err)
	}
	return &mysqlStore{db: db}, nil
}

func (m *mysqlStore) Ping(ctx context.Context) error { return m.db.PingContext(ctx) }
func (m *mysqlStore) DB() *sql.DB                    { return m.db }

type redisStore struct {
	rdb *redis.Client
}

// NewRedisStore 按 URL 建立 Redis 连接。
func NewRedisStore(url string) (CacheStore, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("platform: 解析 Redis URL 失败: %w", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("platform: 连接 Redis 失败: %w", err)
	}
	return &redisStore{rdb: rdb}, nil
}

func (r *redisStore) Ping(ctx context.Context) error { return r.rdb.Ping(ctx).Err() }
