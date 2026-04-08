package prometheus

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type IndexerMetrics struct {
	SyncedBlock prometheus.Counter
	BackfillError   prometheus.Counter
	ForwardfillError prometheus.Counter
	BalancefillError prometheus.Counter
	BackfillIsSyncing prometheus.Gauge
	ForwardfillIsSyncing prometheus.Gauge
	BalancefillIsSyncing prometheus.Gauge
	BackfillLastBlockId prometheus.Gauge
	ForwardfillLastBlockId prometheus.Gauge
	DurationFetchingBlock prometheus.Histogram
	DurationWritingBlockDB prometheus.Histogram
	DurationProcessingBlock prometheus.Histogram
}

type ApiMetrics struct {
	WsActiveConnection prometheus.Gauge
	WsTotalMessageSent prometheus.Counter
	RestProcessedRequest prometheus.Counter
	GraphqlProcessedRequest prometheus.Counter
	RestError prometheus.Counter
	GraphqlError prometheus.Counter
	DurationRestProcessingRequest prometheus.Histogram
	DurationGraphqlProcessingRequest prometheus.Histogram
}

func NewIndexerMetrics(reg prometheus.Registerer) *IndexerMetrics {
	m := &IndexerMetrics{
		SyncedBlock: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "indexer_blocks_synced_total",
			Help: "Total number of blocks synced",
		}),
		BackfillError: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "indexer_backfill_errors_total",
			Help: "Total number of backfill errors",
		}),
		ForwardfillError: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "indexer_forwardfill_errors_total",
			Help: "Total number of forwardfill errors",
		}),
		BalancefillError: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "indexer_balancefill_errors_total",
			Help: "Total number of balancefill errors",
		}),
		BackfillIsSyncing: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "indexer_backfill_syncing",
			Help: "1 if backfill is currently syncing, 0 otherwise",
		}),
		ForwardfillIsSyncing: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "indexer_forwardfill_syncing",
			Help: "1 if forwardfill is currently syncing, 0 otherwise",
		}),
		BalancefillIsSyncing: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "indexer_balancefill_syncing",
			Help: "1 if balancefill is currently syncing, 0 otherwise",
		}),
		BackfillLastBlockId: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "indexer_backfill_last_block_number",
			Help: "Last block number processed by the backfill service",
		}),
		ForwardfillLastBlockId: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "indexer_forwardfill_last_block_number",
			Help: "Last block number processed by the forwardfill service",
		}),
		DurationFetchingBlock: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "indexer_block_fetch_duration_seconds",
			Help:    "Duration of block fetching from RPC in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		DurationWritingBlockDB: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "indexer_block_db_write_duration_seconds",
			Help:    "Duration of block writing to database in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		DurationProcessingBlock: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "indexer_block_processing_duration_seconds",
			Help:    "Total duration of block processing (fetch + write) in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}

	return m
}

func NewApiMetrics(reg prometheus.Registerer) *ApiMetrics {
	m := &ApiMetrics{
		WsActiveConnection: promauto.With(reg).NewGauge(prometheus.GaugeOpts{
			Name: "api_ws_active_connections",
			Help: "Current number of active WebSocket connections",
		}),
		WsTotalMessageSent: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "api_ws_messages_sent_total",
			Help: "Total number of messages sent over WebSocket",
		}),
		RestProcessedRequest: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "api_rest_requests_total",
			Help: "Total number of REST requests processed",
		}),
		GraphqlProcessedRequest: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "api_graphql_requests_total",
			Help: "Total number of GraphQL requests processed",
		}),
		RestError: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "api_rest_errors_total",
			Help: "Total number of REST request errors",
		}),
		GraphqlError: promauto.With(reg).NewCounter(prometheus.CounterOpts{
			Name: "api_graphql_errors_total",
			Help: "Total number of GraphQL request errors",
		}),
		DurationRestProcessingRequest: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "api_rest_request_duration_seconds",
			Help:    "Duration of REST request processing in seconds",
			Buckets: prometheus.DefBuckets,
		}),
		DurationGraphqlProcessingRequest: promauto.With(reg).NewHistogram(prometheus.HistogramOpts{
			Name:    "api_graphql_request_duration_seconds",
			Help:    "Duration of GraphQL request processing in seconds",
			Buckets: prometheus.DefBuckets,
		}),
	}

	return m
}

func NewRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func NewPrometheusServer(reg *prometheus.Registry, port string) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	return &http.Server{
		Addr:   ":" + port,
		Handler: mux,
	}
}

func RunPrometheusServer(ctx context.Context, srv *http.Server) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.ListenAndServe()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}