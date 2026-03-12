package graphql

import (
	"net/http"

	"github/ijusttookadnatest/evm-indexer/internal/core/ports"
	"github/ijusttookadnatest/evm-indexer/internal/handlers/graphql/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/vektah/gqlparser/v2/ast"
)

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

func NewRouter(service ports.QueryService, playgroundEnabled bool) http.Handler {
	mux := http.NewServeMux()
	srv := NewHandler(service)

	if playgroundEnabled {
		mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql/playground"))
	}
	mux.Handle("/", graph.Middleware(service, srv))

	return mux
}
