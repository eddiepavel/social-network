package models

import "time"

type ValidatSession struct {
	UserID    []byte
	Active    bool
	ExpiresAt time.Time
}
