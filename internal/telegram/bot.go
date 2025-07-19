package telegram

import (
	"context"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nurashi/OpenRouterProject/internal/ai"
	"github.com/nurashi/OpenRouterProject/internal/models"
	"github.com/nurashi/OpenRouterProject/internal/repository"
)

var repo = repository.NewMessageRepository()

func RunTelegramBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN not found")
	}

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Bot init failed:", err)
	}
	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		prompt := update.Message.Text

		ctx := context.Background()

		msg := tgbotapi.NewMessage(chatID, "Thinking...")
		sent, err := bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send thinking message: %v", err)
			continue
		}

		// load last 5 messages
		history, err := repo.GetLastMessages(ctx, chatID, 5)
		if err != nil {
			log.Printf("Failed to load history for chat %d: %v", chatID, err)
			history = []*models.Message{}
		}

		// context messages
		var contextMessages []ai.Message
		for i := len(history) - 1; i >= 0; i-- { // reverse order so older messages come first
			contextMessages = append(contextMessages, ai.Message{
				Role:    history[i].Role,
				Content: history[i].Content,
			})
		}

		// new user message adding
		contextMessages = append(contextMessages, ai.Message{
			Role:    "user",
			Content: prompt,
		})

		response, err := ai.AskWithHistory(contextMessages)
		if err != nil {
			response = "ERROR: " + err.Error()
			log.Printf("AI request failed: %v", err)
		}

		userMsg := &models.Message{
			ChatID:    chatID,
			Role:      "user",
			Content:   prompt,
			CreatedAt: time.Now(),
		}
		
		if err := repo.SaveMessage(ctx, userMsg); err != nil {
			log.Printf("ERROR: Failed to save user message for chat %d: %v", chatID, err)
		} else {
			log.Printf("Successfully saved user message for chat %d", chatID)
		}

		aiMsg := &models.Message{
			ChatID:    chatID,
			Role:      "assistant",
			Content:   response,
			CreatedAt: time.Now(),
		}
		
		if err := repo.SaveMessage(ctx, aiMsg); err != nil {
			log.Printf("ERROR: Failed to save AI message for chat %d: %v", chatID, err)
		} else {
			log.Printf("Succesfully saved AI message for chat %d", chatID)
		}

		edit := tgbotapi.NewEditMessageText(chatID, sent.MessageID, response)
		if _, err := bot.Send(edit); err != nil {
			log.Printf("Failed to edit message: %v", err)
		}
	}
}