package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nurashi/Newton/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateOrUpdate(ctx context.Context, telegramUser models.TelegramUser) (*models.User, error) {
	user := &models.User{}
	
	query := `
		INSERT INTO users (id, username, first_name, last_name, is_bot, language_code, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			is_bot = EXCLUDED.is_bot,
			language_code = EXCLUDED.language_code,
			updated_at = CURRENT_TIMESTAMP,
			last_seen = CURRENT_TIMESTAMP
		RETURNING id, username, first_name, last_name, is_bot, language_code, created_at, updated_at, last_seen`
	
	err := r.db.QueryRow(ctx, query,
		telegramUser.ID,
		telegramUser.Username,
		telegramUser.FirstName,
		telegramUser.LastName,
		telegramUser.IsBot,
		telegramUser.LanguageCode,
	).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsBot,
		&user.LanguageCode,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastSeen,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create or update user: %w", err)
	}
	
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	user := &models.User{}
	
	query := `SELECT id, username, first_name, last_name, is_bot, language_code, created_at, updated_at, last_seen FROM users WHERE id = $1`
	 
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.IsBot,
		&user.LanguageCode,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastSeen,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return user, nil
}

func (r *UserRepository) UpdateLastSeen(ctx context.Context, userID int64) error {
	query := `UPDATE users SET last_seen = CURRENT_TIMESTAMP WHERE id = $1`
	
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last seen: %w", err)
	}
	
	return nil
}

func (r *UserRepository) GetStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	stats := map[string]interface{}{
		"user_id":      user.ID,
		"username":     user.Username,
		"first_name":   user.FirstName,
		"member_since": user.CreatedAt.Format("2006-01-02"),
		"last_seen":    user.LastSeen.Format("2006-01-02 15:04:05"),
	}
	
	return stats, nil
}