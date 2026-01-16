package api

import (
	"context"
	"fmt"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// DomainsService handles domain-related operations
type DomainsService struct {
	client *Client
}

// NewDomainsService creates a new domains service
func NewDomainsService(client *Client) *DomainsService {
	return &DomainsService{client: client}
}

// List returns all domains for an environment
func (s *DomainsService) List(ctx context.Context, siteID, envID string) ([]*models.Domain, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/domains", siteID, envID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	var domains []*models.Domain
	if err := DecodeResponse(resp, &domains); err != nil {
		return nil, err
	}

	return domains, nil
}

// Get returns a specific domain
func (s *DomainsService) Get(ctx context.Context, siteID, envID, domainID string) (*models.Domain, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/domains/%s", siteID, envID, domainID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get domain: %w", err)
	}

	var domain models.Domain
	if err := DecodeResponse(resp, &domain); err != nil {
		return nil, err
	}

	return &domain, nil
}

// Add adds a domain to an environment
func (s *DomainsService) Add(ctx context.Context, siteID, envID, domain string) (*models.Domain, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/domains/%s", siteID, envID, domain)

	resp, err := s.client.Put(ctx, path, nil) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to add domain: %w", err)
	}

	var addedDomain models.Domain
	if err := DecodeResponse(resp, &addedDomain); err != nil {
		return nil, err
	}

	return &addedDomain, nil
}

// Remove removes a domain from an environment
func (s *DomainsService) Remove(ctx context.Context, siteID, envID, domainID string) error {
	path := fmt.Sprintf("/sites/%s/environments/%s/domains/%s", siteID, envID, domainID)
	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to remove domain: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("remove domain failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetDNS returns DNS recommendations for a domain
func (s *DomainsService) GetDNS(ctx context.Context, siteID, envID, domainID string) ([]*models.DNSRecord, error) {
	path := fmt.Sprintf("/sites/%s/environments/%s/domains/%s/dns", siteID, envID, domainID)
	resp, err := s.client.Get(ctx, path) //nolint:bodyclose // DecodeResponse closes body
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %w", err)
	}

	var records []*models.DNSRecord
	if err := DecodeResponse(resp, &records); err != nil {
		return nil, err
	}

	return records, nil
}
