package domain

import "time"

type User struct {
	ID         int
	TelegramID int64
	Username   string
	ChatLink   string
	CreatedAt  time.Time
}
