package server

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/hex-zero/MaxwellGoSpine/graph/generated"
	"github.com/hex-zero/MaxwellGoSpine/graph/resolver"
)

// NewServer builds a gqlgen handler server with provided root resolver deps.
func NewExecutableSchema(r *resolver.Resolver) *handler.Server {
	cfg := generated.Config{Resolvers: r}
	return handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
}

func PlaygroundHandler() http.Handler { return playground.Handler("GraphQL", "/v1/graphql") }
