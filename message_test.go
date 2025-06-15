package anthropic

import (
	"testing"
)

func TestRole_String(t *testing.T) {
	tests := []struct {
		role     Role
		expected string
	}{
		{User, "user"},
		{Assistant, "assistant"},
		{System, "system"},
	}

	for _, test := range tests {
		result := test.role.String()
		if result != test.expected {
			t.Errorf("String() = %q, expected %q", result, test.expected)
		}
	}
}

func TestMessage_LastText(t *testing.T) {
	tests := []struct {
		name     string
		message  *Message
		expected string
	}{
		{
			name: "single text content",
			message: &Message{
				Content: []Content{
					&TextContent{Text: "Hello world"},
				},
			},
			expected: "Hello world",
		},
		{
			name: "multiple text content - returns last",
			message: &Message{
				Content: []Content{
					&TextContent{Text: "First"},
					&TextContent{Text: "Last"},
				},
			},
			expected: "Last",
		},
		{
			name: "mixed content - returns last text",
			message: &Message{
				Content: []Content{
					&TextContent{Text: "Text content"},
					&ToolUseContent{ID: "tool1", Name: "test"},
					&TextContent{Text: "Final text"},
				},
			},
			expected: "Final text",
		},
		{
			name: "no text content",
			message: &Message{
				Content: []Content{
					&ToolUseContent{ID: "tool1", Name: "test"},
				},
			},
			expected: "",
		},
		{
			name: "empty content",
			message: &Message{
				Content: []Content{},
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.message.LastText()
			if result != test.expected {
				t.Errorf("LastText() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestMessage_Text(t *testing.T) {
	tests := []struct {
		name     string
		message  *Message
		expected string
	}{
		{
			name: "single text content",
			message: &Message{
				Content: []Content{
					&TextContent{Text: "Hello world"},
				},
			},
			expected: "Hello world",
		},
		{
			name: "multiple text content - concatenated with newlines",
			message: &Message{
				Content: []Content{
					&TextContent{Text: "First line"},
					&TextContent{Text: "Second line"},
				},
			},
			expected: "First line\n\nSecond line",
		},
		{
			name: "mixed content - only text concatenated",
			message: &Message{
				Content: []Content{
					&TextContent{Text: "Start"},
					&ToolUseContent{ID: "tool1", Name: "test"},
					&TextContent{Text: "End"},
				},
			},
			expected: "Start\n\nEnd",
		},
		{
			name: "no text content",
			message: &Message{
				Content: []Content{
					&ToolUseContent{ID: "tool1", Name: "test"},
				},
			},
			expected: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.message.Text()
			if result != test.expected {
				t.Errorf("Text() = %q, expected %q", result, test.expected)
			}
		})
	}
}

func TestMessage_WithText(t *testing.T) {
	msg := &Message{Role: User}

	result := msg.WithText("First", "Second", "Third")

	// Should return the same message instance
	if result != msg {
		t.Error("WithText should return the same message instance")
	}

	// Should have 3 text content blocks
	if len(msg.Content) != 3 {
		t.Errorf("Expected 3 content blocks, got %d", len(msg.Content))
	}

	expectedTexts := []string{"First", "Second", "Third"}
	for i, expected := range expectedTexts {
		textContent, ok := msg.Content[i].(*TextContent)
		if !ok {
			t.Errorf("Content block %d is not TextContent", i)
			continue
		}
		if textContent.Text != expected {
			t.Errorf("Content block %d: expected %q, got %q", i, expected, textContent.Text)
		}
	}
}

func TestMessage_WithContent(t *testing.T) {
	msg := &Message{Role: User}
	textContent := &TextContent{Text: "Hello"}
	toolContent := &ToolUseContent{ID: "tool1", Name: "test"}

	result := msg.WithContent(textContent, toolContent)

	// Should return the same message instance
	if result != msg {
		t.Error("WithContent should return the same message instance")
	}

	// Should have 2 content blocks
	if len(msg.Content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	if msg.Content[0] != textContent {
		t.Error("First content block should be the text content")
	}
	if msg.Content[1] != toolContent {
		t.Error("Second content block should be the tool content")
	}
}

func TestMessage_ImageContent(t *testing.T) {
	// Test with image content
	imageContent := &ImageContent{
		Source: &ContentSource{
			Type:      ContentSourceTypeBase64,
			MediaType: "image/png",
			Data:      "base64data",
		},
	}
	msgWithImage := &Message{
		Content: []Content{
			&TextContent{Text: "Here's an image:"},
			imageContent,
		},
	}

	result, ok := msgWithImage.ImageContent()
	if !ok {
		t.Error("Expected to find image content")
	}
	if result != imageContent {
		t.Error("Returned image content should match the original")
	}

	// Test without image content
	msgWithoutImage := &Message{
		Content: []Content{
			&TextContent{Text: "No image here"},
		},
	}

	result, ok = msgWithoutImage.ImageContent()
	if ok {
		t.Error("Should not find image content")
	}
	if result != nil {
		t.Error("Result should be nil when no image content found")
	}
}

func TestMessage_ThinkingContent(t *testing.T) {
	// Test with thinking content
	thinkingContent := &ThinkingContent{Thinking: "I'm thinking..."}
	msgWithThinking := &Message{
		Content: []Content{
			&TextContent{Text: "Let me think:"},
			thinkingContent,
		},
	}

	result, ok := msgWithThinking.ThinkingContent()
	if !ok {
		t.Error("Expected to find thinking content")
	}
	if result != thinkingContent {
		t.Error("Returned thinking content should match the original")
	}

	// Test without thinking content
	msgWithoutThinking := &Message{
		Content: []Content{
			&TextContent{Text: "No thinking here"},
		},
	}

	result, ok = msgWithoutThinking.ThinkingContent()
	if ok {
		t.Error("Should not find thinking content")
	}
	if result != nil {
		t.Error("Result should be nil when no thinking content found")
	}
}

func TestMessage_DecodeInto(t *testing.T) {
	jsonText := `{"message": "Hello world", "status": "success"}`
	textContent := &TextContent{Text: jsonText}
	msg := &Message{
		Role:    User,
		Content: []Content{textContent},
	}

	var result struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}

	err := msg.DecodeInto(&result)
	if err != nil {
		t.Errorf("DecodeInto failed: %v", err)
	}

	if result.Message != "Hello world" {
		t.Errorf("Expected message 'Hello world', got %q", result.Message)
	}

	if result.Status != "success" {
		t.Errorf("Expected status 'success', got %q", result.Status)
	}
}

func TestMessage_DecodeInto_NoTextContent(t *testing.T) {
	msg := &Message{
		Role: User,
		Content: []Content{
			&ToolUseContent{ID: "tool1", Name: "test"},
		},
	}

	var result struct{}
	err := msg.DecodeInto(&result)
	if err == nil {
		t.Error("Expected error when no text content found")
	}

	expectedError := "no text content found"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestMessage_DecodeInto_InvalidTarget(t *testing.T) {
	msg := &Message{
		Role: User,
		Content: []Content{
			&TextContent{Text: "not json"},
		},
	}

	// Try to decode into a non-pointer
	var result struct{}
	err := msg.DecodeInto(result)
	if err == nil {
		t.Error("Expected error when decoding into non-pointer")
	}
}
