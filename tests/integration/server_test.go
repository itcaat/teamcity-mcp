package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	serverURL = "http://localhost:8123"
	authToken = "test-token"
)

func TestServerHealth(t *testing.T) {
	// Test liveness
	resp, err := http.Get(serverURL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test readiness
	resp, err = http.Get(serverURL + "/readyz")
	require.NoError(t, err)
	defer resp.Body.Close()
	// May return 503 if TeamCity is not available
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, resp.StatusCode)
}

func TestMCPInitialize(t *testing.T) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	resp := makeRequest(t, req)

	assert.Equal(t, "2.0", resp["jsonrpc"])
	assert.Equal(t, float64(1), resp["id"])

	result, ok := resp["result"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "teamcity-mcp", serverInfo["name"])
}

func TestMCPToolsList(t *testing.T) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}

	resp := makeRequest(t, req)

	result, ok := resp["result"].(map[string]interface{})
	require.True(t, ok)

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok)
	assert.Greater(t, len(tools), 0)

	// Check that trigger_build tool exists
	var triggerBuildFound bool
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		require.True(t, ok)
		if toolMap["name"] == "trigger_build" {
			triggerBuildFound = true
			assert.Contains(t, toolMap, "description")
			assert.Contains(t, toolMap, "inputSchema")
		}
	}
	assert.True(t, triggerBuildFound, "trigger_build tool not found")
}

func TestMCPResourcesList(t *testing.T) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "resources/list",
		"params": map[string]interface{}{
			"uri": "teamcity://projects",
		},
	}

	resp := makeRequest(t, req)

	result, ok := resp["result"].(map[string]interface{})
	require.True(t, ok)

	resources, ok := result["resources"].([]interface{})
	require.True(t, ok)
	// Resources may be empty if TeamCity is not configured
	assert.GreaterOrEqual(t, len(resources), 0)
}

func TestInvalidMethod(t *testing.T) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "invalid/method",
	}

	resp := makeRequest(t, req)

	assert.Equal(t, "2.0", resp["jsonrpc"])
	assert.Equal(t, float64(4), resp["id"])

	errorResp, ok := resp["error"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(-32601), errorResp["code"])
	assert.Equal(t, "Method not found", errorResp["message"])
}

func TestAuthenticationRequired(t *testing.T) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      5,
		"method":  "initialize",
	}

	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	// Make request without auth header
	httpReq, err := http.NewRequest("POST", serverURL+"/mcp", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func makeRequest(t *testing.T, req map[string]interface{}) map[string]interface{} {
	reqBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest("POST", serverURL+"/mcp", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	return response
}
