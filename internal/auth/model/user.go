package model

import (
	"time"

	"github.com/google/uuid"
)

// User model
type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"password"`
	CreatedAt      time.Time `json:"created_at"`
}

func (user *User) PopulateValues() *User {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	return user
}
