package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"vpn-bot/internal/domain"
)

type vpnKeyRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewVPNKeyRepository(db *pgxpool.Pool) VPNKeyRepository {
	return &vpnKeyRepositoryImpl{db: db}
}

func (r *vpnKeyRepositoryImpl) FindFreeKey() (*domain.VPNKey, error) {
	query := `SELECT id, key, is_used
              FROM vpn_keys
              WHERE is_used = false
              LIMIT 1`
	row := r.db.QueryRow(context.Background(), query)

	var vk domain.VPNKey
	err := row.Scan(&vk.ID, &vk.Key, &vk.IsUsed)
	if err != nil {
		return nil, errors.New("нет свободных VPN-ключей")
	}
	return &vk, nil
}

func (r *vpnKeyRepositoryImpl) AssignKeyToUser(keyID, userID int, expiresAt time.Time) error {
	query := `UPDATE vpn_keys
              SET is_used = true, user_id = $1, expires_at = $2
              WHERE id = $3`
	_, err := r.db.Exec(context.Background(), query, userID, expiresAt, keyID)
	return err
}

func (r *vpnKeyRepositoryImpl) GetKeysByTelegramID(telegramID int64) ([]domain.VPNKey, error) {
	query := `
        SELECT vk.id, vk.key, vk.is_used, vk.user_id, vk.expires_at
        FROM vpn_keys vk
        INNER JOIN users u ON vk.user_id = u.id
        WHERE u.telegram_id = $1
    `
	rows, err := r.db.Query(context.Background(), query, telegramID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []domain.VPNKey
	for rows.Next() {
		var k domain.VPNKey
		if err := rows.Scan(&k.ID, &k.Key, &k.IsUsed, &k.UserID, &k.ExpiresAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (r *vpnKeyRepositoryImpl) AddKey(key string) error {
	query := `INSERT INTO vpn_keys (key, is_used) VALUES ($1, false)`
	_, err := r.db.Exec(context.Background(), query, key)
	return err
}

func (r *vpnKeyRepositoryImpl) CountFreeKeys() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM vpn_keys WHERE is_used = false`
	err := r.db.QueryRow(context.Background(), query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
