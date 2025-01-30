package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"vpn-bot/internal/domain"
)

type userRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepositoryImpl{db: db}
}

func (r *userRepositoryImpl) CreateUser(telegramID int64, username, chatLink string) error {
	query := `INSERT INTO users (telegram_id, username, chat_link, created_at)
              VALUES ($1, $2, $3, NOW())`
	_, err := r.db.Exec(context.Background(), query, telegramID, username, chatLink)
	return err
}

func (r *userRepositoryImpl) GetByTelegramID(telegramID int64) (*domain.User, error) {
	query := `SELECT id, telegram_id, username, chat_link, created_at
              FROM users WHERE telegram_id = $1 LIMIT 1`
	row := r.db.QueryRow(context.Background(), query, telegramID)

	var u domain.User
	err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.ChatLink, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
