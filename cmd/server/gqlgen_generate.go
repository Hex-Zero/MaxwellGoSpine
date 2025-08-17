package main

//go:generate go run github.com/99designs/gqlgen generate

// This file adds a go:generate directive for gqlgen so running
//   go generate ./...
// will (re)generate the GraphQL schema code under graph/generated.
