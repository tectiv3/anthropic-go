package anthropic

import (
	"testing"
)

func TestNewMessage(t *testing.T) {
	textContent := &TextContent{Text: "Hello"}
	toolContent := &ToolUseContent{ID: "tool1", Name: "test"}
	content := []Content{textContent, toolContent}

	msg := NewMessage(Assistant, content)

	if msg.Role != Assistant {
		t.Errorf("Expected role %v, got %v", Assistant, msg.Role)
	}

	if len(msg.Content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	if msg.Content[0] != textContent {
		t.Error("First content block should match")
	}
	if msg.Content[1] != toolContent {
		t.Error("Second content block should match")
	}
}

func TestNewUserMessage(t *testing.T) {
	textContent := &TextContent{Text: "User message"}
	toolContent := &ToolUseContent{ID: "tool1", Name: "test"}

	msg := NewUserMessage(textContent, toolContent)

	if msg.Role != User {
		t.Errorf("Expected role %v, got %v", User, msg.Role)
	}

	if len(msg.Content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	if msg.Content[0] != textContent {
		t.Error("First content block should match")
	}
	if msg.Content[1] != toolContent {
		t.Error("Second content block should match")
	}
}

func TestNewUserMessage_NoContent(t *testing.T) {
	msg := NewUserMessage()

	if msg.Role != User {
		t.Errorf("Expected role %v, got %v", User, msg.Role)
	}

	if len(msg.Content) != 0 {
		t.Errorf("Expected 0 content blocks, got %d", len(msg.Content))
	}
}

func TestNewUserTextMessage(t *testing.T) {
	text := "Hello, world!"
	msg := NewUserTextMessage(text)

	if msg.Role != User {
		t.Errorf("Expected role %v, got %v", User, msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(msg.Content))
	}

	textContent, ok := msg.Content[0].(*TextContent)
	if !ok {
		t.Error("Content should be TextContent")
	}

	if textContent.Text != text {
		t.Errorf("Expected text %q, got %q", text, textContent.Text)
	}
}

func TestNewTextContent(t *testing.T) {
	text := "Sample text"
	content := NewTextContent(text)

	if content.Text != text {
		t.Errorf("Expected text %q, got %q", text, content.Text)
	}

	if content.Type() != ContentTypeText {
		t.Errorf("Expected type %v, got %v", ContentTypeText, content.Type())
	}
}

func TestNewAssistantMessage(t *testing.T) {
	textContent := &TextContent{Text: "Assistant response"}
	toolContent := &ToolUseContent{ID: "tool1", Name: "test"}

	msg := NewAssistantMessage(textContent, toolContent)

	if msg.Role != Assistant {
		t.Errorf("Expected role %v, got %v", Assistant, msg.Role)
	}

	if len(msg.Content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	if msg.Content[0] != textContent {
		t.Error("First content block should match")
	}
	if msg.Content[1] != toolContent {
		t.Error("Second content block should match")
	}
}

func TestNewAssistantTextMessage(t *testing.T) {
	text := "Assistant response"
	msg := NewAssistantTextMessage(text)

	if msg.Role != Assistant {
		t.Errorf("Expected role %v, got %v", Assistant, msg.Role)
	}

	if len(msg.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(msg.Content))
	}

	textContent, ok := msg.Content[0].(*TextContent)
	if !ok {
		t.Error("Content should be TextContent")
	}

	if textContent.Text != text {
		t.Errorf("Expected text %q, got %q", text, textContent.Text)
	}
}

func TestNewToolResultMessage(t *testing.T) {
	result1 := &ToolResultContent{
		ToolUseID: "tool1",
		Content:   "Result 1",
	}
	result2 := &ToolResultContent{
		ToolUseID: "tool2",
		Content:   "Result 2",
	}

	msg := NewToolResultMessage(result1, result2)

	if msg.Role != User {
		t.Errorf("Expected role %v, got %v", User, msg.Role)
	}

	if len(msg.Content) != 2 {
		t.Errorf("Expected 2 content blocks, got %d", len(msg.Content))
	}

	// Check that the content is properly converted
	firstResult, ok := msg.Content[0].(*ToolResultContent)
	if !ok {
		t.Error("First content block should be ToolResultContent")
	} else if firstResult.ToolUseID != result1.ToolUseID {
		t.Error("First content block should have matching ToolUseID")
	}

	secondResult, ok := msg.Content[1].(*ToolResultContent)
	if !ok {
		t.Error("Second content block should be ToolResultContent")
	} else if secondResult.ToolUseID != result2.ToolUseID {
		t.Error("Second content block should have matching ToolUseID")
	}
}

func TestNewDocumentContent(t *testing.T) {
	source := &ContentSource{
		Type:      ContentSourceTypeBase64,
		MediaType: "application/pdf",
		Data:      "base64data",
	}

	content := NewDocumentContent(source)

	if content.Source != source {
		t.Error("Source should match")
	}

	if content.Type() != ContentTypeDocument {
		t.Errorf("Expected type %v, got %v", ContentTypeDocument, content.Type())
	}
}

func TestEncodedData(t *testing.T) {
	mediaType := "image/png"
	base64Data := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

	source := EncodedData(mediaType, base64Data)

	if source.Type != ContentSourceTypeBase64 {
		t.Errorf("Expected type %v, got %v", ContentSourceTypeBase64, source.Type)
	}

	if source.MediaType != mediaType {
		t.Errorf("Expected media type %q, got %q", mediaType, source.MediaType)
	}

	if source.Data != base64Data {
		t.Errorf("Expected data %q, got %q", base64Data, source.Data)
	}
}

func TestRawData(t *testing.T) {
	mediaType := "image/png"
	rawData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	source := RawData(mediaType, rawData)

	if source.Type != ContentSourceTypeBase64 {
		t.Errorf("Expected type %v, got %v", ContentSourceTypeBase64, source.Type)
	}

	if source.MediaType != mediaType {
		t.Errorf("Expected media type %q, got %q", mediaType, source.MediaType)
	}

	// Should be base64 encoded
	if len(source.Data) == 0 {
		t.Error("Data should not be empty")
	}
}

func TestContentURL(t *testing.T) {
	url := "https://example.com/image.png"

	source := ContentURL(url)

	if source.Type != ContentSourceTypeURL {
		t.Errorf("Expected type %v, got %v", ContentSourceTypeURL, source.Type)
	}

	if source.URL != url {
		t.Errorf("Expected URL %q, got %q", url, source.URL)
	}
}

func TestFileID(t *testing.T) {
	id := "file-123"

	source := FileID(id)

	if source.Type != ContentSourceTypeFile {
		t.Errorf("Expected type %v, got %v", ContentSourceTypeFile, source.Type)
	}

	if source.FileID != id {
		t.Errorf("Expected file ID %q, got %q", id, source.FileID)
	}
}
