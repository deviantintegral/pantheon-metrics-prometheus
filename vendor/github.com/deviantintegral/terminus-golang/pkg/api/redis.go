package api

import (
	"context"
	"fmt"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// RedisService handles Redis-related operations
type RedisService struct {
	client *Client
}

// NewRedisService creates a new Redis service
func NewRedisService(client *Client) *RedisService {
	return &RedisService{client: client}
}

// Enable enables Redis for a site
func (s *RedisService) Enable(ctx context.Context, siteID string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/workflows", siteID)

	workflowReq := map[string]interface{}{
		"type": "enable_addon",
		"params": map[string]interface{}{
			"addon": "cacheserver",
		},
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to enable Redis: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// Disable disables Redis for a site
func (s *RedisService) Disable(ctx context.Context, siteID string) (*models.Workflow, error) {
	path := fmt.Sprintf("/sites/%s/workflows", siteID)

	workflowReq := map[string]interface{}{
		"type": "disable_addon",
		"params": map[string]interface{}{
			"addon": "cacheserver",
		},
	}

	resp, err := s.client.Post(ctx, path, workflowReq) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to disable Redis: %w", err)
	}

	var workflow models.Workflow
	if err := DecodeResponse(resp, &workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}
