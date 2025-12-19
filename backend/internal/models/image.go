package models

import "time"

type FileResponse struct {
	UUId      string    `json:"uuid"`
	ExpiresAt time.Time `json:"expires"`
}
