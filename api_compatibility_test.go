package anthropic

import (
	"encoding/json"
	"testing"
)

// TestAPIRequestStructure verifies that our Request struct matches the expected API format
func TestAPIRequestStructure(t *testing.T) {
	// Test simple message request
	request := &Request{
		Model:     "claude-3-5-haiku-latest",
		MaxTokens: intPtr(1024),
		Messages: []*Message{
			{Role: User, Content: []Content{&TextContent{Text: "Hello, Claude"}}},
			{Role: Assistant, Content: []Content{&TextContent{Text: "Hello!"}}},
			{Role: User, Content: []Content{&TextContent{Text: "Can you describe LLMs to me?"}}},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Verify structure
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	if result["model"] != "claude-3-5-haiku-latest" {
		t.Errorf("Expected model 'claude-3-5-haiku-latest', got %v", result["model"])
	}
	if result["max_tokens"] != float64(1024) {
		t.Errorf("Expected max_tokens 1024, got %v", result["max_tokens"])
	}

	messages, ok := result["messages"].([]interface{})
	if !ok {
		t.Error("Expected messages to be an array")
	}
	if len(messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(messages))
	}
}

// TestAPIResponseStructure verifies that our Response struct matches the expected API format
func TestAPIResponseStructure(t *testing.T) {
	// Simulate an API response
	responseJSON := `{
		"id": "msg_018gCsTGsXkYJVqYPxTgDHBU",
		"type": "message",
		"role": "assistant",
		"content": [
			{
				"type": "text",
				"text": "Sure, I'd be happy to provide..."
			}
		],
		"stop_reason": "end_turn",
		"stop_sequence": null,
		"usage": {
			"input_tokens": 30,
			"output_tokens": 309
		}
	}`

	var response Response
	err := json.Unmarshal([]byte(responseJSON), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.ID != "msg_018gCsTGsXkYJVqYPxTgDHBU" {
		t.Errorf("Expected ID 'msg_018gCsTGsXkYJVqYPxTgDHBU', got %q", response.ID)
	}
	if response.Role != Assistant {
		t.Errorf("Expected role %v, got %v", Assistant, response.Role)
	}
	if response.StopReason != "end_turn" {
		t.Errorf("Expected stop_reason 'end_turn', got %q", response.StopReason)
	}
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(response.Content))
	}

	textContent, ok := response.Content[0].(*TextContent)
	if !ok {
		t.Error("Expected content to be TextContent")
	}
	if textContent.Text != "Sure, I'd be happy to provide..." {
		t.Errorf("Expected text 'Sure, I'd be happy to provide...', got %q", textContent.Text)
	}

	if response.Usage.InputTokens != 30 {
		t.Errorf("Expected input_tokens 30, got %d", response.Usage.InputTokens)
	}
	if response.Usage.OutputTokens != 309 {
		t.Errorf("Expected output_tokens 309, got %d", response.Usage.OutputTokens)
	}
}

// TestVisionAPIStructure verifies that image content works with the API format
func TestVisionAPIStructure(t *testing.T) {
	// Test base64 image content
	base64Content := &ImageContent{
		Source: &ContentSource{
			Type:      ContentSourceTypeBase64,
			MediaType: "image/jpeg",
			Data:      "base64data",
		},
	}

	message := &Message{
		Role: User,
		Content: []Content{
			base64Content,
			&TextContent{Text: "What is in the above image?"},
		},
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	content, ok := result["content"].([]interface{})
	if !ok {
		t.Error("Expected content to be an array")
	}
	if len(content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(content))
	}

	// Check image content structure
	imageContent, ok := content[0].(map[string]interface{})
	if !ok {
		t.Error("Expected first content to be an object")
	}
	if imageContent["type"] != "image" {
		t.Errorf("Expected type 'image', got %v", imageContent["type"])
	}

	source, ok := imageContent["source"].(map[string]interface{})
	if !ok {
		t.Error("Expected source to be an object")
	}
	if source["type"] != "base64" {
		t.Errorf("Expected source type 'base64', got %v", source["type"])
	}
	if source["media_type"] != "image/jpeg" {
		t.Errorf("Expected media_type 'image/jpeg', got %v", source["media_type"])
	}
	if source["data"] != "base64data" {
		t.Errorf("Expected data 'base64data', got %v", source["data"])
	}

	// Test URL image content
	urlContent := &ImageContent{
		Source: &ContentSource{
			Type: ContentSourceTypeURL,
			URL:  "https://example.com/image.jpg",
		},
	}

	data, err = json.Marshal(urlContent)
	if err != nil {
		t.Fatalf("Failed to marshal URL content: %v", err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal URL result: %v", err)
	}

	source = result["source"].(map[string]interface{})
	if source["type"] != "url" {
		t.Errorf("Expected source type 'url', got %v", source["type"])
	}
	if source["url"] != "https://example.com/image.jpg" {
		t.Errorf("Expected url 'https://example.com/image.jpg', got %v", source["url"])
	}
}

// TestToolUseAPIStructure verifies that tool use structures match the API format
func TestToolUseAPIStructure(t *testing.T) {
	// Test tool definition in request
	toolDef := map[string]any{
		"name":        "get_weather",
		"description": "Get the current weather in a given location",
		"input_schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "The city and state, e.g. San Francisco, CA",
				},
			},
			"required": []string{"location"},
		},
	}

	request := &Request{
		Model:     "claude-3-5-haiku-latest",
		MaxTokens: intPtr(1024),
		Tools:     []map[string]any{toolDef},
		Messages: []*Message{
			{Role: User, Content: []Content{&TextContent{Text: "What is the weather like in San Francisco?"}}},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Error("Expected tools to be an array")
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	tool := tools[0].(map[string]interface{})
	if tool["name"] != "get_weather" {
		t.Errorf("Expected tool name 'get_weather', got %v", tool["name"])
	}

	// Test tool use content
	toolUse := &ToolUseContent{
		ID:    "toolu_01A09q90qw90lq917835lq9",
		Name:  "get_weather",
		Input: json.RawMessage(`{"location": "San Francisco, CA", "unit": "celsius"}`),
	}

	data, err = json.Marshal(toolUse)
	if err != nil {
		t.Fatalf("Failed to marshal tool use: %v", err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal tool use result: %v", err)
	}

	if result["type"] != "tool_use" {
		t.Errorf("Expected type 'tool_use', got %v", result["type"])
	}
	if result["id"] != "toolu_01A09q90qw90lq917835lq9" {
		t.Errorf("Expected id 'toolu_01A09q90qw90lq917835lq9', got %v", result["id"])
	}
	if result["name"] != "get_weather" {
		t.Errorf("Expected name 'get_weather', got %v", result["name"])
	}

	// Test tool result content
	toolResult := &ToolResultContent{
		ToolUseID: "toolu_01A09q90qw90lq917835lq9",
		Content:   "15 degrees",
	}

	data, err = json.Marshal(toolResult)
	if err != nil {
		t.Fatalf("Failed to marshal tool result: %v", err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal tool result: %v", err)
	}

	if result["type"] != "tool_result" {
		t.Errorf("Expected type 'tool_result', got %v", result["type"])
	}
	if result["tool_use_id"] != "toolu_01A09q90qw90lq917835lq9" {
		t.Errorf("Expected tool_use_id 'toolu_01A09q90qw90lq917835lq9', got %v", result["tool_use_id"])
	}
	if result["content"] != "15 degrees" {
		t.Errorf("Expected content '15 degrees', got %v", result["content"])
	}
}

// TestStreamingAPIStructure verifies that streaming event structures match the API format
func TestStreamingAPIStructure(t *testing.T) {
	// Test message_start event
	messageStartJSON := `{
		"type": "message_start",
		"message": {
			"id": "msg_1nZdL29xx5MUA1yADyHTEsnR8uuvGzszyY",
			"type": "message",
			"role": "assistant",
			"content": [],
			"model": "claude-3-5-haiku-latest",
			"stop_reason": null,
			"stop_sequence": null,
			"usage": {
				"input_tokens": 25,
				"output_tokens": 1
			}
		}
	}`

	var event Event
	err := json.Unmarshal([]byte(messageStartJSON), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal message_start event: %v", err)
	}

	if event.Type != EventTypeMessageStart {
		t.Errorf("Expected type %v, got %v", EventTypeMessageStart, event.Type)
	}
	if event.Message == nil {
		t.Error("Expected message to be present")
	}
	if event.Message.ID != "msg_1nZdL29xx5MUA1yADyHTEsnR8uuvGzszyY" {
		t.Errorf("Expected message ID 'msg_1nZdL29xx5MUA1yADyHTEsnR8uuvGzszyY', got %q", event.Message.ID)
	}

	// Test content_block_start event
	contentBlockStartJSON := `{
		"type": "content_block_start",
		"index": 0,
		"content_block": {
			"type": "text",
			"text": ""
		}
	}`

	err = json.Unmarshal([]byte(contentBlockStartJSON), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal content_block_start event: %v", err)
	}

	if event.Type != EventTypeContentBlockStart {
		t.Errorf("Expected type %v, got %v", EventTypeContentBlockStart, event.Type)
	}
	if event.Index == nil || *event.Index != 0 {
		t.Errorf("Expected index 0, got %v", event.Index)
	}
	if event.ContentBlock == nil {
		t.Error("Expected content_block to be present")
	}
	if event.ContentBlock.Type != ContentTypeText {
		t.Errorf("Expected content_block type %v, got %v", ContentTypeText, event.ContentBlock.Type)
	}

	// Test content_block_delta event
	contentBlockDeltaJSON := `{
		"type": "content_block_delta",
		"index": 0,
		"delta": {
			"type": "text_delta",
			"text": "Hello"
		}
	}`

	err = json.Unmarshal([]byte(contentBlockDeltaJSON), &event)
	if err != nil {
		t.Fatalf("Failed to unmarshal content_block_delta event: %v", err)
	}

	if event.Type != EventTypeContentBlockDelta {
		t.Errorf("Expected type %v, got %v", EventTypeContentBlockDelta, event.Type)
	}
	if event.Delta == nil {
		t.Error("Expected delta to be present")
	}
	if event.Delta.Type != EventDeltaTypeText {
		t.Errorf("Expected delta type %v, got %v", EventDeltaTypeText, event.Delta.Type)
	}
	if event.Delta.Text != "Hello" {
		t.Errorf("Expected delta text 'Hello', got %q", event.Delta.Text)
	}
}

// TestWebSearchToolStructure verifies that web search tool structures match the API format
func TestWebSearchToolStructure(t *testing.T) {
	// Test web search tool definition
	webSearchTool := map[string]any{
		"type":     "web_search_20250305",
		"name":     "web_search",
		"max_uses": 5,
	}

	request := &Request{
		Model:     "claude-3-5-haiku-latest",
		MaxTokens: intPtr(1024),
		Tools:     []map[string]any{webSearchTool},
		Messages: []*Message{
			{Role: User, Content: []Content{&TextContent{Text: "How do I update a web app to TypeScript 5.5?"}}},
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Error("Expected tools to be an array")
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	tool := tools[0].(map[string]interface{})
	if tool["type"] != "web_search_20250305" {
		t.Errorf("Expected tool type 'web_search_20250305', got %v", tool["type"])
	}
	if tool["name"] != "web_search" {
		t.Errorf("Expected tool name 'web_search', got %v", tool["name"])
	}
	if tool["max_uses"] != float64(5) {
		t.Errorf("Expected max_uses 5, got %v", tool["max_uses"])
	}

	// Test server tool use content (streaming response)
	serverToolUse := &ServerToolUseContent{
		ID:   "srvtoolu_014hJH82Qum7Td6UV8gDXThB",
		Name: "web_search",
		Input: map[string]any{
			"query": "weather NYC today",
		},
	}

	data, err = json.Marshal(serverToolUse)
	if err != nil {
		t.Fatalf("Failed to marshal server tool use: %v", err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal server tool use result: %v", err)
	}

	if result["type"] != "server_tool_use" {
		t.Errorf("Expected type 'server_tool_use', got %v", result["type"])
	}
	if result["id"] != "srvtoolu_014hJH82Qum7Td6UV8gDXThB" {
		t.Errorf("Expected id 'srvtoolu_014hJH82Qum7Td6UV8gDXThB', got %v", result["id"])
	}
	if result["name"] != "web_search" {
		t.Errorf("Expected name 'web_search', got %v", result["name"])
	}

	input := result["input"].(map[string]interface{})
	if input["query"] != "weather NYC today" {
		t.Errorf("Expected query 'weather NYC today', got %v", input["query"])
	}

	// Test web search tool result content
	webSearchResult := &WebSearchToolResultContent{
		ToolUseID: "srvtoolu_014hJH82Qum7Td6UV8gDXThB",
		Content: []*WebSearchResult{
			{
				Title:            "Weather in New York City",
				URL:              "https://weather.example.com/nyc",
				EncryptedContent: "encrypted_content_data",
			},
		},
	}

	data, err = json.Marshal(webSearchResult)
	if err != nil {
		t.Fatalf("Failed to marshal web search result: %v", err)
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal web search result: %v", err)
	}

	if result["type"] != "web_search_tool_result" {
		t.Errorf("Expected type 'web_search_tool_result', got %v", result["type"])
	}
	if result["tool_use_id"] != "srvtoolu_014hJH82Qum7Td6UV8gDXThB" {
		t.Errorf("Expected tool_use_id 'srvtoolu_014hJH82Qum7Td6UV8gDXThB', got %v", result["tool_use_id"])
	}

	content := result["content"].([]interface{})
	if len(content) != 1 {
		t.Errorf("Expected 1 search result, got %d", len(content))
	}

	searchResult := content[0].(map[string]interface{})
	if searchResult["title"] != "Weather in New York City" {
		t.Errorf("Expected title 'Weather in New York City', got %v", searchResult["title"])
	}
	if searchResult["url"] != "https://weather.example.com/nyc" {
		t.Errorf("Expected url 'https://weather.example.com/nyc', got %v", searchResult["url"])
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
