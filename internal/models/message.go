package models

import "time"

type Message struct {
	ChatID    int64     // ID Telegram чата
	Role      string    // "user" или "assistant"
	Content   string    // текст сообщения
	CreatedAt time.Time // когда было отправлено сообщение
}
