package anthropic

import (
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	client := New()

	if client == nil {
		t.Error("New() should not return nil")
	}

	// Check default values
	if client.model != DefaultModel {
		t.Errorf("Expected default model %q, got %q", DefaultModel, client.model)
	}

	if client.maxTokens != DefaultMaxTokens {
		t.Errorf("Expected default max tokens %d, got %d", DefaultMaxTokens, client.maxTokens)
	}

	if client.endpoint != DefaultEndpoint {
		t.Errorf("Expected default endpoint %q, got %q", DefaultEndpoint, client.endpoint)
	}

	if client.version != DefaultVersion {
		t.Errorf("Expected default version %q, got %q", DefaultVersion, client.version)
	}
}

func TestNew_WithOptions(t *testing.T) {
	customAPIKey := "test-api-key"
	customModel := "claude-3-haiku"
	customMaxTokens := 2000
	customEndpoint := "https://custom.example.com"
	customVersion := "2024-01-01"
	customMaxRetries := 5
	customBaseWait := 5 * time.Second
	customClient := &http.Client{Timeout: 30 * time.Second}

	client := New(
		WithAPIKey(customAPIKey),
		WithModel(customModel),
		WithMaxTokens(customMaxTokens),
		WithEndpoint(customEndpoint),
		WithVersion(customVersion),
		WithMaxRetries(customMaxRetries),
		WithBaseWait(customBaseWait),
		WithClient(customClient),
	)

	if client.apiKey != customAPIKey {
		t.Errorf("Expected API key %q, got %q", customAPIKey, client.apiKey)
	}

	if client.model != customModel {
		t.Errorf("Expected model %q, got %q", customModel, client.model)
	}

	if client.maxTokens != customMaxTokens {
		t.Errorf("Expected max tokens %d, got %d", customMaxTokens, client.maxTokens)
	}

	if client.endpoint != customEndpoint {
		t.Errorf("Expected endpoint %q, got %q", customEndpoint, client.endpoint)
	}

	if client.version != customVersion {
		t.Errorf("Expected version %q, got %q", customVersion, client.version)
	}

	if client.maxRetries != customMaxRetries {
		t.Errorf("Expected max retries %d, got %d", customMaxRetries, client.maxRetries)
	}

	if client.retryBaseWait != customBaseWait {
		t.Errorf("Expected base wait %v, got %v", customBaseWait, client.retryBaseWait)
	}

	if client.client != customClient {
		t.Error("Expected custom HTTP client to be set")
	}
}

func TestClient_Name(t *testing.T) {
	client := New()

	name := client.Name()
	if name != ProviderName {
		t.Errorf("Expected name %q, got %q", ProviderName, name)
	}
}

func TestConvertMessages_EmptyMessages(t *testing.T) {
	var messages []*Message
	result, err := convertMessages(messages)

	if err == nil {
		t.Error("Expected error for empty messages")
	}

	if result != nil {
		t.Error("Expected nil result for empty messages")
	}
}

func TestConvertMessages_EmptyContent(t *testing.T) {
	messages := []*Message{
		{Role: User, Content: []Content{}}, // Empty content
	}

	result, err := convertMessages(messages)

	if err == nil {
		t.Error("Expected error for message with empty content")
	}

	if result != nil {
		t.Error("Expected nil result for invalid messages")
	}
}

func TestConvertMessages_SystemMessage(t *testing.T) {
	messages := []*Message{
		{Role: System, Content: []Content{&TextContent{Text: "System prompt"}}},
		{Role: User, Content: []Content{&TextContent{Text: "User message"}}},
	}

	result, err := convertMessages(messages)
	if err != nil {
		t.Errorf("convertMessages failed: %v", err)
	}

	// All messages should be preserved (system messages are not filtered)
	if len(result) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(result))
	}

	if result[0].Role != System {
		t.Errorf("Expected System role, got %v", result[0].Role)
	}
	if result[1].Role != User {
		t.Errorf("Expected User role, got %v", result[1].Role)
	}
}

func TestConvertMessages_MessageReordering(t *testing.T) {
	// Create a message with mixed content that should be reordered
	messages := []*Message{
		{
			Role: Assistant,
			Content: []Content{
				&ToolUseContent{ID: "tool1", Name: "test_tool"},
				&TextContent{Text: "Response text"},
				&ToolUseContent{ID: "tool2", Name: "test_tool2"},
			},
		},
	}

	result, err := convertMessages(messages)
	if err != nil {
		t.Errorf("convertMessages failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 message, got %d", len(result))
	}

	// Content should be reordered: text first, then tool use
	msg := result[0]
	if len(msg.Content) != 3 {
		t.Errorf("Expected 3 content blocks, got %d", len(msg.Content))
	}

	// First content should be text
	if _, ok := msg.Content[0].(*TextContent); !ok {
		t.Error("First content block should be TextContent")
	}

	// Second and third should be tool use
	if _, ok := msg.Content[1].(*ToolUseContent); !ok {
		t.Error("Second content block should be ToolUseContent")
	}
	if _, ok := msg.Content[2].(*ToolUseContent); !ok {
		t.Error("Third content block should be ToolUseContent")
	}
}

func TestNew_WithTools(t *testing.T) {
	// Create a web search tool
	webSearchTool := NewWebSearchTool(WebSearchToolOptions{
		MaxUses: 3,
	})

	// Create another tool
	testTool := NewToolDefinition().
		WithName("test_tool").
		WithDescription("A test tool")

	client := New(
		WithAPIKey("test-key"),
		WithTools(webSearchTool, testTool),
		WithSystemPrompt("You are helpful."),
	)

	if len(client.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(client.Tools))
	}

	if client.Tools[0] != webSearchTool {
		t.Error("First tool should be web search tool")
	}

	if client.Tools[1] != testTool {
		t.Error("Second tool should be test tool")
	}

	if client.SystemPrompt != "You are helpful." {
		t.Errorf("Expected SystemPrompt 'You are helpful.', got %q", client.SystemPrompt)
	}
}
