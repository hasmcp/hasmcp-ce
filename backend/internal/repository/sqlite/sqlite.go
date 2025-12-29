package sqlite

import (
	"context"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/hasmcp/hasmcp-ce/backend/internal/repository/base"
	"github.com/hasmcp/hasmcp-ce/backend/internal/service/config"
	zlog "github.com/rs/zerolog/log"

	"gorm.io/gorm"
)

type (
	Params struct {
		Config config.Service
	}

	repository struct {
		db *gorm.DB
	}

	sqliteConfig struct {
		Enabled         bool          `yaml:"enabled"`
		DSN             string        `yaml:"dsn"`
		MaxIdleConns    int           `yaml:"maxIdleConns"`
		MaxOpenConns    int           `yaml:"maxOpenConns"`
		MaxConnLifetime time.Duration `yaml:"maxConnLifetime"`
	}
)

const (
	_cfgKey = "sqlite"

	_logPrefix = "[sqlite] "
)

func New(p Params) (base.Repository, error) {
	var cfg sqliteConfig

	err := p.Config.Populate(_cfgKey, &cfg)
	if err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		zlog.Warn().Msg(_logPrefix + "sqlite repository is not enabled, skipping")
		return nil, nil
	}

	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{TranslateError: true})
	if err != nil {
		zlog.Error().Err(err).Msg(_logPrefix + "failed to connect")
		return nil, err
	}

	zlog.Info().Msg(_logPrefix + "connected")
	zlog.Debug().Str("dsn", cfg.DSN).Msg(_logPrefix + "connected")

	dbi, err := db.DB()
	if err != nil {
		zlog.Error().Err(err).Msg(_logPrefix + "failed to get database connection")
		return nil, err
	}

	dbi.SetMaxIdleConns(cfg.MaxIdleConns)
	dbi.SetMaxOpenConns(cfg.MaxOpenConns)
	dbi.SetConnMaxLifetime(cfg.MaxConnLifetime)
	dbi.Stats()
	r := &repository{
		db: db,
	}

	return r, nil
}

func (r *repository) Conn(ctx context.Context) *gorm.DB {
	if ctx.Value(base.CtxKeyDBTx) != nil {
		return ctx.Value(base.CtxKeyDBTx).(*gorm.DB)
	}
	return r.db
}

func (r *repository) TxBegin(ctx context.Context) *gorm.DB {
	return r.db.Begin()
}

func (r *repository) TxCommit(ctx context.Context) error {
	return r.Conn(ctx).Commit().Error
}

func (r *repository) TxRollback(ctx context.Context) error {
	return r.Conn(ctx).Rollback().Error
}

func (r *repository) Close(ctx context.Context) {
	dbi, err := r.db.DB()
	if err != nil {
		return
	}
	dbi.Close()
}
