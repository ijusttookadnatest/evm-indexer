package graphql

import (
	"net/http"
	"time"

	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/graphql/graph"
	custmetrics "github/ijusttookadnatest/evm-indexer/internal/metrics"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(metrics *custmetrics.ApiMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rr, r)
		metrics.GraphqlProcessedRequest.Inc()
		metrics.DurationGraphqlProcessingRequest.Observe(time.Since(start).Seconds())
		if rr.status >= 400 {
			metrics.GraphqlError.Inc()
		}
	})
}

func NewHandler(service ports.QueryService) *handler.Server {
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{Service: service}}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return srv
}

func NewRouter(service ports.QueryService, playgroundEnabled bool, metrics *custmetrics.ApiMetrics) http.Handler {
	mux := http.NewServeMux()
	srv := NewHandler(service)

	if playgroundEnabled {
		mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql/playground"))
	}
	mux.Handle("/", graph.Middleware(service, srv))

	return metricsMiddleware(metrics, mux)
}
