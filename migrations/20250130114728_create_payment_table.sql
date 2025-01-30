-- +goose Up

CREATE TABLE IF NOT EXISTS payments (
                                        id SERIAL PRIMARY KEY,
                                        user_id INT REFERENCES users(id),
                                        amount NUMERIC(10,2) NOT NULL,
                                        status TEXT NOT NULL,
                                        payment_id TEXT NOT NULL,
                                        created_at TIMESTAMP DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS payments;