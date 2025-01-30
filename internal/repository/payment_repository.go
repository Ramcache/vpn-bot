package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"vpn-bot/internal/domain"
)

type paymentRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) PaymentRepository {
	return &paymentRepositoryImpl{db: db}
}

func (r *paymentRepositoryImpl) CreatePayment(userID int, amount float64, status, paymentID string) error {
	query := `INSERT INTO payments (user_id, amount, status, payment_id, created_at)
              VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.Exec(context.Background(), query, userID, amount, status, paymentID)
	return err
}

func (r *paymentRepositoryImpl) GetByPaymentID(paymentID string) (*domain.Payment, error) {
	query := `SELECT id, user_id, amount, status, payment_id, created_at
              FROM payments
              WHERE payment_id = $1
              LIMIT 1`
	row := r.db.QueryRow(context.Background(), query, paymentID)

	var p domain.Payment
	err := row.Scan(&p.ID, &p.UserID, &p.Amount, &p.Status, &p.PaymentID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepositoryImpl) UpdatePaymentStatus(paymentID int, status string) error {
	query := `UPDATE payments SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(context.Background(), query, status, paymentID)
	return err
}
