package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// UpstreamsService handles upstream-related operations
type UpstreamsService struct {
	client *Client
}

// NewUpstreamsService creates a new upstreams service
func NewUpstreamsService(client *Client) *UpstreamsService {
	return &UpstreamsService{client: client}
}

// ResolveToID converts an upstream identifier (machine name or UUID) to a UUID
func (s *UpstreamsService) ResolveToID(ctx context.Context, upstreamIdentifier, userID string) (string, error) {
	// If it's already a UUID, return it
	if IsUUID(upstreamIdentifier) {
		return upstreamIdentifier, nil
	}

	// Otherwise, search for it by machine name
	upstreams, err := s.List(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to list upstreams: %w", err)
	}

	for _, upstream := range upstreams {
		if upstream.MachineName == upstreamIdentifier {
			return upstream.ID, nil
		}
	}

	return "", fmt.Errorf("upstream not found: %s", upstreamIdentifier)
}

// Get returns a specific upstream by ID
func (s *UpstreamsService) Get(ctx context.Context, upstreamID string) (*models.Upstream, error) {
	path := fmt.Sprintf("/upstreams/%s", upstreamID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream: %w", err)
	}

	var upstream models.Upstream
	if err := DecodeResponse(resp, &upstream); err != nil {
		return nil, err
	}

	return &upstream, nil
}

// List returns all upstreams accessible to the authenticated user
func (s *UpstreamsService) List(ctx context.Context, userID string) ([]*models.Upstream, error) {
	path := fmt.Sprintf("/users/%s/upstreams", userID)

	rawResults, err := s.client.GetPaged(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list upstreams: %w", err)
	}

	upstreams := make([]*models.Upstream, 0, len(rawResults))
	for _, raw := range rawResults {
		var upstream models.Upstream
		if err := json.Unmarshal(raw, &upstream); err != nil {
			return nil, fmt.Errorf("failed to decode upstream: %w", err)
		}
		upstreams = append(upstreams, &upstream)
	}

	return upstreams, nil
}

// ListUpdates returns the list of upstream update commits for a site environment
func (s *UpstreamsService) ListUpdates(ctx context.Context, siteID, envID string) ([]*models.UpstreamUpdateCommit, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/code-upstream-updates", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to list upstream updates: %w", err)
	}

	var updates []*models.UpstreamUpdateCommit
	if err := DecodeResponse(resp, &updates); err != nil {
		return nil, err
	}

	return updates, nil
}
