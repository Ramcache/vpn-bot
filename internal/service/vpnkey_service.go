package service

import (
	"errors"
	"log"
	"time"
	"vpn-bot/internal/domain"
	"vpn-bot/internal/repository"
)

type vpnKeyServiceImpl struct {
	repo repository.VPNKeyRepository
}

func NewVPNKeyService(r repository.VPNKeyRepository) VPNKeyService {
	return &vpnKeyServiceImpl{repo: r}
}

// Назначаем свободный ключ пользователю
func (s *vpnKeyServiceImpl) AssignFreeKeyToUser(userID int) (string, error) {
	key, err := s.repo.FindFreeKey()
	if err != nil {
		log.Println("❌ Ошибка при поиске VPN-ключа:", err)
		return "", err
	}
	if key == nil {
		log.Println("⚠️ Нет свободных VPN-ключей. Добавьте новые в базу!")
		return "", errors.New("нет свободных VPN-ключей")
	}

	// Привязываем ключ к пользователю
	err = s.repo.AssignKeyToUser(key.ID, userID, time.Now().Add(30*24*time.Hour)) // 30 дней
	if err != nil {
		log.Println("❌ Ошибка при назначении VPN-ключа:", err)
		return "", err
	}

	log.Printf("✅ VPN-ключ %s назначен пользователю %d", key.Key, userID)
	return key.Key, nil
}

// Получаем ключи по TelegramID
func (s *vpnKeyServiceImpl) GetKeysByUserTelegramID(telegramID int64) ([]domain.VPNKey, error) {
	return s.repo.GetKeysByTelegramID(telegramID)
}

func (s *vpnKeyServiceImpl) AddNewKey(key string) error {
	return s.repo.AddKey(key)
}

func (s *vpnKeyServiceImpl) HasFreeKeys() (bool, error) {
	count, err := s.repo.CountFreeKeys()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
