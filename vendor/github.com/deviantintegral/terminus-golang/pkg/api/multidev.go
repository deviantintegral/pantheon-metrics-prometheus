package api

import (
	"context"
	"fmt"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// MultidevService handles multidev environment operations
type MultidevService struct {
	client *Client
}

// NewMultidevService creates a new multidev service
func NewMultidevService(client *Client) *MultidevService {
	return &MultidevService{client: client}
}

// CreateMultidevRequest represents a multidev creation request
type CreateMultidevRequest struct {
	FromEnvironment               string `json:"from_environment"`
	CloudDevelopmentEnvironmentID string `json:"cloud_development_environment_id"`
}

// Create creates a new multidev environment
func (s *MultidevService) Create(ctx context.Context, siteID, envName, fromEnv string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments", siteID)

	req := map[string]interface{}{
		"type": "create_cloud_development_environment",
		"params": map[string]interface{}{
			"environment_id": envName,
			"deploy": map[string]interface{}{
				"clone_database": map[string]interface{}{
					"from_environment": fromEnv,
				},
				"clone_files": map[string]interface{}{
					"from_environment": fromEnv,
				},
				"annotation": fmt.Sprintf("Create the %q environment.", envName),
			},
		},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to create multidev: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Delete deletes a multidev environment
func (s *MultidevService) Delete(ctx context.Context, siteID, envID string, deleteBranch bool) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/workflows", siteID)

	params := map[string]interface{}{
		"environment_id": envID,
		"delete_branch":  deleteBranch,
	}

	req := map[string]interface{}{
		"type":   "delete_cloud_development_environment",
		"params": params,
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to delete multidev: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// MergeToDev merges a multidev to dev
func (s *MultidevService) MergeToDev(ctx context.Context, siteID, envID string, updateDB bool) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	req := map[string]interface{}{
		"type": "merge_cloud_development_environment_into_dev",
		"params": map[string]interface{}{
			"updatedb": updateDB,
		},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to merge to dev: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// MergeFromDev merges dev into a multidev
func (s *MultidevService) MergeFromDev(ctx context.Context, siteID, envID string, updateDB bool) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/workflows", siteID, envID)

	req := map[string]interface{}{
		"type": "merge_dev_into_cloud_development_environment",
		"params": map[string]interface{}{
			"updatedb": updateDB,
		},
	}

	resp, err := s.client.Post(ctx, path, req) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to merge from dev: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// List returns multidev environments for a site (excludes dev, test, live)
func (s *MultidevService) List(ctx context.Context, siteID string) ([]*models.Environment, error) {
	path := fmt.Sprintf("/sites/%s/environments", siteID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	var allEnvs []*models.Environment
	if err := DecodeResponse(resp, &allEnvs); err != nil {
		return nil, err
	}

	// Filter for multidev environments only (not dev, test, or live)
	multidevs := make([]*models.Environment, 0)
	for _, env := range allEnvs {
		if env.ID != "dev" && env.ID != "test" && env.ID != "live" {
			multidevs = append(multidevs, env)
		}
	}

	return multidevs, nil
}
