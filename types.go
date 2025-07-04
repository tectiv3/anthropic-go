package anthropic

import (
	"fmt"
	"net/http"
	"time"
)

type CacheControlType string

const (
	CacheControlTypeEphemeral  CacheControlType = "ephemeral"
	CacheControlTypePersistent CacheControlType = "persistent"
)

// ReasoningEffort defines the effort level for reasoning aka extended thinking.
type ReasoningEffort string

const (
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortHigh   ReasoningEffort = "high"
)

// IsValid returns true if the reasoning effort is a known, valid value.
func (r ReasoningEffort) IsValid() bool {
	return r == ReasoningEffortLow ||
		r == ReasoningEffortMedium ||
		r == ReasoningEffortHigh
}

func (c CacheControlType) String() string {
	return string(c)
}

type ImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type Thinking struct {
	Type         string `json:"type"` // "enabled"
	BudgetTokens int    `json:"budget_tokens"`
}

type Request struct {
	Model       string            `json:"model"`
	Messages    []*Message        `json:"messages"`
	MaxTokens   *int              `json:"max_tokens,omitempty"`
	Temperature *float64          `json:"temperature,omitempty"`
	System      string            `json:"system,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Tools       []map[string]any  `json:"tools,omitempty"`
	ToolChoice  *ToolChoice       `json:"tool_choice,omitempty"`
	Thinking    *Thinking         `json:"thinking,omitempty"`
	MCPServers  []MCPServerConfig `json:"mcp_servers,omitempty"`
}

// Usage contains token usage information for an LLM response.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
}

// Copy returns a deep copy of the usage data.
func (u *Usage) Copy() *Usage {
	return &Usage{
		InputTokens:              u.InputTokens,
		OutputTokens:             u.OutputTokens,
		CacheCreationInputTokens: u.CacheCreationInputTokens,
		CacheReadInputTokens:     u.CacheReadInputTokens,
	}
}

// Add incremental usage to this usage object.
func (u *Usage) Add(other *Usage) {
	u.InputTokens += other.InputTokens
	u.OutputTokens += other.OutputTokens
	u.CacheCreationInputTokens += other.CacheCreationInputTokens
	u.CacheReadInputTokens += other.CacheReadInputTokens
}

// Option is a function that is used to adjust LLM configuration.
type Option func(*Client)

func WithAPIKey(apiKey string) Option {
	return func(p *Client) {
		p.apiKey = apiKey
	}
}

func WithEndpoint(endpoint string) Option {
	return func(p *Client) {
		p.endpoint = endpoint
	}
}

func WithClient(client *http.Client) Option {
	return func(p *Client) {
		p.client = client
	}
}

func WithMaxTokens(maxTokens int) Option {
	return func(p *Client) {
		p.maxTokens = maxTokens
	}
}

func WithModel(model string) Option {
	return func(p *Client) {
		p.model = model
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(p *Client) {
		p.maxRetries = maxRetries
	}
}

func WithBaseWait(baseWait time.Duration) Option {
	return func(p *Client) {
		p.retryBaseWait = baseWait
	}
}

func WithVersion(version string) Option {
	return func(p *Client) {
		p.version = version
	}
}

// WithTools sets the tools available to the client.
func WithTools(tools ...ToolInterface) Option {
	return func(p *Client) {
		p.Tools = tools
	}
}

// WithToolChoice sets the tool choice strategy for the client.
func WithToolChoice(choice *ToolChoice) Option {
	return func(p *Client) {
		p.ToolChoice = choice
	}
}

// WithSystemPrompt sets the system prompt for the client.
func WithSystemPrompt(prompt string) Option {
	return func(p *Client) {
		p.SystemPrompt = prompt
	}
}

type ClientError struct {
	statusCode int
	body       string
}

func (e *ClientError) Error() string {
	return fmt.Sprintf("provider api error (status %d): %s", e.statusCode, e.body)
}

func (e *ClientError) StatusCode() int {
	return e.statusCode
}

func (e *ClientError) IsRecoverable() bool {
	return ShouldRetry(e.statusCode)
}

func NewError(statusCode int, body string) *ClientError {
	return &ClientError{statusCode: statusCode, body: body}
}

// ShouldRetry determines if the given status code should trigger a retry
func ShouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || // 429
		statusCode == http.StatusInternalServerError || // 500
		statusCode == http.StatusServiceUnavailable || // 503
		statusCode == http.StatusGatewayTimeout || // 504
		statusCode == 520 // Cloudflare
}

type Client struct {
	client             *http.Client
	apiKey             string
	endpoint           string
	model              string
	maxTokens          int
	maxRetries         int
	retryBaseWait      time.Duration
	version            string
	SystemPrompt       string                   `json:"system_prompt,omitempty"`
	Tools              []ToolInterface          `json:"tools,omitempty"`
	ToolChoice         *ToolChoice              `json:"tool_choice,omitempty"`
	ParallelToolCalls  *bool                    `json:"parallel_tool_calls,omitempty"`
	MCPServers         []MCPServerConfig        `json:"mcp_servers,omitempty"`
	Prefill            string                   `json:"prefill,omitempty"`
	PrefillClosingTag  string                   `json:"prefill_closing_tag,omitempty"`
	MaxTokens          *int                     `json:"max_tokens,omitempty"`
	Temperature        *float64                 `json:"temperature,omitempty"`
	PresencePenalty    *float64                 `json:"presence_penalty,omitempty"`
	FrequencyPenalty   *float64                 `json:"frequency_penalty,omitempty"`
	ReasoningBudget    *int                     `json:"reasoning_budget,omitempty"`
	ReasoningEffort    ReasoningEffort          `json:"reasoning_effort,omitempty"`
	Features           []string                 `json:"features,omitempty"`
	RequestHeaders     http.Header              `json:"request_headers,omitempty"`
	Caching            *bool                    `json:"caching,omitempty"`
	PreviousResponseID string                   `json:"previous_response_id,omitempty"`
	ServiceTier        string                   `json:"service_tier,omitempty"`
	ClientOptions      map[string]interface{}   `json:"client_options,omitempty"`
	ResponseFormat     *ResponseFormat          `json:"response_format,omitempty"`
	Messages           Messages                 `json:"messages"`
	Client             *http.Client             `json:"-"`
	SSECallback        ServerSentEventsCallback `json:"-"`
}

// Apply applies the given options to the config.
func (c *Client) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
}
