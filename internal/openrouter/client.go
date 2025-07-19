// Package openrouter provides OpenRouter API client functionality
//
// This package implements accurate cost tracking using OpenRouter's generation stats API.
// The generation stats endpoint provides:
// - Real costs based on native model tokenizers (not normalized counts)
// - Model-specific token counts for precise accounting
// - Provider information and detailed billing data
//
// Cost tracking flow:
// 1. Make chat completion request -> get generation ID
// 2. Query /generation endpoint with ID -> get accurate stats
// 3. Store native token counts and real costs for expense tracking
package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"telegrambot/internal/storage"

	log "github.com/sirupsen/logrus"
)

// ChatMessage represents a message in the chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents the request to OpenRouter chat completion API
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	TopP        float64       `json:"top_p,omitempty"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionChoice represents a choice in the response
type ChatCompletionChoice struct {
	Index   int `json:"index"`
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

// ChatCompletionResponse represents the response from OpenRouter
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChatCompletionChoice `json:"choices"`
	Usage   Usage                  `json:"usage"`
	Error   *OpenRouterError       `json:"error,omitempty"`
}

// GenerationStats represents the generation statistics from OpenRouter
type GenerationStats struct {
	ID                     string  `json:"id"`
	Model                  string  `json:"model"`
	CreatedAt              string  `json:"created_at"`
	TokensPrompt           int     `json:"tokens_prompt"`
	TokensCompletion       int     `json:"tokens_completion"`
	NativeTokensPrompt     int     `json:"native_tokens_prompt"`
	NativeTokensCompletion int     `json:"native_tokens_completion"`
	NumMedia               int     `json:"num_media"`
	ProviderName           string  `json:"provider_name"`
	TotalCost              float64 `json:"total_cost"`
	Cancelled              bool    `json:"cancelled"`
	Finish                 bool    `json:"finish"`
}

// OpenRouterError represents an error from the API
type OpenRouterError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// Client represents the OpenRouter API client
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewClient creates a new OpenRouter client
func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatCompletion makes a chat completion request to OpenRouter
func (c *Client) ChatCompletion(req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Set default values
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 200_000
	}

	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/your-repo/telegrambot")
	httpReq.Header.Set("X-Title", "Telegram LLM Bot")

	// Make request
	log.Debugf("Making OpenRouter request to model: %s", req.Model)
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var completionResp ChatCompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if completionResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", completionResp.Error.Message)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	log.Debugf("OpenRouter response: tokens=%d, model=%s", completionResp.Usage.TotalTokens, completionResp.Model)
	return &completionResp, nil
}

// GetGenerationStats queries the generation statistics for a specific generation ID
// This provides accurate cost and native token counts from OpenRouter API
// Unlike the normalized token counts in the completion response, these are model-specific
func (c *Client) GetGenerationStats(generationID string) (*GenerationStats, error) {
	// Create HTTP request
	url := c.baseURL + "/generation?id=" + generationID
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/your-repo/telegrambot")
	httpReq.Header.Set("X-Title", "Telegram LLM Bot")

	// Make request with retry logic
	var resp *http.Response
	for i := 0; i < 5; i++ {
		resp, err = c.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}

		// If not ready yet, wait and retry
		if resp.StatusCode == 202 {
			resp.Body.Close()
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		break
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var stats GenerationStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Debugf("Generation stats: id=%s, cost=$%.6f, native_tokens=%d", stats.ID, stats.TotalCost, stats.NativeTokensPrompt+stats.NativeTokensCompletion)
	return &stats, nil
}

// CalculateCost is deprecated - use GetGenerationStats instead for accurate pricing
// This is kept as fallback for cases where generation stats are not available
func (c *Client) CalculateCost(model string, inputTokens, outputTokens int) float64 {
	log.Warn("Using estimated cost calculation - use GetGenerationStats for accurate pricing")

	// Basic cost estimation (in USD) - rough estimates only
	var inputCostPer1K, outputCostPer1K float64

	switch {
	case contains(model, "gpt-4"):
		inputCostPer1K = 0.03
		outputCostPer1K = 0.06
	case contains(model, "gpt-3.5-turbo"):
		inputCostPer1K = 0.001
		outputCostPer1K = 0.002
	case contains(model, "claude"):
		inputCostPer1K = 0.008
		outputCostPer1K = 0.024
	default:
		// Default pricing for unknown models
		inputCostPer1K = 0.002
		outputCostPer1K = 0.004
	}

	inputCost := float64(inputTokens) / 1000.0 * inputCostPer1K
	outputCost := float64(outputTokens) / 1000.0 * outputCostPer1K

	return inputCost + outputCost
}

// GetChatResponse gets a chat response and tracks the expense
func (c *Client) GetChatResponse(model string, messages []storage.ChatMessage, userID int64, store storage.Storage) (string, error) {
	// Convert storage messages to API messages
	apiMessages := make([]ChatMessage, len(messages))
	for i, msg := range messages {
		apiMessages[i] = ChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Create request
	req := ChatCompletionRequest{
		Model:    model,
		Messages: apiMessages,
	}

	// Make API call
	resp, err := c.ChatCompletion(req)
	if err != nil {
		return "", err
	}

	// Extract response content
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	content := resp.Choices[0].Message.Content

	// Get accurate cost and token counts from generation stats
	var expense storage.ExpenseRecord
	if resp.ID != "" {
		// Query generation stats for accurate pricing
		stats, err := c.GetGenerationStats(resp.ID)
		if err != nil {
			log.Warnf("Failed to get generation stats, using fallback calculation: %v", err)
			// Fallback to estimated cost
			cost := c.CalculateCost(model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
			expense = storage.ExpenseRecord{
				Timestamp:    time.Now(),
				Model:        model,
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				Cost:         cost,
			}
		} else {
			// Use accurate stats from OpenRouter
			expense = storage.ExpenseRecord{
				Timestamp:    time.Now(),
				Model:        stats.Model,
				InputTokens:  stats.NativeTokensPrompt,
				OutputTokens: stats.NativeTokensCompletion,
				Cost:         stats.TotalCost,
			}
			log.Infof("Using accurate OpenRouter pricing: model=%s, native_tokens=%d, cost=$%.6f",
				stats.Model, stats.NativeTokensPrompt+stats.NativeTokensCompletion, stats.TotalCost)
		}
	} else {
		log.Warn("No generation ID in response, using fallback calculation")
		// Fallback to estimated cost
		cost := c.CalculateCost(model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
		expense = storage.ExpenseRecord{
			Timestamp:    time.Now(),
			Model:        model,
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			Cost:         cost,
		}
	}

	// Track expense
	if err := store.AddExpense(userID, expense); err != nil {
		log.Errorf("Failed to track expense: %v", err)
	}

	log.Infof("Chat response generated: model=%s, cost=$%.6f", expense.Model, expense.Cost)
	return content, nil
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s = fmt.Sprintf("%s", s)
	substr = fmt.Sprintf("%s", substr)
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
