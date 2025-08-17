package server

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/hex-zero/MaxwellGoSpine/graph/generated"
	"github.com/hex-zero/MaxwellGoSpine/graph/resolver"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
)

func NewExecutableSchema(userSvc core.UserService) *handler.Server {
	base := &resolver.Resolver{UserService: userSvc}
	cfg := generated.Config{Resolvers: base}
	return handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
}

func PlaygroundHandler() http.Handler {
	return playground.Handler("GraphQL", "/v1/graphql")
}
