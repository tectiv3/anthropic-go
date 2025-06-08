package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
)

// ToolInterface is an interface that defines a tool that can be called by an LLM.
type ToolInterface interface {
	// Name of the tool.
	Name() string

	// Description of the tool.
	Description() string

	// Schema describes the parameters used to call the tool.
	Schema() *Schema
}

type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	InputSchema Schema `json:"input_schema"`
}

type ToolAnnotations struct {
	Title           string         `json:"title,omitempty"`
	ReadOnlyHint    bool           `json:"readOnlyHint,omitempty"`
	DestructiveHint bool           `json:"destructiveHint,omitempty"`
	IdempotentHint  bool           `json:"idempotentHint,omitempty"`
	OpenWorldHint   bool           `json:"openWorldHint,omitempty"`
	Extra           map[string]any `json:"extra,omitempty"`
}

func (a *ToolAnnotations) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"title":           a.Title,
		"readOnlyHint":    a.ReadOnlyHint,
		"destructiveHint": a.DestructiveHint,
		"idempotentHint":  a.IdempotentHint,
		"openWorldHint":   a.OpenWorldHint,
	}
	if a.Extra != nil {
		for k, v := range a.Extra {
			data[k] = v
		}
	}
	return json.Marshal(data)
}

func (a *ToolAnnotations) UnmarshalJSON(data []byte) error {
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}
	// Extract known fields
	if title, ok := rawMap["title"]; ok {
		json.Unmarshal(title, &a.Title)
		delete(rawMap, "title")
	}
	// Handle boolean hints
	boolFields := map[string]*bool{
		"readOnlyHint":    &a.ReadOnlyHint,
		"destructiveHint": &a.DestructiveHint,
		"idempotentHint":  &a.IdempotentHint,
		"openWorldHint":   &a.OpenWorldHint,
	}
	for name, field := range boolFields {
		if val, ok := rawMap[name]; ok {
			json.Unmarshal(val, field)
			delete(rawMap, name)
		}
	}
	// Remaining fields go to Extra
	a.Extra = make(map[string]any)
	for k, v := range rawMap {
		var val any
		json.Unmarshal(v, &val)
		a.Extra[k] = val
	}
	return nil
}

type ToolResultContentType string

const (
	ToolResultContentTypeText  ToolResultContentType = "text"
	ToolResultContentTypeImage ToolResultContentType = "image"
	ToolResultContentTypeAudio ToolResultContentType = "audio"
)

func (t ToolResultContentType) String() string {
	return string(t)
}

type ToolResultContents struct {
	Type        ToolResultContentType `json:"type"`
	Text        string                `json:"text,omitempty"`
	Data        string                `json:"data,omitempty"`
	MimeType    string                `json:"mimeType,omitempty"`
	Annotations map[string]any        `json:"annotations,omitempty"`
}

// ToolResult is the output from a tool call.
type ToolResult struct {
	Content []*ToolResultContents `json:"content"`
	IsError bool                  `json:"isError,omitempty"`
}

// NewToolResultError creates a new ToolResult containing an error message.
func NewToolResultError(text string) *ToolResult {
	return &ToolResult{
		IsError: true,
		Content: []*ToolResultContents{
			{
				Type: ToolResultContentTypeText,
				Text: text,
			},
		},
	}
}

// NewToolResult creates a new ToolResult with the given content.
func NewToolResult(content ...*ToolResultContents) *ToolResult {
	return &ToolResult{Content: content}
}

// NewToolResultText creates a new ToolResult with the given text content.
func NewToolResultText(text string) *ToolResult {
	return NewToolResult(&ToolResultContents{
		Type: ToolResultContentTypeText,
		Text: text,
	})
}

// TypedTool is a tool that can be called with a specific type of input.
type TypedTool[T any] interface {
	// Name of the tool.
	Name() string

	// Description of the tool.
	Description() string

	// Schema describes the parameters used to call the tool.
	Schema() *Schema

	// Annotations returns optional properties that describe tool behavior.
	Annotations() *ToolAnnotations

	// Call is the function that is called to use the tool.
	Call(ctx context.Context, input T) (*ToolResult, error)
}

// ToolAdapter creates a new TypedToolAdapter for the given tool.
func ToolAdapter[T any](tool TypedTool[T]) *TypedToolAdapter[T] {
	return &TypedToolAdapter[T]{tool: tool}
}

// TypedToolAdapter is an adapter that allows a TypedTool to be used as a regular Tool.
// Specifically the Call method accepts `input any` and then internally unmarshals the input
// to the correct type and passes it to the TypedTool.
type TypedToolAdapter[T any] struct {
	tool TypedTool[T]
}

func (t *TypedToolAdapter[T]) Name() string {
	return t.tool.Name()
}

func (t *TypedToolAdapter[T]) Description() string {
	return t.tool.Description()
}

func (t *TypedToolAdapter[T]) Schema() *Schema {
	return t.tool.Schema()
}

func (t *TypedToolAdapter[T]) Annotations() *ToolAnnotations {
	return t.tool.Annotations()
}

func (t *TypedToolAdapter[T]) Call(ctx context.Context, input any) (*ToolResult, error) {
	// Pass through if the input is already the correct type
	if converted, ok := input.(T); ok {
		return t.tool.Call(ctx, converted)
	}

	// Access the raw JSON
	var data []byte
	var err error
	if raw, ok := input.(json.RawMessage); ok {
		data = raw
	} else if raw, ok := input.([]byte); ok {
		data = raw
	} else {
		data, err = json.Marshal(input)
		if err != nil {
			errMessage := fmt.Sprintf("invalid json for tool %s: %v", t.Name(), err)
			return NewToolResultError(errMessage), nil
		}
	}

	// Unmarshal into the typed input
	var typedInput T
	err = json.Unmarshal(data, &typedInput)
	if err != nil {
		errMessage := fmt.Sprintf("invalid json for tool %s: %v", t.Name(), err)
		return NewToolResultError(errMessage), nil
	}
	return t.tool.Call(ctx, typedInput)
}

// Unwrap returns the underlying TypedTool.
func (t *TypedToolAdapter[T]) Unwrap() TypedTool[T] {
	return t.tool
}

func (t *TypedToolAdapter[T]) ToolConfiguration(providerName string) map[string]any {
	if toolWithConfig, ok := t.tool.(ToolConfiguration); ok {
		return toolWithConfig.ToolConfiguration(providerName)
	}
	return nil
}

// ToolCallResult is a tool call that has been made. This is used to understand
// what calls have happened during an LLM interaction.
type ToolCallResult struct {
	ID     string
	Name   string
	Input  any
	Result *ToolUseResult
	Error  error
}

// ToolUseResult contains the result of a tool call
type ToolUseResult struct {
	ToolUseID string `json:"tool_use_id"`
	Output    string `json:"output,omitempty"`
	Error     error  `json:"error,omitempty"`
}

// ToolChoiceType is used to guide the LLM's choice of which tool to use.
type ToolChoiceType string

const (
	ToolChoiceTypeAuto ToolChoiceType = "auto"
	ToolChoiceTypeAny  ToolChoiceType = "any"
	ToolChoiceTypeTool ToolChoiceType = "tool"
	ToolChoiceTypeNone ToolChoiceType = "none"
)

// IsValid returns true if the ToolChoiceType is a known, valid value.
func (t ToolChoiceType) IsValid() bool {
	return t == ToolChoiceTypeAuto ||
		t == ToolChoiceTypeAny ||
		t == ToolChoiceTypeTool ||
		t == ToolChoiceTypeNone
}

// ToolChoiceAuto is a ToolChoice with type "auto".
var ToolChoiceAuto = &ToolChoice{Type: ToolChoiceTypeAuto}

// ToolChoiceAny is a ToolChoice with type "any".
var ToolChoiceAny = &ToolChoice{Type: ToolChoiceTypeAny}

// ToolChoiceNone is a ToolChoice with type "none".
var ToolChoiceNone = &ToolChoice{Type: ToolChoiceTypeNone}

// ToolChoice influences the behavior of the LLM when choosing which tool to use.
type ToolChoice struct {
	Type                   ToolChoiceType `json:"type"`
	Name                   string         `json:"name,omitempty"`
	DisableParallelToolUse bool           `json:"disable_parallel_tool_use,omitempty"`
}

// ToolConfiguration is an interface that may be implemented by a Tool to
// provide explicit JSON configuration to pass to the LLM provider.
type ToolConfiguration interface {
	// ToolConfiguration returns a map of configuration for the tool, when used
	// with the given provider.
	ToolConfiguration(providerName string) map[string]any
}

// NewToolDefinition creates a new ToolDefinition.
func NewToolDefinition() *ToolDefinition {
	return &ToolDefinition{}
}

// ToolDefinition is a concrete implementation of the Tool interface. Note this
// does not provide a mechanism for calling the tool, but only for describing
// what the tool does so the LLM can understand it. You might not use this
// implementation if you use a full dive.Tool implementation in your app.
type ToolDefinition struct {
	name        string
	description string
	schema      *Schema
}

// Name returns the name of the tool, per the Tool interface.
func (t *ToolDefinition) Name() string {
	return t.name
}

// Description returns the description of the tool, per the Tool interface.
func (t *ToolDefinition) Description() string {
	return t.description
}

// Schema returns the schema of the tool, per the Tool interface.
func (t *ToolDefinition) Schema() *Schema {
	return t.schema
}

// WithName sets the name of the tool.
func (t *ToolDefinition) WithName(name string) *ToolDefinition {
	t.name = name
	return t
}

// WithDescription sets the description of the tool.
func (t *ToolDefinition) WithDescription(description string) *ToolDefinition {
	t.description = description
	return t
}

// WithSchema sets the schema of the tool.
func (t *ToolDefinition) WithSchema(schema *Schema) *ToolDefinition {
	t.schema = schema
	return t
}
