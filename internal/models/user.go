package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     *string   `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     *string   `json:"last_name"`
	IsBot        bool      `json:"is_bot"`
	LanguageCode *string   `json:"language_code"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastSeen     time.Time `json:"last_seen"`
}
type TelegramUser struct {
	ID           int64   `json:"id"`
	Username     *string `json:"username"`
	FirstName    string  `json:"first_name"`
	LastName     *string `json:"last_name"`
	IsBot        bool    `json:"is_bot"`
	LanguageCode *string `json:"language_code"`
}
