package testdata

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

// MockHTTPClient implements a mock HTTP client for testing
type MockHTTPClient struct {
	Requests  []MockRequest
	Responses []MockResponse
	Index     int
}

type MockRequest struct {
	Method string
	URL    string
	Body   string
	Headers map[string]string
}

type MockResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		Requests:  []MockRequest{},
		Responses: []MockResponse{},
		Index:     0,
	}
}

func (m *MockHTTPClient) AddResponse(statusCode int, body string, headers map[string]string) {
	if headers == nil {
		headers = make(map[string]string)
	}
	m.Responses = append(m.Responses, MockResponse{
		StatusCode: statusCode,
		Body:       body,
		Headers:    headers,
	})
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Record the request
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	
	headers := make(map[string]string)
	for k, v := range req.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	
	m.Requests = append(m.Requests, MockRequest{
		Method:  req.Method,
		URL:     req.URL.String(),
		Body:    string(bodyBytes),
		Headers: headers,
	})
	
	// Return mock response
	if m.Index >= len(m.Responses) {
		// Default response if no more mocked responses
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader(`{"error": {"message": "No mock response configured"}}`)),
			Header:     make(http.Header),
		}, nil
	}
	
	resp := m.Responses[m.Index]
	m.Index++
	
	header := make(http.Header)
	for k, v := range resp.Headers {
		header.Set(k, v)
	}
	
	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(strings.NewReader(resp.Body)),
		Header:     header,
	}, nil
}

func (m *MockHTTPClient) GetLastRequest() *MockRequest {
	if len(m.Requests) == 0 {
		return nil
	}
	return &m.Requests[len(m.Requests)-1]
}

func (m *MockHTTPClient) GetAllRequests() []MockRequest {
	return m.Requests
}

func (m *MockHTTPClient) Reset() {
	m.Requests = []MockRequest{}
	m.Responses = []MockResponse{}
	m.Index = 0
}

// HTTPClientInterface defines the interface for HTTP clients
type HTTPClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// Default HTTP client wrapper
type DefaultHTTPClient struct {
	client *http.Client
}

func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (d *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return d.client.Do(req)
}

// Common test responses
func GetSuccessfulOpenAIResponse() string {
	return `{
		"choices": [
			{
				"message": {
					"role": "assistant",
					"content": "これは音声の要約です。\n\n主なポイント:\n- 重要な話題について議論\n- 具体的な提案が行われた\n- 次のステップが決定された"
				}
			}
		]
	}`
}

func GetOpenAIErrorResponse() string {
	return `{
		"error": {
			"message": "Invalid API key provided",
			"type": "invalid_request_error",
			"code": "invalid_api_key"
		}
	}`
}

func GetOpenAIRateLimitResponse() string {
	return `{
		"error": {
			"message": "Rate limit exceeded",
			"type": "rate_limit_error",
			"code": "rate_limit_exceeded"
		}
	}`
}

func GetOpenAITokenLimitResponse() string {
	return `{
		"error": {
			"message": "This model's maximum context length is 4096 tokens",
			"type": "invalid_request_error",
			"code": "context_length_exceeded"
		}
	}`
}