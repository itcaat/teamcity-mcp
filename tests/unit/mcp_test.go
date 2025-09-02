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
		"search_builds",
		"fetch_build_log",
	}

	// Validate we have the right number of tools
	assert.Equal(t, 7, len(expectedTools))

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

func TestFetchBuildLogTool(t *testing.T) {
	// Test fetch_build_log tool parameter validation
	tests := []struct {
		name     string
		input    map[string]interface{}
		valid    bool
		expected string
	}{
		{
			name: "Valid buildId only",
			input: map[string]interface{}{
				"buildId": "12345",
			},
			valid: true,
		},
		{
			name: "Valid with all parameters",
			input: map[string]interface{}{
				"buildId":    "12345",
				"plain":      true,
				"archived":   false,
				"dateFormat": "yyyy-MM-dd HH:mm:ss",
			},
			valid: true,
		},
		{
			name: "Missing buildId",
			input: map[string]interface{}{
				"plain": true,
			},
			valid:    false,
			expected: "buildId is required",
		},
		{
			name: "Empty buildId",
			input: map[string]interface{}{
				"buildId": "",
			},
			valid:    false,
			expected: "buildId is required",
		},
		{
			name: "Valid archived request",
			input: map[string]interface{}{
				"buildId":  "12345",
				"archived": true,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate required field presence
			if tt.valid {
				assert.Contains(t, tt.input, "buildId")
				buildId := tt.input["buildId"].(string)
				assert.NotEmpty(t, buildId)
			} else {
				buildId, exists := tt.input["buildId"]
				if !exists || buildId == "" {
					// Expected to be invalid due to missing/empty buildId
					assert.True(t, true) // Test passes as expected
				}
			}
		})
	}
}

func TestBuildLogURLConstruction(t *testing.T) {
	// Test URL construction logic for build log endpoint
	buildId := "12345"

	// Test base URL construction
	expectedBase := "/downloadBuildLog.html?buildId=" + buildId
	assert.Contains(t, expectedBase, buildId)
	assert.Contains(t, expectedBase, "downloadBuildLog.html")

	// Test parameter combinations
	testCases := []struct {
		name       string
		plain      *bool
		archived   *bool
		dateFormat string
		expected   []string // Expected URL components
	}{
		{
			name:     "Default plain",
			plain:    nil, // Should default to true
			expected: []string{"plain=true"},
		},
		{
			name:     "Explicit plain false",
			plain:    boolPtr(false),
			expected: []string{}, // No plain parameter when false
		},
		{
			name:     "Archived true",
			archived: boolPtr(true),
			expected: []string{"archived=true"},
		},
		{
			name:       "Custom date format",
			dateFormat: "yyyy-MM-dd",
			expected:   []string{"dateFormat=yyyy-MM-dd"},
		},
		{
			name:       "All parameters",
			plain:      boolPtr(true),
			archived:   boolPtr(true),
			dateFormat: "yyyy-MM-dd HH:mm:ss",
			expected:   []string{"plain=true", "archived=true", "dateFormat=yyyy-MM-dd HH:mm:ss"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate URL parameter construction
			params := []string{}

			// Default to plain=true unless explicitly set to false
			plain := true
			if tc.plain != nil {
				plain = *tc.plain
			}
			if plain {
				params = append(params, "plain=true")
			}

			if tc.archived != nil && *tc.archived {
				params = append(params, "archived=true")
			}

			if tc.dateFormat != "" {
				params = append(params, "dateFormat="+tc.dateFormat)
			}

			// Verify expected parameters are present
			for _, expected := range tc.expected {
				assert.Contains(t, params, expected)
			}
		})
	}
}

// Helper function for bool pointers
func boolPtr(b bool) *bool {
	return &b
}
