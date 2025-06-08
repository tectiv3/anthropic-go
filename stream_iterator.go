package anthropic

import (
	"io"
	"strings"
	"sync"
)

// StreamIterator implements the StreamIterator interface
type StreamIterator struct {
	reader            *ServerSentEventsReader[Event]
	body              io.ReadCloser
	err               error
	currentEvent      *Event
	prefill           string
	prefillClosingTag string
	closeOnce         sync.Once
}

// Next advances to the next event in the stream. Returns true if an event was
// successfully read, false when the stream is complete or an error occurs.
func (s *StreamIterator) Next() bool {
	for {
		event, ok := s.reader.Next()
		if !ok {
			s.err = s.reader.Err()
			s.Close()
			return false
		}
		processedEvent := s.processEvent(&event)
		if processedEvent != nil {
			s.currentEvent = processedEvent
			return true
		}
	}
}

// Event returns the current event. Should only be called after a successful Next().
func (s *StreamIterator) Event() *Event {
	return s.currentEvent
}

// processEvent processes an Anthropic event and applies prefill logic if needed
func (s *StreamIterator) processEvent(event *Event) *Event {
	if event.Type == "" {
		return nil
	}

	// Apply prefill logic for the first text content block
	if s.prefill != "" && event.Type == EventTypeContentBlockStart {
		if event.ContentBlock != nil && event.ContentBlock.Type == ContentTypeText {
			// Add prefill to the beginning of the text
			if s.prefillClosingTag == "" || strings.Contains(event.ContentBlock.Text, s.prefillClosingTag) {
				event.ContentBlock.Text = s.prefill + event.ContentBlock.Text
				s.prefill = "" // Only apply prefill once
			}
		}
	}

	return event
}

func (s *StreamIterator) Close() error {
	var err error
	s.closeOnce.Do(func() { err = s.body.Close() })
	return err
}

func (s *StreamIterator) Err() error {
	return s.err
}
