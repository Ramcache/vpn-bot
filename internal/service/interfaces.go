package service

import "vpn-bot/internal/domain"

// UserService описывает логику работы с пользователями
type UserService interface {
	RegisterUser(telegramID int64, username, chatLink string) error
	GetUserByTelegramID(telegramID int64) (*domain.User, error)
}

// VPNKeyService описывает логику работы с VPN-ключами
type VPNKeyService interface {
	AssignFreeKeyToUser(userID int) (string, error)
	GetKeysByUserTelegramID(telegramID int64) ([]domain.VPNKey, error)
	AddNewKey(key string) error
	HasFreeKeys() (bool, error)
}

// PaymentService описывает логику оплаты
type PaymentService interface {
	CreatePayment(userID int, amount float64, description string) (string, error)
	ConfirmPayment(paymentID string) error
}
