package bun

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/infra/clients"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bunotel"
)

// DB wraps the Bun DB client
type DB struct {
	*bun.DB
}

// NewDB creates a new Bun DB from existing Database
// Since Database already embeds *bun.DB, we just wrap it
func NewDB(dataDB *clients.Database) *DB {
	// Add query hooks for observability to the existing DB
	dataDB.AddQueryHook(bunotel.NewQueryHook())

	return &DB{DB: dataDB.DB}
}

// NewDBWithConfig creates a new Bun DB with direct configuration
// Use this if you want to use Bun independently of Ent
func NewDBWithConfig(driver, dsn string, debug bool, pool *PoolConfig) (*DB, func(), error) {
	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("opening sql db: %w", err)
	}

	if pool != nil {
		if pool.MaxOpenConns > 0 {
			sqlDB.SetMaxOpenConns(pool.MaxOpenConns)
		}
		if pool.MaxIdleConns > 0 {
			sqlDB.SetMaxIdleConns(pool.MaxIdleConns)
		}
		if pool.ConnMaxLifetime > 0 {
			sqlDB.SetConnMaxLifetime(pool.ConnMaxLifetime)
		}
		if pool.ConnMaxIdleTime > 0 {
			sqlDB.SetConnMaxIdleTime(pool.ConnMaxIdleTime)
		}
	}

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	if debug {
		bunDB.AddQueryHook(&DebugQueryHook{})
	}

	bunDB.AddQueryHook(bunotel.NewQueryHook())

	return &DB{DB: bunDB}, func() {
		_ = sqlDB.Close()
	}, nil
}

// PoolConfig defines database connection pool settings
type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DebugQueryHook logs all SQL queries for debugging
type DebugQueryHook struct{}

func (h *DebugQueryHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
	return ctx
}

func (h *DebugQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	if event.Err != nil {
		log.Printf("[BUN ERROR] %s: %v", event.Query, event.Err)
	} else {
		log.Printf("[BUN] %s %s", time.Since(event.StartTime).Round(time.Millisecond), event.Query)
	}
}
