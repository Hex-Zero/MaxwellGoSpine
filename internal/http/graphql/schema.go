package graphql

import (
	"context"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
)

// BuildSchema constructs a GraphQL schema using the existing UserService.
func BuildSchema(userSvc core.UserService) (graphql.Schema, error) {
	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "User",
		Fields: graphql.Fields{
			"id":        &graphql.Field{Type: graphql.NewNonNull(graphql.ID)},
			"name":      &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"email":     &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"createdAt": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"updatedAt": &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"deletedAt": &graphql.Field{Type: graphql.String},
		},
	})

	// Query root
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"users": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(userType))),
				Args: graphql.FieldConfigArgument{
					"page":     &graphql.ArgumentConfig{Type: graphql.Int},
					"pageSize": &graphql.ArgumentConfig{Type: graphql.Int},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					page, _ := p.Args["page"].(int)
					pageSize, _ := p.Args["pageSize"].(int)
					if page == 0 {
						page = 1
					}
					if pageSize == 0 {
						pageSize = 20
					}
					users, _, err := userSvc.List(p.Context, page, pageSize)
					if err != nil {
						return nil, err
					}
					out := make([]map[string]any, 0, len(users))
					for _, u := range users {
						out = append(out, convertUser(u))
					}
					return out, nil
				},
			},
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					idStr := p.Args["id"].(string)
					uid, err := uuid.Parse(idStr)
					if err != nil {
						return nil, err
					}
					u, err := userSvc.Get(p.Context, uid)
					if err != nil {
						return nil, err
					}
					return convertUser(u), nil
				},
			},
		},
	})

	// Mutation root
	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"email": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					name := p.Args["name"].(string)
					email := p.Args["email"].(string)
					u, err := userSvc.Create(p.Context, name, email)
					if err != nil {
						return nil, err
					}
					return convertUser(u), nil
				},
			},
			"updateUser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id":    &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
					"name":  &graphql.ArgumentConfig{Type: graphql.String},
					"email": &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					idStr := p.Args["id"].(string)
					uid, err := uuid.Parse(idStr)
					if err != nil {
						return nil, err
					}
					var namePtr, emailPtr *string
					if v, ok := p.Args["name"].(string); ok {
						namePtr = &v
					}
					if v, ok := p.Args["email"].(string); ok {
						emailPtr = &v
					}
					u, err := userSvc.Update(p.Context, uid, namePtr, emailPtr)
					if err != nil {
						return nil, err
					}
					return convertUser(u), nil
				},
			},
			"deleteUser": &graphql.Field{
				Type: graphql.NewNonNull(graphql.Boolean),
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					idStr := p.Args["id"].(string)
					uid, err := uuid.Parse(idStr)
					if err != nil {
						return nil, err
					}
					if err := userSvc.Delete(p.Context, uid); err != nil {
						return nil, err
					}
					return true, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{Query: queryType, Mutation: mutationType})
}

func convertUser(u *core.User) map[string]any {
	if u == nil {
		return nil
	}
	m := map[string]any{
		"id":        u.ID.String(),
		"name":      u.Name,
		"email":     u.Email,
		"createdAt": u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		"updatedAt": u.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if u.DeletedAt != nil {
		m["deletedAt"] = u.DeletedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return m
}

// Execute helper for tests or handlers
func Execute(ctx context.Context, schema graphql.Schema, query string, variables map[string]any) *graphql.Result {
	return graphql.Do(graphql.Params{Schema: schema, RequestString: query, VariableValues: variables, Context: ctx})
}
