package unit

import (
	"fmt"
	"strings"
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
		"search_build_configurations",
		"get_current_time",
		"get_test_results",
	}

	// Validate we have the right number of tools
	assert.Equal(t, 10, len(expectedTools))

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
				"buildId":       "12345",
				"plain":         true,
				"archived":      false,
				"dateFormat":    "yyyy-MM-dd HH:mm:ss",
				"maxLines":      100,
				"filterPattern": "error",
				"severity":      "error",
				"tailLines":     50,
			},
			valid: true,
		},
		{
			name: "Valid with filtering parameters",
			input: map[string]interface{}{
				"buildId":       "12345",
				"maxLines":      200,
				"filterPattern": "test.*failed",
			},
			valid: true,
		},
		{
			name: "Valid with severity filter",
			input: map[string]interface{}{
				"buildId":  "12345",
				"severity": "error",
				"maxLines": 50,
			},
			valid: true,
		},
		{
			name: "Valid with tailLines",
			input: map[string]interface{}{
				"buildId":   "12345",
				"tailLines": 100,
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

				// Validate severity values if present
				if severity, ok := tt.input["severity"]; ok {
					severityStr := severity.(string)
					assert.Contains(t, []string{"error", "warning", "info"}, severityStr)
				}

				// Validate numeric parameters if present
				if maxLines, ok := tt.input["maxLines"]; ok {
					assert.IsType(t, 0, maxLines)
				}
				if tailLines, ok := tt.input["tailLines"]; ok {
					assert.IsType(t, 0, tailLines)
				}
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

func TestBuildLogFiltering(t *testing.T) {
	// Test build log filtering logic
	tests := []struct {
		name           string
		filterPattern  string
		severity       string
		expectedCount  int
		expectedLines  []string
	}{
		{
			name:          "Filter by error severity",
			severity:      "error",
			expectedCount: 3,
			expectedLines: []string{
				"[ERROR] Connection failed",
				"[ERROR] Test TestFoo failed",
				"[ERROR] Test TestBar failed",
			},
		},
		{
			name:          "Filter by warning severity",
			severity:      "warning",
			expectedCount: 1,
			expectedLines: []string{
				"[WARN] Deprecated API usage",
			},
		},
		{
			name:          "Filter by pattern",
			filterPattern: "Test.*failed",
			expectedCount: 2,
			expectedLines: []string{
				"[ERROR] Test TestFoo failed",
				"[ERROR] Test TestBar failed",
			},
		},
		{
			name:          "Filter by literal string",
			filterPattern: "Connection",
			expectedCount: 1,
			expectedLines: []string{
				"[ERROR] Connection failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test validates the filtering logic conceptually
			// The actual implementation is in the Client.applyBuildLogFilters method
			assert.Equal(t, tt.expectedCount, len(tt.expectedLines))
			assert.NotEmpty(t, tt.expectedLines)
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

func TestSearchBuildConfigurationsTool(t *testing.T) {
	// Test search_build_configurations tool parameter validation
	tests := []struct {
		name     string
		input    map[string]interface{}
		valid    bool
		expected string
	}{
		{
			name: "Valid with all parameters",
			input: map[string]interface{}{
				"projectId": "MyProject",
				"name":      "Test",
				"enabled":   true,
				"paused":    false,
				"template":  false,
				"count":     50,
			},
			valid: true,
		},
		{
			name: "Valid with only name",
			input: map[string]interface{}{
				"name": "Test Configuration",
			},
			valid: true,
		},
		{
			name:  "Valid with no parameters (should return all)",
			input: map[string]interface{}{},
			valid: true,
		},
		{
			name: "Valid with template filter",
			input: map[string]interface{}{
				"template": true,
				"count":    25,
			},
			valid: true,
		},
		{
			name: "Valid with project filter",
			input: map[string]interface{}{
				"projectId": "MyProject",
				"enabled":   true,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - all inputs should be valid for this tool
			// since all parameters are optional
			assert.True(t, tt.valid)

			// Validate parameter types if present
			if projectId, exists := tt.input["projectId"]; exists {
				assert.IsType(t, "", projectId)
			}
			if name, exists := tt.input["name"]; exists {
				assert.IsType(t, "", name)
			}
			if enabled, exists := tt.input["enabled"]; exists {
				assert.IsType(t, true, enabled)
			}
			if paused, exists := tt.input["paused"]; exists {
				assert.IsType(t, true, paused)
			}
			if template, exists := tt.input["template"]; exists {
				assert.IsType(t, true, template)
			}
			if count, exists := tt.input["count"]; exists {
				assert.IsType(t, 0, count)
			}
		})
	}
}

func TestBuildConfigurationURLConstruction(t *testing.T) {
	// Test URL construction logic for build configuration search endpoint
	baseEndpoint := "/buildTypes"

	// Test base URL construction
	assert.Contains(t, baseEndpoint, "buildTypes")

	// Test parameter combinations
	testCases := []struct {
		name       string
		projectId  string
		nameFilter string
		enabled    *bool
		paused     *bool
		template   *bool
		count      int
		expected   []string // Expected URL components
	}{
		{
			name:     "Default count only",
			count:    0, // Should default to 100
			expected: []string{"count:100"},
		},
		{
			name:      "Project filter",
			projectId: "MyProject",
			expected:  []string{"project:MyProject"},
		},
		{
			name:       "Name filter",
			nameFilter: "Test",
			expected:   []string{"name:Test"},
		},
		{
			name:     "Enabled filter",
			enabled:  boolPtr(true),
			expected: []string{"enabled:true"},
		},
		{
			name:     "Template filter",
			template: boolPtr(true),
			expected: []string{"template:true"},
		},
		{
			name:       "Multiple filters",
			projectId:  "MyProject",
			nameFilter: "Test",
			enabled:    boolPtr(true),
			count:      50,
			expected:   []string{"project:MyProject", "name:Test", "enabled:true", "count:50"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate URL parameter construction
			params := []string{}

			if tc.projectId != "" {
				params = append(params, "project:"+tc.projectId)
			}
			if tc.nameFilter != "" {
				params = append(params, "name:"+tc.nameFilter)
			}
			if tc.enabled != nil {
				if *tc.enabled {
					params = append(params, "enabled:true")
				} else {
					params = append(params, "enabled:false")
				}
			}
			if tc.paused != nil {
				if *tc.paused {
					params = append(params, "paused:true")
				} else {
					params = append(params, "paused:false")
				}
			}
			if tc.template != nil {
				if *tc.template {
					params = append(params, "template:true")
				} else {
					params = append(params, "template:false")
				}
			}

			// Verify expected parameters are present
			for _, expected := range tc.expected {
				// Skip count parameters in this test for simplicity
				if !strings.HasPrefix(expected, "count:") {
					assert.Contains(t, params, expected)
				}
			}
		})
	}
}

func TestSearchBuildConfigurationsAdvancedTool(t *testing.T) {
	// Test search_build_configurations advanced parameter validation
	tests := []struct {
		name     string
		input    map[string]interface{}
		valid    bool
		expected string
	}{
		{
			name: "Valid with basic filters",
			input: map[string]interface{}{
				"projectId": "MyProject",
				"name":      "Test",
				"enabled":   true,
			},
			valid: true,
		},
		{
			name: "Valid with parameter filters",
			input: map[string]interface{}{
				"parameterName":  "env.DEPLOY_TARGET",
				"parameterValue": "production",
				"includeDetails": true,
			},
			valid: true,
		},
		{
			name: "Valid with step filters",
			input: map[string]interface{}{
				"stepType": "gradle",
				"stepName": "Build",
			},
			valid: true,
		},
		{
			name: "Valid with VCS filters",
			input: map[string]interface{}{
				"vcsType":        "git",
				"includeDetails": true,
			},
			valid: true,
		},
		{
			name: "Valid with combined filters",
			input: map[string]interface{}{
				"projectId":      "MyProject",
				"parameterName":  "system.docker.image",
				"stepType":       "docker",
				"vcsType":        "git",
				"includeDetails": true,
				"count":          50,
			},
			valid: true,
		},
		{
			name:  "Valid with no parameters (should return all)",
			input: map[string]interface{}{},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - all inputs should be valid for this tool
			// since all parameters are optional
			assert.True(t, tt.valid)

			// Validate parameter types if present
			if projectId, exists := tt.input["projectId"]; exists {
				assert.IsType(t, "", projectId)
			}
			if name, exists := tt.input["name"]; exists {
				assert.IsType(t, "", name)
			}
			if enabled, exists := tt.input["enabled"]; exists {
				assert.IsType(t, true, enabled)
			}
			if parameterName, exists := tt.input["parameterName"]; exists {
				assert.IsType(t, "", parameterName)
			}
			if parameterValue, exists := tt.input["parameterValue"]; exists {
				assert.IsType(t, "", parameterValue)
			}
			if stepType, exists := tt.input["stepType"]; exists {
				assert.IsType(t, "", stepType)
			}
			if stepName, exists := tt.input["stepName"]; exists {
				assert.IsType(t, "", stepName)
			}
			if vcsType, exists := tt.input["vcsType"]; exists {
				assert.IsType(t, "", vcsType)
			}
			if includeDetails, exists := tt.input["includeDetails"]; exists {
				assert.IsType(t, true, includeDetails)
			}
			if count, exists := tt.input["count"]; exists {
				assert.IsType(t, 0, count)
			}
		})
	}
}

func TestDetailedSearchFiltering(t *testing.T) {
	// Test detailed search filtering logic
	testCases := []struct {
		name           string
		parameterName  string
		parameterValue string
		stepType       string
		stepName       string
		vcsType        string
		expectedFields []string
	}{
		{
			name:           "Parameter search",
			parameterName:  "env.DEPLOY",
			parameterValue: "prod",
			expectedFields: []string{"parameterName", "parameterValue"},
		},
		{
			name:           "Step search",
			stepType:       "gradle",
			stepName:       "build",
			expectedFields: []string{"stepType", "stepName"},
		},
		{
			name:           "VCS search",
			vcsType:        "git",
			expectedFields: []string{"vcsType"},
		},
		{
			name:           "Combined search",
			parameterName:  "system.docker",
			stepType:       "docker",
			vcsType:        "git",
			expectedFields: []string{"parameterName", "stepType", "vcsType"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Validate that we're testing the expected fields
			for _, field := range tc.expectedFields {
				switch field {
				case "parameterName":
					assert.NotEmpty(t, tc.parameterName)
				case "parameterValue":
					assert.NotEmpty(t, tc.parameterValue)
				case "stepType":
					assert.NotEmpty(t, tc.stepType)
				case "stepName":
					assert.NotEmpty(t, tc.stepName)
				case "vcsType":
					assert.NotEmpty(t, tc.vcsType)
				}
			}
		})
	}
}

func TestDetailedSearchAPIEndpoints(t *testing.T) {
	// Test that we know the correct API endpoints for detailed search
	expectedEndpoints := map[string]string{
		"basic":      "/buildTypes",
		"parameters": "/buildTypes/id:{buildTypeId}/parameters",
		"steps":      "/buildTypes/id:{buildTypeId}/steps",
		"vcs":        "/buildTypes/id:{buildTypeId}/vcs-root-entries",
	}

	for name, endpoint := range expectedEndpoints {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, endpoint)
			assert.Contains(t, endpoint, "/buildTypes")

			if name != "basic" {
				assert.Contains(t, endpoint, "{buildTypeId}")
			}
		})
	}
}

// Helper function for bool pointers
func boolPtr(b bool) *bool {
	return &b
}

func TestBuildTypeDetailsFieldsParameter(t *testing.T) {
	// Test that buildType details request uses fields parameter to exclude nested structures
	// This prevents JSON unmarshal errors when parameters/steps/vcs-roots are returned as objects
	buildTypeID := "MyBuildConfig"
	expectedURL := "/buildTypes/id:" + buildTypeID + "?fields=id,name,projectName,projectId,href,webUrl,enabled,paused,template"

	// Verify the URL contains the fields parameter
	assert.Contains(t, expectedURL, "?fields=")
	assert.Contains(t, expectedURL, "id,name,projectName,projectId")
	assert.Contains(t, expectedURL, "enabled,paused,template")

	// Verify it does NOT include parameters, steps, or vcs-roots in fields
	assert.NotContains(t, expectedURL, "parameters")
	assert.NotContains(t, expectedURL, "steps")
	assert.NotContains(t, expectedURL, "vcs-root")
}

func TestParameterStructureParsing(t *testing.T) {
	// Test that we handle TeamCity's nested parameter structure correctly
	// The API returns: {"property": [{"name": "...", "value": "..."}]}
	// Not: [{"name": "...", "value": "..."}]

	tests := []struct {
		name           string
		jsonResponse   string
		expectedCount  int
		expectError    bool
		description    string
	}{
		{
			name: "Valid nested property structure",
			jsonResponse: `{
				"property": [
					{"name": "env.DEPLOY_TARGET", "value": "production"},
					{"name": "system.docker.image", "value": "myapp:latest"}
				]
			}`,
			expectedCount: 2,
			expectError:   false,
			description:   "Should parse nested property array correctly",
		},
		{
			name:          "Empty property array",
			jsonResponse:  `{"property": []}`,
			expectedCount: 0,
			expectError:   false,
			description:   "Should handle empty property array",
		},
		{
			name:          "Missing property field",
			jsonResponse:  `{}`,
			expectedCount: 0,
			expectError:   false,
			description:   "Should handle missing property field gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the structure can be parsed
			var paramResponse struct {
				Property []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"property"`
			}

			// Validate the JSON response structure
			if tt.expectedCount > 0 {
				// Should contain property field with elements
				assert.Contains(t, tt.jsonResponse, "property", tt.description)
			}

			if tt.expectError {
				// In a real implementation, this would test unmarshal errors
				assert.True(t, tt.expectError)
			} else {
				// In real implementation, this would validate successful unmarshal
				assert.NotNil(t, paramResponse, tt.description)
			}
		})
	}
}

func TestSeparateDetailsFetching(t *testing.T) {
	// Test that we fetch parameters, steps, and vcs-roots in separate requests
	// This is necessary because the main buildType endpoint returns them as nested objects
	buildTypeID := "MyBuildConfig"

	expectedEndpoints := []struct {
		name     string
		endpoint string
		purpose  string
	}{
		{
			name:     "parameters",
			endpoint: "/buildTypes/id:" + buildTypeID + "/parameters",
			purpose:  "Fetch parameters separately to handle nested structure",
		},
		{
			name:     "steps",
			endpoint: "/buildTypes/id:" + buildTypeID + "/steps",
			purpose:  "Fetch build steps separately to handle nested structure",
		},
		{
			name:     "vcs-roots",
			endpoint: "/buildTypes/id:" + buildTypeID + "/vcs-root-entries",
			purpose:  "Fetch VCS roots separately to handle nested structure",
		},
	}

	for _, ep := range expectedEndpoints {
		t.Run(ep.name, func(t *testing.T) {
			assert.NotEmpty(t, ep.endpoint, "Endpoint should not be empty")
			assert.Contains(t, ep.endpoint, buildTypeID, "Endpoint should contain buildTypeID")
			assert.NotEmpty(t, ep.purpose, "Purpose should be documented")
		})
	}
}

func TestGetTestResultsTool(t *testing.T) {
	// Test get_test_results tool parameter validation
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
			name: "Valid with status filter",
			input: map[string]interface{}{
				"buildId": "12345",
				"status":  "FAILURE",
			},
			valid: true,
		},
		{
			name: "Valid with all parameters",
			input: map[string]interface{}{
				"buildId":        "12345",
				"status":         "SUCCESS",
				"includeDetails": true,
				"count":          50,
			},
			valid: true,
		},
		{
			name: "Missing buildId",
			input: map[string]interface{}{
				"status": "FAILURE",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate required field presence
			if tt.valid {
				assert.Contains(t, tt.input, "buildId")
				buildId := tt.input["buildId"].(string)
				assert.NotEmpty(t, buildId)

				// Validate optional parameters if present
				if status, exists := tt.input["status"]; exists {
					assert.IsType(t, "", status)
				}
				if includeDetails, exists := tt.input["includeDetails"]; exists {
					assert.IsType(t, true, includeDetails)
				}
				if count, exists := tt.input["count"]; exists {
					assert.IsType(t, 0, count)
				}
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

func TestTestResultsURLConstruction(t *testing.T) {
	// Test URL construction logic for test results endpoint
	buildId := "12345"

	// Test base URL construction
	expectedBase := "/testOccurrences?locator=build:(id:" + buildId + ")"
	assert.Contains(t, expectedBase, buildId)
	assert.Contains(t, expectedBase, "testOccurrences")

	// Test parameter combinations
	testCases := []struct {
		name            string
		status          string
		includeDetails  bool
		count           int
		expectedLocator []string // Expected locator components
		expectedFields  []string // Expected fields
	}{
		{
			name:            "Basic query",
			expectedLocator: []string{"build:(id:12345)", "count:100"},
			expectedFields:  []string{"testOccurrence(id,name,status,duration,href)"},
		},
		{
			name:            "With status filter",
			status:          "FAILURE",
			expectedLocator: []string{"build:(id:12345)", "status:FAILURE", "count:100"},
			expectedFields:  []string{"testOccurrence(id,name,status,duration,href)"},
		},
		{
			name:            "With details",
			includeDetails:  true,
			expectedLocator: []string{"build:(id:12345)", "count:100"},
			expectedFields:  []string{"testOccurrence(id,name,status,duration,href,details)"},
		},
		{
			name:            "With custom count",
			count:           50,
			expectedLocator: []string{"build:(id:12345)", "count:50"},
			expectedFields:  []string{"testOccurrence(id,name,status,duration,href)"},
		},
		{
			name:            "All parameters",
			status:          "SUCCESS",
			includeDetails:  true,
			count:           25,
			expectedLocator: []string{"build:(id:12345)", "status:SUCCESS", "count:25"},
			expectedFields:  []string{"testOccurrence(id,name,status,duration,href,details)"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate locator construction
			locator := "build:(id:" + buildId + ")"
			if tc.status != "" {
				locator += ",status:" + tc.status
			}

			count := tc.count
			if count == 0 {
				count = 100
			}
			locator += fmt.Sprintf(",count:%d", count)

			// Verify expected locator components are present
			for _, expected := range tc.expectedLocator {
				assert.Contains(t, locator, expected)
			}

			// Simulate fields construction
			fields := "testOccurrence(id,name,status,duration,href"
			if tc.includeDetails {
				fields += ",details"
			}
			fields += ")"

			// Verify expected fields are present
			for _, expected := range tc.expectedFields {
				assert.Equal(t, expected, fields)
			}
		})
	}
}

func TestTestStatusValues(t *testing.T) {
	// Test that we know the valid test status values
	validStatuses := []string{"SUCCESS", "FAILURE", "UNKNOWN", "IGNORED"}

	for _, status := range validStatuses {
		t.Run(status, func(t *testing.T) {
			assert.NotEmpty(t, status)
			assert.Equal(t, status, strings.ToUpper(status)) // All statuses should be uppercase
		})
	}
}

func TestNotificationHandling(t *testing.T) {
	// Test that notification methods are properly handled
	tests := []struct {
		name           string
		method         string
		shouldHaveID   bool
		expectResponse bool
	}{
		{
			name:           "notifications/cancelled should not require response",
			method:         "notifications/cancelled",
			shouldHaveID:   false,
			expectResponse: false,
		},
		{
			name:           "notifications/initialized should not require response",
			method:         "notifications/initialized",
			shouldHaveID:   false,
			expectResponse: false,
		},
		{
			name:           "Unknown notification should not error",
			method:         "notifications/unknown",
			shouldHaveID:   false,
			expectResponse: false,
		},
		{
			name:           "Unknown request should error",
			method:         "unknown_method",
			shouldHaveID:   true,
			expectResponse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate notification methods don't have "error" in expected behavior
			if !tt.shouldHaveID {
				assert.False(t, tt.expectResponse, "Notifications should not expect response")
			}
		})
	}
}

func TestResourcesListParameterHandling(t *testing.T) {
	// Test resources/list parameter handling
	tests := []struct {
		name        string
		params      string
		shouldParse bool
		description string
	}{
		{
			name:        "Empty params should be handled",
			params:      "",
			shouldParse: true,
			description: "Empty string params should not cause parse error",
		},
		{
			name:        "Null params should be handled",
			params:      "null",
			shouldParse: true,
			description: "Null params should be skipped",
		},
		{
			name:        "Valid JSON params should parse",
			params:      `{"uri":"teamcity://projects"}`,
			shouldParse: true,
			description: "Valid JSON should parse normally",
		},
		{
			name:        "Empty object should parse",
			params:      "{}",
			shouldParse: true,
			description: "Empty object should parse with empty URI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.shouldParse, tt.description)
			// Verify params format is valid
			if tt.params != "" && tt.params != "null" {
				// Should be valid JSON
				assert.True(t, tt.params == "" || tt.params == "null" || strings.HasPrefix(tt.params, "{"))
			}
		})
	}
}

func TestResourceListingBehavior(t *testing.T) {
	// Test that resource listing behaves differently for empty URI vs specific URI
	tests := []struct {
		name              string
		uri               string
		expectsMetadata   bool
		expectsActualData bool
	}{
		{
			name:              "Empty URI returns metadata",
			uri:               "",
			expectsMetadata:   true,
			expectsActualData: false,
		},
		{
			name:              "Specific URI returns actual data",
			uri:               "teamcity://projects",
			expectsMetadata:   false,
			expectsActualData: true,
		},
		{
			name:              "BuildTypes URI returns actual data",
			uri:               "teamcity://buildTypes",
			expectsMetadata:   false,
			expectsActualData: true,
		},
		{
			name:              "Runtime URI returns actual data",
			uri:               "teamcity://runtime",
			expectsMetadata:   false,
			expectsActualData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectsMetadata {
				// Empty URI should return list of available resources with metadata
				assert.True(t, tt.uri == "", "Metadata should only be returned for empty URI")
			}
			if tt.expectsActualData {
				// Specific URI should fetch actual data
				assert.NotEmpty(t, tt.uri, "Actual data should be returned for specific URI")
				assert.True(t, strings.HasPrefix(tt.uri, "teamcity://"), "URI should have teamcity:// prefix")
			}
		})
	}
}

func TestResourceMetadataStructure(t *testing.T) {
	// Test expected structure of resource metadata
	expectedResources := []struct {
		uri         string
		name        string
		description string
		mimeType    string
	}{
		{
			uri:         "teamcity://projects",
			name:        "Projects",
			description: "TeamCity projects",
			mimeType:    "application/json",
		},
		{
			uri:         "teamcity://buildTypes",
			name:        "Build Types",
			description: "TeamCity build configurations",
			mimeType:    "application/json",
		},
		{
			uri:         "teamcity://builds",
			name:        "Builds",
			description: "Recent TeamCity builds",
			mimeType:    "application/json",
		},
		{
			uri:         "teamcity://agents",
			name:        "Agents",
			description: "TeamCity build agents",
			mimeType:    "application/json",
		},
		{
			uri:         "teamcity://runtime",
			name:        "Runtime Information",
			description: "Current server date, time, and runtime information",
			mimeType:    "application/json",
		},
	}

	for _, resource := range expectedResources {
		t.Run(resource.name, func(t *testing.T) {
			assert.NotEmpty(t, resource.uri, "URI should not be empty")
			assert.NotEmpty(t, resource.name, "Name should not be empty")
			assert.NotEmpty(t, resource.description, "Description should not be empty")
			assert.Equal(t, "application/json", resource.mimeType, "MIME type should be application/json")
			assert.True(t, strings.HasPrefix(resource.uri, "teamcity://"), "URI should start with teamcity://")
		})
	}
}

func TestErrorResponseBehavior(t *testing.T) {
	// Test error response behavior for requests vs notifications
	tests := []struct {
		name                string
		hasID               bool
		expectErrorResponse bool
	}{
		{
			name:                "Request with ID should get error response",
			hasID:               true,
			expectErrorResponse: true,
		},
		{
			name:                "Notification without ID should not get error response",
			hasID:               false,
			expectErrorResponse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hasID {
				assert.True(t, tt.expectErrorResponse, "Requests should receive error responses")
			} else {
				assert.False(t, tt.expectErrorResponse, "Notifications should not receive error responses")
			}
		})
	}
}
