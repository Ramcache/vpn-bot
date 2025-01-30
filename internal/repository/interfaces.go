package repository

import (
	"time"
	"vpn-bot/internal/domain"
)

type UserRepository interface {
	CreateUser(telegramID int64, username, chatLink string) error
	GetByTelegramID(telegramID int64) (*domain.User, error)
}

type VPNKeyRepository interface {
	FindFreeKey() (*domain.VPNKey, error)
	AssignKeyToUser(keyID, userID int, expiresAt time.Time) error
	GetKeysByTelegramID(telegramID int64) ([]domain.VPNKey, error)
	AddKey(key string) error
	CountFreeKeys() (int, error)
}

type PaymentRepository interface {
	CreatePayment(userID int, amount float64, status, paymentID string) error
	GetByPaymentID(paymentID string) (*domain.Payment, error)
	UpdatePaymentStatus(paymentID int, status string) error
}
