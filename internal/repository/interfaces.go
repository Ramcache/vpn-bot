package repository

import (
	"time"
	"vpn-bot/internal/domain"
)

// UserRepository описывает методы работы с таблицей users
type UserRepository interface {
	CreateUser(telegramID int64, username, chatLink string) error
	GetByTelegramID(telegramID int64) (*domain.User, error)
}

// VPNKeyRepository описывает методы работы с таблицей vpn_keys
type VPNKeyRepository interface {
	FindFreeKey() (*domain.VPNKey, error)
	AssignKeyToUser(keyID, userID int, expiresAt time.Time) error
	GetKeysByTelegramID(telegramID int64) ([]domain.VPNKey, error)
	AddKey(key string) error
	CountFreeKeys() (int, error)
}

// PaymentRepository описывает методы работы с таблицей payments
type PaymentRepository interface {
	CreatePayment(userID int, amount float64, status, paymentID string) error
	GetByPaymentID(paymentID string) (*domain.Payment, error)
	UpdatePaymentStatus(paymentID int, status string) error
}
