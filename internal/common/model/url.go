package model

import "time"

// Url model
type Url struct {
	ShortId   string    `json:"short_id"`
	LongUrl   string    `json:"long_url"`
	UserId    *string   `json:"user_id"` // *string as it can be null
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
