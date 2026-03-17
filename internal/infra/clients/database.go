package clients

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Database struct {
	*bun.DB
}

type DataClients struct {
	DB  *Database
	RDB *redis.Client
}

func newDatabase(source string, debug bool, pool *config.Database_Pool) (*Database, func(), error) {
	if err := validateDatabaseSource(source); err != nil {
		return nil, func() {}, err
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(source)))
	if pool != nil {
		sqldb.SetMaxIdleConns(int(pool.MaxIdleConns))
		sqldb.SetMaxOpenConns(int(pool.MaxOpenConns))
		if d, err := time.ParseDuration(pool.ConnMaxIdleTime); err == nil {
			sqldb.SetConnMaxIdleTime(d)
		}
		if d, err := time.ParseDuration(pool.ConnMaxLifetime); err == nil {
			sqldb.SetConnMaxLifetime(d)
		}
	} else {
		maxOpenConns := 4 * runtime.GOMAXPROCS(0)
		sqldb.SetMaxOpenConns(maxOpenConns)
		sqldb.SetMaxIdleConns(maxOpenConns)
	}

	db := bun.NewDB(sqldb, pgdialect.New())
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, func() {}, err
	}
	db = db.WithQueryHook(bundebug.NewQueryHook(bundebug.WithEnabled(debug)))
	return &Database{db}, func() {
		_ = db.Close()
	}, nil
}

func NewDataClients(cfg *config.AppConfig) (*DataClients, func(), error) {
	c := cfg.GetData()
	ctx := context.Background()
	db, closeDb, err := newDatabase(c.Database.Source, c.Database.Debug, c.Database.GetPool())
	if err != nil {
		return nil, nil, err
	}

	rdb := newRedis(c)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, err
	}
	return &DataClients{
			DB:  db,
			RDB: rdb,
		}, func() {
			closeDb()
			_ = rdb.Close()
		}, nil
}

func newRedis(cfg *config.Data) *redis.Client {
	c := cfg.Redis
	return redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       int(c.Db),
	})
}

// GetBunDB returns the underlying Bun DB instance
func (d *DataClients) GetBunDB() *bun.DB {
	return d.DB.DB
}

func validateDatabaseSource(source string) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return fmt.Errorf("database dsn is empty: set data.database.source or SKYLINE_DATA_DATABASE_SOURCE")
	}

	u, err := url.Parse(source)
	if err != nil {
		return fmt.Errorf("invalid database dsn %q: %w", source, err)
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return fmt.Errorf("invalid database dsn scheme %q: expected postgres or postgresql", u.Scheme)
	}
	return nil
}
