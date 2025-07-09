package telegram

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nurashi/internal/ai"
)

func RunTelegramBot() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(botToken)

	if err != nil {
		panic(err)
	}

	bot.Debug = true

	var userHistory = make(map[int64][]ai.Message)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		prompt := update.Message.Text

		msg := tgbotapi.NewMessage(chatID, "Thinking...")
		sent, _ := bot.Send(msg)

		userHistory[chatID] = append(userHistory[chatID], ai.Message{
			Role:    "user",
			Content: prompt,
		})

		response, err := ai.AskWithHistory(userHistory[chatID])
		if err != nil {
			response = "Ошибка: " + err.Error()
		} else {
			userHistory[chatID] = append(userHistory[chatID], ai.Message{
				Role:    "assistant",
				Content: response,
			})
		}

		edit := tgbotapi.NewEditMessageText(chatID, sent.MessageID, response)
		bot.Send(edit)

	}
}
