package service

import (
	"fmt"
	"vpn-bot/internal/domain"
	"vpn-bot/internal/repository"
)

type userServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userServiceImpl{repo: r}
}

func (s *userServiceImpl) RegisterUser(telegramID int64, username, chatLink string) error {
	_, err := s.repo.GetByTelegramID(telegramID)
	if err == nil {
		return nil
	}
	err = s.repo.CreateUser(telegramID, username, chatLink)
	if err != nil {
		return err
	}
	fmt.Println("Новый пользователь зарегистрирован:", telegramID)
	return nil
}

func (s *userServiceImpl) GetUserByTelegramID(telegramID int64) (*domain.User, error) {
	return s.repo.GetByTelegramID(telegramID)
}
