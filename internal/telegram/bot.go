package telegram

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nurashi/Newton/internal/ai"
	"github.com/nurashi/Newton/internal/models"
	"github.com/nurashi/Newton/internal/repository"
)

type Bot struct {
	api         *tgbotapi.BotAPI
	userRepo *repository.UserRepository
	userHistory map[int64][]ai.Message
}

func NewBot(userRepo *repository.UserRepository) (*Bot, error) {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}

	api, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	api.Debug = false

	return &Bot{
		api:         api,
		userRepo: userRepo,
		userHistory: make(map[int64][]ai.Message),
	}, nil
}

func (b *Bot) Run() error {
	log.Printf("Bot @%s started successfully", b.api.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		go b.handleUpdate(update)
	}

	return nil
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// Save or update user in database
	ctx := context.Background()
	telegramUser := b.extractTelegramUser(update.Message.From)

	user, err := b.userRepo.CreateOrUpdate(ctx, telegramUser)
	if err != nil {
		log.Printf("Failed to save user %d: %v", update.Message.From.ID, err)
		// Continue processing even if user save fails
	} else {
		log.Printf("User saved/updated: ID=%d, Username=%s, FirstName=%s",
			user.ID,
			b.stringPtrToString(user.Username),
			user.FirstName)
	}

	// Handle different message types
	switch {
	case update.Message.IsCommand():
		b.handleCommand(update.Message)
	case update.Message.Text != "":
		b.handleTextMessage(update.Message)
	default:
		b.sendMessage(update.Message.Chat.ID, "I only support text messages for now.")
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	// Update last seen
	ctx := context.Background()
	b.userRepo.UpdateLastSeen(ctx, int64(userID))

	switch message.Command() {
	case "start":
		// Get user info for personalized welcome
		user, err := b.userRepo.GetByID(ctx, int64(userID))
		var firstName string
		if err == nil {
			firstName = user.FirstName
		} else {
			firstName = message.From.FirstName
		}

		welcomeMsg := fmt.Sprintf(`Hello %s! Welcome to Newton AI Bot! 

Commands:
/help - Show this help message
/clear - Clear conversation history
/profile - Show your profile information
/stats - Show your usage statistics

Just send me any message and I'll respond using AI!`, firstName)

		b.sendMessage(chatID, welcomeMsg)

	case "help":
		b.handleCommand(message) 

	case "clear":
		delete(b.userHistory, chatID)
		b.sendMessage(chatID, "Conversation history cleared!")

	case "profile":
		b.handleProfileCommand(chatID, int64(userID))

	case "stats":
		b.handleStatsCommand(chatID, int64(userID))

	default:
		b.sendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handleProfileCommand(chatID, userID int64) {
	ctx := context.Background()
	user, err := b.userRepo.GetByID(ctx, userID)

	if err != nil {
		log.Printf("Failed to get user profile %d: %v", userID, err)
		b.sendMessage(chatID, "Sorry, couldn't retrieve your profile information.")
		return
	}

	profileMsg := fmt.Sprintf(`Your Profile

ID: %d
Name: %s %s
Username: %s
Language: %s
Member since: %s
Last seen: %s`,
		user.ID,
		user.FirstName,
		b.stringPtrToString(&user.LastName),
		b.stringPtrToString(user.Username),
		b.stringPtrToString(user.LanguageCode),
		user.CreatedAt.Format("2006-01-02"),
		user.LastSeen.Format("2006-01-02 15:04:05"))

	b.sendMessage(chatID, profileMsg)
}

func (b *Bot) handleStatsCommand(chatID, userID int64) {
	ctx := context.Background()
	stats, err := b.userRepo.GetStats(ctx, userID)

	if err != nil {
		log.Printf("Failed to get user stats %d: %v", userID, err)
		b.sendMessage(chatID, "Sorry, couldn't retrieve your statistics.")
		return
	}

	count := len(b.userHistory[chatID])

	statsMsg := fmt.Sprintf(`Your Statistics

Messages in current session: %d
Member since: %s
Last seen: %s

More detailed statistics coming soon!`,
		count,
		stats["member_since"],
		stats["last_seen"])

	b.sendMessage(chatID, statsMsg)
}

func (b *Bot) handleTextMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID
	prompt := message.Text

	ctx := context.Background()
	b.userRepo.UpdateLastSeen(ctx, int64(userID))

	log.Printf("User %d (%s) in chat %d: %s",
		userID, message.From.UserName, chatID, prompt)

	typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping)
	b.api.Send(typing)

	thinkingMsg := tgbotapi.NewMessage(chatID, "Thinking...")
	sent, err := b.api.Send(thinkingMsg)
	if err != nil {
		log.Printf("Failed to send thinking message: %v", err)
		return
	}

	b.userHistory[chatID] = append(b.userHistory[chatID], ai.Message{
		Role:    "user",
		Content: prompt,
	})

	if len(b.userHistory[chatID]) > 20 {
		b.userHistory[chatID] = b.userHistory[chatID][len(b.userHistory[chatID])-20:]
	}

	start := time.Now()
	response, err := ai.AskWithHistory(b.userHistory[chatID])
	duration := time.Since(start)

	if err != nil {
		log.Printf("AI request failed: %v", err)
		response = "Sorry, I'm having trouble processing your request. Please try again later."

		if len(b.userHistory[chatID]) > 0 {
			b.userHistory[chatID] = b.userHistory[chatID][:len(b.userHistory[chatID])-1]
		}
	} else {
		// Add AI response to history
		b.userHistory[chatID] = append(b.userHistory[chatID], ai.Message{
			Role:    "assistant",
			Content: response,
		})

		log.Printf("AI responded in %v for user %d", duration, userID)
	}

	edit := tgbotapi.NewEditMessageText(chatID, sent.MessageID, response)
	edit.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(edit); err != nil {
		log.Printf("Failed to edit message: %v", err)
		b.sendMessage(chatID, response)
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send message to chat %d: %v", chatID, err)
	}
}

// Helper function to extract Telegram user data
func (b *Bot) extractTelegramUser(from *tgbotapi.User) models.TelegramUser {
	var username, lastName, languageCode *string

	if from.UserName != "" {
		username = &from.UserName
	}
	if from.LastName != "" {
		lastName = &from.LastName
	}
	if from.LanguageCode != "" {
		languageCode = &from.LanguageCode
	}

	return models.TelegramUser{
		ID:           int64(from.ID),
		Username:     username,
		FirstName:    from.FirstName,
		LastName:     lastName,
		IsBot:        from.IsBot,
		LanguageCode: languageCode,
	}
}

// Helper function to safely convert string pointer to string
func (b *Bot) stringPtrToString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

// Updated RunTelegramBot function
func RunTelegramBot(userService *repository.UserRepository) {
	bot, err := NewBot(userService)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	if err := bot.Run(); err != nil {
		log.Fatalf("Bot failed: %v", err)
	}
}
