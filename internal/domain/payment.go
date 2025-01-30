package domain

import "time"

type Payment struct {
	ID        int
	UserID    int
	Amount    float64
	Status    string
	PaymentID string
	CreatedAt time.Time
}
