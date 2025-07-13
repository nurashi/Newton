package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/OpenRouterProject/internal/db"
	"github.com/nurashi/OpenRouterProject/internal/models"
)



type MessageRepository struct {
	conn *pgxpool.Pool
}


func NewMessageRepository() *MessageRepository {
	if db.Pool == nil {
		panic("database pool is not initialized")
	}
	return &MessageRepository{
		conn: db.Pool,
	}
}


func (r *MessageRepository) SaveMessage(ctx context.Context, msg *models.Message) error {
	query := `INSERT INTO messages (chat_id, role, content, created_at) VALUES ($1, $2, $3, $4)`
 
	_, err := r.conn.Exec(ctx, query, msg.ChatID, msg.Role, msg.Content, msg.CreatedAt)
	
	return err
}

func (r *MessageRepository) GetLastMessages(ctx context.Context, chatID int64, limit int) ([]*models.Message, error) {
	query := `SELECT chat_id, role, content, created_at 
		FROM messages WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.conn.Query(ctx, query, chatID, limit)
	if err != nil {
		return nil, err
	}

	var messages []*models.Message

	for rows.Next(){ 
		var msg models.Message

		if err := rows.Scan(&msg.ChatID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil { 
			return nil, err
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}