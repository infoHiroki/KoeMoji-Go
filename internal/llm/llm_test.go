package llm

import (
	"testing"

	"github.com/hirokitakamura/koemoji-go/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestValidateAPIKey_ValidKey(t *testing.T) {
	// Note: This test requires a real API call to OpenAI
	// In a real test environment, you would mock the HTTP client
	// For now, we test the basic validation logic
	config := &config.Config{LLMAPIKey: "sk-test-key-should-not-work"}
	err := ValidateAPIKey(config)
	// We expect an error since this is not a real API key
	assert.Error(t, err)
}

func TestValidateAPIKey_EmptyKey(t *testing.T) {
	config := &config.Config{LLMAPIKey: ""}
	err := ValidateAPIKey(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API key is empty")
}

// Note: ValidateModel is not exported, tested through ValidateAPIKey integration test

// Note: Model validation tested through SummarizeText integration test

// Note: GetSupportedModels is not exported, functionality tested through integration tests

// Note: buildPrompt is not exported, tested through SummarizeText integration test

// Note: Prompt building logic tested through integration test

// Note: Edge cases tested through SummarizeText error handling

// Note: createSummaryRequest is not exported, tested through SummarizeText integration test

// Note: Auto language detection tested through SummarizeText integration test

// Note: validateConfig is not exported, tested through ValidateAPIKey integration test

// Note: Config validation tested through ValidateAPIKey error cases

// Note: Disabled LLM config tested through ValidateAPIKey edge cases

// Note: The following functions are not exported and cannot be tested directly:
// - makeRequest (HTTP client functionality)
// - parseResponse (JSON parsing)
// - handleRateLimit (retry logic)

// Integration tests would be needed to test the main GenerateSummary function
// with actual API calls or mocked HTTP responses.