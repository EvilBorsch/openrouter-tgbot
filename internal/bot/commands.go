package bot

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// handleStartCommand handles the /start and /help commands
func (b *Bot) handleStartCommand(userID int64) {
	welcomeMessage := `🤖 *Welcome to LLM Chat Bot!*

This bot provides access to various LLM models through OpenRouter.

*Quick Start:*
• Just type any message to chat with the AI
• Use the menu buttons below for settings and commands
• The bot supports advanced formatting

*Features:*
✅ Multiple LLM models support
✅ Chat history modes
✅ Expense tracking
✅ Custom model management
✅ Message splitting for long responses

Use the buttons below to get started!`

	keyboard := b.createMainMenuKeyboard()
	b.sendMessageWithKeyboard(userID, welcomeMessage, "Markdown", keyboard)
}

// handleMenuCommand handles the /menu command
func (b *Bot) handleMenuCommand(userID int64) {
	menuMessage := `📋 *Main Menu*

Welcome to your AI assistant! Choose an option below or just start typing to chat with the AI.

*Quick Actions:*
• 💬 Just type a message to chat
• ⚙️ Settings - Configure chat mode and models
• 📊 View expenses and usage statistics
• 🤖 Browse and change AI models

*Features:*
✅ Interactive button controls
✅ Multiple LLM models
✅ Chat history management
✅ Real-time expense tracking

Choose an option below:`

	keyboard := b.createMainMenuKeyboard()
	b.sendMessageWithKeyboard(userID, menuMessage, "Markdown", keyboard)
}

// handleModeCommand handles the /mode command
func (b *Bot) handleModeCommand(userID int64, args string) {
	if args == "" {
		settings, err := b.storage.GetUserSettings(userID)
		if err != nil {
			log.Errorf("Failed to get user settings: %v", err)
			b.sendMessage(userID, "Error retrieving your settings.")
			return
		}

		message := fmt.Sprintf("🔧 *Current chat mode:* `%s`\n\n", settings.ChatMode)
		message += "*Available modes:*\n"
		message += "• `with_history` - AI remembers previous messages\n"
		message += "• `without_history` - Each message is independent\n\n"
		message += "*Usage:* `/mode with_history` or `/mode without_history`"

		b.sendMessage(userID, message)
		return
	}

	mode := strings.ToLower(strings.TrimSpace(args))
	if mode != "with_history" && mode != "without_history" {
		b.sendMessage(userID, "❌ Invalid mode. Use: `with_history` or `without_history`")
		return
	}

	// Get current settings
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	// Update mode
	settings.ChatMode = mode
	if err := b.storage.SaveUserSettings(settings); err != nil {
		log.Errorf("Failed to save user settings: %v", err)
		b.sendMessage(userID, "Error saving your settings.")
		return
	}

	message := fmt.Sprintf("✅ Chat mode changed to: `%s`", mode)
	if mode == "with_history" {
		message += "\n\n*Note:* The AI will now remember your previous messages in this session."
	} else {
		message += "\n\n*Note:* Each message will be processed independently."
	}

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "Markdown", keyboard)
}

// handleModelCommand handles the /model command
func (b *Bot) handleModelCommand(userID int64, args string) {
	if args == "" {
		settings, err := b.storage.GetUserSettings(userID)
		if err != nil {
			log.Errorf("Failed to get user settings: %v", err)
			b.sendMessage(userID, "Error retrieving your settings.")
			return
		}

		message := fmt.Sprintf("🤖 *Current model:* `%s`\n\n", settings.CurrentModel)
		message += "*Popular models:*\n"
		message += "• `openai/gpt-4` - Most capable, higher cost\n"
		message += "• `openai/gpt-3.5-turbo` - Fast and affordable\n"
		message += "• `anthropic/claude-3-sonnet` - Great for analysis\n"
		message += "• `google/gemini-pro` - Google's latest model\n\n"
		message += "*Usage:* `/model openai/gpt-4`\n"
		message += "*See all:* `/listmodels`"

		b.sendMessage(userID, message)
		return
	}

	model := strings.TrimSpace(args)
	if model == "" {
		b.sendMessage(userID, "❌ Please specify a model name.")
		return
	}

	// Get current settings
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	// Update model
	settings.CurrentModel = model
	if err := b.storage.SaveUserSettings(settings); err != nil {
		log.Errorf("Failed to save user settings: %v", err)
		b.sendMessage(userID, "Error saving your settings.")
		return
	}

	message := fmt.Sprintf("✅ Model changed to: `%s`\n\n", model)
	message += "*Tip:* The pricing and capabilities may vary between models. Check expenses to monitor usage."

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "Markdown", keyboard)
}

// handleAddModelCommand handles the /addmodel command
func (b *Bot) handleAddModelCommand(userID int64, args string) {
	if args == "" {
		message := "🔧 *Add Custom Model*\n\n"
		message += "*Usage:* `/addmodel model-provider/model-name`\n\n"
		message += "*Examples:*\n"
		message += "• `/addmodel mistralai/mistral-7b-instruct`\n"
		message += "• `/addmodel meta-llama/llama-2-70b-chat`\n"
		message += "• `/addmodel cohere/command-r-plus`\n\n"
		message += "*Note:* Make sure the model is available on OpenRouter."

		b.sendMessage(userID, message)
		return
	}

	model := strings.TrimSpace(args)
	if model == "" {
		b.sendMessage(userID, "❌ Please specify a model name.")
		return
	}

	// Get current settings
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	// Check if model already exists
	for _, existingModel := range settings.CustomModels {
		if existingModel == model {
			b.sendMessage(userID, fmt.Sprintf("❌ Model `%s` is already in your list.", model))
			return
		}
	}

	// Add model
	settings.CustomModels = append(settings.CustomModels, model)
	if err := b.storage.SaveUserSettings(settings); err != nil {
		log.Errorf("Failed to save user settings: %v", err)
		b.sendMessage(userID, "Error saving your settings.")
		return
	}

	message := fmt.Sprintf("✅ Added model: `%s`\n\n", model)
	message += "You can now use it with: `/model " + model + "`"

	b.sendMessage(userID, message)
}

// handleListModelsCommand handles the /listmodels command
func (b *Bot) handleListModelsCommand(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "🤖 *Available Models*\n\n"
	message += fmt.Sprintf("*Current:* `%s` ✅\n\n", settings.CurrentModel)

	message += "*Popular Models:*\n"
	popularModels := []string{
		"openai/gpt-4",
		"openai/gpt-3.5-turbo",
		"anthropic/claude-3-sonnet",
		"google/gemini-pro",
		"mistralai/mistral-7b-instruct",
		"meta-llama/llama-2-70b-chat",
	}

	for _, model := range popularModels {
		if model == settings.CurrentModel {
			continue // Skip current model as it's already shown
		}
		message += fmt.Sprintf("• `%s`\n", model)
	}

	if len(settings.CustomModels) > 0 {
		message += "\n*Your Custom Models:*\n"
		for _, model := range settings.CustomModels {
			if model == settings.CurrentModel {
				continue // Skip current model as it's already shown
			}
			message += fmt.Sprintf("• `%s`\n", model)
		}
	}

	message += "\n*Usage:* Click a model button above or type `/model model-name`"

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "Markdown", keyboard)
}

// handleExpensesCommand handles the /expenses command
func (b *Bot) handleExpensesCommand(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "💰 *Your Usage Statistics*\n\n"
	message += fmt.Sprintf("*Total Expenses:* $%.6f\n", settings.TotalExpenses)
	message += fmt.Sprintf("*Total Requests:* %d\n", len(settings.ExpenseHistory))
	message += "_Using accurate OpenRouter pricing & native token counts_\n"

	if len(settings.ExpenseHistory) > 0 {
		// Calculate stats
		var totalTokens int
		modelUsage := make(map[string]int)
		var recentExpenses float64

		// Get recent expenses (last 7 days)
		weekAgo := time.Now().AddDate(0, 0, -7)

		for _, expense := range settings.ExpenseHistory {
			totalTokens += expense.InputTokens + expense.OutputTokens
			modelUsage[expense.Model]++

			if expense.Timestamp.After(weekAgo) {
				recentExpenses += expense.Cost
			}
		}

		message += fmt.Sprintf("*Total Tokens:* %d\n", totalTokens)
		message += fmt.Sprintf("*Last 7 Days:* $%.6f\n\n", recentExpenses)

		// Show model usage
		message += "*Model Usage:*\n"
		for model, count := range modelUsage {
			message += fmt.Sprintf("• `%s`: %d requests\n", model, count)
		}

		// Show recent transactions (last 5)
		message += "\n*Recent Transactions:*\n"
		start := len(settings.ExpenseHistory) - 5
		if start < 0 {
			start = 0
		}

		for i := start; i < len(settings.ExpenseHistory); i++ {
			expense := settings.ExpenseHistory[i]
			message += fmt.Sprintf("• %s: $%.6f (%s)\n",
				expense.Timestamp.Format("01/02 15:04"),
				expense.Cost,
				expense.Model)
		}
	} else {
		message += "\n*No usage data yet.* Start chatting to see your statistics!"
	}

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "Markdown", keyboard)
}

// handleClearCommand handles the /clear command
func (b *Bot) handleClearCommand(userID int64) {
	if err := b.storage.ClearChatHistory(userID); err != nil {
		log.Errorf("Failed to clear chat history: %v", err)
		b.sendMessage(userID, "Error clearing chat history.")
		return
	}

	message := "🗑️ *Chat history cleared!*\n\n"
	message += "Your conversation history has been deleted. The AI will start fresh with your next message."

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "Markdown", keyboard)
}

// handleStatusCommand handles the /status command
func (b *Bot) handleStatusCommand(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "📊 *Your Current Settings*\n\n"
	message += fmt.Sprintf("*User ID:* `%d`\n", settings.UserID)
	message += fmt.Sprintf("*Current Model:* `%s`\n", settings.CurrentModel)
	message += fmt.Sprintf("*Chat Mode:* `%s`\n", settings.ChatMode)
	message += fmt.Sprintf("*Total Expenses:* $%.6f\n", settings.TotalExpenses)
	message += fmt.Sprintf("*Chat History:* %d messages\n", len(settings.ChatHistory))
	message += fmt.Sprintf("*Custom Models:* %d\n", len(settings.CustomModels))
	message += fmt.Sprintf("*Last Updated:* %s\n", settings.LastUpdated.Format("2006-01-02 15:04:05"))

	if len(settings.ExpenseHistory) > 0 {
		lastExpense := settings.ExpenseHistory[len(settings.ExpenseHistory)-1]
		message += fmt.Sprintf("*Last Activity:* %s\n", lastExpense.Timestamp.Format("2006-01-02 15:04:05"))
	}

	message += "\n*Quick Actions:*\n"
	message += "Use the buttons below for easy navigation."

	keyboard := b.createMainMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "Markdown", keyboard)
}
