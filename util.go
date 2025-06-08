package anthropic

import (
	"fmt"
	"strings"
)

func reorderMessageContent(messages []*Message) {
	// For each assistant message, reorder content blocks so that all tool_use
	// content blocks appear after non-tool_use content blocks, while preserving
	// relative ordering within each group. This works-around an Anthropic bug.
	// https://github.com/anthropics/claude-code/issues/473
	for _, msg := range messages {
		if msg.Role != Assistant || len(msg.Content) < 2 {
			continue
		}
		// Separate blocks into tool use and non-tool use
		var toolUseBlocks []Content
		var otherBlocks []Content
		for _, block := range msg.Content {
			if block.Type() == ContentTypeToolUse {
				toolUseBlocks = append(toolUseBlocks, block)
			} else {
				otherBlocks = append(otherBlocks, block)
			}
		}
		// If we found any tool use blocks and other blocks, reorder them
		if len(toolUseBlocks) > 0 && len(otherBlocks) > 0 {
			// Combine slices with non-tool-use blocks first
			msg.Content = append(otherBlocks, toolUseBlocks...)
		}
	}
}

func addPrefill(blocks []Content, prefill, closingTag string) error {
	if prefill == "" {
		return nil
	}
	for _, block := range blocks {
		content, ok := block.(*TextContent)
		if ok {
			if closingTag == "" || strings.Contains(content.Text, closingTag) {
				content.Text = prefill + content.Text
				return nil
			}
			return fmt.Errorf("prefill closing tag not found")
		}
	}
	return fmt.Errorf("no text content found in message")
}
