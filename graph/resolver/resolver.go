package resolver

import "github.com/hex-zero/MaxwellGoSpine/internal/core"

// Resolver serves as dependency injection for your app, add services here.
type Resolver struct {
	UserService core.UserService
}

// Root objects for gqlgen default layout
type Query struct{ *Resolver }
type Mutation struct{ *Resolver }
