package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// BackupsService handles backup-related operations
type BackupsService struct {
	client *Client
}

// NewBackupsService creates a new backups service
func NewBackupsService(client *Client) *BackupsService {
	return &BackupsService{client: client}
}

// List returns all backups for an environment
func (s *BackupsService) List(ctx context.Context, siteID, envID string) ([]*models.Backup, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/backups/catalog", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	var backups []*models.Backup
	if err := DecodeResponse(resp, &backups); err != nil {
		return nil, err
	}

	return backups, nil
}

// Get returns a specific backup
func (s *BackupsService) Get(ctx context.Context, siteID, envID, backupID string) (*models.Backup, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/backups/catalog/%s", siteID, envID, backupID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get backup: %w", err)
	}

	var backup models.Backup
	if err := DecodeResponse(resp, &backup); err != nil {
		return nil, err
	}

	return &backup, nil
}

// CreateBackupRequest represents a backup creation request
type CreateBackupRequest struct {
	KeepFor int `json:"ttl,omitempty"` // TTL in days
}

// Create creates a new backup
func (s *BackupsService) Create(ctx context.Context, siteID, envID string, req *CreateBackupRequest) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	params := map[string]interface{}{
		"code":       true,
		"database":   true,
		"files":      true,
		"entry_type": "backup",
	}
	if req != nil && req.KeepFor > 0 {
		// Convert days to seconds
		params["ttl"] = req.KeepFor * 86400
	}

	workflowReq := map[string]interface{}{
		"type":   "do_export",
		"params": params,
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// CreateElement creates a backup of a specific element (code, database, files)
func (s *BackupsService) CreateElement(ctx context.Context, siteID, envID, element string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	params := map[string]interface{}{
		"code":       element == "code",
		"database":   element == "database",
		"files":      element == "files",
		"entry_type": "backup",
	}

	workflowReq := map[string]interface{}{
		"type":   "do_export",
		"params": params,
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to create %s backup: %w", element, err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetDownloadURL returns the download URL for a backup element
func (s *BackupsService) GetDownloadURL(ctx context.Context, siteID, envID, backupID, element string) (string, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/backups/catalog/%s/downloads/%s", siteID, envID, backupID, element)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := DecodeResponse(resp, &result); err != nil {
		return "", err
	}

	return result.URL, nil
}

// Download downloads a backup element to a file
func (s *BackupsService) Download(ctx context.Context, siteID, envID, backupID, element, outputPath string) error {
	// Get download URL
	downloadURL, err := s.GetDownloadURL(ctx, siteID, envID, backupID, element)
	if err != nil {
		return err
	}

	// Download file
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download backup: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create output file
	out, err := os.Create(outputPath) //nolint:gosec // User-specified output path
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = out.Close() }()

	// Copy data
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write backup data: %w", err)
	}

	return nil
}

// Restore restores a backup
func (s *BackupsService) Restore(ctx context.Context, siteID, envID, backupID string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	workflowReq := map[string]interface{}{
		"type": "restore",
		"params": map[string]interface{}{
			"backup_id": backupID,
		},
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to restore backup: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// GetSchedule returns the backup schedule for an environment
func (s *BackupsService) GetSchedule(ctx context.Context, siteID, envID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/backups/schedule", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get backup schedule: %w", err)
	}

	var schedule map[string]interface{}
	if err := DecodeResponse(resp, &schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

// SetSchedule sets the backup schedule for an environment
func (s *BackupsService) SetSchedule(ctx context.Context, siteID, envID string, enabled bool, day int) error {
	path := fmt.Sprintf("/sites/%s/environments/%s/backups/schedule", siteID, envID)

	req := map[string]interface{}{
		"enabled": enabled,
	}
	if enabled && day > 0 {
		req["day"] = day
	}

	resp, err := s.client.Put(ctx, path, req)
	if err != nil {
		return fmt.Errorf("failed to set backup schedule: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("set backup schedule failed with status %d", resp.StatusCode)
	}

	return nil
}
