package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func ConnectDB(ctx context.Context, logger *slog.Logger, databaseUrl string) *Repo {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		logger.Error("unable to parse database url", "error", err.Error())
		panic(err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Error("unable to create connection pool", "error", err.Error())
		panic(err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		logger.Error("unable to ping database", "error", err.Error())
		panic(err)
	}
	logger.Info("PostgreSQL connected successfully")
	return &Repo{
		pool: pool,
	}
}

func (r *Repo) CloseConn() {
	if r.pool != nil {
		r.pool.Close()
	}
}
