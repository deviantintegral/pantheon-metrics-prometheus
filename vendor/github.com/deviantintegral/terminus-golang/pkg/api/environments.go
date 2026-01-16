package api

import (
	"context"
	"fmt"
	"time"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// EnvironmentsService handles environment-related operations
type EnvironmentsService struct {
	client *Client
}

// NewEnvironmentsService creates a new environments service
func NewEnvironmentsService(client *Client) *EnvironmentsService {
	return &EnvironmentsService{client: client}
}

// List returns all environments for a site
func (s *EnvironmentsService) List(ctx context.Context, siteIdentifier string) ([]*models.Environment, error) {
	siteID, err := EnsureSiteUUID(ctx, s.client, siteIdentifier)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/sites/%s/environments", siteID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	// The API returns a map of environment_id -> environment object
	var envsMap map[string]*models.Environment
	if err := DecodeResponse(resp, &envsMap); err != nil {
		return nil, err
	}

	// Convert map to slice
	envs := make([]*models.Environment, 0, len(envsMap))
	for _, env := range envsMap {
		envs = append(envs, env)
	}

	return envs, nil
}

// Get returns a specific environment
func (s *EnvironmentsService) Get(ctx context.Context, siteID, envID string) (*models.Environment, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}

	var env models.Environment
	if err := DecodeResponse(resp, &env); err != nil {
		return nil, err
	}

	return &env, nil
}

// ClearCache clears the cache for an environment
func (s *EnvironmentsService) ClearCache(ctx context.Context, siteID, envID string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	req := map[string]interface{}{
		"type":   "clear_cache",
		"params": map[string]interface{}{},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to clear cache: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// DeployRequest represents a deploy request
type DeployRequest struct {
	UpdateDB   bool   `json:"updatedb,omitempty"`
	Note       string `json:"annotation,omitempty"`
	ClearCache bool   `json:"clear_cache,omitempty"`
}

// Deploy deploys code to an environment
func (s *EnvironmentsService) Deploy(ctx context.Context, siteID, envID string, req *DeployRequest) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	workflowReq := map[string]interface{}{
		"type": "deploy",
		"params": map[string]interface{}{
			"updatedb":    req.UpdateDB,
			"annotation":  req.Note,
			"clear_cache": req.ClearCache,
		},
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to deploy: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// CloneContentRequest represents a clone content request
type CloneContentRequest struct {
	FromEnvironment string `json:"from_environment"`
	Database        bool   `json:"db,omitempty"`
	Files           bool   `json:"files,omitempty"`
}

// CloneContent clones database and/or files from one environment to another
func (s *EnvironmentsService) CloneContent(ctx context.Context, siteID, envID string, req *CloneContentRequest) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	workflowReq := map[string]interface{}{
		"type": "clone_database_files",
		"params": map[string]interface{}{
			"from_environment": req.FromEnvironment,
			"db":               req.Database,
			"files":            req.Files,
		},
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to clone content: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// ChangeConnectionMode changes the connection mode (git or sftp)
func (s *EnvironmentsService) ChangeConnectionMode(ctx context.Context, siteID, envID, mode string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	req := map[string]interface{}{
		"type": "connection_mode_change",
		"params": map[string]interface{}{
			"mode": mode,
		},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to change connection mode: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// CommitRequest represents a commit request
type CommitRequest struct {
	Message string `json:"message"`
}

// Commit commits changes in an environment
func (s *EnvironmentsService) Commit(ctx context.Context, siteID, envID string, req *CommitRequest) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	workflowReq := map[string]interface{}{
		"type": "commit_and_push_on_server_changes",
		"params": map[string]interface{}{
			"message": req.Message,
		},
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Wipe wipes an environment's content
func (s *EnvironmentsService) Wipe(ctx context.Context, siteID, envID string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	req := map[string]interface{}{
		"type":   "wipe",
		"params": map[string]interface{}{},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to wipe environment: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetConnectionInfo returns connection information for an environment
func (s *EnvironmentsService) GetConnectionInfo(ctx context.Context, siteID, envID string) (*models.ConnectionInfo, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/connection-info", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get connection info: %w", err)
	}

	var info models.ConnectionInfo
	if err := DecodeResponse(resp, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// GetUpstreamUpdates returns upstream update information
func (s *EnvironmentsService) GetUpstreamUpdates(ctx context.Context, siteID, envID string) (*models.UpstreamUpdate, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/upstream-updates", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream updates: %w", err)
	}

	var updates models.UpstreamUpdate
	if err := DecodeResponse(resp, &updates); err != nil {
		return nil, err
	}

	return &updates, nil
}

// ApplyUpstreamUpdates applies upstream updates
func (s *EnvironmentsService) ApplyUpstreamUpdates(ctx context.Context, siteID, envID string, updateDB, acceptUpstream bool) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	req := map[string]interface{}{
		"type": "apply_upstream_updates",
		"params": map[string]interface{}{
			"updatedb":        updateDB,
			"accept_upstream": acceptUpstream,
		},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to apply upstream updates: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetLock returns lock information for an environment
func (s *EnvironmentsService) GetLock(ctx context.Context, siteID, envID string) (*models.Lock, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/lock", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get lock: %w", err)
	}

	var lock models.Lock
	if err := DecodeResponse(resp, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// SetLock sets or updates the lock for an environment
func (s *EnvironmentsService) SetLock(ctx context.Context, siteID, envID, username, password string) error {
	path := fmt.Sprintf("/sites/%s/environments/%s/lock", siteID, envID)

	req := map[string]interface{}{
		"username": username,
		"password": password,
	}

	resp, err := s.client.Put(ctx, path, req)
	if err != nil {
		return fmt.Errorf("failed to set lock: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("set lock failed with status %d", resp.StatusCode)
	}

	return nil
}

// RemoveLock removes the lock from an environment
func (s *EnvironmentsService) RemoveLock(ctx context.Context, siteID, envID string) error {
	path := fmt.Sprintf("/sites/%s/environments/%s/lock", siteID, envID)
	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to remove lock: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("remove lock failed with status %d", resp.StatusCode)
	}

	return nil
}

// metricsResponse represents the API response for traffic metrics
type metricsResponse struct {
	Timeseries []metricsDatapoint `json:"timeseries"`
}

// metricsDatapoint represents a single data point from the API
type metricsDatapoint struct {
	Timestamp   int64 `json:"timestamp"`
	Visits      int64 `json:"visits"`
	PagesServed int64 `json:"pages_served"`
	CacheHits   int64 `json:"cache_hits"`
	CacheMisses int64 `json:"cache_misses"`
}

// GetMetrics returns traffic metrics for an environment
func (s *EnvironmentsService) GetMetrics(ctx context.Context, siteIdentifier, envID, duration string) ([]*models.Metrics, error) {
	siteID, err := EnsureSiteUUID(ctx, s.client, siteIdentifier)
	if err != nil {
		return nil, err
	}

	var path string
	if envID == "" {
		// Site-level metrics (all environments combined)
		path = fmt.Sprintf("/sites/%s/traffic?duration=%s", siteID, duration)
	} else {
		// Environment-level metrics
		path = fmt.Sprintf("/sites/%s/environments/%s/traffic?duration=%s", siteID, envID, duration)
	}

	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	var metricsResp metricsResponse
	if err := DecodeResponse(resp, &metricsResp); err != nil {
		return nil, err
	}

	// Convert API response to Metrics model
	metrics := make([]*models.Metrics, 0, len(metricsResp.Timeseries))
	for _, dp := range metricsResp.Timeseries {
		// Convert timestamp to ISO 8601 datetime format (matching PHP terminus)
		datetime := formatTimestampISO8601(dp.Timestamp)

		// Calculate cache hit ratio as string with 2 decimal places
		// Show "--" when pages_served is 0
		var cacheHitRatio string
		if dp.PagesServed > 0 {
			ratio := float64(dp.CacheHits) / float64(dp.PagesServed)
			cacheHitRatio = fmt.Sprintf("%.2f%%", ratio*100)
		} else {
			cacheHitRatio = "--"
		}

		metrics = append(metrics, &models.Metrics{
			Timestamp:     dp.Timestamp,
			Datetime:      datetime,
			Visits:        dp.Visits,
			PagesServed:   dp.PagesServed,
			CacheHits:     dp.CacheHits,
			CacheMisses:   dp.CacheMisses,
			CacheHitRatio: cacheHitRatio,
		})
	}

	return metrics, nil
}

// formatTimestampISO8601 converts a Unix timestamp to ISO 8601 datetime format
func formatTimestampISO8601(ts int64) string {
	t := time.Unix(ts, 0).UTC()
	return t.Format("2006-01-02T15:04:05")
}
