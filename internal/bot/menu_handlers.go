package bot

import (
	"fmt"
)

// handleSettingsMenu shows the settings menu with buttons
func (b *Bot) handleSettingsMenu(userID int64) {
	message := "⚙️ <i>Settings Menu</i>\n\n"
	message += "Choose what you'd like to configure:"

	keyboard := b.createSettingsKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleChatModeMenu shows the chat mode selection menu
func (b *Bot) handleChatModeMenu(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "💬 <i>Chat Mode Settings</i>\n\n"
	message += fmt.Sprintf("<i>Current mode:</i> <code>%s</code>\n\n", settings.ChatMode)
	message += "<i>Available modes:</i>\n"
	message += "• <b>With History</b> - AI remembers previous messages\n"
	message += "• <b>Without History</b> - Each message is independent\n\n"
	message += "Select your preferred mode:"

	keyboard := b.createChatModeKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleModelSelectionMenu shows the model selection menu
func (b *Bot) handleModelSelectionMenu(userID int64) {
	settings, err := b.storage.GetUserSettings(userID)
	if err != nil {
		b.sendMessage(userID, "Error retrieving your settings.")
		return
	}

	message := "🤖 <i>Model Selection</i>\n\n"
	message += fmt.Sprintf("<i>Current model:</i> <code>%s</code>\n\n", settings.CurrentModel)
	message += "Choose from popular models or view all available models:"

	keyboard := b.createModelSelectionKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleClearWithConfirmation shows confirmation before clearing
func (b *Bot) handleClearWithConfirmation(userID int64) {
	message := "🗑️ <i>Clear Chat History</i>\n\n"
	message += "Are you sure you want to clear your chat history?\n"
	message += "This action cannot be undone."

	keyboard := b.createConfirmationKeyboard("clear")
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}

// handleAddModelPrompt prompts user to add a custom model
func (b *Bot) handleAddModelPrompt(userID int64) {
	message := "➕ <i>Add Custom Model</i>\n\n"
	message += "To add a custom model, use this command format:\n"
	message += "<code>/addmodel provider/model-name</code>\n\n"
	message += "<i>Examples:</i>\n"
	message += "• <code>/addmodel mistralai/mistral-7b-instruct</code>\n"
	message += "• <code>/addmodel meta-llama/llama-2-70b-chat</code>\n"
	message += "• <code>/addmodel cohere/command-r-plus</code>\n\n"
	message += "<i>Note:</i> Make sure the model is available on OpenRouter."

	keyboard := b.createBackToMenuKeyboard()
	b.sendMessageWithKeyboard(userID, message, "HTML", keyboard)
}
