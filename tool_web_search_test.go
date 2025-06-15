package anthropic

import (
	"encoding/json"
	"testing"
)

func TestNewWebSearchTool(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{
		MaxUses: 10,
	})

	if tool == nil {
		t.Error("NewWebSearchTool should not return nil")
	}

	if tool.Name() != "web_search" {
		t.Errorf("Expected name 'web_search', got %q", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// Schema should be nil for server-side tools
	if tool.Schema() != nil {
		t.Error("Expected nil schema for server-side tools")
	}
}

func TestNewWebSearchTool_DefaultValues(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{})

	config := tool.ToolConfiguration("anthropic")
	if config == nil {
		t.Error("Expected tool configuration")
	}

	if config["type"] != "web_search_20250305" {
		t.Errorf("Expected default type 'web_search_20250305', got %v", config["type"])
	}

	if config["max_uses"] != 5 {
		t.Errorf("Expected default max_uses 5, got %v", config["max_uses"])
	}

	if config["name"] != "web_search" {
		t.Errorf("Expected name 'web_search', got %v", config["name"])
	}
}

func TestNewWebSearchTool_CustomOptions(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{
		Type:           "web_search_custom",
		MaxUses:        3,
		AllowedDomains: []string{"example.com", "trusted.org"},
		BlockedDomains: []string{"spam.com"},
		UserLocation: &UserLocation{
			Type:     "approximate",
			City:     "San Francisco",
			Region:   "California",
			Country:  "US",
			Timezone: "America/Los_Angeles",
		},
	})

	config := tool.ToolConfiguration("anthropic")

	if config["type"] != "web_search_custom" {
		t.Errorf("Expected custom type 'web_search_custom', got %v", config["type"])
	}

	if config["max_uses"] != 3 {
		t.Errorf("Expected max_uses 3, got %v", config["max_uses"])
	}

	allowedDomains, ok := config["allowed_domains"].([]string)
	if !ok {
		t.Error("Expected allowed_domains to be []string")
	} else if len(allowedDomains) != 2 {
		t.Errorf("Expected 2 allowed domains, got %d", len(allowedDomains))
	}

	blockedDomains, ok := config["blocked_domains"].([]string)
	if !ok {
		t.Error("Expected blocked_domains to be []string")
	} else if len(blockedDomains) != 1 {
		t.Errorf("Expected 1 blocked domain, got %d", len(blockedDomains))
	}

	userLocation, ok := config["user_location"].(*UserLocation)
	if !ok {
		t.Error("Expected user_location to be *UserLocation")
	} else {
		if userLocation.City != "San Francisco" {
			t.Errorf("Expected city 'San Francisco', got %q", userLocation.City)
		}
		if userLocation.Country != "US" {
			t.Errorf("Expected country 'US', got %q", userLocation.Country)
		}
	}
}

func TestWebSearchTool_ToolConfiguration_Serialization(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{
		MaxUses:        5,
		AllowedDomains: []string{"example.com"},
		UserLocation: &UserLocation{
			Type:    "approximate",
			City:    "New York",
			Country: "US",
		},
	})

	config := tool.ToolConfiguration("anthropic")
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal tool configuration: %v", err)
	}

	// Verify the JSON structure matches API examples
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal result: %v", err)
	}

	expectedFields := []string{"type", "name", "max_uses", "allowed_domains", "user_location"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field %q in serialized config", field)
		}
	}

	t.Logf("Serialized config: %s", string(data))
}

func TestWebSearchTool_ImplementsInterfaces(t *testing.T) {
	tool := NewWebSearchTool(WebSearchToolOptions{})

	// Should implement ToolInterface
	var _ ToolInterface = tool

	// Should implement ToolConfiguration
	var _ ToolConfiguration = tool

	// Verify interface methods work
	if tool.Name() != "web_search" {
		t.Errorf("Name() should return 'web_search', got %q", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("Description() should return non-empty string")
	}

	if tool.Schema() != nil {
		t.Error("Schema() should return nil for server-side tools")
	}

	config := tool.ToolConfiguration("anthropic")
	if config == nil {
		t.Error("ToolConfiguration() should return non-nil config")
	}
}
