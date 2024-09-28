package model

import (
	"time"
)

// User model
type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"password"`
	CreatedAt      time.Time `json:"created_at"`
}
