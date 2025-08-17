package model

//go:generate go run github.com/99designs/gqlgen generate

// User GraphQL model mirrors core user entity but flattens time fields to RFC3339 strings.
type User struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	DeletedAt *string `json:"deletedAt"`
}
