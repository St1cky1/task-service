package metrics

import (
	"net/http"
	"time"
)

// HTTPMiddleware возвращает middleware для логирования HTTP метрик
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter для перехвата статуса
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Вызываем следующий handler
		next.ServeHTTP(wrapped, r)

		// Записываем метрики
		duration := time.Since(start).Seconds()
		metrics := GetMetrics()

		metrics.HTTPRequestsTotal.Inc()
		metrics.HTTPRequestDuration.Observe(duration)

		// Считаем ошибки (status >= 400)
		if wrapped.statusCode >= 400 {
			metrics.HTTPErrors.Inc()
		}
	})
}

// responseWriter обертка для отслеживания status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}