//go:build integration
// +build integration

package anthropic

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// TestWebSearchStreamingIntegration tests the web search tool with streaming using the real API
// Run with: go test -tags=integration -run TestWebSearchStreamingIntegration
// Make sure to set ANTHROPIC_API_KEY environment variable
func TestWebSearchStreamingIntegration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	// Create client with API key
	client := New(
		WithAPIKey(apiKey),
		WithModel("claude-3-5-haiku-latest"),
	)

	// Test the exact web search tool format from the API examples
	messages := Messages{
		NewUserTextMessage("What is the current weather in New York City today? Answer briefly."),
	}

	// Configure client with web search tool
	webSearchTool := NewWebSearchTool(WebSearchToolOptions{
		MaxUses: 5,
	})

	// Apply the tool to the client using the proper option
	client.Apply(WithTools(webSearchTool))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test streaming with web search
	stream, err := client.Stream(ctx, messages)
	if err != nil {
		t.Fatalf("Failed to start streaming: %v", err)
	}
	defer stream.Close()

	// Track what we see in the stream
	var (
		messageStartSeen    = false
		textContentSeen     = false
		serverToolUseSeen   = false
		webSearchResultSeen = false

		messageStopSeen = false
		collectedText   strings.Builder
		toolUseID       string
		searchQuery     string
	)

	for stream.Next() {
		event := stream.Event()
		t.Logf("Event: %s", event.Type)

		switch event.Type {
		case EventTypePing:
			// Just log pings, nothing to validate
			// t.Logf("Ping received")

		case EventTypeMessageStart:
			messageStartSeen = true
			if event.Message == nil {
				t.Error("message_start event should have message")
			}
			if event.Message.Role != Assistant {
				t.Errorf("Expected assistant role, got %v", event.Message.Role)
			}
			t.Logf("Message started: ID=%s, Model=%s", event.Message.ID, event.Message.Model)

		case EventTypeContentBlockStart:
			if event.ContentBlock == nil {
				t.Error("content_block_start should have content_block")
				continue
			}

			switch event.ContentBlock.Type {
			case ContentTypeText:
				textContentSeen = true
				t.Logf("Text content block started at index %d", *event.Index)

			case ContentTypeServerToolUse:
				serverToolUseSeen = true
				toolUseID = event.ContentBlock.ID
				t.Logf("Server tool use started: ID=%s, Name=%s", event.ContentBlock.ID, event.ContentBlock.Name)
				if event.ContentBlock.Name != "web_search" {
					t.Errorf("Expected tool name 'web_search', got %s", event.ContentBlock.Name)
				}

			case ContentTypeWebSearchToolResult:
				webSearchResultSeen = true
				t.Logf("Web search result block started at index %d", *event.Index)
			}

		case EventTypeContentBlockDelta:
			if event.Delta == nil {
				t.Error("content_block_delta should have delta")
				continue
			}

			switch event.Delta.Type {
			case EventDeltaTypeText:
				collectedText.WriteString(event.Delta.Text)
				t.Logf("Text delta: %q", event.Delta.Text)

			case EventDeltaTypeInputJSON:
				searchQuery += event.Delta.PartialJSON
				t.Logf("Input JSON delta: %q", event.Delta.PartialJSON)
			}

		case EventTypeContentBlockStop:
			t.Logf("Content block stopped at index %d", *event.Index)

		case EventTypeMessageDelta:
			if event.Delta != nil {
				t.Logf("Message delta: stop_reason=%s", event.Delta.StopReason)
			}

		case EventTypeMessageStop:
			messageStopSeen = true
			t.Logf("Message stopped")
		}
	}

	// Check for any streaming errors
	if err := stream.Err(); err != nil {
		t.Fatalf("Streaming error: %v", err)
	}

	// Validate we saw the expected events
	if !messageStartSeen {
		t.Error("Expected to see message_start event")
	}
	if !messageStopSeen {
		t.Error("Expected to see message_stop event")
	}
	if !textContentSeen {
		t.Error("Expected to see text content blocks")
	}

	// The model might or might not use web search depending on the query
	// But if it does, we should see the proper structure
	if serverToolUseSeen {
		t.Logf("✅ Server tool use detected with ID: %s", toolUseID)
		if toolUseID == "" {
			t.Error("Tool use should have an ID")
		}
		if searchQuery == "" {
			t.Error("Expected to collect search query from input_json_delta events")
		} else {
			t.Logf("Search query collected: %s", searchQuery)
		}
	}

	if webSearchResultSeen {
		t.Logf("✅ Web search results detected")
	}

	// Validate we got some text response
	finalText := collectedText.String()
	if finalText == "" {
		t.Error("Expected to collect some text from the response")
	} else {
		t.Logf("Collected text length: %d characters", len(finalText))
		t.Logf("Response: %s\n", finalText)
	}

	// Validate the response contains weather-related content
	if serverToolUseSeen && !strings.Contains(strings.ToLower(finalText), "weather") {
		t.Error("Expected weather-related content in response")
	}
}

// TestWebSearchToolDefinitionIntegration tests creating a request with web search tool
func TestWebSearchToolDefinitionIntegration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel("claude-3-5-haiku-latest"),
	)

	// Test that we can create a request with web search tool definition
	// This matches the API example format
	messages := Messages{
		NewUserTextMessage("How do I update a web app to TypeScript 5.5?"),
	}

	// Configure the client with web search tool using the proper constructor
	webSearchTool := NewWebSearchTool(WebSearchToolOptions{
		Type:    "web_search_20250305",
		MaxUses: 5,
	})

	// Apply the tool to the client using the proper option
	client.Apply(WithTools(webSearchTool))

	// Test streaming with the configured web search tool
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.Stream(ctx, messages)
	if err != nil {
		t.Fatalf("Failed to start streaming: %v", err)
	}
	defer stream.Close()

	// Consume a few events to validate the request was processed
	var eventCount int
	for stream.Next() && eventCount < 3 {
		event := stream.Event()
		t.Logf("Event %d: %s", eventCount, event.Type)
		eventCount++
	}

	if err := stream.Err(); err != nil {
		t.Fatalf("Streaming error: %v", err)
	}

	// Test the tool configuration separately
	toolConfig := webSearchTool.ToolConfiguration("anthropic")
	data, err := json.Marshal(toolConfig)
	if err != nil {
		t.Fatalf("Failed to marshal tool config: %v", err)
	}

	t.Logf("Request size: %d bytes", len(data))

	// Verify the request contains expected fields
	requestStr := string(data)
	if !strings.Contains(requestStr, "web_search_20250305") {
		t.Error("Request should contain web_search_20250305 tool type")
	}
	if !strings.Contains(requestStr, "max_uses") {
		t.Error("Request should contain max_uses field")
	}
	if !strings.Contains(requestStr, "stream") {
		t.Error("Request should contain stream field")
	}

	// Validate the tool configuration structure
	var toolResult map[string]interface{}
	err = json.Unmarshal(data, &toolResult)
	if err != nil {
		t.Fatalf("Failed to unmarshal tool config: %v", err)
	}

	if toolResult["type"] != "web_search_20250305" {
		t.Errorf("Expected tool type 'web_search_20250305', got %v", toolResult["type"])
	}
	if toolResult["name"] != "web_search" {
		t.Errorf("Expected tool name 'web_search', got %v", toolResult["name"])
	}
	if toolResult["max_uses"] != float64(5) {
		t.Errorf("Expected max_uses 5, got %v", toolResult["max_uses"])
	}

	t.Logf("✅ Web search tool configuration validated: %s", string(data))
	t.Logf("✅ Streaming request with web search tool successful")
}

// TestSimpleStreamingIntegration tests basic streaming without tools
func TestSimpleStreamingIntegration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	client := New(
		WithAPIKey(apiKey),
		WithModel("claude-3-5-haiku-latest"),
	)

	messages := Messages{
		NewUserTextMessage("Hello! Please respond with exactly: 'Hello from Claude'"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stream, err := client.Stream(ctx, messages)
	if err != nil {
		t.Fatalf("Failed to start streaming: %v", err)
	}
	defer stream.Close()

	var (
		messageStartSeen = false
		textContentSeen  = false
		messageStopSeen  = false
		collectedText    strings.Builder
	)

	for stream.Next() {
		event := stream.Event()

		switch event.Type {
		case EventTypeMessageStart:
			messageStartSeen = true

		case EventTypeContentBlockStart:
			if event.ContentBlock != nil && event.ContentBlock.Type == ContentTypeText {
				textContentSeen = true
			}

		case EventTypeContentBlockDelta:
			if event.Delta != nil && event.Delta.Type == EventDeltaTypeText {
				collectedText.WriteString(event.Delta.Text)
			}

		case EventTypeMessageStop:
			messageStopSeen = true
		}
	}

	if err := stream.Err(); err != nil {
		t.Fatalf("Streaming error: %v", err)
	}

	// Validate basic streaming worked
	if !messageStartSeen {
		t.Error("Expected to see message_start event")
	}
	if !textContentSeen {
		t.Error("Expected to see text content")
	}
	if !messageStopSeen {
		t.Error("Expected to see message_stop event")
	}

	finalText := collectedText.String()
	if finalText == "" {
		t.Error("Expected to collect some text from the response")
	}

	t.Logf("✅ Simple streaming test completed. Response: %q", finalText)
}
