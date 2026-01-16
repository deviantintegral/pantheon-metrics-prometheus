package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// OrganizationsService handles organization-related operations
type OrganizationsService struct {
	client *Client
}

// NewOrganizationsService creates a new organizations service
func NewOrganizationsService(client *Client) *OrganizationsService {
	return &OrganizationsService{client: client}
}

// List returns all organizations for the authenticated user
func (s *OrganizationsService) List(ctx context.Context, userID string) ([]*models.Organization, error) {
	path := fmt.Sprintf("/users/%s/memberships/organizations", userID)

	rawResults, err := s.client.GetPaged(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	orgs := make([]*models.Organization, 0, len(rawResults))
	for _, raw := range rawResults {
		var membership struct {
			Organization *models.Organization `json:"organization"`
		}
		if err := json.Unmarshal(raw, &membership); err != nil {
			return nil, err
		} else if membership.Organization != nil {
			orgs = append(orgs, membership.Organization)
		}
	}

	return orgs, nil
}

// Get returns a specific organization
func (s *OrganizationsService) Get(ctx context.Context, orgID string) (*models.Organization, error) {
	path := fmt.Sprintf("/organizations/%s", orgID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	var org models.Organization
	if err := DecodeResponse(resp, &org); err != nil {
		return nil, err
	}

	return &org, nil
}

// ListMembers returns members of an organization
func (s *OrganizationsService) ListMembers(ctx context.Context, orgID string) ([]*models.User, error) {
	path := fmt.Sprintf("/organizations/%s/memberships/users", orgID)

	rawResults, err := s.client.GetPaged(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list organization members: %w", err)
	}

	members := make([]*models.User, 0, len(rawResults))
	for _, raw := range rawResults {
		var membership struct {
			User *models.User `json:"user"`
			Role string       `json:"role"`
		}
		if err := json.Unmarshal(raw, &membership); err != nil {
			return nil, fmt.Errorf("failed to decode member: %w", err)
		}
		if membership.User != nil {
			members = append(members, membership.User)
		}
	}

	return members, nil
}

// ListUpstreams returns upstreams for an organization
func (s *OrganizationsService) ListUpstreams(ctx context.Context, orgID string) ([]*models.Upstream, error) {
	path := fmt.Sprintf("/organizations/%s/upstreams", orgID)

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
