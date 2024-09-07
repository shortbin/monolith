package model

import (
	"time"

	"shortbin/pkg/config"
)

// Url model
type Url struct {
	ShortId   string    `json:"short_id"`
	LongUrl   string    `json:"long_url"`
	UserId    *string   `json:"user_id"` // *string as it can be null
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// PopulateValues populates the values of the url model
func (url *Url) PopulateValues() {
	url.CreatedAt = time.Now()
	url.ExpiresAt = time.Now().AddDate(
		config.GetConfig().ExpirationInYears,
		0,
		0,
	)
}
