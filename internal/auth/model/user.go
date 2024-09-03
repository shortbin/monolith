package model

import (
	"github.com/google/uuid"
	"time"
)

// User model
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

func (user *User) PopulateValues() *User {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	return user
}
