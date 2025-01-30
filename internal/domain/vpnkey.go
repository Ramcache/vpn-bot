package domain

import "time"

type VPNKey struct {
	ID        int
	Key       string
	IsUsed    bool
	UserID    *int
	ExpiresAt *time.Time
}
