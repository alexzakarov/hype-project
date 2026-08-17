package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"main/config"
)

var (
	ConnStr string
)

const (
	maxOpenConns    = 250
	connMaxLifetime = 120
	maxIdleConns    = 30
	connMaxIdleTime = 20
)

// NewPostgresDB Return new Postgres client
func NewPostgresDB(cfg *config.Config) (db *pgxpool.Pool, err error) {
	println("Driver PostgreSQL Initialized")
	ConnStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s pool_max_conns=%d", cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DefaultDb, cfg.Postgres.MaxConnections)

	db, err = pgxpool.Connect(context.Background(), ConnStr)
	if err != nil {
		println(err.Error())
		return
	} else {
		print("conn ok")
	}

	if err = db.Ping(context.Background()); err != nil {
		println(err.Error())
	}

	return
}
