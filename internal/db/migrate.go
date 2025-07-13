package db

import "context"

// InitSchema creates the "messages" table in the database if it does not already exist.
// It returns an error if the table could not be created.
func InitSchema() error {
	// SQL statement to create the messages table with columns:
	// id (auto-incrementing primary key), chat_id, role, content, and created_at timestamp.
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		chat_id BIGINT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT now()
	);`

	// "Pool" refers to the database connection pool, which manages and reuses connections to the database efficiently.
	// Execute the SQL statement using the database connection pool.
	_, err := Pool.Exec(context.Background(), schema)

	return err
}