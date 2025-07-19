package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ChatMessage represents a message in chat history
type ChatMessage struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ExpenseRecord represents an expense record for API calls
type ExpenseRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	Model        string    `json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	Cost         float64   `json:"cost"`
}

// UserSettings represents user-specific settings
type UserSettings struct {
	UserID         int64           `json:"user_id"`
	CurrentModel   string          `json:"current_model"`
	ChatMode       string          `json:"chat_mode"` // "with_history" or "without_history"
	CustomModels   []string        `json:"custom_models"`
	TotalExpenses  float64         `json:"total_expenses"`
	ExpenseHistory []ExpenseRecord `json:"expense_history"`
	ChatHistory    []ChatMessage   `json:"chat_history"`
	LastUpdated    time.Time       `json:"last_updated"`
}

// Storage interface defines methods for data persistence
type Storage interface {
	GetUserSettings(userID int64) (*UserSettings, error)
	SaveUserSettings(settings *UserSettings) error
	AddExpense(userID int64, expense ExpenseRecord) error
	GetTotalExpenses(userID int64) (float64, error)
	AddChatMessage(userID int64, message ChatMessage) error
	GetChatHistory(userID int64) ([]ChatMessage, error)
	ClearChatHistory(userID int64) error
	Close() error
}

// FileStorage implements Storage interface using file system
type FileStorage struct {
	dataDir string
	mutex   sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(dataDir string) (*FileStorage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	return &FileStorage{
		dataDir: dataDir,
	}, nil
}

// getUserFilePath returns the file path for user settings
func (fs *FileStorage) getUserFilePath(userID int64) string {
	return filepath.Join(fs.dataDir, fmt.Sprintf("user_%d.json", userID))
}

// GetUserSettings retrieves user settings
func (fs *FileStorage) GetUserSettings(userID int64) (*UserSettings, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	filePath := fs.getUserFilePath(userID)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return default settings for new user
		return &UserSettings{
			UserID:         userID,
			CurrentModel:   "openai/gpt-3.5-turbo",
			ChatMode:       "without_history",
			CustomModels:   []string{},
			TotalExpenses:  0,
			ExpenseHistory: []ExpenseRecord{},
			ChatHistory:    []ChatMessage{},
			LastUpdated:    time.Now(),
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user settings: %w", err)
	}

	var settings UserSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse user settings: %w", err)
	}

	return &settings, nil
}

// SaveUserSettings saves user settings to file
func (fs *FileStorage) SaveUserSettings(settings *UserSettings) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	settings.LastUpdated = time.Now()

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user settings: %w", err)
	}

	filePath := fs.getUserFilePath(settings.UserID)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write user settings: %w", err)
	}

	return nil
}

// AddExpense adds an expense record to user's history
func (fs *FileStorage) AddExpense(userID int64, expense ExpenseRecord) error {
	settings, err := fs.GetUserSettings(userID)
	if err != nil {
		return err
	}

	settings.ExpenseHistory = append(settings.ExpenseHistory, expense)
	settings.TotalExpenses += expense.Cost

	return fs.SaveUserSettings(settings)
}

// GetTotalExpenses returns total expenses for a user
func (fs *FileStorage) GetTotalExpenses(userID int64) (float64, error) {
	settings, err := fs.GetUserSettings(userID)
	if err != nil {
		return 0, err
	}

	return settings.TotalExpenses, nil
}

// AddChatMessage adds a message to chat history
func (fs *FileStorage) AddChatMessage(userID int64, message ChatMessage) error {
	settings, err := fs.GetUserSettings(userID)
	if err != nil {
		return err
	}

	settings.ChatHistory = append(settings.ChatHistory, message)

	// Keep only last 50 messages to avoid too large files
	if len(settings.ChatHistory) > 50 {
		settings.ChatHistory = settings.ChatHistory[len(settings.ChatHistory)-50:]
	}

	return fs.SaveUserSettings(settings)
}

// GetChatHistory returns chat history for a user
func (fs *FileStorage) GetChatHistory(userID int64) ([]ChatMessage, error) {
	settings, err := fs.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}

	return settings.ChatHistory, nil
}

// ClearChatHistory clears chat history for a user
func (fs *FileStorage) ClearChatHistory(userID int64) error {
	settings, err := fs.GetUserSettings(userID)
	if err != nil {
		return err
	}

	settings.ChatHistory = []ChatMessage{}
	return fs.SaveUserSettings(settings)
}

// Close closes the storage (no-op for file storage)
func (fs *FileStorage) Close() error {
	return nil
}
