package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds all configuration for the bot
type Config struct {
	// Telegram Bot Token
	TelegramToken string `json:"telegram_token"`

	// OpenRouter API Key
	OpenRouterAPIKey string `json:"openrouter_api_key"`

	// OpenRouter Base URL
	OpenRouterBaseURL string `json:"openrouter_base_url"`

	// List of allowed Telegram user IDs
	AllowedUsers []int64 `json:"allowed_users"`

	// Default model for new users
	DefaultModel string `json:"default_model"`

	// Default chat mode (with_history or without_history)
	DefaultChatMode string `json:"default_chat_mode"`

	// Maximum message length before splitting
	MaxMessageLength int `json:"max_message_length"`

	// Log level
	LogLevel string `json:"log_level"`

	// Data directory for persistence
	DataDirectory string `json:"data_directory"`
}

// Load loads configuration from a JSON file
func Load(filename string) (*Config, error) {
	// Default configuration
	config := &Config{
		OpenRouterBaseURL: "https://openrouter.ai/api/v1",
		DefaultModel:      "openai/gpt-3.5-turbo",
		DefaultChatMode:   "without_history",
		MaxMessageLength:  4096,
		LogLevel:          "info",
		DataDirectory:     "data",
	}

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Create a default config file
		if err := config.Save(filename); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, fmt.Errorf("config file %s created with defaults. Please fill in required values (telegram_token, openrouter_api_key, allowed_users)", filename)
	}

	// Read configuration file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if config.TelegramToken == "" {
		return nil, fmt.Errorf("telegram_token is required")
	}
	if config.OpenRouterAPIKey == "" {
		return nil, fmt.Errorf("openrouter_api_key is required")
	}
	if len(config.AllowedUsers) == 0 {
		return nil, fmt.Errorf("allowed_users list cannot be empty")
	}

	return config, nil
}

// Save saves the configuration to a JSON file
func (c *Config) Save(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IsUserAllowed checks if a user ID is in the allowed users list
func (c *Config) IsUserAllowed(userID int64) bool {
	for _, id := range c.AllowedUsers {
		if id == userID {
			return true
		}
	}
	return false
}
