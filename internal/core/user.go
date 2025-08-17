package core

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
