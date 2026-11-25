package repository

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

var embedMigrations embed.FS


// DBTX is satisfied by both *pgxpool.Pool and pgx.Tx
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}

// OpenPool configures, pings, and runs Goose migrations, returning *pgxpool.Pool.
func OpenPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	/*
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse DATABASE_URL: %v", err)
	}*/

	// Optional: Adjust pool settings
	// poolConfig.MaxConns = 10
	// poolConfig.MinConns = 2

	/*
	// 3. Create the connection pool
	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)

	if err != nil {
		panic(err)
	}*/

	//defer dbPool.Close()

	// 4. Verify the connection
	//if err := dbPool.Ping(ctx); err != nil {
	//	log.Fatalf("Unable to ping database: %v", err)
	//}

	


	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.HealthCheckPeriod = 1 * time.Minute
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	/* pool already as a pointer here */
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	if err != nil {
		panic(err)
	}

	
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

/*
func runGooseMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Bridge *pgxpool.Pool to *sql.DB for Goose without creating a separate pool
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}

	if err := goose.UpContext(ctx, sqlDB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}*/
func RunMigration(
	ctx context.Context,
	pool *pgxpool.Pool,
	command string,
	args ...string,
) error {
	// Wrap pgxpool.Pool as a standard *sql.DB for Goose
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.RunContext(
		ctx, 
		command, 
		sqlDB, 
		"migrations", 
		args...
	)
}
