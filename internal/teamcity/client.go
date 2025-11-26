package teamcity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/itcaat/teamcity-mcp/internal/config"
	"github.com/itcaat/teamcity-mcp/internal/metrics"
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

// Parameter represents a TeamCity build configuration parameter
type Parameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// BuildStep represents a TeamCity build step
type BuildStep struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Disabled   bool              `json:"disabled"`
	Properties map[string]string `json:"properties,omitempty"`
}

// VCSRoot represents a TeamCity VCS root
type VCSRoot struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	VcsName    string            `json:"vcsName"`
	Properties map[string]string `json:"properties,omitempty"`
}

// DetailedBuildType represents a TeamCity build configuration with detailed information
type DetailedBuildType struct {
	BuildType
	Parameters []Parameter `json:"parameters,omitempty"`
	Steps      []BuildStep `json:"steps,omitempty"`
	VcsRoots   []VCSRoot   `json:"vcs-roots,omitempty"`
	Enabled    bool        `json:"enabled"`
	Paused     bool        `json:"paused"`
	Template   bool        `json:"template"`
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

// SearchBuilds searches for builds with various filters
func (c *Client) SearchBuilds(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildTypeID string   `json:"buildTypeId"`
		Status      string   `json:"status"`
		State       string   `json:"state"`
		Branch      string   `json:"branch"`
		Agent       string   `json:"agent"`
		User        string   `json:"user"`
		SinceBuild  string   `json:"sinceBuild"`
		SinceDate   string   `json:"sinceDate"`
		UntilDate   string   `json:"untilDate"`
		Tags        []string `json:"tags"`
		Personal    *bool    `json:"personal"`
		Pinned      *bool    `json:"pinned"`
		Count       int      `json:"count"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("search_builds", "success", time.Since(start).Seconds())
	}()

	// Build query parameters
	params := make([]string, 0)

	if req.BuildTypeID != "" {
		params = append(params, fmt.Sprintf("buildType:%s", req.BuildTypeID))
	}
	if req.Status != "" {
		params = append(params, fmt.Sprintf("status:%s", req.Status))
	}
	if req.State != "" {
		params = append(params, fmt.Sprintf("state:%s", req.State))
	}
	if req.Branch != "" {
		params = append(params, fmt.Sprintf("branch:%s", req.Branch))
	}
	if req.Agent != "" {
		params = append(params, fmt.Sprintf("agent:%s", req.Agent))
	}
	if req.User != "" {
		params = append(params, fmt.Sprintf("user:%s", req.User))
	}
	if req.SinceBuild != "" {
		params = append(params, fmt.Sprintf("sinceBuild:%s", req.SinceBuild))
	}
	if req.SinceDate != "" {
		params = append(params, fmt.Sprintf("sinceDate:%s", req.SinceDate))
	}
	if req.UntilDate != "" {
		params = append(params, fmt.Sprintf("untilDate:%s", req.UntilDate))
	}
	if req.Personal != nil {
		params = append(params, fmt.Sprintf("personal:%t", *req.Personal))
	}
	if req.Pinned != nil {
		params = append(params, fmt.Sprintf("pinned:%t", *req.Pinned))
	}

	for _, tag := range req.Tags {
		params = append(params, fmt.Sprintf("tag:%s", tag))
	}

	// Set default count if not specified
	count := req.Count
	if count == 0 {
		count = 100
	}

	// Build endpoint with locator
	endpoint := "/builds"
	if len(params) > 0 {
		locator := fmt.Sprintf("count:%d", count)
		for _, param := range params {
			locator += "," + param
		}
		endpoint += "?locator=" + locator
	} else {
		endpoint += fmt.Sprintf("?locator=count:%d", count)
	}

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to search builds: %w", err)
	}

	var response struct {
		Count int     `json:"count"`
		Build []Build `json:"build"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse builds response: %w", err)
	}

	// Format response
	result := fmt.Sprintf("Found %d builds:\n\n", response.Count)
	for _, build := range response.Build {
		result += fmt.Sprintf("Build #%s (ID: %d)\n", build.Number, build.ID)
		result += fmt.Sprintf("  Status: %s\n", build.Status)
		result += fmt.Sprintf("  State: %s\n", build.State)
		result += fmt.Sprintf("  Build Type: %s (%s)\n", build.BuildType.Name, build.BuildTypeID)
		if build.BranchName != "" {
			result += fmt.Sprintf("  Branch: %s\n", build.BranchName)
		}

		// Enhanced time information with duration calculation
		if build.QueuedDate != "" {
			result += fmt.Sprintf("  Queued: %s\n", c.formatTeamCityDate(build.QueuedDate))
		}
		if build.StartDate != "" {
			result += fmt.Sprintf("  Started: %s\n", c.formatTeamCityDate(build.StartDate))
		}
		if build.FinishDate != "" {
			result += fmt.Sprintf("  Finished: %s\n", c.formatTeamCityDate(build.FinishDate))
		}

		// Calculate and display durations
		if build.QueuedDate != "" && build.StartDate != "" {
			if queueTime := c.calculateDuration(build.QueuedDate, build.StartDate); queueTime != "" {
				result += fmt.Sprintf("  Queue Time: %s\n", queueTime)
			}
		}
		if build.StartDate != "" && build.FinishDate != "" {
			if buildTime := c.calculateDuration(build.StartDate, build.FinishDate); buildTime != "" {
				result += fmt.Sprintf("  Build Time: %s\n", buildTime)
			}
		}
		if build.QueuedDate != "" && build.FinishDate != "" {
			if totalTime := c.calculateDuration(build.QueuedDate, build.FinishDate); totalTime != "" {
				result += fmt.Sprintf("  Total Time: %s\n", totalTime)
			}
		}

		result += "\n"
	}

	if response.Count == 0 {
		result = "No builds found matching the specified criteria."
	}

	return result, nil
}

// formatTeamCityDate formats TeamCity date string to a more readable format
func (c *Client) formatTeamCityDate(tcDate string) string {
	// TeamCity format: 20241226T143022+0300
	if tcDate == "" {
		return ""
	}

	// Parse TeamCity date format
	t, err := time.Parse("20060102T150405-0700", tcDate)
	if err != nil {
		// Try alternative format without timezone
		t, err = time.Parse("20060102T150405", tcDate)
		if err != nil {
			// If parsing fails, return original
			return tcDate
		}
	}

	// Return in more readable format
	return t.Format("2006-01-02 15:04:05")
}

// calculateDuration calculates duration between two TeamCity date strings
func (c *Client) calculateDuration(startDate, endDate string) string {
	if startDate == "" || endDate == "" {
		return ""
	}

	// Parse start date
	start, err := time.Parse("20060102T150405-0700", startDate)
	if err != nil {
		start, err = time.Parse("20060102T150405", startDate)
		if err != nil {
			return ""
		}
	}

	// Parse end date
	end, err := time.Parse("20060102T150405-0700", endDate)
	if err != nil {
		end, err = time.Parse("20060102T150405", endDate)
		if err != nil {
			return ""
		}
	}

	duration := end.Sub(start)

	// Format duration in human-readable format
	if duration < 0 {
		return ""
	}

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		if seconds == 0 {
			return fmt.Sprintf("%dm", minutes)
		}
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// FetchBuildLog fetches the build log for a specific build
func (c *Client) FetchBuildLog(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildID       string `json:"buildId"`
		Plain         *bool  `json:"plain,omitempty"`
		Archived      *bool  `json:"archived,omitempty"`
		DateFormat    string `json:"dateFormat,omitempty"`
		MaxLines      *int   `json:"maxLines,omitempty"`
		FilterPattern string `json:"filterPattern,omitempty"`
		Severity      string `json:"severity,omitempty"`
		TailLines     *int   `json:"tailLines,omitempty"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	if req.BuildID == "" {
		return "", fmt.Errorf("buildId is required")
	}

	// Validate severity if provided
	if req.Severity != "" {
		validSeverities := map[string]bool{
			"error":   true,
			"warning": true,
			"info":    true,
		}
		if !validSeverities[strings.ToLower(req.Severity)] {
			return "", fmt.Errorf("invalid severity: must be 'error', 'warning', or 'info'")
		}
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("fetch_build_log", "success", time.Since(start).Seconds())
	}()

	// Build the log download URL with parameters
	endpoint := fmt.Sprintf("/downloadBuildLog.html?buildId=%s", req.BuildID)

	// Add query parameters based on the request
	params := make([]string, 0)

	// Default to plain=true unless explicitly set to false
	plain := true
	if req.Plain != nil {
		plain = *req.Plain
	}
	if plain {
		params = append(params, "plain=true")
	}

	if req.Archived != nil && *req.Archived {
		params = append(params, "archived=true")
	}

	if req.DateFormat != "" {
		// URL encode the date format parameter
		params = append(params, fmt.Sprintf("dateFormat=%s", req.DateFormat))
	}

	// Add parameters to endpoint
	if len(params) > 0 {
		endpoint += "&" + strings.Join(params, "&")
	}

	// Make the request using the custom endpoint (not REST API)
	url := c.baseURL + endpoint

	reqObj, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Set authentication
	if c.cfg.Token != "" {
		reqObj.Header.Set("Authorization", "Bearer "+c.cfg.Token)
	}

	resp, err := c.httpClient.Do(reqObj)
	if err != nil {
		return "", fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	// If archived, we get binary data - indicate this in the response
	if req.Archived != nil && *req.Archived {
		return fmt.Sprintf("Build log for build %s downloaded as archive (%d bytes). Archive content is binary data.",
			req.BuildID, len(respBody)), nil
	}

	// For plain text logs, apply filtering
	logContent := string(respBody)
	lines := strings.Split(logContent, "\n")
	totalLines := len(lines)

	// Apply filters
	filteredLines := c.applyBuildLogFilters(lines, req.FilterPattern, req.Severity)

	// Apply tail if requested
	if req.TailLines != nil && *req.TailLines > 0 {
		tailCount := *req.TailLines
		if tailCount < len(filteredLines) {
			filteredLines = filteredLines[len(filteredLines)-tailCount:]
		}
	}

	// Apply max lines limit
	if req.MaxLines != nil && *req.MaxLines > 0 {
		maxLines := *req.MaxLines
		if maxLines < len(filteredLines) {
			filteredLines = filteredLines[:maxLines]
		}
	}

	// Build result
	result := fmt.Sprintf("Build log for build %s\n", req.BuildID)
	result += fmt.Sprintf("Total lines: %d", totalLines)

	if req.FilterPattern != "" || req.Severity != "" || req.TailLines != nil {
		result += fmt.Sprintf(", Filtered lines: %d", len(filteredLines))
	}

	result += fmt.Sprintf(", Showing: %d lines\n\n", len(filteredLines))

	if len(filteredLines) > 0 {
		result += strings.Join(filteredLines, "\n")
	} else {
		result += "(No lines match the specified filters)"
	}

	return result, nil
}

// applyBuildLogFilters applies pattern and severity filters to log lines
func (c *Client) applyBuildLogFilters(lines []string, pattern string, severity string) []string {
	filtered := lines

	// Apply pattern filter
	if pattern != "" {
		matched := make([]string, 0)
		// Compile regex pattern
		re, err := regexp.Compile(pattern)
		if err != nil {
			// If regex compilation fails, treat as literal string search
			for _, line := range filtered {
				if strings.Contains(line, pattern) {
					matched = append(matched, line)
				}
			}
		} else {
			for _, line := range filtered {
				if re.MatchString(line) {
					matched = append(matched, line)
				}
			}
		}
		filtered = matched
	}

	// Apply severity filter
	if severity != "" {
		matched := make([]string, 0)
		severityLower := strings.ToLower(severity)

		// Common patterns for different severity levels
		errorPatterns := []string{"error", "fail", "exception", "fatal", "[e]", "[error]"}
		warningPatterns := []string{"warn", "warning", "[w]", "[warn]"}

		var patterns []string
		switch severityLower {
		case "error":
			patterns = errorPatterns
		case "warning":
			patterns = warningPatterns
		case "info":
			// For info, we exclude errors and warnings
			for _, line := range filtered {
				lineLower := strings.ToLower(line)
				isErrorOrWarning := false

				for _, p := range append(errorPatterns, warningPatterns...) {
					if strings.Contains(lineLower, p) {
						isErrorOrWarning = true
						break
					}
				}

				if !isErrorOrWarning && strings.TrimSpace(line) != "" {
					matched = append(matched, line)
				}
			}
			filtered = matched
			return filtered
		}

		// For error and warning filters
		for _, line := range filtered {
			lineLower := strings.ToLower(line)
			for _, p := range patterns {
				if strings.Contains(lineLower, p) {
					matched = append(matched, line)
					break
				}
			}
		}
		filtered = matched
	}

	return filtered
}

// SearchBuildConfigurations searches for build configurations with comprehensive filters including parameters, steps, and VCS roots
func (c *Client) SearchBuildConfigurations(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		// Basic filters
		ProjectID string `json:"projectId"`
		Name      string `json:"name"`
		Enabled   *bool  `json:"enabled"`
		Paused    *bool  `json:"paused"`
		Template  *bool  `json:"template"`
		Count     int    `json:"count"`

		// Advanced filters for detailed search
		ParameterName  string `json:"parameterName"`
		ParameterValue string `json:"parameterValue"`
		StepType       string `json:"stepType"`
		StepName       string `json:"stepName"`
		VcsType        string `json:"vcsType"`
		IncludeDetails bool   `json:"includeDetails"` // Whether to fetch detailed info
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("search_build_configurations", "success", time.Since(start).Seconds())
	}()

	// First, get basic build configurations matching basic criteria
	basicConfigs, err := c.getBasicBuildConfigurations(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get basic configurations: %w", err)
	}

	var matchingConfigs []DetailedBuildType

	// For each configuration, check detailed criteria if requested
	for _, config := range basicConfigs {
		if req.IncludeDetails || req.ParameterName != "" || req.ParameterValue != "" ||
			req.StepType != "" || req.StepName != "" || req.VcsType != "" {

			detailed, err := c.getBuildConfigurationDetails(ctx, config.ID)
			if err != nil {
				c.logger.Warn("Failed to get details for build configuration", "id", config.ID, "error", err)
				continue
			}

			// Apply detailed filters
			if c.matchesDetailedCriteria(detailed, req) {
				matchingConfigs = append(matchingConfigs, *detailed)
			}
		} else {
			// If no detailed criteria, just convert basic to detailed
			matchingConfigs = append(matchingConfigs, DetailedBuildType{
				BuildType: config,
			})
		}
	}

	// Format response
	return c.formatDetailedSearchResults(matchingConfigs, req.IncludeDetails), nil
}

// getBasicBuildConfigurations gets configurations using basic filters
func (c *Client) getBasicBuildConfigurations(ctx context.Context, req struct {
	ProjectID      string `json:"projectId"`
	Name           string `json:"name"`
	Enabled        *bool  `json:"enabled"`
	Paused         *bool  `json:"paused"`
	Template       *bool  `json:"template"`
	Count          int    `json:"count"`
	ParameterName  string `json:"parameterName"`
	ParameterValue string `json:"parameterValue"`
	StepType       string `json:"stepType"`
	StepName       string `json:"stepName"`
	VcsType        string `json:"vcsType"`
	IncludeDetails bool   `json:"includeDetails"`
}) ([]BuildType, error) {
	// Build query parameters
	params := make([]string, 0)

	if req.ProjectID != "" {
		params = append(params, fmt.Sprintf("project:%s", req.ProjectID))
	}
	if req.Name != "" {
		params = append(params, fmt.Sprintf("name:%s", req.Name))
	}
	if req.Enabled != nil {
		params = append(params, fmt.Sprintf("enabled:%t", *req.Enabled))
	}
	if req.Paused != nil {
		params = append(params, fmt.Sprintf("paused:%t", *req.Paused))
	}
	if req.Template != nil {
		params = append(params, fmt.Sprintf("template:%t", *req.Template))
	}

	// Set default count if not specified
	count := req.Count
	if count == 0 {
		count = 100
	}

	// Build endpoint with locator
	endpoint := "/buildTypes"
	if len(params) > 0 {
		locator := fmt.Sprintf("count:%d", count)
		for _, param := range params {
			locator += "," + param
		}
		endpoint += "?locator=" + locator
	} else {
		endpoint += fmt.Sprintf("?locator=count:%d", count)
	}

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search build configurations: %w", err)
	}

	var response struct {
		Count     int         `json:"count"`
		BuildType []BuildType `json:"buildType"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse build configurations response: %w", err)
	}

	return response.BuildType, nil
}

// getBuildConfigurationDetails gets detailed information for a specific build configuration
func (c *Client) getBuildConfigurationDetails(ctx context.Context, buildTypeID string) (*DetailedBuildType, error) {
	// Get basic build type info, excluding parameters/steps/vcs-roots since we fetch them separately
	respBody, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/buildTypes/id:%s?fields=id,name,projectName,projectId,href,webUrl,enabled,paused,template", buildTypeID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get build type details: %w", err)
	}

	var buildType DetailedBuildType
	if err := json.Unmarshal(respBody, &buildType); err != nil {
		return nil, fmt.Errorf("failed to parse build type details: %w", err)
	}

	// Get parameters
	paramResp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/buildTypes/id:%s/parameters", buildTypeID), nil)
	if err != nil {
		c.logger.Warn("Failed to get parameters", "buildTypeId", buildTypeID, "error", err)
	} else {
		var paramResponse struct {
			Property []Parameter `json:"property"`
		}
		if err := json.Unmarshal(paramResp, &paramResponse); err == nil {
			buildType.Parameters = paramResponse.Property
		}
	}

	// Get build steps
	stepsResp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/buildTypes/id:%s/steps", buildTypeID), nil)
	if err != nil {
		c.logger.Warn("Failed to get steps", "buildTypeId", buildTypeID, "error", err)
	} else {
		var stepsResponse struct {
			Step []BuildStep `json:"step"`
		}
		if err := json.Unmarshal(stepsResp, &stepsResponse); err == nil {
			buildType.Steps = stepsResponse.Step
		}
	}

	// Get VCS roots
	vcsResp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/buildTypes/id:%s/vcs-root-entries", buildTypeID), nil)
	if err != nil {
		c.logger.Warn("Failed to get VCS roots", "buildTypeId", buildTypeID, "error", err)
	} else {
		var vcsResponse struct {
			VcsRootEntry []struct {
				VcsRoot VCSRoot `json:"vcs-root"`
			} `json:"vcs-root-entry"`
		}
		if err := json.Unmarshal(vcsResp, &vcsResponse); err == nil {
			for _, entry := range vcsResponse.VcsRootEntry {
				buildType.VcsRoots = append(buildType.VcsRoots, entry.VcsRoot)
			}
		}
	}

	return &buildType, nil
}

// matchesDetailedCriteria checks if a configuration matches detailed search criteria
func (c *Client) matchesDetailedCriteria(config *DetailedBuildType, req struct {
	ProjectID      string `json:"projectId"`
	Name           string `json:"name"`
	Enabled        *bool  `json:"enabled"`
	Paused         *bool  `json:"paused"`
	Template       *bool  `json:"template"`
	Count          int    `json:"count"`
	ParameterName  string `json:"parameterName"`
	ParameterValue string `json:"parameterValue"`
	StepType       string `json:"stepType"`
	StepName       string `json:"stepName"`
	VcsType        string `json:"vcsType"`
	IncludeDetails bool   `json:"includeDetails"`
}) bool {
	// Check parameter criteria
	if req.ParameterName != "" || req.ParameterValue != "" {
		paramMatch := false
		for _, param := range config.Parameters {
			nameMatch := req.ParameterName == "" || strings.Contains(strings.ToLower(param.Name), strings.ToLower(req.ParameterName))
			valueMatch := req.ParameterValue == "" || strings.Contains(strings.ToLower(param.Value), strings.ToLower(req.ParameterValue))

			if nameMatch && valueMatch {
				paramMatch = true
				break
			}
		}
		if !paramMatch {
			return false
		}
	}

	// Check step criteria
	if req.StepType != "" || req.StepName != "" {
		stepMatch := false
		for _, step := range config.Steps {
			typeMatch := req.StepType == "" || strings.Contains(strings.ToLower(step.Type), strings.ToLower(req.StepType))
			nameMatch := req.StepName == "" || strings.Contains(strings.ToLower(step.Name), strings.ToLower(req.StepName))

			if typeMatch && nameMatch {
				stepMatch = true
				break
			}
		}
		if !stepMatch {
			return false
		}
	}

	// Check VCS criteria
	if req.VcsType != "" {
		vcsMatch := false
		for _, vcs := range config.VcsRoots {
			if strings.Contains(strings.ToLower(vcs.VcsName), strings.ToLower(req.VcsType)) {
				vcsMatch = true
				break
			}
		}
		if !vcsMatch {
			return false
		}
	}

	return true
}

// formatDetailedSearchResults formats the search results
func (c *Client) formatDetailedSearchResults(configs []DetailedBuildType, includeDetails bool) string {
	if len(configs) == 0 {
		return "No build configurations found matching the specified criteria."
	}

	result := fmt.Sprintf("Found %d build configurations:\n\n", len(configs))

	for _, config := range configs {
		result += fmt.Sprintf("Configuration: %s (%s)\n", config.Name, config.ID)
		result += fmt.Sprintf("  Project: %s (%s)\n", config.Project.Name, config.ProjectID)

		if config.Description != "" {
			result += fmt.Sprintf("  Description: %s\n", config.Description)
		}

		if includeDetails {
			// Add parameters
			if len(config.Parameters) > 0 {
				result += "  Parameters:\n"
				for _, param := range config.Parameters {
					result += fmt.Sprintf("    %s = %s\n", param.Name, param.Value)
				}
			}

			// Add steps
			if len(config.Steps) > 0 {
				result += "  Build Steps:\n"
				for i, step := range config.Steps {
					status := ""
					if step.Disabled {
						status = " (disabled)"
					}
					result += fmt.Sprintf("    %d. %s [%s]%s\n", i+1, step.Name, step.Type, status)
				}
			}

			// Add VCS roots
			if len(config.VcsRoots) > 0 {
				result += "  VCS Roots:\n"
				for _, vcs := range config.VcsRoots {
					result += fmt.Sprintf("    %s (%s)\n", vcs.Name, vcs.VcsName)
				}
			}
		}

		result += "\n"
	}

	return result
}

// GetTestFailures returns failing tests for a specific build
func (c *Client) GetTestFailures(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		BuildID string `json:"buildId"`
	}
	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	if req.BuildID == "" {
		return "", fmt.Errorf("buildId is required")
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("get_test_failures", "success", time.Since(start).Seconds())
	}()

	endpoint := fmt.Sprintf("/testOccurrences?locator=build:(id:%s),status:FAILURE", req.BuildID)
	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get test failures: %w", err)
	}

	var response struct {
		Count          int `json:"count"`
		TestOccurrence []struct {
			Name     string `json:"name"`
			Status   string `json:"status"`
			Duration int    `json:"duration"`
			Message  string `json:"details"`
		} `json:"testOccurrence"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse test failures response: %w", err)
	}

	if response.Count == 0 {
		return "No failing tests found for this build.", nil
	}

	result := fmt.Sprintf("%d failing tests:\n", response.Count)
	for _, test := range response.TestOccurrence {
		result += fmt.Sprintf("- %s (duration: %d ms)", test.Name, test.Duration)
		if test.Message != "" {
			result += fmt.Sprintf(": %s", test.Message)
		}
		result += "\n"
	}
	return result, nil
}
