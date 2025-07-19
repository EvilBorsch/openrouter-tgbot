package bot

import (
	"fmt"
)

// handleSettingsMenu shows the settings menu with buttons
func (b *Bot) handleSettingsMenu(userID int64) {
	message := "⚙️ *Settings Menu*\n\n"
	message += "Choose what you'd like to configure:"

	keyboard := b.createSettingsKeyboard()
	b.sendMessageWithKeyboard(userID, message, "MarkdownV2", keyboard)
}

// handleChatModeMenu shows the chat mode selection menu
func (b *Bot) handleChatModeMenu(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "💬 *Chat Mode Settings*\n\n"
	message += fmt.Sprintf("*Current mode:* `%s`\n\n", settings.ChatMode)
	message += "*Available modes:*\n"
	message += "• **With History** - AI remembers previous messages\n"
	message += "• **Without History** - Each message is independent\n\n"
	message += "Select your preferred mode:"

	keyboard := b.createChatModeKeyboard()
	b.sendMessageWithKeyboard(userID, message, "MarkdownV2", keyboard)
}

// handleModelSelectionMenu shows the model selection menu
func (b *Bot) handleModelSelectionMenu(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "🤖 *Model Selection*\n\n"
	message += fmt.Sprintf("*Current model:* `%s`\n\n", settings.CurrentModel)
	message += "Choose from popular models or view all available models:"

	keyboard := b.createModelSelectionKeyboard()
	b.sendMessageWithKeyboard(userID, message, "MarkdownV2", keyboard)
}

// handleClearWithConfirmation shows confirmation before clearing
func (b *Bot) handleClearWithConfirmation(userID int64) {
	message := "🗑️ *Clear Chat History*\n\n"
	message += "Are you sure you want to clear your chat history?\n"
	message += "This action cannot be undone."

	keyboard := b.createConfirmationKeyboard("clear")
	b.sendMessageWithKeyboard(userID, message, "MarkdownV2", keyboard)
}

// handleAddModelPrompt prompts user to add a custom model
func (b *Bot) handleAddModelPrompt(userID int64) {
	message := "➕ *Add Custom Model*\n\n"
	message += "To add a custom model, use this command format:\n"
	message += "`/addmodel provider/model-name`\n\n"
	message += "*Examples:*\n"
	message += "• `/addmodel mistralai/mistral-7b-instruct`\n"
	message += "• `/addmodel meta-llama/llama-2-70b-chat`\n"
	message += "• `/addmodel cohere/command-r-plus`\n\n"
	message += "*Note:* Make sure the model is available on OpenRouter."

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "MarkdownV2", keyboard)
}
