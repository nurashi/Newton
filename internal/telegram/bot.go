package telegram

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nurashi/Newton/internal/ai"
	"github.com/nurashi/Newton/internal/handlers"
	"github.com/nurashi/Newton/internal/models"
	"github.com/nurashi/Newton/internal/repository"
)

// Bot represents the Telegram bot instance
type Bot struct {
	api         *tgbotapi.BotAPI
	userRepo    *repository.UserRepository
	userHistory map[int64][]ai.Message
	pdfContext  map[int64]string
}

// NewBot creates a new Telegram bot instance
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
		userRepo:    userRepo,
		userHistory: make(map[int64][]ai.Message),
		pdfContext:  make(map[int64]string),
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

	ctx := context.Background()
	telegramUser := b.extractTelegramUser(update.Message.From)

	user, err := b.userRepo.CreateOrUpdate(ctx, telegramUser)
	if err != nil {
		log.Printf("Failed to save user %d: %v", update.Message.From.ID, err)
	} else {
		log.Printf("User saved/updated: ID=%d, Username=%s, FirstName=%s",
			user.ID,
			b.stringPtrToString(user.Username),
			user.FirstName)
	}

	switch {
	case update.Message.IsCommand():
		b.handleCommand(update.Message)
	case update.Message.Text != "":
		b.handleTextMessage(update.Message)
	case update.Message.Document != nil:
		b.handleDocument(update.Message)
	default:
		b.sendMessage(update.Message.Chat.ID, "I only support text messages, documents, and commands for now.")
	}
}

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID

	ctx := context.Background()
	b.userRepo.UpdateLastSeen(ctx, int64(userID))

	switch message.Command() {
	case "start":
		user, err := b.userRepo.GetByID(ctx, int64(userID))
		var firstName string
		if err == nil {
			firstName = user.FirstName
		} else {
			firstName = message.From.FirstName
		}

		welcomeMsg := fmt.Sprintf(`Hello %s! Welcome to Newton AI Bot! 


use /help to see all available commands

Just send me any message and I'll respond using AI!`, firstName)

		b.sendMessage(chatID, welcomeMsg)

	case "help":

		helpMsg := `
	
	Commands:
/help - Show this help message.
/clear - Clear conversation history(ai will forget all messanges).
/profile - Show your profile information.
/stats - Show your usage statistics.
/weather <city> - provides weather.
/pitch <topic> - provides idea to pitch by following topic.
/photo <topic> - shows some photo of provided topic by some author.
/image <topic> - generates image by provided topic. 
	
	`

		b.sendMessage(chatID, helpMsg)

	case "clear":
		delete(b.userHistory, chatID)
		b.sendMessage(chatID, "Conversation history cleared!")

	case "profile":
		b.handleProfileCommand(chatID, int64(userID))

	case "stats":
		b.handleStatsCommand(chatID, int64(userID))
	case "weather":
		args := message.CommandArguments()
		if args == "" {
			b.sendMessage(chatID, "Please provide a city name. Example: /weather London")
			return
		}

		weatherInfo, err := handlers.GetWeather(args)
		if err != nil {
			log.Printf("Weather API error: %v", err)
			b.sendMessage(chatID, "Sorry, I couldn't fetch the weather right now.")
			return
		}

		b.sendMessage(chatID, weatherInfo)
	case "pitch":
		args := message.CommandArguments()
		if args == "" {
			b.sendMessage(chatID, "Please provide your startup idea. Example: /pitch AI tool for lawyers")
			return
		}
		thinkingMsg := tgbotapi.NewMessage(chatID, "Generating your pitch, please wait...")
		sent, err := b.api.Send(thinkingMsg)
		if err != nil {
			log.Printf("ERROR: Failed to send thinking message: %v", err)
		}

		pitch, err := ai.GeneratePitch(args)
		if err != nil {
			log.Printf("ERROR: failed to generate pitch: %v", err)
			edit := tgbotapi.NewEditMessageText(chatID, sent.MessageID, "Sorry, I couldn't generate pitch right now.")
			b.api.Send(edit)
			return
		}

		edit := tgbotapi.NewEditMessageText(chatID, sent.MessageID, pitch)
		b.api.Send(edit)

	case "photo":
		query := message.CommandArguments()
		if query == "" {
			b.sendMessage(message.Chat.ID, "Usage: /photo <query>")
			return
		}

		url, caption, err := handlers.SendUnsplashPhoto(chatID, query)
		if err != nil {
			b.sendMessage(chatID, "can't find photo for now")
			return
		}

		msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url))
		msg.Caption = caption
		if _, err := b.api.Send(msg); err != nil {
			log.Printf("ERROR: failed to send msg from unsplash: %v", err)
			return
		}
	case "image":
		prompt := message.CommandArguments()

		url, caption, err := handlers.SendAIImage(prompt)
		if err != nil {
			b.sendMessage(chatID, "can't generate image for now")
		}
		msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(url))
		msg.Caption = caption
		b.api.Send(msg)

	default:
		b.sendMessage(chatID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handleProfileCommand(chatID, userID int64) {
	log.Printf("DEBUG: handleProfileCommand called for chatID=%d userID=%d", chatID, userID)
	ctx := context.Background()
	user, err := b.userRepo.GetByID(ctx, userID)
	log.Printf("DEBUG: user from DB: %+v", user)
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
		escapeMarkdownV2(user.FirstName),
		escapeMarkdownV2(b.stringPtrToString(user.LastName)),
		escapeMarkdownV2(b.stringPtrToString(user.Username)),
		escapeMarkdownV2(b.stringPtrToString(user.LanguageCode)),
		user.CreatedAt.Format("2006-01-02"),
		user.LastSeen.Format("2006-01-02 15:04:05"),
	)

	b.sendMessage(chatID, profileMsg)
	log.Printf("DEBUG: sending profile message to chat %d", chatID)
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
	messageCount := stats["message_count"].(int)

	statsMsg := fmt.Sprintf(`Your Statistics

Total messages sent: %d
Messages in current session: %d
Member since: %s
Last seen: %s`,
		messageCount,
		count,
		stats["member_since"],
		stats["last_seen"])

	b.sendMessage(chatID, statsMsg)
}

func (b *Bot) handleTextMessageLM(message *tgbotapi.Message) { // just exist, to test work with LM studio
	ctx := context.Background()

	b.userRepo.UpdateLastSeen(ctx, int64(message.From.ID))

	log.Printf("User: %d (%s) in chat %d: %s", message.Chat.ID, message.From.UserName, message.Chat.ID, message.Text)

	typing := tgbotapi.NewChatAction(message.Chat.ID, tgbotapi.ChatTyping)

	b.api.Send(typing)

	thinkingMsg := tgbotapi.NewMessage(message.Chat.ID, "Thinking...")
	send, err := b.api.Send(thinkingMsg)
	if err != nil {
		log.Printf("Failed to send thinking message: %v", err)
		return
	}

	start := time.Now()
	response, err := ai.LMStudioAPICall(message.Text) // message.text = prompt
	duration := time.Since(start)

	if err != nil {
		log.Printf("ERROR: failed to get LM studio response: %v", err)
		return
	}

	log.Printf("AI responded in %v", duration)

	edit := tgbotapi.NewEditMessageText(message.Chat.ID, send.MessageID, response)

	if _, err := b.api.Send(edit); err != nil {
		log.Printf("Failed to edit message: %v", err)
		b.sendMessage(message.Chat.ID, response)
	}

}

func (b *Bot) handleTextMessage(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	userID := message.From.ID
	prompt := message.Text

	ctx := context.Background()

	if err := b.userRepo.IncrementMessageCount(ctx, int64(userID)); err != nil {
		log.Printf("Failed to increment message count for user %d: %v", userID, err)
	}

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
	response, err := ai.AskGeminiWithHistory(b.userHistory[chatID])
	duration := time.Since(start)

	if err != nil {
		log.Printf("AI request failed: %v", err)
		response = "Sorry, I'm having trouble processing your request. Please try again later."

		if len(b.userHistory[chatID]) > 0 {
			b.userHistory[chatID] = b.userHistory[chatID][:len(b.userHistory[chatID])-1]
		}
	} else {
		b.userHistory[chatID] = append(b.userHistory[chatID], ai.Message{
			Role:    "assistant",
			Content: response,
		})

		log.Printf("AI responded in %v for user %d", duration, userID)
	}

	// Send response (handles long messages and markdown)
	if err := b.sendLongMessage(chatID, sent.MessageID, response, true); err != nil {
		log.Printf("Failed to send response: %v", err)
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

// sanitizeMarkdown fixes common Markdown issues that break Telegram parsing
func sanitizeMarkdown(text string) string {
	lines := strings.Split(text, "\n")
	var result []string
	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			result = append(result, line)
			continue
		}

		if inCodeBlock {
			result = append(result, line)
			continue
		}

		backtickCount := strings.Count(line, "`")
		if backtickCount%2 != 0 {
			line += "`"
		}

		result = append(result, line)
	}

	finalText := strings.Join(result, "\n")

	codeBlockCount := strings.Count(finalText, "```")
	if codeBlockCount%2 != 0 {
		finalText += "\n```"
	}

	return finalText
}

// splitLongMessage splits message into chunks under maxLen chars
func splitLongMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	var currentChunk strings.Builder
	inCodeBlock := false
	codeBlockLang := ""

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if !inCodeBlock {
				inCodeBlock = true
				codeBlockLang = strings.TrimPrefix(strings.TrimSpace(line), "```")
			} else {
				inCodeBlock = false
				codeBlockLang = ""
			}
		}

		if currentChunk.Len()+len(line)+1 > maxLen {
			chunkText := currentChunk.String()
			if inCodeBlock {
				chunkText += "\n```"
			}
			chunks = append(chunks, chunkText)
			currentChunk.Reset()

			if inCodeBlock {
				currentChunk.WriteString("```" + codeBlockLang + "\n")
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n")
		}
		currentChunk.WriteString(line)
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

// sendLongMessage sends a message, splitting if necessary
func (b *Bot) sendLongMessage(chatID int64, messageID int, text string, isEdit bool) error {
	const maxTelegramLength = 4000

	text = sanitizeMarkdown(text)

	chunks := splitLongMessage(text, maxTelegramLength)

	if len(chunks) == 1 {
		if isEdit {
			edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
			edit.ParseMode = "Markdown"

			if _, err := b.api.Send(edit); err != nil {
				log.Printf("Markdown failed, trying plain text: %v", err)
				edit.ParseMode = ""
				if _, err := b.api.Send(edit); err != nil {
					return b.sendPlainMessage(chatID, text)
				}
			}
			return nil
		}
		return b.sendPlainMessage(chatID, text)
	}

	if isEdit {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
		b.api.Send(deleteMsg)
	}

	for i, chunk := range chunks {
		var msgText string
		if len(chunks) > 1 {
			msgText = fmt.Sprintf("*Part %d/%d*\n\n%s", i+1, len(chunks), chunk)
		} else {
			msgText = chunk
		}

		if err := b.sendPlainMessage(chatID, msgText); err != nil {
			return err
		}

		if i < len(chunks)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	return nil
}

// sendPlainMessage sends a single message with markdown fallback
func (b *Bot) sendPlainMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Markdown failed, sending plain text: %v", err)
		msg.ParseMode = ""

		msg.Text = strings.ReplaceAll(text, "```", "")
		msg.Text = strings.ReplaceAll(msg.Text, "`", "")
		msg.Text = strings.ReplaceAll(msg.Text, "*", "")
		msg.Text = strings.ReplaceAll(msg.Text, "_", "")
		_, err = b.api.Send(msg)
	}

	return err
}

func RunTelegramBot(userService *repository.UserRepository) {
	bot, err := NewBot(userService)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	if err := bot.Run(); err != nil {
		log.Fatalf("Bot failed: %v", err)
	}
}

// my username is nurasyl_orazbek, and "_" gives error at query.
func escapeMarkdownV2(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(s)
}

func (b *Bot) handleDocument(message *tgbotapi.Message) {
	chatID := message.Chat.ID
	file := message.Document
	filename := strings.ToLower(file.FileName)

	// Check supported file types
	ext := handlers.GetFileExtension(filename)
	if ext != "pdf" && ext != "pptx" {
		b.sendMessage(chatID, "Supported formats: PDF and PPTX files only")
		return
	}

	typing := tgbotapi.NewChatAction(chatID, tgbotapi.ChatUploadDocument)
	b.api.Send(typing)

	fileTypeLabel := strings.ToUpper(ext)
	processingMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Processing %s...", fileTypeLabel))
	send, err := b.api.Send(processingMsg)

	if err != nil {
		log.Printf("Failed to send processing message: %v", err)
	}

	fileConfig := tgbotapi.FileConfig{FileID: file.FileID}
	tgFile, err := b.api.GetFile(fileConfig)
	if err != nil {
		b.editOrSendMessage(chatID, send.MessageID, "Failed to get file from Telegram")
		log.Printf("ERROR: failed to get file: %v", err)
		return
	}

	fileURL := tgFile.Link(b.api.Token)
	localPath := fmt.Sprintf("/tmp/%d_%s", time.Now().Unix(), file.FileName)

	if err := handlers.DownloadFile(localPath, fileURL); err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Failed to download %s file", fileTypeLabel))
		return
	}

	defer os.Remove(localPath)

	var text string
	switch ext {
	case "pdf":
		text, err = handlers.ExtractPDFText(localPath)
	case "pptx":
		text, err = handlers.ExtractPPTXText(localPath)
	}

	if err != nil {
		b.editOrSendMessage(chatID, send.MessageID, fmt.Sprintf("❌ Failed to read %s: %v", fileTypeLabel, err))
		return
	}

	if strings.TrimSpace(text) == "" {
		b.editOrSendMessage(chatID, send.MessageID, fmt.Sprintf("📭 No text found in %s. It might be image-based.", fileTypeLabel))
		return
	}

	edit := tgbotapi.NewEditMessageText(chatID, send.MessageID, "🎓 Creating Educational Guide...")
	b.api.Send(edit)

	b.createEducationalGuide(chatID, send.MessageID, text, file.FileName, fileTypeLabel)
}

func (b *Bot) createEducationalGuide(chatID int64, messageID int, documentText string, filename string, fileType string) {
	startTime := time.Now()

	response, err := ai.GenerateEducationalGuide(documentText, filename, fileType)
	if err != nil {
		log.Printf("Educational guide generation failed: %v", err)
		b.editOrSendMessage(chatID, messageID, fmt.Sprintf("Failed to generate guide: %v", err))
		return
	}

	fullResponse := fmt.Sprintf("*Educational Guide*\n `%s`\n\n%s\n\n_You can now ask me questions about this document!_", filename, response)

	// Store context for follow-up questions
	b.pdfContext[chatID] = documentText

	if err := b.sendLongMessage(chatID, messageID, fullResponse, true); err != nil {
		log.Printf("Failed to send educational guide: %v", err)
	}

	log.Printf("Educational guide created in %v for chat %d", time.Since(startTime), chatID)
}

func (b *Bot) editOrSendMessage(chatID int64, messageID int, text string) {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	edit.ParseMode = "Markdown"

	if _, err := b.api.Send(edit); err != nil {
		log.Printf("Failed to edit message, sending new: %v", err)
		b.sendMessage(chatID, text)
	}
}
