package anthropic

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListModels(t *testing.T) {
	response := ModelListResponse{
		Data: []Model{
			{
				ID:             "claude-sonnet-4-20250514",
				Type:           "model",
				DisplayName:    "Claude Sonnet 4",
				CreatedAt:      "2025-05-14T00:00:00Z",
				MaxInputTokens: 200000,
				MaxTokens:      8192,
				Capabilities: ModelCapabilities{
					Thinking: ThinkingCapability{
						Supported: true,
						Types: ThinkingTypes{
							Enabled:  CapabilitySupport{Supported: true},
							Adaptive: CapabilitySupport{Supported: true},
						},
					},
					ImageInput: CapabilitySupport{Supported: true},
					PDFInput:   CapabilitySupport{Supported: true},
					Batch:      CapabilitySupport{Supported: true},
				},
			},
		},
		FirstID: "claude-sonnet-4-20250514",
		LastID:  "claude-sonnet-4-20250514",
		HasMore: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/models" {
			t.Errorf("expected path /v1/models, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key test-key, got %s", r.Header.Get("x-api-key"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithEndpoint(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	result, err := client.ListModels(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 model, got %d", len(result.Data))
	}
	if result.Data[0].ID != "claude-sonnet-4-20250514" {
		t.Errorf("expected model ID claude-sonnet-4-20250514, got %s", result.Data[0].ID)
	}
	if result.Data[0].DisplayName != "Claude Sonnet 4" {
		t.Errorf("expected display name 'Claude Sonnet 4', got %s", result.Data[0].DisplayName)
	}
	if !result.Data[0].Capabilities.Thinking.Supported {
		t.Error("expected thinking to be supported")
	}
	if !result.Data[0].Capabilities.PDFInput.Supported {
		t.Error("expected pdf_input to be supported")
	}
	if result.HasMore {
		t.Error("expected HasMore to be false")
	}
}

func TestListModels_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("after_id") == "" {
			json.NewEncoder(w).Encode(ModelListResponse{
				Data: []Model{
					{ID: "claude-opus-4-20250514", Type: "model", DisplayName: "Claude Opus 4"},
				},
				FirstID: "claude-opus-4-20250514",
				LastID:  "claude-opus-4-20250514",
				HasMore: true,
			})
		} else {
			if r.URL.Query().Get("after_id") != "claude-opus-4-20250514" {
				t.Errorf("expected after_id=claude-opus-4-20250514, got %s", r.URL.Query().Get("after_id"))
			}
			json.NewEncoder(w).Encode(ModelListResponse{
				Data: []Model{
					{ID: "claude-haiku-4-20250514", Type: "model", DisplayName: "Claude Haiku 4"},
				},
				FirstID: "claude-haiku-4-20250514",
				LastID:  "claude-haiku-4-20250514",
				HasMore: false,
			})
		}
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithEndpoint(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	// First page
	result, err := client.ListModels(context.Background(), &ModelListParams{Limit: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasMore {
		t.Error("expected HasMore to be true on first page")
	}
	if result.Data[0].ID != "claude-opus-4-20250514" {
		t.Errorf("expected claude-opus-4-20250514, got %s", result.Data[0].ID)
	}

	// Second page
	result, err = client.ListModels(context.Background(), &ModelListParams{
		Limit:   1,
		AfterID: result.LastID,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasMore {
		t.Error("expected HasMore to be false on second page")
	}
	if result.Data[0].ID != "claude-haiku-4-20250514" {
		t.Errorf("expected claude-haiku-4-20250514, got %s", result.Data[0].ID)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestGetModel(t *testing.T) {
	model := Model{
		ID:             "claude-sonnet-4-20250514",
		Type:           "model",
		DisplayName:    "Claude Sonnet 4",
		CreatedAt:      "2025-05-14T00:00:00Z",
		MaxInputTokens: 200000,
		MaxTokens:      8192,
		Capabilities: ModelCapabilities{
			Thinking: ThinkingCapability{
				Supported: true,
				Types: ThinkingTypes{
					Enabled:  CapabilitySupport{Supported: true},
					Adaptive: CapabilitySupport{Supported: true},
				},
			},
			Effort: EffortCapability{
				Supported: true,
				Low:       CapabilitySupport{Supported: true},
				Medium:    CapabilitySupport{Supported: true},
				High:      CapabilitySupport{Supported: true},
			},
			CodeExecution: CapabilitySupport{Supported: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/models/claude-sonnet-4-20250514" {
			t.Errorf("expected path /v1/models/claude-sonnet-4-20250514, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model)
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithEndpoint(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	result, err := client.GetModel(context.Background(), "claude-sonnet-4-20250514")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "claude-sonnet-4-20250514" {
		t.Errorf("expected ID claude-sonnet-4-20250514, got %s", result.ID)
	}
	if result.MaxInputTokens != 200000 {
		t.Errorf("expected MaxInputTokens 200000, got %d", result.MaxInputTokens)
	}
	if !result.Capabilities.Effort.Supported {
		t.Error("expected effort to be supported")
	}
	if !result.Capabilities.CodeExecution.Supported {
		t.Error("expected code_execution to be supported")
	}
}

func TestGetModel_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": {"type": "not_found_error", "message": "model not found"}}`))
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithEndpoint(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	_, err := client.GetModel(context.Background(), "nonexistent-model")
	if err == nil {
		t.Fatal("expected error for nonexistent model")
	}

	clientErr, ok := err.(*ClientError)
	if !ok {
		t.Fatalf("expected *ClientError, got %T", err)
	}
	if clientErr.StatusCode() != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", clientErr.StatusCode())
	}
}

func TestWithEndpoint_BackwardsCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "full URL with /messages suffix",
			input:    "https://api.anthropic.com/v1/messages",
			expected: "https://api.anthropic.com/v1",
		},
		{
			name:     "base URL without /messages",
			input:    "https://api.anthropic.com/v1",
			expected: "https://api.anthropic.com/v1",
		},
		{
			name:     "custom endpoint without /messages",
			input:    "https://custom.example.com/api",
			expected: "https://custom.example.com/api",
		},
		{
			name:     "custom endpoint with /messages",
			input:    "https://custom.example.com/v1/messages",
			expected: "https://custom.example.com/v1",
		},
		{
			name:     "trailing slash after /messages",
			input:    "https://api.anthropic.com/v1/messages/",
			expected: "https://api.anthropic.com/v1",
		},
		{
			name:     "trailing slash without /messages",
			input:    "https://api.anthropic.com/v1/",
			expected: "https://api.anthropic.com/v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := New(WithEndpoint(tt.input))
			if client.baseURL != tt.expected {
				t.Errorf("WithEndpoint(%q): expected baseURL %q, got %q", tt.input, tt.expected, client.baseURL)
			}
		})
	}
}

func TestGetModel_EmptyID(t *testing.T) {
	client := New(WithAPIKey("test-key"), WithMaxRetries(0))
	_, err := client.GetModel(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty modelID")
	}
}

func TestGetModel_InvalidID(t *testing.T) {
	client := New(WithAPIKey("test-key"), WithMaxRetries(0))
	for _, id := range []string{"../messages", "foo/bar", "id?x=1", "id#frag"} {
		_, err := client.GetModel(context.Background(), id)
		if err == nil {
			t.Errorf("expected error for modelID %q", id)
		}
	}
}

func TestWithBaseURL(t *testing.T) {
	client := New(WithBaseURL("https://custom.example.com/v1"))
	if client.baseURL != "https://custom.example.com/v1" {
		t.Errorf("expected baseURL https://custom.example.com/v1, got %s", client.baseURL)
	}

	client2 := New(WithBaseURL("https://custom.example.com/v1/"))
	if client2.baseURL != "https://custom.example.com/v1" {
		t.Errorf("expected trailing slash stripped, got %s", client2.baseURL)
	}
}

func TestListModels_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("limit") != "5" {
			t.Errorf("expected limit=5, got %s", q.Get("limit"))
		}
		if q.Get("after_id") != "cursor-abc" {
			t.Errorf("expected after_id=cursor-abc, got %s", q.Get("after_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ModelListResponse{Data: []Model{}})
	}))
	defer server.Close()

	client := New(
		WithAPIKey("test-key"),
		WithEndpoint(server.URL+"/v1/messages"),
		WithMaxRetries(0),
	)

	_, err := client.ListModels(context.Background(), &ModelListParams{
		Limit:   5,
		AfterID: "cursor-abc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
