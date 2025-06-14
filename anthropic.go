package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tectiv3/anthropic-go/retry"
)

const ProviderName = "anthropic"

var (
	DefaultModel         = "claude-sonnet-4-20250514"
	DefaultEndpoint      = "https://api.anthropic.com/v1/messages"
	DefaultMaxTokens     = 4096
	DefaultClient        = &http.Client{Timeout: 300 * time.Second}
	DefaultMaxRetries    = 6
	DefaultRetryBaseWait = 2 * time.Second
	DefaultVersion       = "2023-06-01"
)

func New(opts ...Option) *Client {
	p := &Client{
		apiKey:        os.Getenv("ANTHROPIC_API_KEY"),
		endpoint:      DefaultEndpoint,
		client:        DefaultClient,
		model:         DefaultModel,
		maxTokens:     DefaultMaxTokens,
		maxRetries:    DefaultMaxRetries,
		retryBaseWait: DefaultRetryBaseWait,
		version:       DefaultVersion,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (p *Client) Name() string {
	return ProviderName
}

func (p *Client) Generate(ctx context.Context, messages Messages) (*Response, error) {
	var request Request
	if err := p.applyRequestConfig(&request); err != nil {
		return nil, err
	}
	msgs, err := convertMessages(messages)
	if err != nil {
		return nil, err
	}
	request.Messages = msgs

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	var result Response
	err = retry.Do(ctx, func() error {
		req, err := p.createRequest(ctx, body, false)
		if err != nil {
			return err
		}
		resp, err := p.client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == 429 {
				log.Println("rate limit exceeded, status: %s, body: %s", resp.StatusCode, string(body))
			}
			return NewError(resp.StatusCode, string(body))
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("error decoding response: %w", err)
		}
		return nil
	}, retry.WithMaxRetries(p.maxRetries), retry.WithBaseWait(p.retryBaseWait))
	if err != nil {
		return nil, err
	}
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response from anthropic api")
	}

	return &result, nil
}

func (p *Client) Stream(ctx context.Context, messages Messages) (*StreamIterator, error) {
	var request Request
	if err := p.applyRequestConfig(&request); err != nil {
		return nil, err
	}
	msgs, err := convertMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("error converting messages: %w", err)
	}

	request.Messages = msgs
	request.Stream = true

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	var stream *StreamIterator
	err = retry.Do(ctx, func() error {
		req, err := p.createRequest(ctx, body, true)
		if err != nil {
			return err
		}
		resp, err := p.client.Do(req)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == 429 {
				log.Println("rate limit exceeded, status: %s, body: %s", resp.StatusCode, string(body))
			}
			return NewError(resp.StatusCode, string(body))
		}
		stream = &StreamIterator{
			body:   resp.Body,
			reader: NewServerSentEventsReader[Event](resp.Body),
		}
		return nil
	}, retry.WithMaxRetries(p.maxRetries), retry.WithBaseWait(p.retryBaseWait))
	if err != nil {
		return nil, err
	}
	return stream, nil
}

func convertMessages(messages []*Message) ([]*Message, error) {
	messageCount := len(messages)
	if messageCount == 0 {
		return nil, fmt.Errorf("no messages provided")
	}
	for i, message := range messages {
		if len(message.Content) == 0 {
			return nil, fmt.Errorf("empty message detected (index %d)", i)
		}
	}
	// Workaround for Anthropic bug
	reorderMessageContent(messages)
	// Anthropic errors if a message ID is set, so make a copy of the messages
	// and omit the ID field
	copied := make([]*Message, len(messages))
	for i, message := range messages {
		// The "name" field in tool results can't be set either
		var copiedContent []Content
		for _, content := range message.Content {
			switch c := content.(type) {
			case *ToolResultContent:
				copiedContent = append(copiedContent, &ToolResultContent{
					Content:      c.Content,
					ToolUseID:    c.ToolUseID,
					IsError:      c.IsError,
					CacheControl: c.CacheControl,
				})
			case *DocumentContent:
				// Handle DocumentContent with file IDs for Anthropic API compatibility
				if c.Source != nil && c.Source.Type == ContentSourceTypeFile && c.Source.FileID != "" {
					// For Anthropic API, file IDs are passed in the source structure
					docContent := &DocumentContent{
						Title:        c.Title,
						Context:      c.Context,
						Citations:    c.Citations,
						CacheControl: c.CacheControl,
						Source: &ContentSource{
							Type:   c.Source.Type,
							FileID: c.Source.FileID,
						},
					}
					copiedContent = append(copiedContent, docContent)
				} else {
					// Pass through other DocumentContent as-is
					copiedContent = append(copiedContent, content)
				}
			default:
				copiedContent = append(copiedContent, content)
			}
		}
		copied[i] = &Message{
			Role:    message.Role,
			Content: copiedContent,
		}
	}
	return copied, nil
}

func (p *Client) applyRequestConfig(req *Request) error {
	req.Model = p.model
	req.MaxTokens = &p.maxTokens

	if len(p.Tools) > 0 {
		var tools []map[string]any
		for _, tool := range p.Tools {
			// Handle tools that explicitly provide a configuration
			if toolWithConfig, ok := tool.(ToolConfiguration); ok {
				toolConfig := toolWithConfig.ToolConfiguration(p.Name())
				// nil means no configuration is specified and to use the default
				if toolConfig != nil {
					tools = append(tools, toolConfig)
					continue
				}
			}
			// Handle tools with the default configuration behavior
			schema := tool.Schema()
			toolConfig := map[string]any{
				"name":        tool.Name(),
				"description": tool.Description(),
			}
			if schema.Type != "" {
				toolConfig["input_schema"] = schema
			}
			tools = append(tools, toolConfig)
		}
		req.Tools = tools
	}

	if p.ToolChoice != nil && len(p.Tools) > 0 {
		req.ToolChoice = &ToolChoice{
			Type: ToolChoiceType(p.ToolChoice.Type),
			Name: p.ToolChoice.Name,
		}
		if p.ParallelToolCalls != nil && !*p.ParallelToolCalls {
			req.ToolChoice.DisableParallelToolUse = true
		}
	}

	if len(p.MCPServers) > 0 {
		req.MCPServers = p.MCPServers
	}

	req.Temperature = p.Temperature
	req.System = p.SystemPrompt

	return nil
}

// createRequest creates an HTTP request with appropriate headers for Anthropic API calls
func (p *Client) createRequest(ctx context.Context, body []byte, isStreaming bool) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", p.version)
	req.Header.Set("content-type", "application/json")

	if isStreaming {
		req.Header.Set("accept", "text/event-stream")
	}

	for key, values := range p.RequestHeaders {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return req, nil
}
