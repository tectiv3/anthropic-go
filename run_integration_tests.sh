#!/bin/bash

# Script to run integration tests for anthropic-go library
# Usage: ./run_integration_tests.sh
# Make sure to set ANTHROPIC_API_KEY environment variable

set -e

echo "🧪 Running Anthropic API Integration Tests"
echo "==========================================="

if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "❌ Error: ANTHROPIC_API_KEY environment variable is not set"
    echo "Please set your API key: export ANTHROPIC_API_KEY=your_key_here"
    exit 1
fi

echo "✅ API key found"
echo ""

echo "🔧 Running simple streaming test..."
go test -tags=integration -run TestSimpleStreamingIntegration -v .

echo ""
echo "🔧 Running web search tool definition test..."
go test -tags=integration -run TestWebSearchToolDefinitionIntegration -v .

echo ""
echo "🔧 Running web search streaming integration test..."
echo "⚠️  Note: This test depends on Claude deciding to use web search for the weather query"
go test -tags=integration -run TestWebSearchStreamingIntegration -v .

echo ""
echo "🎉 All integration tests completed!"
echo "📊 To run all tests (unit + integration): go test -tags=integration -v ./..."
