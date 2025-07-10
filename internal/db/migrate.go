package db

import "context"


func InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		chat_id BIGINT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT now()
	);`

	_, err := Pool.Exec(context.Background(), schema)

	return err
}