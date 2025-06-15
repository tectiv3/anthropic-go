package anthropic

import (
	"net/http"
	"testing"
	"time"
)

func TestReasoningEffort_IsValid(t *testing.T) {
	tests := []struct {
		effort   ReasoningEffort
		expected bool
	}{
		{ReasoningEffortLow, true},
		{ReasoningEffortMedium, true},
		{ReasoningEffortHigh, true},
		{ReasoningEffort("invalid"), false},
		{ReasoningEffort(""), false},
	}

	for _, test := range tests {
		result := test.effort.IsValid()
		if result != test.expected {
			t.Errorf("IsValid(%q) = %v, expected %v", test.effort, result, test.expected)
		}
	}
}

func TestCacheControlType_String(t *testing.T) {
	tests := []struct {
		cacheType CacheControlType
		expected  string
	}{
		{CacheControlTypeEphemeral, "ephemeral"},
		{CacheControlTypePersistent, "persistent"},
	}

	for _, test := range tests {
		result := test.cacheType.String()
		if result != test.expected {
			t.Errorf("String() = %q, expected %q", result, test.expected)
		}
	}
}

func TestUsage_Copy(t *testing.T) {
	original := &Usage{
		CacheCreationInputTokens: 100,
		CacheReadInputTokens:     200,
		InputTokens:              300,
		OutputTokens:             400,
	}

	copy := original.Copy()

	if copy == original {
		t.Error("Copy() should return a different instance")
	}

	if *copy != *original {
		t.Error("Copy() should have the same values as original")
	}

	// Modify original to ensure copy is independent
	original.InputTokens = 999
	if copy.InputTokens == 999 {
		t.Error("Copy should be independent of original")
	}
}

func TestUsage_Add(t *testing.T) {
	usage1 := &Usage{
		CacheCreationInputTokens: 100,
		CacheReadInputTokens:     200,
		InputTokens:              300,
		OutputTokens:             400,
	}

	usage2 := &Usage{
		CacheCreationInputTokens: 50,
		CacheReadInputTokens:     75,
		InputTokens:              125,
		OutputTokens:             150,
	}

	usage1.Add(usage2)

	expected := &Usage{
		CacheCreationInputTokens: 150,
		CacheReadInputTokens:     275,
		InputTokens:              425,
		OutputTokens:             550,
	}

	if *usage1 != *expected {
		t.Errorf("Add() result = %+v, expected %+v", usage1, expected)
	}
}

func TestUsage_AddNilUsage(t *testing.T) {
	usage := &Usage{
		InputTokens:  100,
		OutputTokens: 200,
	}

	// This should panic with nil pointer dereference - that's the current behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when adding nil usage")
		}
	}()

	usage.Add(nil)
}

func TestWithAPIKey(t *testing.T) {
	client := &Client{}
	opt := WithAPIKey("test-api-key")
	opt(client)

	if client.apiKey != "test-api-key" {
		t.Errorf("expected apiKey to be 'test-api-key', got %q", client.apiKey)
	}
}

func TestWithEndpoint(t *testing.T) {
	client := &Client{}
	opt := WithEndpoint("https://custom-endpoint.com")
	opt(client)

	if client.endpoint != "https://custom-endpoint.com" {
		t.Errorf("expected endpoint to be 'https://custom-endpoint.com', got %q", client.endpoint)
	}
}

func TestWithClient(t *testing.T) {
	client := &Client{}
	httpClient := &http.Client{}
	opt := WithClient(httpClient)
	opt(client)

	if client.client != httpClient {
		t.Error("expected client to be set")
	}
}

func TestWithMaxTokens(t *testing.T) {
	client := &Client{}
	opt := WithMaxTokens(1000)
	opt(client)

	if client.maxTokens != 1000 {
		t.Errorf("expected maxTokens to be 1000, got %d", client.maxTokens)
	}
}

func TestWithModel(t *testing.T) {
	client := &Client{}
	opt := WithModel("claude-3-sonnet")
	opt(client)

	if client.model != "claude-3-sonnet" {
		t.Errorf("expected model to be 'claude-3-sonnet', got %q", client.model)
	}
}

func TestWithMaxRetries(t *testing.T) {
	client := &Client{}
	opt := WithMaxRetries(5)
	opt(client)

	if client.maxRetries != 5 {
		t.Errorf("expected maxRetries to be 5, got %d", client.maxRetries)
	}
}

func TestWithBaseWait(t *testing.T) {
	client := &Client{}
	baseWait := 5 * time.Second
	opt := WithBaseWait(baseWait)
	opt(client)

	if client.retryBaseWait != baseWait {
		t.Errorf("expected retryBaseWait to be %v, got %v", baseWait, client.retryBaseWait)
	}
}

func TestWithVersion(t *testing.T) {
	client := &Client{}
	opt := WithVersion("2023-06-01")
	opt(client)

	if client.version != "2023-06-01" {
		t.Errorf("expected version to be '2023-06-01', got %q", client.version)
	}
}

func TestClientError_Error(t *testing.T) {
	err := &ClientError{
		statusCode: 400,
		body:       "Bad Request",
	}

	expected := "provider api error (status 400): Bad Request"
	if err.Error() != expected {
		t.Errorf("Error() = %q, expected %q", err.Error(), expected)
	}
}

func TestClientError_StatusCode(t *testing.T) {
	err := &ClientError{statusCode: 404}
	if err.StatusCode() != 404 {
		t.Errorf("StatusCode() = %d, expected 404", err.StatusCode())
	}
}

func TestClientError_IsRecoverable(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{429, true},  // Too Many Requests
		{500, true},  // Internal Server Error
		{502, false}, // Bad Gateway (not retryable)
		{503, true},  // Service Unavailable
		{504, true},  // Gateway Timeout
		{520, true},  // Cloudflare
		{400, false}, // Bad Request
		{401, false}, // Unauthorized
		{403, false}, // Forbidden
		{404, false}, // Not Found
	}

	for _, test := range tests {
		err := &ClientError{statusCode: test.statusCode}
		result := err.IsRecoverable()
		if result != test.expected {
			t.Errorf("IsRecoverable() for status %d = %v, expected %v", test.statusCode, result, test.expected)
		}
	}
}

func TestNewError(t *testing.T) {
	err := NewError(500, "Internal Server Error")

	if err.statusCode != 500 {
		t.Errorf("expected statusCode to be 500, got %d", err.statusCode)
	}
	if err.body != "Internal Server Error" {
		t.Errorf("expected body to be 'Internal Server Error', got %q", err.body)
	}
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{429, true},  // Too Many Requests
		{500, true},  // Internal Server Error
		{502, false}, // Bad Gateway (not retryable)
		{503, true},  // Service Unavailable
		{504, true},  // Gateway Timeout
		{520, true},  // Cloudflare
		{400, false}, // Bad Request
		{401, false}, // Unauthorized
		{403, false}, // Forbidden
		{404, false}, // Not Found
		{200, false}, // OK
	}

	for _, test := range tests {
		result := ShouldRetry(test.statusCode)
		if result != test.expected {
			t.Errorf("ShouldRetry(%d) = %v, expected %v", test.statusCode, result, test.expected)
		}
	}
}

func TestWithTools(t *testing.T) {
	client := &Client{}

	// Create mock tools
	tool1 := &ToolDefinition{}
	tool1.WithName("tool1")
	tool2 := &ToolDefinition{}
	tool2.WithName("tool2")

	opt := WithTools(tool1, tool2)
	opt(client)

	if len(client.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(client.Tools))
	}
	if client.Tools[0] != tool1 {
		t.Error("first tool should match")
	}
	if client.Tools[1] != tool2 {
		t.Error("second tool should match")
	}
}

func TestWithToolChoice(t *testing.T) {
	client := &Client{}
	choice := &ToolChoice{
		Type: ToolChoiceTypeAuto,
	}

	opt := WithToolChoice(choice)
	opt(client)

	if client.ToolChoice != choice {
		t.Error("expected ToolChoice to be set")
	}
	if client.ToolChoice.Type != ToolChoiceTypeAuto {
		t.Errorf("expected ToolChoice type %v, got %v", ToolChoiceTypeAuto, client.ToolChoice.Type)
	}
}

func TestWithSystemPrompt(t *testing.T) {
	client := &Client{}
	prompt := "You are a helpful assistant."

	opt := WithSystemPrompt(prompt)
	opt(client)

	if client.SystemPrompt != prompt {
		t.Errorf("expected SystemPrompt to be %q, got %q", prompt, client.SystemPrompt)
	}
}

func TestClient_Apply(t *testing.T) {
	client := &Client{}

	// Create a web search tool
	webSearchTool := NewWebSearchTool(WebSearchToolOptions{
		MaxUses: 3,
	})

	opts := []Option{
		WithAPIKey("test-key"),
		WithModel("test-model"),
		WithMaxTokens(2000),
		WithTools(webSearchTool),
		WithSystemPrompt("You are helpful."),
	}

	client.Apply(opts...)

	if client.apiKey != "test-key" {
		t.Errorf("expected apiKey to be 'test-key', got %q", client.apiKey)
	}
	if client.model != "test-model" {
		t.Errorf("expected model to be 'test-model', got %q", client.model)
	}
	if client.maxTokens != 2000 {
		t.Errorf("expected maxTokens to be 2000, got %d", client.maxTokens)
	}
	if len(client.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(client.Tools))
	}
	if client.Tools[0] != webSearchTool {
		t.Error("expected web search tool to be set")
	}
	if client.SystemPrompt != "You are helpful." {
		t.Errorf("expected SystemPrompt to be 'You are helpful.', got %q", client.SystemPrompt)
	}
}
