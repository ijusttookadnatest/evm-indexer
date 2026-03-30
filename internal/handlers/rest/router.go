package rest

import (
	"net/http"
	"time"

	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	custprometheus "github/ijusttookadnatest/evm-indexer/internal/prometheus"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(metrics *custprometheus.ApiMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)
		metrics.RestProcessedRequest.Inc()
		metrics.DurationRestProcessingRequest.Observe(time.Since(start).Seconds())
		if rr.status >= 400 {
			metrics.RestError.Inc()
		}
	})
}

func NewRouter(service ports.QueryService, metrics *custprometheus.ApiMetrics) http.Handler {
	mux := http.NewServeMux()
	handler := NewHandler(service)

	mux.HandleFunc("GET /blocks", handler.GetBlock)
	mux.HandleFunc("GET /transactions", handler.GetTransaction)
	mux.HandleFunc("GET /events", handler.GetEvent)
	mux.HandleFunc("GET /events/log", handler.GetEventByTxLog)

	return metricsMiddleware(metrics, mux)
}
