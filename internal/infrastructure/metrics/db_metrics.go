package metrics

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDBMetrics инициализирует мониторинг метрик БД
func InitDBMetrics(db *pgxpool.Pool) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if db != nil {
				stats := db.Stat()
				metrics := GetMetrics()
				metrics.DBConnections.Set(float64(stats.AcquiredConns()))
				metrics.DBConnectionsPool.Set(float64(stats.TotalConns()))
			}
		}
	}()
}
