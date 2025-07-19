package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
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
			if update.Message != nil {
				// Process message in goroutine to avoid blocking
				go b.handleMessage(update.Message)
			} else if update.CallbackQuery != nil {
				// Handle callback query from inline buttons
				go b.handleCallbackQuery(update.CallbackQuery)
			}
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

// sendTypingIndicator sends a typing indicator with a context for cancellation
func (b *Bot) sendTypingIndicator(ctx context.Context, userID int64) {
	ticker := time.NewTicker(4 * time.Second) // Send typing indicator every 4 seconds
	defer ticker.Stop()

	// Send initial typing indicator
	typing := tgbotapi.NewChatAction(userID, tgbotapi.ChatTyping)
	b.api.Send(typing)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			typing := tgbotapi.NewChatAction(userID, tgbotapi.ChatTyping)
			b.api.Send(typing)
		}
	}
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

// handleCallbackQuery handles button presses from inline keyboards
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	// Check if user is allowed
	if !b.config.IsUserAllowed(callback.From.ID) {
		log.Warnf("Unauthorized user %d (%s) tried to use bot buttons", callback.From.ID, callback.From.UserName)
		return
	}

	userID := callback.From.ID
	data := callback.Data

	log.Infof("Button pressed by user %d: %s", userID, data)

	// Answer the callback query to remove loading state
	answerCallback := tgbotapi.NewCallback(callback.ID, "")
	b.api.Request(answerCallback)

	// Handle different button actions
	switch {
	case data == "menu" || data == "back_to_menu":
		b.handleMenuCommand(userID)
	case data == "settings":
		b.handleSettingsMenu(userID)
	case data == "expenses":
		b.handleExpensesCommand(userID)
	case data == "status":
		b.handleStatusCommand(userID)
	case data == "listmodels":
		b.handleListModelsCommand(userID)
	case data == "clear":
		b.handleClearWithConfirmation(userID)
	case data == "confirm_clear":
		b.handleClearCommand(userID)
	case data == "cancel_clear":
		b.sendMessage(userID, "❌ Clear operation cancelled.")
	case data == "help":
		b.handleStartCommand(userID)
	case data == "chat_mode":
		b.handleChatModeMenu(userID)
	case data == "change_model":
		b.handleModelSelectionMenu(userID)
	case data == "add_model":
		b.handleAddModelPrompt(userID)
	case data == "mode_with_history":
		b.handleModeCommand(userID, "with_history")
	case data == "mode_without_history":
		b.handleModeCommand(userID, "without_history")
	case strings.HasPrefix(data, "model_"):
		modelName := strings.TrimPrefix(data, "model_")
		b.handleModelCommand(userID, modelName)
	default:
		b.sendMessage(userID, "Unknown button action. Please try again.")
	}
}

// sendMessage sends a message to a user
func (b *Bot) sendMessage(userID int64, text string) error {
	return b.sendMessageWithKeyboard(userID, text, "HTML", nil)
}

// sendMessageWithKeyboard sends a message with an inline keyboard
func (b *Bot) sendMessageWithKeyboard(userID int64, text, parseMode string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	originalText := text

	// Format text based on parse mode
	if parseMode == "MarkdownV2" {
		text = b.formatForMarkdownV2(text)
		log.Debugf("MarkdownV2 formatting applied - Original length: %d, Formatted length: %d", len(originalText), len(text))
	} else if parseMode == "Markdown" {
		// For regular Markdown, just do basic table conversion
		text = b.convertTablesToMarkdown(text)
		log.Debugf("Markdown formatting applied - Original length: %d, Formatted length: %d", len(originalText), len(text))
	} else if parseMode == "HTML" {
		// For HTML, convert tables and escape HTML entities
		text = b.convertTablesToHTML(text)
		log.Debugf("HTML formatting applied - Original length: %d, Formatted length: %d", len(originalText), len(text))
	}

	msg := tgbotapi.NewMessage(userID, text)
	if parseMode != "" {
		msg.ParseMode = parseMode
		log.Debugf("Sending message with parse mode: %s", parseMode)
	}
	if keyboard != nil {
		msg.ReplyMarkup = keyboard
	}

	_, err := b.api.Send(msg)
	if err != nil {
		log.Errorf("Failed to send message to user %d: %v", userID, err)
	}
	return err
}

// sendLLMResponse sends an LLM response with proper HTML formatting
func (b *Bot) sendLLMResponse(userID int64, response string) error {
	// Format the LLM response for HTML (most reliable for international text)
	formattedResponse := b.convertTablesToHTML(response)

	// Split message if too long
	messages := b.splitMessage(formattedResponse, b.config.MaxMessageLength)

	for _, msgText := range messages {
		msg := tgbotapi.NewMessage(userID, msgText)
		msg.ParseMode = "HTML"

		if _, err := b.api.Send(msg); err != nil {
			log.Errorf("Failed to send LLM response to user %d: %v", userID, err)
			return err
		}

		// Small delay between messages to avoid rate limiting
		if len(messages) > 1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

// sendMessageWithMode sends a message with specific parse mode
func (b *Bot) sendMessageWithMode(userID int64, text, parseMode string) error {
	// Format text based on parse mode
	if parseMode == "MarkdownV2" {
		text = b.formatForMarkdownV2(text)
	} else if parseMode == "Markdown" {
		text = b.convertTablesToMarkdown(text)
	} else if parseMode == "HTML" {
		text = b.convertTablesToHTML(text)
	}

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

// splitMessage splits a long message into smaller chunks while preserving structure
func (b *Bot) splitMessage(text string, maxLength int) []string {
	if len(text) <= maxLength {
		return []string{text}
	}

	var messages []string

	// Split by paragraphs first to preserve structure
	paragraphs := strings.Split(text, "\n\n")
	current := ""

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		// If adding this paragraph would exceed the limit
		if len(current)+len(paragraph)+2 > maxLength {
			if current != "" {
				messages = append(messages, strings.TrimSpace(current))
				current = paragraph
			} else {
				// Paragraph is too long, split by lines
				lines := strings.Split(paragraph, "\n")
				tempCurrent := ""

				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}

					if len(tempCurrent)+len(line)+1 > maxLength {
						if tempCurrent != "" {
							messages = append(messages, strings.TrimSpace(tempCurrent))
							tempCurrent = line
						} else {
							// Line is too long, split by words as last resort
							words := strings.Fields(line)
							wordCurrent := ""

							for _, word := range words {
								if len(wordCurrent)+len(word)+1 > maxLength {
									if wordCurrent != "" {
										messages = append(messages, strings.TrimSpace(wordCurrent))
										wordCurrent = word
									} else {
										// Word is too long, force split
										messages = append(messages, word[:maxLength])
										wordCurrent = word[maxLength:]
									}
								} else {
									if wordCurrent == "" {
										wordCurrent = word
									} else {
										wordCurrent += " " + word
									}
								}
							}
							if wordCurrent != "" {
								tempCurrent = wordCurrent
							}
						}
					} else {
						if tempCurrent == "" {
							tempCurrent = line
						} else {
							tempCurrent += "\n" + line
						}
					}
				}
				current = tempCurrent
			}
		} else {
			if current == "" {
				current = paragraph
			} else {
				current += "\n\n" + paragraph
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

	// Add system message for HTML formatting
	systemMsg := storage.ChatMessage{
		Role:    "system",
		Content: b.createSystemMessageForHTML(),
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

	// Create context for typing indicator
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start typing indicator in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		b.sendTypingIndicator(ctx, userID)
	}()

	// Get LLM response
	log.Infof("Starting LLM request for user %d with model %s", userID, settings.CurrentModel)
	response, err := b.llmClient.GetChatResponse(settings.CurrentModel, messages, userID, b.storage)

	// Stop typing indicator
	cancel()
	wg.Wait()

	if err != nil {
		log.Errorf("Failed to get LLM response: %v", err)
		b.sendMessage(userID, fmt.Sprintf("Sorry, there was an error getting a response: %v", err))
		return
	}

	log.Infof("LLM request completed for user %d", userID)

	// Send response (format LLM response for MarkdownV2)
	if err := b.sendLLMResponse(userID, response); err != nil {
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
