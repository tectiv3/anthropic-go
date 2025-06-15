package anthropic

import (
	"testing"
)

func TestReorderMessageContent_NoMessages(t *testing.T) {
	var messages []*Message
	reorderMessageContent(messages)
	// Should not panic with empty slice
}

func TestReorderMessageContent_UserMessage(t *testing.T) {
	messages := []*Message{
		{
			Role: User,
			Content: []Content{
				&TextContent{Text: "Hello"},
				&ToolUseContent{ID: "tool1", Name: "test_tool"},
			},
		},
	}

	originalContent := make([]Content, len(messages[0].Content))
	copy(originalContent, messages[0].Content)

	reorderMessageContent(messages)

	// User messages should not be reordered
	for i, content := range messages[0].Content {
		if content != originalContent[i] {
			t.Error("User message content should not be reordered")
		}
	}
}

func TestReorderMessageContent_AssistantMessageSingleBlock(t *testing.T) {
	messages := []*Message{
		{
			Role: Assistant,
			Content: []Content{
				&TextContent{Text: "Hello"},
			},
		},
	}

	originalContent := messages[0].Content[0]
	reorderMessageContent(messages)

	// Single block should not be reordered
	if messages[0].Content[0] != originalContent {
		t.Error("Single block message should not be reordered")
	}
}

func TestReorderMessageContent_AssistantMessageReordering(t *testing.T) {
	textContent := &TextContent{Text: "Here's the result:"}
	toolUseContent := &ToolUseContent{ID: "tool1", Name: "test_tool"}
	textContent2 := &TextContent{Text: "Additional text"}
	toolUseContent2 := &ToolUseContent{ID: "tool2", Name: "test_tool2"}

	messages := []*Message{
		{
			Role: Assistant,
			Content: []Content{
				toolUseContent,
				textContent,
				toolUseContent2,
				textContent2,
			},
		},
	}

	reorderMessageContent(messages)

	// Should have all text content first, then all tool use content
	expectedOrder := []Content{textContent, textContent2, toolUseContent, toolUseContent2}

	if len(messages[0].Content) != len(expectedOrder) {
		t.Fatalf("Expected %d content blocks, got %d", len(expectedOrder), len(messages[0].Content))
	}

	for i, expected := range expectedOrder {
		if messages[0].Content[i] != expected {
			t.Errorf("Content block %d: expected %T, got %T", i, expected, messages[0].Content[i])
		}
	}
}

func TestReorderMessageContent_AssistantMessageNoToolUse(t *testing.T) {
	textContent1 := &TextContent{Text: "First text"}
	textContent2 := &TextContent{Text: "Second text"}

	messages := []*Message{
		{
			Role:    Assistant,
			Content: []Content{textContent1, textContent2},
		},
	}

	originalOrder := []Content{textContent1, textContent2}
	reorderMessageContent(messages)

	// Should remain unchanged when no tool use blocks
	for i, expected := range originalOrder {
		if messages[0].Content[i] != expected {
			t.Errorf("Content block %d should remain unchanged", i)
		}
	}
}

func TestReorderMessageContent_AssistantMessageOnlyToolUse(t *testing.T) {
	toolUse1 := &ToolUseContent{ID: "tool1", Name: "test_tool1"}
	toolUse2 := &ToolUseContent{ID: "tool2", Name: "test_tool2"}

	messages := []*Message{
		{
			Role:    Assistant,
			Content: []Content{toolUse1, toolUse2},
		},
	}

	originalOrder := []Content{toolUse1, toolUse2}
	reorderMessageContent(messages)

	// Should remain unchanged when only tool use blocks
	for i, expected := range originalOrder {
		if messages[0].Content[i] != expected {
			t.Errorf("Content block %d should remain unchanged", i)
		}
	}
}

func TestAddPrefill_EmptyPrefill(t *testing.T) {
	blocks := []Content{
		&TextContent{Text: "Original text"},
	}

	err := addPrefill(blocks, "", "")
	if err != nil {
		t.Errorf("Expected no error for empty prefill, got %v", err)
	}

	textContent := blocks[0].(*TextContent)
	if textContent.Text != "Original text" {
		t.Error("Text should remain unchanged when prefill is empty")
	}
}

func TestAddPrefill_NoClosingTag(t *testing.T) {
	blocks := []Content{
		&TextContent{Text: "Original text"},
	}

	err := addPrefill(blocks, "Prefill: ", "")
	if err != nil {
		t.Errorf("Expected no error when no closing tag specified, got %v", err)
	}

	textContent := blocks[0].(*TextContent)
	expected := "Prefill: Original text"
	if textContent.Text != expected {
		t.Errorf("Expected %q, got %q", expected, textContent.Text)
	}
}

func TestAddPrefill_WithClosingTagFound(t *testing.T) {
	blocks := []Content{
		&TextContent{Text: "Some text </result> more text"},
	}

	err := addPrefill(blocks, "Prefill: ", "</result>")
	if err != nil {
		t.Errorf("Expected no error when closing tag is found, got %v", err)
	}

	textContent := blocks[0].(*TextContent)
	expected := "Prefill: Some text </result> more text"
	if textContent.Text != expected {
		t.Errorf("Expected %q, got %q", expected, textContent.Text)
	}
}

func TestAddPrefill_WithClosingTagNotFound(t *testing.T) {
	blocks := []Content{
		&TextContent{Text: "Some text without closing tag"},
	}

	err := addPrefill(blocks, "Prefill: ", "</result>")
	if err == nil {
		t.Error("Expected error when closing tag is not found")
	}

	expectedError := "prefill closing tag not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestAddPrefill_NoTextContent(t *testing.T) {
	blocks := []Content{
		&ToolUseContent{ID: "tool1", Name: "test_tool"},
	}

	err := addPrefill(blocks, "Prefill: ", "")
	if err == nil {
		t.Error("Expected error when no text content found")
	}

	expectedError := "no text content found in message"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestAddPrefill_EmptyBlocks(t *testing.T) {
	var blocks []Content

	err := addPrefill(blocks, "Prefill: ", "")
	if err == nil {
		t.Error("Expected error when no blocks provided")
	}

	expectedError := "no text content found in message"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}

func TestAddPrefill_MultipleTextBlocks(t *testing.T) {
	blocks := []Content{
		&TextContent{Text: "First text"},
		&TextContent{Text: "Second text"},
	}

	err := addPrefill(blocks, "Prefill: ", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Should modify the first text block
	firstText := blocks[0].(*TextContent)
	if firstText.Text != "Prefill: First text" {
		t.Errorf("Expected first block to be modified, got %q", firstText.Text)
	}

	// Second block should remain unchanged
	secondText := blocks[1].(*TextContent)
	if secondText.Text != "Second text" {
		t.Errorf("Expected second block to remain unchanged, got %q", secondText.Text)
	}
}
