package models

import "time"

type FileResponse struct {
	UUId      string    `json:"uuid"`
	Filename  string    `json:"filename"`
	ExpiresAt time.Time `json:"expires"`
	Url       string    `json:"url"`
}
