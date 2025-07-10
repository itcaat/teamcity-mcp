package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/itcaat/teamcity-mcp/internal/cache"
	"github.com/itcaat/teamcity-mcp/internal/metrics"
	"github.com/itcaat/teamcity-mcp/internal/teamcity"
)

// Handler handles MCP protocol messages
type Handler struct {
	tc     *teamcity.Client
	cache  *cache.Cache
	logger *zap.SugaredLogger
}

// NewHandler creates a new MCP handler
func NewHandler(tc *teamcity.Client, cache *cache.Cache, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		tc:     tc,
		cache:  cache,
		logger: logger,
	}
}

// HandleRequest handles an MCP JSON-RPC request
func (h *Handler) HandleRequest(ctx context.Context, req json.RawMessage) (interface{}, error) {
	start := time.Now()

	// Parse basic JSON-RPC structure
	var baseReq struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      interface{}     `json:"id,omitempty"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params,omitempty"`
	}

	if err := json.Unmarshal(req, &baseReq); err != nil {
		return h.errorResponse(nil, -32700, "Parse error", nil), nil
	}

	// Validate JSON-RPC version
	if baseReq.JSONRPC != "2.0" {
		return h.errorResponse(baseReq.ID, -32600, "Invalid Request", nil), nil
	}

	// Record metrics
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordMCPRequest(baseReq.Method, "success", duration)
	}()

	// Route to appropriate handler
	switch baseReq.Method {
	case "initialize":
		return h.handleInitialize(baseReq.ID, baseReq.Params)
	case "initialized":
		return h.handleInitialized(baseReq.ID)
	case "notifications/initialized":
		return h.handleInitialized(baseReq.ID)
	case "resources/list":
		return h.handleResourcesList(ctx, baseReq.ID, baseReq.Params)
	case "resources/read":
		return h.handleResourcesRead(ctx, baseReq.ID, baseReq.Params)
	case "tools/list":
		return h.handleToolsList(baseReq.ID)
	case "tools/call":
		return h.handleToolsCall(ctx, baseReq.ID, baseReq.Params)
	case "ping":
		return h.handlePing(baseReq.ID)
	default:
		h.logger.Warn("Unknown method called", "method", baseReq.Method, "id", baseReq.ID)
		return h.errorResponse(baseReq.ID, -32601, "Method not found", nil), nil
	}
}

// handleInitialize handles the initialize request
func (h *Handler) handleInitialize(id interface{}, params json.RawMessage) (interface{}, error) {
	return h.successResponse(id, map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"resources": map[string]interface{}{
				"subscribe":   false,
				"listChanged": false,
			},
			"tools":   map[string]interface{}{},
			"logging": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "teamcity-mcp",
			"version": "1.0.0",
		},
	}), nil
}

// handleInitialized handles the initialized notification
func (h *Handler) handleInitialized(id interface{}) (interface{}, error) {
	// Notification - no response needed
	return nil, nil
}

// handleResourcesList handles resources/list requests
func (h *Handler) handleResourcesList(ctx context.Context, id interface{}, params json.RawMessage) (interface{}, error) {
	var req struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return h.errorResponse(id, -32602, "Invalid params", nil), nil
	}

	resources, err := h.listResources(ctx, req.URI)
	if err != nil {
		return h.errorResponse(id, -32603, "Internal error", err.Error()), nil
	}

	return h.successResponse(id, map[string]interface{}{
		"resources": resources,
	}), nil
}

// handleResourcesRead handles resources/read requests
func (h *Handler) handleResourcesRead(ctx context.Context, id interface{}, params json.RawMessage) (interface{}, error) {
	var req struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return h.errorResponse(id, -32602, "Invalid params", nil), nil
	}

	resource, err := h.readResource(ctx, req.URI)
	if err != nil {
		return h.errorResponse(id, -32603, "Internal error", err.Error()), nil
	}

	return h.successResponse(id, map[string]interface{}{
		"contents": []interface{}{resource},
	}), nil
}

// handleToolsList handles tools/list requests
func (h *Handler) handleToolsList(id interface{}) (interface{}, error) {
	tools := []map[string]interface{}{
		{
			"name":        "trigger_build",
			"description": "Trigger a new build",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"buildTypeId": map[string]interface{}{
						"type":        "string",
						"description": "Build configuration ID",
					},
					"branchName": map[string]interface{}{
						"type":        "string",
						"description": "Branch name (optional)",
					},
					"properties": map[string]interface{}{
						"type":        "object",
						"description": "Build properties",
					},
				},
				"required": []string{"buildTypeId"},
			},
		},
		{
			"name":        "cancel_build",
			"description": "Cancel a running build",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"buildId": map[string]interface{}{
						"type":        "string",
						"description": "Build ID to cancel",
					},
					"comment": map[string]interface{}{
						"type":        "string",
						"description": "Cancellation comment",
					},
				},
				"required": []string{"buildId"},
			},
		},
		{
			"name":        "pin_build",
			"description": "Pin or unpin a build",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"buildId": map[string]interface{}{
						"type":        "string",
						"description": "Build ID to pin/unpin",
					},
					"pin": map[string]interface{}{
						"type":        "boolean",
						"description": "True to pin, false to unpin",
					},
					"comment": map[string]interface{}{
						"type":        "string",
						"description": "Pin comment",
					},
				},
				"required": []string{"buildId", "pin"},
			},
		},
		{
			"name":        "set_build_tag",
			"description": "Add or remove build tags",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"buildId": map[string]interface{}{
						"type":        "string",
						"description": "Build ID",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to add",
					},
					"removeTags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to remove",
					},
				},
				"required": []string{"buildId"},
			},
		},
		{
			"name":        "download_artifact",
			"description": "Download build artifacts",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"buildId": map[string]interface{}{
						"type":        "string",
						"description": "Build ID",
					},
					"artifactPath": map[string]interface{}{
						"type":        "string",
						"description": "Artifact path",
					},
				},
				"required": []string{"buildId", "artifactPath"},
			},
		},
		{
			"name":        "search_builds",
			"description": "Search for builds with various filters",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"buildTypeId": map[string]interface{}{
						"type":        "string",
						"description": "Build configuration ID to filter by",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Build status: SUCCESS, FAILURE, ERROR, UNKNOWN",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Build state: queued, running, finished",
					},
					"branch": map[string]interface{}{
						"type":        "string",
						"description": "Branch name to filter by",
					},
					"agent": map[string]interface{}{
						"type":        "string",
						"description": "Agent name to filter by",
					},
					"user": map[string]interface{}{
						"type":        "string",
						"description": "User who triggered the build",
					},
					"sinceBuild": map[string]interface{}{
						"type":        "string",
						"description": "Search builds since this build ID",
					},
					"sinceDate": map[string]interface{}{
						"type":        "string",
						"description": "Search builds since this date (YYYYMMDDTHHMMSS+HHMM)",
					},
					"untilDate": map[string]interface{}{
						"type":        "string",
						"description": "Search builds until this date (YYYYMMDDTHHMMSS+HHMM)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Tags to filter by",
					},
					"personal": map[string]interface{}{
						"type":        "boolean",
						"description": "Include personal builds",
					},
					"pinned": map[string]interface{}{
						"type":        "boolean",
						"description": "Filter by pinned status",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of builds to return (default: 100)",
						"minimum":     1,
						"maximum":     1000,
					},
				},
			},
		},
		{
			"name":        "get_projects",
			"description": "Get all projects or filter by specific criteria",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID to get specific project",
					},
					"project": map[string]interface{}{
						"type":        "string",
						"description": "Project ID or name to filter by",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Project name to filter by",
					},
					"archived": map[string]interface{}{
						"type":        "boolean",
						"description": "Include archived projects",
					},
					"virtual": map[string]interface{}{
						"type":        "boolean",
						"description": "Include virtual projects",
					},
					"build": map[string]interface{}{
						"type":        "string",
						"description": "Build ID to filter by",
					},
					"buildType": map[string]interface{}{
						"type":        "string",
						"description": "Build type ID to filter by",
					},
					"defaultTemplate": map[string]interface{}{
						"type":        "string",
						"description": "Default template ID to filter by",
					},
					"vcsRoot": map[string]interface{}{
						"type":        "string",
						"description": "VCS root ID to filter by",
					},
					"projectFeature": map[string]interface{}{
						"type":        "string",
						"description": "Project feature ID to filter by",
					},
					"pool": map[string]interface{}{
						"type":        "string",
						"description": "Agent pool ID to filter by",
					},
					"start": map[string]interface{}{
						"type":        "integer",
						"description": "Start index for pagination",
						"minimum":     0,
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of projects to return",
						"minimum":     1,
						"maximum":     1000,
					},
				},
			},
		},
	}

	return h.successResponse(id, map[string]interface{}{
		"tools": tools,
	}), nil
}

// handleToolsCall handles tools/call requests
func (h *Handler) handleToolsCall(ctx context.Context, id interface{}, params json.RawMessage) (interface{}, error) {
	var req struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if err := json.Unmarshal(params, &req); err != nil {
		return h.errorResponse(id, -32602, "Invalid params", nil), nil
	}

	result, err := h.callTool(ctx, req.Name, req.Arguments)
	if err != nil {
		return h.errorResponse(id, -32603, "Tool execution failed", err.Error()), nil
	}

	return h.successResponse(id, map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": result,
			},
		},
	}), nil
}

// handlePing handles ping requests
func (h *Handler) handlePing(id interface{}) (interface{}, error) {
	return h.successResponse(id, map[string]interface{}{}), nil
}

// successResponse creates a JSON-RPC success response
func (h *Handler) successResponse(id interface{}, result interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
}

// errorResponse creates a JSON-RPC error response
func (h *Handler) errorResponse(id interface{}, code int, message string, data interface{}) map[string]interface{} {
	error := map[string]interface{}{
		"code":    code,
		"message": message,
	}
	if data != nil {
		error["data"] = data
	}

	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error":   error,
	}
}

// listResources lists available resources
func (h *Handler) listResources(ctx context.Context, uri string) ([]interface{}, error) {
	switch {
	case uri == "teamcity://projects" || uri == "":
		return h.listProjects(ctx)
	case uri == "teamcity://buildTypes":
		return h.listBuildTypes(ctx)
	case uri == "teamcity://builds":
		return h.listBuilds(ctx)
	case uri == "teamcity://agents":
		return h.listAgents(ctx)
	default:
		return nil, fmt.Errorf("unsupported resource URI: %s", uri)
	}
}

// readResource reads a specific resource
func (h *Handler) readResource(ctx context.Context, uri string) (interface{}, error) {
	// Parse URI and delegate to appropriate handler
	return h.tc.GetResource(ctx, uri)
}

// callTool executes a tool
func (h *Handler) callTool(ctx context.Context, name string, args json.RawMessage) (string, error) {
	switch name {
	case "trigger_build":
		return h.tc.TriggerBuild(ctx, args)
	case "cancel_build":
		return h.tc.CancelBuild(ctx, args)
	case "pin_build":
		return h.tc.PinBuild(ctx, args)
	case "set_build_tag":
		return h.tc.SetBuildTag(ctx, args)
	case "download_artifact":
		return h.tc.DownloadArtifact(ctx, args)
	case "search_builds":
		return h.tc.SearchBuilds(ctx, args)
	case "get_projects":
		return h.tc.GetProjects(ctx, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

// Placeholder implementations - to be expanded
func (h *Handler) listProjects(ctx context.Context) ([]interface{}, error) {
	return h.tc.ListProjects(ctx)
}

func (h *Handler) listBuildTypes(ctx context.Context) ([]interface{}, error) {
	return h.tc.ListBuildTypes(ctx)
}

func (h *Handler) listBuilds(ctx context.Context) ([]interface{}, error) {
	return h.tc.ListBuilds(ctx)
}

func (h *Handler) listAgents(ctx context.Context) ([]interface{}, error) {
	return h.tc.ListAgents(ctx)
}
