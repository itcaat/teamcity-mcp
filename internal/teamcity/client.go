package teamcity

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		BuildTypeID     string            `json:"buildTypeId"`
		Status          string            `json:"status"`
		State           string            `json:"state"`
		Branch          string            `json:"branch"`
		Agent           string            `json:"agent"`
		User            string            `json:"user"`
		SinceBuild      string            `json:"sinceBuild"`
		SinceDate       string            `json:"sinceDate"`
		UntilDate       string            `json:"untilDate"`
		Tags            []string          `json:"tags"`
		Personal        *bool             `json:"personal"`
		Pinned          *bool             `json:"pinned"`
		Count           int               `json:"count"`
		Project         string            `json:"project"`
		Number          string            `json:"number"`
		Hanging         *bool             `json:"hanging"`
		Canceled        *bool             `json:"canceled"`
		QueuedDate      string            `json:"queuedDate"`
		StartDate       string            `json:"startDate"`
		FinishDate      string            `json:"finishDate"`
		FailedToStart   *bool             `json:"failedToStart"`
		Composite       *bool             `json:"composite"`
		Tag             string            `json:"tag"`
		Property        map[string]string `json:"property"`
		CompatibleAgent string            `json:"compatibleAgent"`
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
	if req.Project != "" {
		params = append(params, fmt.Sprintf("project:%s", req.Project))
	}
	if req.Number != "" {
		params = append(params, fmt.Sprintf("number:%s", req.Number))
	}
	if req.Hanging != nil {
		params = append(params, fmt.Sprintf("hanging:%t", *req.Hanging))
	}
	if req.Canceled != nil {
		params = append(params, fmt.Sprintf("canceled:%t", *req.Canceled))
	}
	if req.QueuedDate != "" {
		params = append(params, fmt.Sprintf("queuedDate:%s", req.QueuedDate))
	}
	if req.StartDate != "" {
		params = append(params, fmt.Sprintf("startDate:%s", req.StartDate))
	}
	if req.FinishDate != "" {
		params = append(params, fmt.Sprintf("finishDate:%s", req.FinishDate))
	}
	if req.FailedToStart != nil {
		params = append(params, fmt.Sprintf("failedToStart:%t", *req.FailedToStart))
	}
	if req.Composite != nil {
		params = append(params, fmt.Sprintf("composite:%t", *req.Composite))
	}
	if req.Tag != "" {
		params = append(params, fmt.Sprintf("tag:%s", req.Tag))
	}
	if req.CompatibleAgent != "" {
		params = append(params, fmt.Sprintf("compatibleAgent:%s", req.CompatibleAgent))
	}

	for _, tag := range req.Tags {
		params = append(params, fmt.Sprintf("tag:%s", tag))
	}

	for key, value := range req.Property {
		params = append(params, fmt.Sprintf("property:%s:%s", key, value))
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
		if build.StartDate != "" {
			result += fmt.Sprintf("  Started: %s\n", build.StartDate)
		}
		if build.FinishDate != "" {
			result += fmt.Sprintf("  Finished: %s\n", build.FinishDate)
		}
		result += "\n"
	}

	if response.Count == 0 {
		result = "No builds found matching the specified criteria."
	}

	return result, nil
}

// GetProjects gets projects with optional filters
func (c *Client) GetProjects(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		ID              string `json:"id"`
		Project         string `json:"project"`
		Name            string `json:"name"`
		Archived        *bool  `json:"archived"`
		Virtual         *bool  `json:"virtual"`
		Build           string `json:"build"`
		BuildType       string `json:"buildType"`
		DefaultTemplate string `json:"defaultTemplate"`
		VcsRoot         string `json:"vcsRoot"`
		ProjectFeature  string `json:"projectFeature"`
		Pool            string `json:"pool"`
		Start           *int   `json:"start"`
		Count           *int   `json:"count"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("get_projects", "success", time.Since(start).Seconds())
	}()

	// Build endpoint
	endpoint := "/projects"
	if req.ID != "" {
		endpoint = fmt.Sprintf("/projects/id:%s", req.ID)
	}

	// Build locator for filtering
	var locatorParts []string

	if req.Project != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("project:%s", req.Project))
	}
	if req.Name != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("name:%s", req.Name))
	}
	if req.Archived != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("archived:%t", *req.Archived))
	}
	if req.Virtual != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("virtual:%t", *req.Virtual))
	}
	if req.Build != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("build:%s", req.Build))
	}
	if req.BuildType != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("buildType:%s", req.BuildType))
	}
	if req.DefaultTemplate != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("defaultTemplate:%s", req.DefaultTemplate))
	}
	if req.VcsRoot != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("vcsRoot:%s", req.VcsRoot))
	}
	if req.ProjectFeature != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("projectFeature:%s", req.ProjectFeature))
	}
	if req.Pool != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("pool:%s", req.Pool))
	}
	if req.Start != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("start:%d", *req.Start))
	}
	if req.Count != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("count:%d", *req.Count))
	}

	// Add locator parameter if we have filters and not getting a specific project by ID
	if len(locatorParts) > 0 && req.ID == "" {
		locator := strings.Join(locatorParts, ",")
		endpoint += fmt.Sprintf("?locator=%s", locator)
	}

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get projects: %w", err)
	}

	// If getting a specific project
	if req.ID != "" {
		var project Project
		if err := json.Unmarshal(respBody, &project); err != nil {
			return "", fmt.Errorf("failed to parse project response: %w", err)
		}

		result := fmt.Sprintf("Project: %s\n", project.Name)
		result += fmt.Sprintf("ID: %s\n", project.ID)
		if project.Description != "" {
			result += fmt.Sprintf("Description: %s\n", project.Description)
		}
		if project.WebURL != "" {
			result += fmt.Sprintf("Web URL: %s\n", project.WebURL)
		}

		return result, nil
	}

	// Getting all projects
	var response struct {
		Count   int       `json:"count"`
		Project []Project `json:"project"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse projects response: %w", err)
	}

	// Format response
	result := fmt.Sprintf("Found %d projects:\n\n", response.Count)
	for _, project := range response.Project {
		result += fmt.Sprintf("Project: %s\n", project.Name)
		result += fmt.Sprintf("  ID: %s\n", project.ID)
		if project.Description != "" {
			result += fmt.Sprintf("  Description: %s\n", project.Description)
		}
		if project.WebURL != "" {
			result += fmt.Sprintf("  Web URL: %s\n", project.WebURL)
		}
		result += "\n"
	}

	if response.Count == 0 {
		result = "No projects found."
	}

	return result, nil
}

// GetBuildTypes gets build types with optional filters
func (c *Client) GetBuildTypes(ctx context.Context, args json.RawMessage) (string, error) {
	var req struct {
		ID              string `json:"id"`
		Project         string `json:"project"`
		AffectedProject string `json:"affectedProject"`
		Name            string `json:"name"`
		TemplateFlag    *bool  `json:"templateFlag"`
		Template        string `json:"template"`
		Paused          *bool  `json:"paused"`
		VcsRoot         string `json:"vcsRoot"`
		VcsRootInstance string `json:"vcsRootInstance"`
		Build           string `json:"build"`
		Start           *int   `json:"start"`
		Count           *int   `json:"count"`
	}

	if err := json.Unmarshal(args, &req); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	start := time.Now()
	defer func() {
		metrics.RecordTeamCityRequest("get_build_types", "success", time.Since(start).Seconds())
	}()

	// Build endpoint
	endpoint := "/buildTypes"
	if req.ID != "" {
		endpoint = fmt.Sprintf("/buildTypes/id:%s", req.ID)
	}

	// Build locator for filtering
	var locatorParts []string

	if req.Project != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("project:%s", req.Project))
	}
	if req.AffectedProject != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("affectedProject:%s", req.AffectedProject))
	}
	if req.Name != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("name:%s", req.Name))
	}
	if req.TemplateFlag != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("templateFlag:%t", *req.TemplateFlag))
	}
	if req.Template != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("template:%s", req.Template))
	}
	if req.Paused != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("paused:%t", *req.Paused))
	}
	if req.VcsRoot != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("vcsRoot:%s", req.VcsRoot))
	}
	if req.VcsRootInstance != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("vcsRootInstance:%s", req.VcsRootInstance))
	}
	if req.Build != "" {
		locatorParts = append(locatorParts, fmt.Sprintf("build:%s", req.Build))
	}
	if req.Start != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("start:%d", *req.Start))
	}
	if req.Count != nil {
		locatorParts = append(locatorParts, fmt.Sprintf("count:%d", *req.Count))
	}

	// Add locator parameter if we have filters and not getting a specific build type by ID
	if len(locatorParts) > 0 && req.ID == "" {
		locator := strings.Join(locatorParts, ",")
		endpoint += fmt.Sprintf("?locator=%s", locator)
	}

	respBody, err := c.makeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get build types: %w", err)
	}

	// If getting a specific build type
	if req.ID != "" {
		var buildType BuildType
		if err := json.Unmarshal(respBody, &buildType); err != nil {
			return "", fmt.Errorf("failed to parse build type response: %w", err)
		}

		result := fmt.Sprintf("Build Type: %s\n", buildType.Name)
		result += fmt.Sprintf("ID: %s\n", buildType.ID)
		if buildType.Description != "" {
			result += fmt.Sprintf("Description: %s\n", buildType.Description)
		}
		result += fmt.Sprintf("Project ID: %s\n", buildType.ProjectID)
		if buildType.Project.Name != "" {
			result += fmt.Sprintf("Project Name: %s\n", buildType.Project.Name)
		}

		return result, nil
	}

	// Getting all build types
	var response struct {
		Count     int         `json:"count"`
		BuildType []BuildType `json:"buildType"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("failed to parse build types response: %w", err)
	}

	// Format response
	result := fmt.Sprintf("Found %d build types:\n\n", response.Count)
	for _, buildType := range response.BuildType {
		result += fmt.Sprintf("Build Type: %s\n", buildType.Name)
		result += fmt.Sprintf("  ID: %s\n", buildType.ID)
		if buildType.Description != "" {
			result += fmt.Sprintf("  Description: %s\n", buildType.Description)
		}
		result += fmt.Sprintf("  Project: %s (%s)\n", buildType.Project.Name, buildType.ProjectID)
		result += "\n"
	}

	if response.Count == 0 {
		result = "No build types found."
	}

	return result, nil
}
