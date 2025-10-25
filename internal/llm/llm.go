package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/infoHiroki/KoeMoji-Go/internal/config"
	"github.com/infoHiroki/KoeMoji-Go/internal/logger"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// OpenAI API structures
type OpenAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// SummarizeText generates a summary of the given text using LLM API
func SummarizeText(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry,
	logMutex *sync.RWMutex, debugMode bool, text string) (string, error) {

	if !config.LLMSummaryEnabled {
		return "", fmt.Errorf("LLM summary is disabled")
	}

	if config.LLMAPIKey == "" {
		return "", fmt.Errorf("LLM API key is not configured")
	}

	// Prepare prompt
	prompt := preparePrompt(config, text)

	// Call API based on provider
	switch config.LLMAPIProvider {
	case "openai":
		return callOpenAI(config, log, logBuffer, logMutex, debugMode, prompt)
	default:
		return "", fmt.Errorf("unsupported LLM provider: %s", config.LLMAPIProvider)
	}
}

func preparePrompt(config *config.Config, text string) string {
	prompt := config.SummaryPromptTemplate
	language := getSummaryLanguage(config)

	// 後方互換性: {text}と{language}プレースホルダーがあれば置換
	hasTextPlaceholder := strings.Contains(prompt, "{text}")
	hasLanguagePlaceholder := strings.Contains(prompt, "{language}")

	if hasTextPlaceholder || hasLanguagePlaceholder {
		// 旧形式: プレースホルダーを置換
		prompt = strings.ReplaceAll(prompt, "{text}", text)
		prompt = strings.ReplaceAll(prompt, "{language}", language)
	} else {
		// 新形式: テンプレートの末尾にテキストを自動追加
		prompt = prompt + "\n\n" + text
	}

	return prompt
}

func getSummaryLanguage(config *config.Config) string {
	switch config.SummaryLanguage {
	case "auto":
		// Use the same language as the transcription
		if config.Language == "ja" {
			return "日本語"
		}
		return "英語"
	case "ja":
		return "日本語"
	case "en":
		return "英語"
	default:
		return "日本語"
	}
}

func callOpenAI(config *config.Config, log *log.Logger, logBuffer *[]logger.LogEntry,
	logMutex *sync.RWMutex, debugMode bool, prompt string) (string, error) {

	// Prepare request
	request := OpenAIRequest{
		Model:     config.LLMModel,
		MaxTokens: config.LLMMaxTokens,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.LogDebug(log, logBuffer, logMutex, debugMode, "OpenAI API request prepared")
	logger.LogDebug(log, logBuffer, logMutex, debugMode, "OpenAI API request JSON: %s", string(jsonData))

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.LLMAPIKey)

	// Make request with retry logic
	var response *http.Response
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		client := &http.Client{
			Timeout: 5 * time.Minute,
		}

		response, err = client.Do(req)
		if err != nil {
			logger.LogError(log, logBuffer, logMutex, "OpenAI API request failed (attempt %d/%d): %v", attempt, maxRetries, err)
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * 10 * time.Second)
				continue
			}
			return "", fmt.Errorf("failed to call OpenAI API after %d attempts: %w", maxRetries, err)
		}

		if response.StatusCode == 429 {
			// Rate limit hit
			logger.LogInfo(log, logBuffer, logMutex, "Rate limit hit, waiting 60 seconds...")
			response.Body.Close()
			if attempt < maxRetries {
				time.Sleep(60 * time.Second)
				continue
			}
		}

		break
	}
	defer response.Body.Close()

	// Read response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	logger.LogDebug(log, logBuffer, logMutex, debugMode, "OpenAI API response received (status: %d)", response.StatusCode)

	// Check HTTP status code before parsing
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		bodyStr := string(body)
		logger.LogError(log, logBuffer, logMutex, "OpenAI API error (status %d). Full response: %s", response.StatusCode, bodyStr)
		return "", fmt.Errorf("OpenAI API request failed with status %d: %s", response.StatusCode, bodyStr)
	}

	// Parse response
	var apiResponse OpenAIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// Log the raw response for debugging
		logger.LogError(log, logBuffer, logMutex, "Failed to parse OpenAI response. Status: %d, Body (first 500 chars): %s", response.StatusCode, string(body[:min(500, len(body))]))
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if apiResponse.Error != nil {
		return "", fmt.Errorf("OpenAI API error: %s", apiResponse.Error.Message)
	}

	if len(apiResponse.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI API")
	}

	summary := apiResponse.Choices[0].Message.Content
	logger.LogDebug(log, logBuffer, logMutex, debugMode, "Summary generated successfully (%d characters)", len(summary))

	return summary, nil
}

// ValidateAPIKey tests the API key by making a simple request
func ValidateAPIKey(config *config.Config) error {
	if config.LLMAPIKey == "" {
		return fmt.Errorf("API key is empty")
	}

	// Simple test request
	request := OpenAIRequest{
		Model:     "gpt-3.5-turbo",
		MaxTokens: 1,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Test",
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal test request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.LLMAPIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to test API key: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == 401 {
		return fmt.Errorf("invalid API key")
	}

	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("API test failed with status %d: %s", response.StatusCode, string(body))
	}

	return nil
}
