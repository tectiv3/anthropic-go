# Testing Guide for anthropic-go

This library includes comprehensive test coverage with both unit tests and integration tests.

## Test Overview

- **🧪 114 Total Tests** across 11 test files
- **📋 Unit Tests**: Test individual components and functions
- **🌐 Integration Tests**: Test real API interactions (requires API key)
- **✅ API Compatibility Tests**: Verify JSON structures match Anthropic's format

## Running Tests

### Unit Tests Only
```bash
go test ./...
```

### Integration Tests (Requires API Key)
```bash
# Set your API key
export ANTHROPIC_API_KEY="your_anthropic_api_key_here"

# Run integration tests
go test -tags=integration -v ./...

# Or use the helper script
./run_integration_tests.sh
```

### Specific Test Categories
```bash
# API compatibility tests
go test -run "API" -v .

# Streaming tests
go test -run "Stream" -v .

# Tool tests
go test -run "Tool" -v .

# Message handling tests
go test -run "Message" -v .

# Retry logic tests
go test -v ./retry
```

## Test Files

| File | Purpose | Tests |
|------|---------|-------|
| `anthropic_test.go` | Client initialization and core functions | 7 |
| `message_test.go` | Message structures and methods | 12 |
| `message_helpers_test.go` | Message creation helpers | 14 |
| `types_test.go` | Type definitions and configurations | 17 |
| `util_test.go` | Utility functions | 13 |
| `tool_test.go` | Tool system and adapters | 16 |
| `stream_test.go` | Streaming event processing | 7 |
| `retry/retry_test.go` | Retry logic with backoff | 8 |
| `retry/recoverable_error_test.go` | Error handling | 9 |
| `api_compatibility_test.go` | API format validation | 6 |
| `integration_test.go` | Real API tests | 3 |

## Integration Test Details

The integration tests validate real API interactions:

### 1. Simple Streaming Test
- Tests basic streaming without tools
- Validates event sequence: `message_start` → `content_block_start` → `content_block_delta` → `message_stop`
- Confirms text collection works correctly

### 2. Web Search Tool Definition Test
- Tests web search tool configuration
- Validates request structure matches API format
- Confirms tool definition serialization

### 3. Web Search Streaming Integration Test
- **Most Comprehensive Test**: Tests real web search with streaming
- Sends weather query that should trigger web search
- Validates complete streaming flow:
  - Initial text response
  - Server tool use (`server_tool_use`)
  - Input JSON delta events building search query
  - Web search results (`web_search_tool_result`)
  - Final synthesized response
- Confirms all event types and structures match API specification

## API Compatibility

The API compatibility tests verify our structures match Anthropic's exact format:

- ✅ **Request Format**: Model, messages, tools, streaming parameters
- ✅ **Response Format**: ID, role, content, usage, stop reason
- ✅ **Vision Support**: Base64 and URL image formats
- ✅ **Tool Use**: Tool definitions, tool use content, tool results
- ✅ **Streaming Events**: All event types with correct field structures
- ✅ **Web Search**: Server tool use and web search result formats

## Test Data Examples

The tests use realistic data matching the official API documentation:

```go
// Simple message
messages := Messages{
    NewUserTextMessage("Hello, Claude"),
}

// Vision with base64 image
imageContent := &ImageContent{
    Source: &ContentSource{
        Type:      ContentSourceTypeBase64,
        MediaType: "image/jpeg",
        Data:      "base64data",
    },
}

// Tool definition
toolDef := map[string]any{
    "name": "get_weather",
    "description": "Get the current weather in a given location",
    "input_schema": map[string]any{
        "type": "object",
        "properties": map[string]any{
            "location": map[string]any{
                "type": "string",
                "description": "The city and state, e.g. San Francisco, CA",
            },
        },
        "required": []string{"location"},
    },
}

// Web search tool (using proper constructor)
webSearchTool := NewWebSearchTool(WebSearchToolOptions{
    MaxUses: 5,
})

// Create client with tools
client := New(
    WithAPIKey("your-key"),
    WithTools(webSearchTool),
    WithSystemPrompt("You are a helpful assistant."),
)
```

## Coverage Areas

### Core Functionality
- ✅ Client configuration and initialization
- ✅ Message creation and manipulation
- ✅ Content type handling (text, images, tools, etc.)
- ✅ Request/response serialization
- ✅ Error handling and validation

### Advanced Features
- ✅ Streaming with all event types
- ✅ Tool system with adapters
- ✅ Vision (image) support
- ✅ Web search integration
- ✅ Retry logic with exponential backoff
- ✅ Cache control and usage tracking

### Edge Cases
- ✅ Empty messages and content
- ✅ Invalid configurations
- ✅ Network timeouts and cancellation
- ✅ Malformed responses
- ✅ Tool errors and recovery

## Running in CI/CD

For automated testing without API keys:
```bash
# Run only unit tests
go test ./...

# Skip integration tests automatically
go test -short ./...
```

For testing with API key in CI:
```bash
# Set API key in CI environment
export ANTHROPIC_API_KEY="$CI_ANTHROPIC_API_KEY"
go test -tags=integration -v ./...
```

## Test Performance

- **Unit Tests**: ~100ms (no network calls)
- **Integration Tests**: ~30-60s (depends on API response time)
- **Memory Usage**: Minimal (streaming tests validate memory efficiency)
