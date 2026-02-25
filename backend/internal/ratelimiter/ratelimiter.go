package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.RWMutex
	Requests map[string]*RequestInfo
	Rate     int
	Window   time.Duration
}

type RequestInfo struct {
	Count       int
	WindowStart time.Time
	ExpiresAt   time.Time
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		Requests: make(map[string]*RequestInfo),
		Rate:     120,
		Window:   120 * time.Second,
	}

	go rl.cleanupExpired()

	return rl
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	info, exists := rl.Requests[key]

	if !exists || now.After(info.ExpiresAt) {
		rl.Requests[key] = &RequestInfo{
			Count:       1,
			WindowStart: now,
			ExpiresAt:   now.Add(rl.Window),
		}
		return true
	}

	if now.Sub(info.WindowStart) > rl.Window {

		info.Count = 1
		info.WindowStart = now
		info.ExpiresAt = now.Add(rl.Window)
		return true
	}


	if info.Count >= rl.Rate {
		return false 
	}

	info.Count++
	return true
}


func (rl *RateLimiter) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.Requests {
			if now.After(info.ExpiresAt) {
				delete(rl.Requests, key)
			}
		}
		rl.mu.Unlock()
	}
}
