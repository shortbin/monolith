package model

import "time"

// Url model
type Url struct {
	Short     string    `json:"short"`
	Long      string    `json:"long"`
	UserId    *string   `json:"user_id"` // *string as it can be null
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
