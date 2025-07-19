package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"

	"telegrambot/internal/config"
	"telegrambot/internal/openrouter"
	"telegrambot/internal/storage"
)

// Bot represents the Telegram bot
type Bot struct {
	api       *tgbotapi.BotAPI
	config    *config.Config
	storage   storage.Storage
	llmClient *openrouter.Client
	updates   tgbotapi.UpdatesChannel
}

// New creates a new bot instance
func New(cfg *config.Config, store storage.Storage) (*Bot, error) {
	// Initialize Telegram bot API
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	// Set debug mode based on log level
	api.Debug = strings.ToLower(cfg.LogLevel) == "debug"

	// Initialize OpenRouter client
	llmClient := openrouter.NewClient(cfg.OpenRouterAPIKey, cfg.OpenRouterBaseURL)

	log.Infof("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:       api,
		config:    cfg,
		storage:   store,
		llmClient: llmClient,
	}, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	// Set up update configuration
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Get updates channel
	b.updates = b.api.GetUpdatesChan(u)

	log.Info("Bot started, waiting for messages...")

	// Process updates
	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-b.updates:
			if update.Message == nil {
				continue
			}

			// Process message in goroutine to avoid blocking
			go b.handleMessage(update.Message)
		}
	}
}

// Stop stops the bot
func (b *Bot) Stop() {
	if b.updates != nil {
		b.api.StopReceivingUpdates()
	}
	log.Info("Bot stopped")
}

// handleMessage handles incoming messages
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// Check if user is allowed
	if !b.config.IsUserAllowed(message.From.ID) {
		log.Warnf("Unauthorized user %d (%s) tried to use bot", message.From.ID, message.From.UserName)
		return
	}

	userID := message.From.ID
	log.Infof("Message from user %d: %s", userID, message.Text)

	// Handle commands
	if message.IsCommand() {
		b.handleCommand(message)
		return
	}

	// Handle regular messages (chat with LLM)
	b.handleChatMessage(message)
}

// handleCommand handles bot commands
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	userID := message.From.ID
	command := message.Command()
	args := message.CommandArguments()

	switch command {
	case "start", "help":
		b.handleStartCommand(userID)
	case "menu":
		b.handleMenuCommand(userID)
	case "mode":
		b.handleModeCommand(userID, args)
	case "model":
		b.handleModelCommand(userID, args)
	case "addmodel":
		b.handleAddModelCommand(userID, args)
	case "listmodels":
		b.handleListModelsCommand(userID)
	case "expenses":
		b.handleExpensesCommand(userID)
	case "clear":
		b.handleClearCommand(userID)
	case "status":
		b.handleStatusCommand(userID)
	default:
		b.sendMessage(userID, "Unknown command. Type /menu to see available commands.")
	}
}

// sendMessage sends a message to a user
func (b *Bot) sendMessage(userID int64, text string) error {
	return b.sendMessageWithMode(userID, text, "Markdown")
}

// sendMessageWithMode sends a message with specific parse mode
func (b *Bot) sendMessageWithMode(userID int64, text, parseMode string) error {
	// Split message if too long
	messages := b.splitMessage(text, b.config.MaxMessageLength)

	for _, msgText := range messages {
		msg := tgbotapi.NewMessage(userID, msgText)
		if parseMode != "" {
			msg.ParseMode = parseMode
		}

		if _, err := b.api.Send(msg); err != nil {
			log.Errorf("Failed to send message to user %d: %v", userID, err)
			return err
		}

		// Small delay between messages to avoid rate limiting
		if len(messages) > 1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

// splitMessage splits a long message into smaller chunks
func (b *Bot) splitMessage(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	var messages []string
	words := strings.Fields(text)
	current := ""

	for _, word := range words {
		// If adding this word would exceed the limit
		if len(current)+len(word)+1 > maxLength {
			if current != "" {
				messages = append(messages, current)
				current = word
			} else {
				// Word is too long, split it
				messages = append(messages, word[:maxLength])
				current = word[maxLength:]
			}
		} else {
			if current == "" {
				current = word
			} else {
				current += " " + word
			}
		}
	}

	if current != "" {
		messages = append(messages, current)
	}

	return messages
}

// handleChatMessage handles regular chat messages
func (b *Bot) handleChatMessage(message *tgbotapi.Message) {
	userID := message.From.ID
	userText := message.Text

	// Get user settings
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Sorry, there was an error processing your request.")
		return
	}

	// Add user message to storage
	userMsg := storage.ChatMessage{
		Role:      "user",
		Content:   userText,
		Timestamp: time.Now(),
	}

	if err := b.storage.AddChatMessage(userID, userMsg); err != nil {
		log.Errorf("Failed to save user message: %v", err)
	}

	// Prepare messages for LLM
	var messages []storage.ChatMessage

	// Add system message for markdown formatting
	systemMsg := storage.ChatMessage{
		Role:    "system",
		Content: "You are a helpful assistant. Format your responses using Markdown syntax for better readability in Telegram. Use **bold**, *italic*, `code`, and other Markdown features appropriately.",
	}
	messages = append(messages, systemMsg)

	// Add chat history if mode is with_history
	if settings.ChatMode == "with_history" {
		history, err := b.storage.GetChatHistory(userID)
		if err != nil {
			log.Errorf("Failed to get chat history: %v", err)
		} else {
			// Add last 10 messages for context (excluding the current message)
			start := len(history) - 11
			if start < 0 {
				start = 0
			}
			for i := start; i < len(history)-1; i++ {
				messages = append(messages, history[i])
			}
		}
	}

	// Add current user message
	messages = append(messages, userMsg)

	// Send typing indicator
	typing := tgbotapi.NewChatAction(userID, tgbotapi.ChatTyping)
	b.api.Send(typing)

	// Get LLM response
	response, err := b.llmClient.GetChatResponse(settings.CurrentModel, messages, userID, b.storage)
	if err != nil {
		log.Errorf("Failed to get LLM response: %v", err)
		b.sendMessage(userID, fmt.Sprintf("Sorry, there was an error getting a response: %v", err))
		return
	}

	// Send response
	if err := b.sendMessage(userID, response); err != nil {
		log.Errorf("Failed to send response: %v", err)
		return
	}

	// Save assistant response
	assistantMsg := storage.ChatMessage{
		Role:      "assistant",
		Content:   response,
		Timestamp: time.Now(),
	}

	if err := b.storage.AddChatMessage(userID, assistantMsg); err != nil {
		log.Errorf("Failed to save assistant message: %v", err)
	}
}
