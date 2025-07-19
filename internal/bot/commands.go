package bot

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// handleStartCommand handles the /start and /help commands
func (b *Bot) handleStartCommand(userID int64) {
	welcomeMessage := `🤖 <i>Welcome to LLM Chat Bot!</i>

This bot provides access to various LLM models through OpenRouter.

<i>Quick Start:</i>
• Just type any message to chat with the AI
• Use the menu buttons below for settings and commands
• The bot supports advanced formatting

<i>Features:</i>
✅ Multiple LLM models support
✅ Chat history modes
✅ Expense tracking
✅ Custom model management
✅ Message splitting for long responses

Use the buttons below to get started!`

	keyboard := b.createMainMenuKeyboard()
	b.sendMessageWithKeyboard(userID, welcomeMessage, "HTML", keyboard)
}

// handleMenuCommand handles the /menu command
func (b *Bot) handleMenuCommand(userID int64) {
	menuMessage := `📋 <i>Main Menu</i>

Welcome to your AI assistant! Choose an option below or just start typing to chat with the AI.

<i>Quick Actions:</i>
• 💬 Just type a message to chat
• ⚙️ Settings - Configure chat mode and models
• 📊 View expenses and usage statistics
• 🤖 Browse and change AI models

<i>Features:</i>
✅ Interactive button controls
✅ Multiple LLM models
✅ Chat history management
✅ Real-time expense tracking

Choose an option below:`

	keyboard := b.createMainMenuKeyboard()
	b.sendMessageWithKeyboard(userID, menuMessage, "HTML", keyboard)
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

		message := fmt.Sprintf("🔧 <i>Current chat mode:</i> <code>%s</code>\n\n", settings.ChatMode)
		message += "<i>Available modes:</i>\n"
		message += "• <code>with_history</code> - AI remembers previous messages\n"
		message += "• <code>without_history</code> - Each message is independent\n\n"
		message += "<i>Usage:</i> <code>/mode with_history</code> or <code>/mode without_history</code>"

		b.sendMessage(userID, message)
		return
	}

	mode := strings.ToLower(strings.TrimSpace(args))
	if mode != "with_history" && mode != "without_history" {
		b.sendMessage(userID, "❌ Invalid mode. Use: <code>with_history</code> or <code>without_history</code>")
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

	message := fmt.Sprintf("✅ Chat mode changed to: <code>%s</code>", mode)
	if mode == "with_history" {
		message += "\n\n<i>Note:</i> The AI will now remember your previous messages in this session."
	} else {
		message += "\n\n<i>Note:</i> Each message will be processed independently."
	}

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
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

		message := fmt.Sprintf("🤖 <i>Current model:</i> <code>%s</code>\n\n", settings.CurrentModel)
		message += "<i>Popular models:</i>\n"
		message += "• <code>openai/gpt-4</code> - Most capable, higher cost\n"
		message += "• <code>openai/gpt-3.5-turbo</code> - Fast and affordable\n"
		message += "• <code>anthropic/claude-3-sonnet</code> - Great for analysis\n"
		message += "• <code>google/gemini-pro</code> - Google's latest model\n\n"
		message += "<i>Usage:</i> <code>/model openai/gpt-4</code>\n"
		message += "<i>See all:</i> <code>/listmodels</code>"

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

	message := fmt.Sprintf("✅ Model changed to: <code>%s</code>\n\n", model)
	message += "<i>Tip:</i> The pricing and capabilities may vary between models. Check expenses to monitor usage."

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleAddModelCommand handles the /addmodel command
func (b *Bot) handleAddModelCommand(userID int64, args string) {
	if args == "" {
		message := "🔧 <i>Add Custom Model</i>\n\n"
		message += "<i>Usage:</i> <code>/addmodel model-provider/model-name</code>\n\n"
		message += "<i>Examples:</i>\n"
		message += "• <code>/addmodel mistralai/mistral-7b-instruct</code>\n"
		message += "• <code>/addmodel meta-llama/llama-2-70b-chat</code>\n"
		message += "• <code>/addmodel cohere/command-r-plus</code>\n\n"
		message += "<i>Note:</i> Make sure the model is available on OpenRouter."

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
			b.sendMessage(userID, fmt.Sprintf("❌ Model <code>%s</code> is already in your list.", model))
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

	message := fmt.Sprintf("✅ Added model: <code>%s</code>\n\n", model)
	message += "You can now use it with: <code>/model " + model + "</code>"

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

	message := "🤖 <i>Available Models</i>\n\n"
	message += fmt.Sprintf("<i>Current:</i> <code>%s</code> ✅\n\n", settings.CurrentModel)

	message += "<i>Popular Models:</i>\n"
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
		message += fmt.Sprintf("• <code>%s</code>\n", model)
	}

	if len(settings.CustomModels) > 0 {
		message += "\n<i>Your Custom Models:</i>\n"
		for _, model := range settings.CustomModels {
			if model == settings.CurrentModel {
				continue // Skip current model as it's already shown
			}
			message += fmt.Sprintf("• <code>%s</code>\n", model)
		}
	}

	message += "\n<i>Usage:</i> Click a model button above or type <code>/model model-name</code>"

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleExpensesCommand handles the /expenses command
func (b *Bot) handleExpensesCommand(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "💰 <i>Your Usage Statistics</i>\n\n"
	message += fmt.Sprintf("<i>Total Expenses:</i> $%.6f\n", settings.TotalExpenses)
	message += fmt.Sprintf("<i>Total Requests:</i> %d\n", len(settings.ExpenseHistory))
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

		message += fmt.Sprintf("<i>Total Tokens:</i> %d\n", totalTokens)
		message += fmt.Sprintf("<i>Last 7 Days:</i> $%.6f\n\n", recentExpenses)

		// Show model usage
		message += "<i>Model Usage:</i>\n"
		for model, count := range modelUsage {
			message += fmt.Sprintf("• <code>%s</code>: %d requests\n", model, count)
		}

		// Show recent transactions (last 5)
		message += "\n<i>Recent Transactions:</i>\n"
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
		message += "\n<i>No usage data yet.</i> Start chatting to see your statistics!"
	}

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleClearCommand handles the /clear command
func (b *Bot) handleClearCommand(userID int64) {
	if err := b.storage.ClearChatHistory(userID); err != nil {
		log.Errorf("Failed to clear chat history: %v", err)
		b.sendMessage(userID, "Error clearing chat history.")
		return
	}

	message := "🗑️ <i>Chat history cleared!</i>\n\n"
	message += "Your conversation history has been deleted. The AI will start fresh with your next message."

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleStatusCommand handles the /status command
func (b *Bot) handleStatusCommand(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		log.Errorf("Failed to get user settings: %v", err)
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "📊 <i>Your Current Settings</i>\n\n"
	message += fmt.Sprintf("<i>User ID:</i> <code>%d</code>\n", settings.UserID)
	message += fmt.Sprintf("<i>Current Model:</i> <code>%s</code>\n", settings.CurrentModel)
	message += fmt.Sprintf("<i>Chat Mode:</i> <code>%s</code>\n", settings.ChatMode)
	message += fmt.Sprintf("<i>Total Expenses:</i> $%.6f\n", settings.TotalExpenses)
	message += fmt.Sprintf("<i>Chat History:</i> %d messages\n", len(settings.ChatHistory))
	message += fmt.Sprintf("<i>Custom Models:</i> %d\n", len(settings.CustomModels))
	message += fmt.Sprintf("<i>Last Updated:</i> %s\n", settings.LastUpdated.Format("2006-01-02 15:04:05"))

	if len(settings.ExpenseHistory) > 0 {
		lastExpense := settings.ExpenseHistory[len(settings.ExpenseHistory)-1]
		message += fmt.Sprintf("<i>Last Activity:</i> %s\n", lastExpense.Timestamp.Format("2006-01-02 15:04:05"))
	}

	message += "\n<i>Quick Actions:</i>\n"
	message += "Use the buttons below for easy navigation."

	keyboard := b.createMainMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}
