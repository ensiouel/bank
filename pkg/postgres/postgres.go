package postgres

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

func (config Config) String() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", config.User, config.Password, config.Host,
		config.Port, config.DB)
}

type Client interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type Tx interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	LargeObjects() pgx.LargeObjects
	Prepare(ctx context.Context, name string, sql string) (*pgconn.StatementDescription, error)
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Conn() *pgx.Conn
}

type client struct {
	*pgxpool.Pool
}

type Txw struct {
	pgx.Tx
}

func (t Txw) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, t.Tx, dest, query, args...)
}

func (t Txw) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, t.Tx, dest, query, args...)
}

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	config, err := pgxpool.ParseConfig(cfg.String())
	if err != nil {
		return nil, err
	}

	var pool *pgxpool.Pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &client{
		Pool: pool,
	}, nil
}

func (c *client) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.Pool.Begin(ctx)
}

func (c *client) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return c.Pool.Exec(ctx, query, args...)
}

func (c *client) Get(ctx context.Context, dest interface{}, query string, args ...any) error {
	return pgxscan.Get(ctx, c.Pool, dest, query, args...)
}

func (c *client) Select(ctx context.Context, dest interface{}, query string, args ...any) error {
	return pgxscan.Select(ctx, c.Pool, dest, query, args...)
}

func (c *client) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	return c.Pool.Query(ctx, query, args...)
}

func (c *client) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return c.Pool.QueryRow(ctx, query, args...)
}
