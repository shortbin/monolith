package model

import (
	"time"
)

// URL model
type URL struct {
	ShortID   string    `json:"short_id"`
	LongURL   string    `json:"long_url"`
	UserID    *string   `json:"user_id"` // *string as it can be null
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
