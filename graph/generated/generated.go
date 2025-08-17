package generated

import (
    "github.com/99designs/gqlgen/graphql"
    "github.com/hex-zero/MaxwellGoSpine/graph/resolver"
)

type Config struct { Resolvers *resolver.Resolver }

func NewExecutableSchema(cfg Config) graphql.ExecutableSchema { return nil }
