package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tectiv3/anthropic-go/retry"
)

type CapabilitySupport struct {
	Supported bool `json:"supported"`
}

type ThinkingTypes struct {
	Enabled  CapabilitySupport `json:"enabled"`
	Adaptive CapabilitySupport `json:"adaptive"`
}

type ThinkingCapability struct {
	Supported bool          `json:"supported"`
	Types     ThinkingTypes `json:"types"`
}

type EffortCapability struct {
	Supported bool              `json:"supported"`
	Low       CapabilitySupport `json:"low"`
	Medium    CapabilitySupport `json:"medium"`
	High      CapabilitySupport `json:"high"`
	Max       CapabilitySupport `json:"max"`
	XHigh     CapabilitySupport `json:"xhigh"`
}

type ContextManagementCapability struct {
	Supported           bool              `json:"supported"`
	ClearThinking       CapabilitySupport `json:"clear_thinking_20251015"`
	ClearToolUses       CapabilitySupport `json:"clear_tool_uses_20250919"`
	Compact             CapabilitySupport `json:"compact_20260112"`
}

type ModelCapabilities struct {
	Batch             CapabilitySupport           `json:"batch"`
	Citations         CapabilitySupport           `json:"citations"`
	CodeExecution     CapabilitySupport           `json:"code_execution"`
	ContextManagement ContextManagementCapability `json:"context_management"`
	Effort            EffortCapability            `json:"effort"`
	ImageInput        CapabilitySupport           `json:"image_input"`
	PDFInput          CapabilitySupport           `json:"pdf_input"`
	StructuredOutputs CapabilitySupport           `json:"structured_outputs"`
	Thinking          ThinkingCapability          `json:"thinking"`
}

type Model struct {
	ID             string            `json:"id"`
	Type           string            `json:"type"`
	DisplayName    string            `json:"display_name"`
	CreatedAt      string            `json:"created_at"`
	MaxInputTokens int               `json:"max_input_tokens"`
	MaxTokens      int               `json:"max_tokens"`
	Capabilities   ModelCapabilities `json:"capabilities"`
}

type ModelListResponse struct {
	Data    []Model `json:"data"`
	FirstID string  `json:"first_id"`
	LastID  string  `json:"last_id"`
	HasMore bool    `json:"has_more"`
}

type ModelListParams struct {
	Limit    int
	AfterID  string
	BeforeID string
}

func (p *Client) ListModels(ctx context.Context, params *ModelListParams) (*ModelListResponse, error) {
	path := "/models"
	if params != nil {
		q := url.Values{}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.AfterID != "" {
			q.Set("after_id", params.AfterID)
		}
		if params.BeforeID != "" {
			q.Set("before_id", params.BeforeID)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var result ModelListResponse
	err := retry.Do(ctx, func() error {
		req, err := p.createRequest(ctx, "GET", path, nil, false)
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
				log.Printf("rate limit exceeded, status: %d, body: %s", resp.StatusCode, string(body))
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

	return &result, nil
}

func (p *Client) GetModel(ctx context.Context, modelID string) (*Model, error) {
	if modelID == "" {
		return nil, fmt.Errorf("modelID must not be empty")
	}
	if strings.ContainsAny(modelID, "/?#") {
		return nil, fmt.Errorf("modelID contains invalid characters: %q", modelID)
	}
	path := "/models/" + url.PathEscape(modelID)

	var result Model
	err := retry.Do(ctx, func() error {
		req, err := p.createRequest(ctx, "GET", path, nil, false)
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
				log.Printf("rate limit exceeded, status: %d, body: %s", resp.StatusCode, string(body))
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

	return &result, nil
}
