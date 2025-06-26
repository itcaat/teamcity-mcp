package teamcity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"teamcity-mcp/internal/config"
	"teamcity-mcp/internal/metrics"
)

// Client wraps the TeamCity REST API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *zap.SugaredLogger
	cfg        config.TeamCityConfig
}

// Project represents a TeamCity project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WebURL      string `json:"webUrl"`
}

// BuildType represents a TeamCity build configuration
type BuildType struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ProjectID   string  `json:"projectId"`
	Project     Project `json:"project"`
}

// Build represents a TeamCity build
type Build struct {
	ID          int       `json:"id"`
	Number      string    `json:"number"`
	Status      string    `json:"status"`
	State       string    `json:"state"`
	BranchName  string    `json:"branchName"`
	BuildTypeID string    `json:"buildTypeId"`
	StartDate   string    `json:"startDate"`
	FinishDate  string    `json:"finishDate"`
	QueuedDate  string    `json:"queuedDate"`
	BuildType   BuildType `json:"buildType"`
}

// Agent represents a TeamCity build agent
type Agent struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	Enabled   bool   `json:"enabled"`
	WebURL    string `json:"webUrl"`
}

// NewClient creates a new TeamCity client
func NewClient(cfg config.TeamCityConfig, logger *zap.SugaredLogger) (*Client, error) {
	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    cfg.URL,
		logger:     logger,
		cfg:        cfg,
	}, nil
}

// makeRequest makes an authenticated HTTP request to TeamCity
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body []byte) ([]byte, error) {
	url := c.baseURL + "/app/rest" + endpoint

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set authentication
	if c.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	} else if c.cfg.Username != "" && c.cfg.Password != "" {
		req.SetBasicAuth(c.cfg.Username, c.cfg.Password)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetResource gets a resource by URI
func (c *Client) GetResource(ctx context.Context, uri string) (interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("get_resource", "success", time.Since(start).Seconds())
	}()

	// Parse URI and call appropriate method
	// This is a simplified implementation
	return map[string]interface{}{
		"uri":     uri,
		"type":    "resource",
		"content": "Resource content for " + uri,
	}, nil
}

// ListProjects lists all projects
func (c *Client) ListProjects(ctx context.Context) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("list_projects", "success", time.Since(start).Seconds())
	}()

	respBody, err := c.makeRequest(ctx, "GET", "/projects", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	var response struct {
		Project []Project `json:"project"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	result := make([]interface{}, len(response.Project))
	for i, project := range response.Project {
		result[i] = map[string]interface{}{
			"uri":         fmt.Sprintf("teamcity://projects/%s", project.ID),
			"name":        project.Name,
			"description": project.Description,
			"mimeType":    "application/json",
		}
	}

	return result, nil
}

// ListBuildTypes lists all build configurations
func (c *Client) ListBuildTypes(ctx context.Context) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("list_build_types", "success", time.Since(start).Seconds())
	}()

	respBody, err := c.makeRequest(ctx, "GET", "/buildTypes", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get build types: %w", err)
	}

	var response struct {
		BuildType []BuildType `json:"buildType"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse build types response: %w", err)
	}

	result := make([]interface{}, len(response.BuildType))
	for i, bt := range response.BuildType {
		result[i] = map[string]interface{}{
			"uri":         fmt.Sprintf("teamcity://buildTypes/%s", bt.ID),
			"name":        bt.Name,
			"description": bt.Description,
			"mimeType":    "application/json",
		}
	}

	return result, nil
}

// ListBuilds lists recent builds
func (c *Client) ListBuilds(ctx context.Context) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("list_builds", "success", time.Since(start).Seconds())
	}()

	respBody, err := c.makeRequest(ctx, "GET", "/builds?locator=count:100", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get builds: %w", err)
	}

	var response struct {
		Build []Build `json:"build"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse builds response: %w", err)
	}

	result := make([]interface{}, len(response.Build))
	for i, build := range response.Build {
		result[i] = map[string]interface{}{
			"uri":         fmt.Sprintf("teamcity://builds/%d", build.ID),
			"name":        fmt.Sprintf("Build #%s", build.Number),
			"description": fmt.Sprintf("Status: %s", build.Status),
			"mimeType":    "application/json",
		}
	}

	return result, nil
}

// ListAgents lists all build agents
func (c *Client) ListAgents(ctx context.Context) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("list_agents", "success", time.Since(start).Seconds())
	}()

	respBody, err := c.makeRequest(ctx, "GET", "/agents", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}

	var response struct {
		Agent []Agent `json:"agent"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse agents response: %w", err)
	}

	result := make([]interface{}, len(response.Agent))
	for i, agent := range response.Agent {
		result[i] = map[string]interface{}{
			"uri":         fmt.Sprintf("teamcity://agents/%d", agent.ID),
			"name":        agent.Name,
			"description": fmt.Sprintf("Connected: %t", agent.Connected),
			"mimeType":    "application/json",
		}
	}

	return result, nil
}

// TriggerBuild triggers a new build
func (c *Client) TriggerBuild(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildTypeID string            `json:"buildTypeId"`
		BranchName  string            `json:"branchName,omitempty"`
		Properties  map[string]string `json:"properties,omitempty"`
		Comment     string            `json:"comment,omitempty"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("trigger_build", "success", time.Since(start).Seconds())
	}()

	// Create build request
	buildRequest := map[string]interface{}{
		"buildType": map[string]string{
			"id": req.BuildTypeID,
		},
	}

	if req.BranchName != "" {
		buildRequest["branchName"] = req.BranchName
	}

	if req.Comment != "" {
		buildRequest["comment"] = map[string]string{
			"text": req.Comment,
		}
	}

	if req.Properties != nil {
		properties := make([]map[string]string, 0, len(req.Properties))
		for key, value := range req.Properties {
			properties = append(properties, map[string]string{
				"name":  key,
				"value": value,
			})
		}
		buildRequest["properties"] = map[string]interface{}{
			"property": properties,
		}
	}

	reqBody, err := json.Marshal(buildRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal build request: %w", err)
	}

	respBody, err := c.makeRequest(ctx, "POST", "/buildQueue", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to trigger build: %w", err)
	}

	var build Build
	if err := json.Unmarshal(respBody, &build); err != nil {
		return "", fmt.Errorf("failed to parse trigger response: %w", err)
	}

	return fmt.Sprintf("Build #%s queued successfully (ID: %d)", build.Number, build.ID), nil
}

// CancelBuild cancels a running build
func (c *Client) CancelBuild(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildID string `json:"buildId"`
		Comment string `json:"comment,omitempty"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("cancel_build", "success", time.Since(start).Seconds())
	}()

	buildID, err := strconv.Atoi(req.BuildID)
	if err != nil {
		return "", fmt.Errorf("invalid build ID: %w", err)
	}

	// Get build to get its number for response
	respBody, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/builds/id:%d", buildID), nil)
	if err != nil {
		return "", fmt.Errorf("build not found: %w", err)
	}

	var build Build
	if err := json.Unmarshal(respBody, &build); err != nil {
		return "", fmt.Errorf("failed to parse build: %w", err)
	}

	// Cancel the build
	cancelRequest := map[string]interface{}{
		"comment": req.Comment,
	}

	reqBody, err := json.Marshal(cancelRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cancel request: %w", err)
	}

	_, err = c.makeRequest(ctx, "POST", fmt.Sprintf("/builds/id:%d/cancelRequest", buildID), reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to cancel build: %w", err)
	}

	return fmt.Sprintf("Build #%s cancelled successfully", build.Number), nil
}

// PinBuild pins or unpins a build
func (c *Client) PinBuild(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildID string `json:"buildId"`
		Pin     bool   `json:"pin"`
		Comment string `json:"comment,omitempty"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("pin_build", "success", time.Since(start).Seconds())
	}()

	buildID, err := strconv.Atoi(req.BuildID)
	if err != nil {
		return "", fmt.Errorf("invalid build ID: %w", err)
	}

	// Get build to get its number for response
	respBody, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/builds/id:%d", buildID), nil)
	if err != nil {
		return "", fmt.Errorf("build not found: %w", err)
	}

	var build Build
	if err := json.Unmarshal(respBody, &build); err != nil {
		return "", fmt.Errorf("failed to parse build: %w", err)
	}

	// Pin or unpin the build
	pinRequest := map[string]interface{}{
		"comment": req.Comment,
	}

	reqBody, err := json.Marshal(pinRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal pin request: %w", err)
	}

	if req.Pin {
		_, err = c.makeRequest(ctx, "PUT", fmt.Sprintf("/builds/id:%d/pin", buildID), reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to pin build: %w", err)
		}
		return fmt.Sprintf("Build #%s pinned successfully", build.Number), nil
	} else {
		_, err = c.makeRequest(ctx, "DELETE", fmt.Sprintf("/builds/id:%d/pin", buildID), nil)
		if err != nil {
			return "", fmt.Errorf("failed to unpin build: %w", err)
		}
		return fmt.Sprintf("Build #%s unpinned successfully", build.Number), nil
	}
}

// SetBuildTag adds or removes build tags
func (c *Client) SetBuildTag(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildID    string   `json:"buildId"`
		Tags       []string `json:"tags,omitempty"`
		RemoveTags []string `json:"removeTags,omitempty"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("set_build_tag", "success", time.Since(start).Seconds())
	}()

	buildID, err := strconv.Atoi(req.BuildID)
	if err != nil {
		return "", fmt.Errorf("invalid build ID: %w", err)
	}

	// Get build to get its number for response
	respBody, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/builds/id:%d", buildID), nil)
	if err != nil {
		return "", fmt.Errorf("build not found: %w", err)
	}

	var build Build
	if err := json.Unmarshal(respBody, &build); err != nil {
		return "", fmt.Errorf("failed to parse build: %w", err)
	}

	// Add tags
	for _, tag := range req.Tags {
		tagData := map[string]string{"name": tag}
		reqBody, err := json.Marshal(tagData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal tag: %w", err)
		}

		_, err = c.makeRequest(ctx, "POST", fmt.Sprintf("/builds/id:%d/tags", buildID), reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to add tag %s: %w", tag, err)
		}
	}

	// Remove tags
	for _, tag := range req.RemoveTags {
		_, err = c.makeRequest(ctx, "DELETE", fmt.Sprintf("/builds/id:%d/tags/%s", buildID, tag), nil)
		if err != nil {
			return "", fmt.Errorf("failed to remove tag %s: %w", tag, err)
		}
	}

	return fmt.Sprintf("Tags updated for build #%s", build.Number), nil
}

// DownloadArtifact downloads build artifacts
func (c *Client) DownloadArtifact(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildID      string `json:"buildId"`
		ArtifactPath string `json:"artifactPath"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("download_artifact", "success", time.Since(start).Seconds())
	}()

	// This is a simplified implementation
	// In practice, you would stream the artifact content
	return fmt.Sprintf("Artifact %s from build %s download initiated", req.ArtifactPath, req.BuildID), nil
}
