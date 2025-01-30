-- +goose Up
CREATE TABLE IF NOT EXISTS vpn_keys (
                                        id SERIAL PRIMARY KEY,
                                        key TEXT NOT NULL,
                                        is_used BOOLEAN DEFAULT false,
                                        user_id INT REFERENCES users(id),
                                        expires_at TIMESTAMP
);
-- +goose Down
DROP TABLE IF EXISTS vpn_keys;