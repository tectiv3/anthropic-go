package anthropic

import (
	"context"
	"encoding/json"
	"testing"
)

func TestToolAnnotations_MarshalJSON(t *testing.T) {
	annotations := &ToolAnnotations{
		Title:           "Test Tool",
		ReadOnlyHint:    true,
		DestructiveHint: false,
		IdempotentHint:  true,
		OpenWorldHint:   false,
		Extra: map[string]any{
			"custom_field": "custom_value",
			"number_field": 42,
		},
	}

	data, err := annotations.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Errorf("Failed to unmarshal result: %v", err)
	}

	if result["title"] != "Test Tool" {
		t.Errorf("Expected title 'Test Tool', got %v", result["title"])
	}
	if result["readOnlyHint"] != true {
		t.Errorf("Expected readOnlyHint true, got %v", result["readOnlyHint"])
	}
	if result["destructiveHint"] != false {
		t.Errorf("Expected destructiveHint false, got %v", result["destructiveHint"])
	}
	if result["custom_field"] != "custom_value" {
		t.Errorf("Expected custom_field 'custom_value', got %v", result["custom_field"])
	}
	if result["number_field"] != float64(42) {
		t.Errorf("Expected number_field 42, got %v", result["number_field"])
	}
}

func TestToolAnnotations_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"title": "Test Tool",
		"readOnlyHint": true,
		"destructiveHint": false,
		"idempotentHint": true,
		"openWorldHint": false,
		"custom_field": "custom_value",
		"number_field": 42
	}`

	var annotations ToolAnnotations
	err := annotations.UnmarshalJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if annotations.Title != "Test Tool" {
		t.Errorf("Expected title 'Test Tool', got %s", annotations.Title)
	}
	if annotations.ReadOnlyHint != true {
		t.Errorf("Expected readOnlyHint true, got %v", annotations.ReadOnlyHint)
	}
	if annotations.DestructiveHint != false {
		t.Errorf("Expected destructiveHint false, got %v", annotations.DestructiveHint)
	}
	if annotations.IdempotentHint != true {
		t.Errorf("Expected idempotentHint true, got %v", annotations.IdempotentHint)
	}
	if annotations.OpenWorldHint != false {
		t.Errorf("Expected openWorldHint false, got %v", annotations.OpenWorldHint)
	}
	if annotations.Extra["custom_field"] != "custom_value" {
		t.Errorf("Expected custom_field 'custom_value', got %v", annotations.Extra["custom_field"])
	}
	if annotations.Extra["number_field"] != float64(42) {
		t.Errorf("Expected number_field 42, got %v", annotations.Extra["number_field"])
	}
}

func TestToolResultContentType_String(t *testing.T) {
	tests := []struct {
		contentType ToolResultContentType
		expected    string
	}{
		{ToolResultContentTypeText, "text"},
		{ToolResultContentTypeImage, "image"},
		{ToolResultContentTypeAudio, "audio"},
		{ToolResultContentType("unknown"), "unknown"},
	}

	for _, test := range tests {
		result := test.contentType.String()
		if result != test.expected {
			t.Errorf("String() = %q, expected %q", result, test.expected)
		}
	}
}

func TestNewToolResultError(t *testing.T) {
	text := "Error occurred"
	result := NewToolResultError(text)

	if !result.IsError {
		t.Error("Expected IsError to be true")
	}

	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}

	content := result.Content[0]
	if content.Type != ToolResultContentTypeText {
		t.Errorf("Expected content type %v, got %v", ToolResultContentTypeText, content.Type)
	}
	if content.Text != text {
		t.Errorf("Expected text %q, got %q", text, content.Text)
	}
}

func TestNewToolResult(t *testing.T) {
	content1 := &ToolResultContents{
		Type: ToolResultContentTypeText,
		Text: "Result 1",
	}
	content2 := &ToolResultContents{
		Type: ToolResultContentTypeText,
		Text: "Result 2",
	}

	result := NewToolResult(content1, content2)

	if result.IsError {
		t.Error("Expected IsError to be false")
	}

	if len(result.Content) != 2 {
		t.Errorf("Expected 2 content items, got %d", len(result.Content))
	}

	if result.Content[0] != content1 {
		t.Error("First content item should match")
	}
	if result.Content[1] != content2 {
		t.Error("Second content item should match")
	}
}

func TestNewToolResultText(t *testing.T) {
	text := "Test result text"
	result := NewToolResultText(text)

	if result.IsError {
		t.Error("Expected IsError to be false")
	}

	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}

	content := result.Content[0]
	if content.Type != ToolResultContentTypeText {
		t.Errorf("Expected content type %v, got %v", ToolResultContentTypeText, content.Type)
	}
	if content.Text != text {
		t.Errorf("Expected text %q, got %q", text, content.Text)
	}
}

type mockTypedTool struct {
	name        string
	description string
	schema      *Schema
	annotations *ToolAnnotations
}

func (m *mockTypedTool) Name() string {
	return m.name
}

func (m *mockTypedTool) Description() string {
	return m.description
}

func (m *mockTypedTool) Schema() *Schema {
	return m.schema
}

func (m *mockTypedTool) Annotations() *ToolAnnotations {
	return m.annotations
}

func (m *mockTypedTool) Call(ctx context.Context, input string) (*ToolResult, error) {
	return NewToolResultText("mock result: " + input), nil
}

func TestToolAdapter(t *testing.T) {
	mockTool := &mockTypedTool{
		name:        "mock_tool",
		description: "A mock tool for testing",
		schema:      &Schema{Type: "object"},
		annotations: &ToolAnnotations{Title: "Mock Tool"},
	}

	adapter := ToolAdapter(mockTool)

	if adapter.Name() != "mock_tool" {
		t.Errorf("Expected name 'mock_tool', got %q", adapter.Name())
	}
	if adapter.Description() != "A mock tool for testing" {
		t.Errorf("Expected description 'A mock tool for testing', got %q", adapter.Description())
	}
	if adapter.Schema() != mockTool.schema {
		t.Error("Schema should match")
	}
	if adapter.Annotations() != mockTool.annotations {
		t.Error("Annotations should match")
	}
	if adapter.Unwrap() != mockTool {
		t.Error("Unwrap should return the original tool")
	}
}

func TestTypedToolAdapter_Call(t *testing.T) {
	mockTool := &mockTypedTool{
		name: "mock_tool",
	}

	adapter := ToolAdapter(mockTool)

	result, err := adapter.Call(context.Background(), "test input")
	if err != nil {
		t.Errorf("Call failed: %v", err)
	}

	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}

	if result.Content[0].Text != "mock result: test input" {
		t.Errorf("Expected 'mock result: test input', got %q", result.Content[0].Text)
	}
}

func TestToolChoiceType_IsValid(t *testing.T) {
	tests := []struct {
		choice   ToolChoiceType
		expected bool
	}{
		{ToolChoiceTypeAuto, true},
		{ToolChoiceTypeAny, true},
		{ToolChoiceTypeTool, true},
		{ToolChoiceTypeNone, true},
		{ToolChoiceType("invalid"), false},
		{ToolChoiceType(""), false},
	}

	for _, test := range tests {
		result := test.choice.IsValid()
		if result != test.expected {
			t.Errorf("IsValid(%q) = %v, expected %v", test.choice, result, test.expected)
		}
	}
}

func TestNewToolDefinition(t *testing.T) {
	tool := NewToolDefinition()

	if tool == nil {
		t.Error("NewToolDefinition should not return nil")
	}

	// Default values should be empty
	if tool.Name() != "" {
		t.Errorf("Expected empty name, got %q", tool.Name())
	}
	if tool.Description() != "" {
		t.Errorf("Expected empty description, got %q", tool.Description())
	}
	if tool.Schema() != nil {
		t.Error("Expected nil schema")
	}
}

func TestToolDefinition_WithName(t *testing.T) {
	tool := NewToolDefinition().WithName("test_tool")

	if tool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got %q", tool.Name())
	}
}

func TestToolDefinition_WithDescription(t *testing.T) {
	tool := NewToolDefinition().WithDescription("Test description")

	if tool.Description() != "Test description" {
		t.Errorf("Expected description 'Test description', got %q", tool.Description())
	}
}

func TestToolDefinition_WithSchema(t *testing.T) {
	schema := &Schema{Type: "object"}
	tool := NewToolDefinition().WithSchema(schema)

	if tool.Schema() != schema {
		t.Error("Schema should match")
	}
}

func TestToolDefinition_Chaining(t *testing.T) {
	schema := &Schema{Type: "object"}
	tool := NewToolDefinition().
		WithName("chained_tool").
		WithDescription("A tool created with method chaining").
		WithSchema(schema)

	if tool.Name() != "chained_tool" {
		t.Errorf("Expected name 'chained_tool', got %q", tool.Name())
	}
	if tool.Description() != "A tool created with method chaining" {
		t.Errorf("Expected description 'A tool created with method chaining', got %q", tool.Description())
	}
	if tool.Schema() != schema {
		t.Error("Schema should match")
	}
}
