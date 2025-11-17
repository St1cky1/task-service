package client

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	Pool *pgxpool.Pool
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresClient(cfg Config) (*PostgresClient, error) {
	connString := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Настройки пула соединений
	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresClient{Pool: pool}, nil
}

func (c *PostgresClient) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}

// HealthCheck проверяет состояние базы данных
func (c *PostgresClient) HealthCheck(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}

// GetPool возвращает пул соединений для использования в репозиториях
func (c *PostgresClient) GetPool() *pgxpool.Pool {
	return c.Pool
}
