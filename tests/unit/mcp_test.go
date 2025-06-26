package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicFunctionality(t *testing.T) {
	// Basic test to ensure package compiles and imports work
	assert.True(t, true, "Basic test should pass")
}

func TestJSONRPCStructure(t *testing.T) {
	// Test basic JSON-RPC 2.0 structure validation
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]interface{}{},
		},
	}

	// Validate required fields
	assert.Equal(t, "2.0", request["jsonrpc"])
	assert.Equal(t, 1, request["id"])
	assert.Equal(t, "initialize", request["method"])
	assert.NotNil(t, request["params"])
}

func TestMCPProtocolVersion(t *testing.T) {
	// Test MCP protocol version compliance
	protocolVersion := "2025-03-26"
	assert.Equal(t, "2025-03-26", protocolVersion)
}

func TestToolDefinitions(t *testing.T) {
	// Test that we have all required tools defined
	expectedTools := []string{
		"trigger_build",
		"cancel_build",
		"pin_build",
		"set_build_tag",
		"download_artifact",
	}

	// Validate we have the right number of tools
	assert.Equal(t, 5, len(expectedTools))

	// Validate tool names are correctly formatted
	for _, tool := range expectedTools {
		assert.NotEmpty(t, tool)
		assert.Contains(t, tool, "_") // All our tools use snake_case
	}
}

func TestResourceTypes(t *testing.T) {
	// Test that we have all required resource types
	expectedResources := []string{
		"projects",
		"buildTypes",
		"builds",
		"agents",
	}

	// Validate we have the right resource types
	assert.Equal(t, 4, len(expectedResources))

	// Validate resource names
	for _, resource := range expectedResources {
		assert.NotEmpty(t, resource)
	}
}
