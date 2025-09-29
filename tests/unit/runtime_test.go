package unit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/itcaat/teamcity-mcp/internal/cache"
	"github.com/itcaat/teamcity-mcp/internal/config"
	"github.com/itcaat/teamcity-mcp/internal/mcp"
	"github.com/itcaat/teamcity-mcp/internal/teamcity"
)

func TestRuntimeFunctionality(t *testing.T) {
	// Create logger
	logger := zaptest.NewLogger(t).Sugar()

	// Create cache
	cacheConfig := config.CacheConfig{TTL: "10s"}
	cache, err := cache.New(cacheConfig)
	require.NoError(t, err)

	// Create TeamCity client (mock is fine for this test)
	tcConfig := config.TeamCityConfig{
		URL:     "http://localhost:8111",
		Token:   "test-token",
		Timeout: "30s",
	}
	tc, err := teamcity.NewClient(tcConfig, logger)
	require.NoError(t, err)

	// Create handler
	handler := mcp.NewHandler(tc, cache, logger)

	t.Run("initialize includes current time", func(t *testing.T) {
		resp, err := handler.HandleRequest(context.Background(), json.RawMessage(`{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "initialize",
			"params": {
				"protocolVersion": "2024-11-05",
				"capabilities": {}
			}
		}`))
		require.NoError(t, err)

		respMap, ok := resp.(map[string]interface{})
		require.True(t, ok)

		result, ok := respMap["result"].(map[string]interface{})
		require.True(t, ok)

		serverInfo, ok := result["serverInfo"].(map[string]interface{})
		require.True(t, ok)

		// Check that current time fields are present
		assert.Contains(t, serverInfo, "currentTime")
		assert.Contains(t, serverInfo, "currentDate")
		assert.Contains(t, serverInfo, "timezone")

		// Validate date format
		currentTime := serverInfo["currentTime"].(string)
		_, err = time.Parse(time.RFC3339, currentTime)
		assert.NoError(t, err)

		currentDate := serverInfo["currentDate"].(string)
		_, err = time.Parse("2006-01-02", currentDate)
		assert.NoError(t, err)
	})

	t.Run("list runtime resource", func(t *testing.T) {
		resp, err := handler.HandleRequest(context.Background(), json.RawMessage(`{
			"jsonrpc": "2.0",
			"id": 2,
			"method": "resources/list",
			"params": {
				"uri": "teamcity://runtime"
			}
		}`))
		require.NoError(t, err)

		respMap, ok := resp.(map[string]interface{})
		require.True(t, ok)

		result, ok := respMap["result"].(map[string]interface{})
		require.True(t, ok)

		resources, ok := result["resources"].([]interface{})
		require.True(t, ok)
		assert.Len(t, resources, 1)

		resource := resources[0].(map[string]interface{})
		assert.Equal(t, "teamcity://runtime", resource["uri"])
		assert.Equal(t, "Runtime Information", resource["name"])
		assert.Contains(t, resource["description"], "Current server date")
	})

	t.Run("read runtime resource", func(t *testing.T) {
		resp, err := handler.HandleRequest(context.Background(), json.RawMessage(`{
			"jsonrpc": "2.0",
			"id": 3,
			"method": "resources/read",
			"params": {
				"uri": "teamcity://runtime"
			}
		}`))
		require.NoError(t, err)

		respMap, ok := resp.(map[string]interface{})
		require.True(t, ok)

		result, ok := respMap["result"].(map[string]interface{})
		require.True(t, ok)

		contents, ok := result["contents"].([]interface{})
		require.True(t, ok)
		assert.Len(t, contents, 1)

		runtimeInfo := contents[0].(map[string]interface{})
		assert.Equal(t, "runtime-info", runtimeInfo["type"])
		assert.Contains(t, runtimeInfo, "currentTime")
		assert.Contains(t, runtimeInfo, "currentDate")
		assert.Contains(t, runtimeInfo, "currentTimestamp")
		assert.Contains(t, runtimeInfo, "utcTime")
		assert.Contains(t, runtimeInfo, "note")

		// Validate the note contains the important message
		note := runtimeInfo["note"].(string)
		assert.Contains(t, note, "REAL current date and time")
		assert.Contains(t, note, "Do not use any training data dates")
	})

	t.Run("get_current_time tool", func(t *testing.T) {
		resp, err := handler.HandleRequest(context.Background(), json.RawMessage(`{
			"jsonrpc": "2.0",
			"id": 4,
			"method": "tools/call",
			"params": {
				"name": "get_current_time",
				"arguments": {
					"format": "date",
					"timezone": "UTC"
				}
			}
		}`))
		require.NoError(t, err)

		respMap, ok := resp.(map[string]interface{})
		require.True(t, ok)

		result, ok := respMap["result"].(map[string]interface{})
		require.True(t, ok)

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		assert.Len(t, content, 1)

		textContent := content[0].(map[string]interface{})
		text := textContent["text"].(string)

		// Check that response contains current time and important note
		assert.Contains(t, text, "Current time:")
		assert.Contains(t, text, "REAL current date/time")
		assert.Contains(t, text, "Use this for all time-based operations")

		// Check that it contains a valid date in the expected format
		lines := strings.Split(text, "\n")
		timeLine := lines[0]
		assert.Contains(t, timeLine, "Current time:")

		// Extract date part and validate format
		dateStr := strings.TrimPrefix(timeLine, "Current time: ")
		_, err = time.Parse("2006-01-02", dateStr)
		assert.NoError(t, err)
	})

	t.Run("get_current_time with default params", func(t *testing.T) {
		resp, err := handler.HandleRequest(context.Background(), json.RawMessage(`{
			"jsonrpc": "2.0",
			"id": 5,
			"method": "tools/call",
			"params": {
				"name": "get_current_time",
				"arguments": {}
			}
		}`))
		require.NoError(t, err)

		respMap, ok := resp.(map[string]interface{})
		require.True(t, ok)

		result, ok := respMap["result"].(map[string]interface{})
		require.True(t, ok)

		content, ok := result["content"].([]interface{})
		require.True(t, ok)
		assert.Len(t, content, 1)

		textContent := content[0].(map[string]interface{})
		text := textContent["text"].(string)

		// Check default format (RFC3339)
		assert.Contains(t, text, "Current time:")
		lines := strings.Split(text, "\n")
		timeLine := lines[0]
		timeStr := strings.TrimPrefix(timeLine, "Current time: ")

		// Should be RFC3339 format by default
		_, err = time.Parse(time.RFC3339, timeStr)
		assert.NoError(t, err)
	})

	t.Run("tools list includes get_current_time", func(t *testing.T) {
		resp, err := handler.HandleRequest(context.Background(), json.RawMessage(`{
			"jsonrpc": "2.0",
			"id": 6,
			"method": "tools/list",
			"params": {}
		}`))
		require.NoError(t, err)

		respMap, ok := resp.(map[string]interface{})
		require.True(t, ok)

		result, ok := respMap["result"].(map[string]interface{})
		require.True(t, ok)

		tools, ok := result["tools"].([]map[string]interface{})
		require.True(t, ok)

		// Find get_current_time tool
		var getCurrentTimeTool map[string]interface{}
		for _, tool := range tools {
			if tool["name"] == "get_current_time" {
				getCurrentTimeTool = tool
				break
			}
		}

		require.NotNil(t, getCurrentTimeTool)
		assert.Equal(t, "get_current_time", getCurrentTimeTool["name"])
		assert.Contains(t, getCurrentTimeTool["description"], "current server date and time")
		assert.Contains(t, getCurrentTimeTool["description"], "real current date/time")
		assert.Contains(t, getCurrentTimeTool, "inputSchema")
	})
}
