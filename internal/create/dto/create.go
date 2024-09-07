package dto

import (
	"time"
)

type CreateReq struct {
	LongUrl   string     `json:"long_url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type CreateRes struct {
	ShortId   string    `json:"short_id"`
	LongUrl   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
