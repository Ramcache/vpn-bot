package service

import "vpn-bot/internal/domain"

type UserService interface {
	RegisterUser(telegramID int64, username, chatLink string) error
	GetUserByTelegramID(telegramID int64) (*domain.User, error)
}

type VPNKeyService interface {
	AssignFreeKeyToUser(userID int) (string, error)
	GetKeysByUserTelegramID(telegramID int64) ([]domain.VPNKey, error)
	AddNewKey(key string) error
	HasFreeKeys() (bool, error)
}

type PaymentService interface {
	CreatePayment(userID int, amount float64, description string) (string, error)
	ConfirmPayment(paymentID string) error
}
