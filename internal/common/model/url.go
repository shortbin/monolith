package model

import (
	"time"

	"shortbin/pkg/config"
)

// URL model
type URL struct {
	ShortID   string    `json:"short_id"`
	LongURL   string    `json:"long_url"`
	UserID    *string   `json:"user_id"` // *string as it can be null
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PopulateValues populates the values of the url model
func (url *URL) PopulateValues() {
	url.CreatedAt = time.Now()
	if url.ExpiresAt.IsZero() {
		url.ExpiresAt = time.Now().AddDate(
			config.GetConfig().ExpirationInYears,
			0,
			0,
		)
	}
}
