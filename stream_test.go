package anthropic

import (
	"testing"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventTypePing, "ping"},
		{EventTypeMessageStart, "message_start"},
		{EventTypeMessageDelta, "message_delta"},
		{EventTypeMessageStop, "message_stop"},
		{EventTypeContentBlockStart, "content_block_start"},
		{EventTypeContentBlockDelta, "content_block_delta"},
		{EventTypeContentBlockStop, "content_block_stop"},
	}

	for _, test := range tests {
		result := test.eventType.String()
		if result != test.expected {
			t.Errorf("String() = %q, expected %q", result, test.expected)
		}
	}
}

func TestEventDeltaType_String(t *testing.T) {
	tests := []struct {
		deltaType EventDeltaType
		expected  string
	}{
		{EventDeltaTypeText, "text_delta"},
		{EventDeltaTypeInputJSON, "input_json_delta"},
		{EventDeltaTypeThinking, "thinking_delta"},
		{EventDeltaTypeSignature, "signature_delta"},
		{EventDeltaTypeCitations, "citations_delta"},
	}

	for _, test := range tests {
		result := test.deltaType.String()
		if result != test.expected {
			t.Errorf("String() = %q, expected %q", result, test.expected)
		}
	}
}

func TestNewResponseAccumulator(t *testing.T) {
	accumulator := NewResponseAccumulator()

	if accumulator == nil {
		t.Error("NewResponseAccumulator should not return nil")
	}

	if accumulator.IsComplete() {
		t.Error("New accumulator should not be complete")
	}

	// Response should be nil until message_start event is added
	if accumulator.Response() != nil {
		t.Error("Response should be nil initially")
	}

	// Note: Usage() will panic if called before message_start, which is expected behavior
}

func TestResponseAccumulator_AddEvent_MessageStart(t *testing.T) {
	accumulator := NewResponseAccumulator()

	messageStartEvent := &Event{
		Type: EventTypeMessageStart,
		Message: &Response{
			ID:   "msg_123",
			Role: Assistant,
			Content: []Content{
				&TextContent{Text: ""},
			},
		},
	}

	err := accumulator.AddEvent(messageStartEvent)
	if err != nil {
		t.Errorf("AddEvent failed: %v", err)
	}

	response := accumulator.Response()
	if response.ID != "msg_123" {
		t.Errorf("Expected ID 'msg_123', got %q", response.ID)
	}
	if response.Role != Assistant {
		t.Errorf("Expected role %v, got %v", Assistant, response.Role)
	}
}

func TestResponseAccumulator_AddEvent_ContentBlockStart(t *testing.T) {
	accumulator := NewResponseAccumulator()

	// First add message start
	messageStartEvent := &Event{
		Type: EventTypeMessageStart,
		Message: &Response{
			ID:      "msg_123",
			Role:    Assistant,
			Content: []Content{},
		},
	}
	err := accumulator.AddEvent(messageStartEvent)
	if err != nil {
		t.Errorf("AddEvent (message_start) failed: %v", err)
	}

	// Then add content block start
	index := 0
	contentBlockStartEvent := &Event{
		Type:  EventTypeContentBlockStart,
		Index: &index,
		ContentBlock: &EventContentBlock{
			Type: ContentTypeText,
			Text: "",
		},
	}

	err = accumulator.AddEvent(contentBlockStartEvent)
	if err != nil {
		t.Errorf("AddEvent (content_block_start) failed: %v", err)
	}

	response := accumulator.Response()
	if len(response.Content) != 1 {
		t.Errorf("Expected 1 content block, got %d", len(response.Content))
	}
}

func TestResponseAccumulator_AddEvent_MessageStop(t *testing.T) {
	accumulator := NewResponseAccumulator()

	// Setup with message start
	messageStartEvent := &Event{
		Type: EventTypeMessageStart,
		Message: &Response{
			ID:      "msg_123",
			Role:    Assistant,
			Content: []Content{},
		},
	}
	accumulator.AddEvent(messageStartEvent)

	// Add message stop
	messageStopEvent := &Event{
		Type: EventTypeMessageStop,
	}

	err := accumulator.AddEvent(messageStopEvent)
	if err != nil {
		t.Errorf("AddEvent (message_stop) failed: %v", err)
	}

	if !accumulator.IsComplete() {
		t.Error("Accumulator should be complete after message_stop")
	}
}
