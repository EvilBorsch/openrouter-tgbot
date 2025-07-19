package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// createMainMenuKeyboard creates the main menu inline keyboard
func (b *Bot) createMainMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Settings", "settings"),
			tgbotapi.NewInlineKeyboardButtonData("📊 Expenses", "expenses"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🤖 Models", "listmodels"),
			tgbotapi.NewInlineKeyboardButtonData("📈 Status", "status"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Clear History", "clear"),
			tgbotapi.NewInlineKeyboardButtonData("❓ Help", "help"),
		),
	)
	return &keyboard
}

// createSettingsKeyboard creates the settings menu keyboard
func (b *Bot) createSettingsKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💬 Chat Mode", "chat_mode"),
			tgbotapi.NewInlineKeyboardButtonData("🤖 Change Model", "change_model"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Add Custom Model", "add_model"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Menu", "back_to_menu"),
		),
	)
	return &keyboard
}

// createChatModeKeyboard creates the chat mode selection keyboard
func (b *Bot) createChatModeKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📝 With History", "mode_with_history"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Without History", "mode_without_history"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Settings", "settings"),
		),
	)
	return &keyboard
}

// createModelSelectionKeyboard creates a model selection keyboard with popular models
func (b *Bot) createModelSelectionKeyboard() *tgbotapi.InlineKeyboardMarkup {
	// Popular models with shortened display names
	models := []struct {
		display string
		value   string
	}{
		{"GPT-4", "openai/gpt-4"},
		{"GPT-3.5 Turbo", "openai/gpt-3.5-turbo"},
		{"Claude Sonnet", "anthropic/claude-3-sonnet"},
		{"Gemini Pro", "google/gemini-pro"},
		{"Mistral 7B", "mistralai/mistral-7b-instruct"},
		{"Llama 2 70B", "meta-llama/llama-2-70b-chat"},
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	// Create rows of 2 buttons each
	for i := 0; i < len(models); i += 2 {
		var row []tgbotapi.InlineKeyboardButton
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(models[i].display, "model_"+models[i].value))

		if i+1 < len(models) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(models[i+1].display, "model_"+models[i+1].value))
		}
		rows = append(rows, row)
	}

	// Add navigation buttons
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📋 All Models", "listmodels"),
	))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Settings", "settings"),
	))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	return &keyboard
}

// createConfirmationKeyboard creates a yes/no confirmation keyboard
func (b *Bot) createConfirmationKeyboard(action string) *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Yes", "confirm_"+action),
			tgbotapi.NewInlineKeyboardButtonData("❌ No", "cancel_"+action),
		),
	)
	return &keyboard
}

// createBackToMenuKeyboard creates a simple back to menu button
func (b *Bot) createBackToMenuKeyboard() *tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Back to Menu", "back_to_menu"),
		),
	)
	return &keyboard
}
