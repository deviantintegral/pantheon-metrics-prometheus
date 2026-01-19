// Package pantheon provides types and client functions for interacting with Pantheon via the terminus-golang library.
package pantheon

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/deviantintegral/terminus-golang/pkg/api"
)

// Client wraps the terminus-golang library for Pantheon API access.
type Client struct {
	sessionManager *SessionManager
}

// NewClient creates a new Pantheon API client.
func NewClient() *Client {
	return &Client{
		sessionManager: NewSessionManager(),
	}
}

// GetAccountID returns an account identifier from a machine token (last 8 chars).
// This is used as a fallback when email retrieval fails.
func GetAccountID(token string) string {
	// Return last 8 characters of token for identification
	if len(token) >= 8 {
		return token[len(token)-8:]
	}
	return token
}

// Authenticate authenticates with a machine token and returns the account email.
func (c *Client) Authenticate(ctx context.Context, machineToken string) (string, error) {
	log.Printf("Authenticating with machine token...")
	session, err := c.sessionManager.Authenticate(ctx, machineToken)
	if err != nil {
		return "", err
	}
	return session.Email, nil
}

// GetEmail returns the email for the given machine token (cached from session).
func (c *Client) GetEmail(ctx context.Context, machineToken string) (string, error) {
	return c.sessionManager.GetEmail(ctx, machineToken)
}

// FetchAllSites fetches all sites accessible to the user, including:
// 1. Sites from direct user memberships
// 2. Sites from all organizations the user is a member of
func (c *Client) FetchAllSites(ctx context.Context, machineToken string) (map[string]SiteListEntry, error) {
	log.Printf("Fetching all sites from Pantheon API...")

	session, err := c.sessionManager.GetSession(ctx, machineToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	sitesService := api.NewSitesService(session.Client)
	orgsService := api.NewOrganizationsService(session.Client)

	// Track unique sites by ID to avoid duplicates
	siteMap := make(map[string]SiteListEntry)

	// 1. Get sites from direct user memberships
	userSites, err := sitesService.List(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user sites: %w", err)
	}
	for _, site := range userSites {
		siteMap[site.ID] = ConvertSite(site)
	}
	log.Printf("Found %d sites from direct user memberships", len(userSites))

	// 2. Get user's organization memberships
	orgs, err := orgsService.List(ctx, session.UserID)
	if err != nil {
		// Log warning but continue - user sites were already fetched
		log.Printf("Warning: failed to list user organizations: %v", err)
	} else {
		log.Printf("Found %d organizations", len(orgs))

		// 3. For each organization, get all sites
		for _, org := range orgs {
			orgSites, err := sitesService.ListByOrganization(ctx, org.ID)
			if err != nil {
				// Continue on error to get sites from other orgs
				orgName := org.ID
				if org.Label != "" {
					orgName = org.Label
				}
				log.Printf("Warning: failed to list sites for organization %s: %v", orgName, err)
				continue
			}

			// Add org sites to the map (deduplicating by ID)
			// Don't overwrite if site already exists (to preserve direct team membership info)
			orgSiteCount := 0
			for _, site := range orgSites {
				if _, exists := siteMap[site.ID]; !exists {
					siteMap[site.ID] = ConvertSite(site)
					orgSiteCount++
				}
			}
			if orgSiteCount > 0 {
				orgName := org.ID
				if org.Label != "" {
					orgName = org.Label
				}
				log.Printf("Found %d additional sites from organization %s", orgSiteCount, orgName)
			}
		}
	}

	log.Printf("Total unique sites found: %d", len(siteMap))
	return siteMap, nil
}

// FetchMetricsData fetches metrics data for a site.
// duration should be "28d" for initial fetch or "1d" for subsequent refreshes.
func (c *Client) FetchMetricsData(ctx context.Context, machineToken, siteID, environment, duration string) (map[string]MetricData, error) {
	log.Printf("Fetching metrics for site %s.%s (duration: %s)...", siteID, environment, duration)

	session, err := c.sessionManager.GetSession(ctx, machineToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	envsService := api.NewEnvironmentsService(session.Client)
	metrics, err := envsService.GetMetrics(ctx, siteID, environment, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	return ConvertMetricsToMap(metrics), nil
}

// InvalidateSession removes a session, forcing re-authentication on next use.
func (c *Client) InvalidateSession(machineToken string) {
	c.sessionManager.InvalidateSession(machineToken)
}

// ----- Test helper functions (kept for testing with JSON files) -----

// parseMetricsData parses metrics JSON data
func parseMetricsData(data []byte) (map[string]MetricData, error) {
	var response MetricsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return response.Timeseries, nil
}

// parseSiteInfo parses site info JSON data
func parseSiteInfo(data []byte) (*SiteInfo, error) {
	var siteInfo SiteInfo
	if err := json.Unmarshal(data, &siteInfo); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &siteInfo, nil
}

// parseSiteList parses site list JSON data
func parseSiteList(data []byte) (map[string]SiteListEntry, error) {
	var siteList map[string]SiteListEntry
	if err := json.Unmarshal(data, &siteList); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return siteList, nil
}

// LoadMetricsData loads metrics data from a JSON file (used for testing)
func LoadMetricsData(filename string) (map[string]MetricData, error) {
	data, err := os.ReadFile(filename) // #nosec G304 - test helper function, filename from test data
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseMetricsData(data)
}

// LoadSiteConfig loads site config from a JSON file (legacy format, used for testing)
func LoadSiteConfig(filename string) (*SiteConfig, error) {
	data, err := os.ReadFile(filename) // #nosec G304 - test helper function, filename from test data
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var config SiteConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &config, nil
}

// LoadSiteInfo loads site info from a JSON file (used for testing)
func LoadSiteInfo(filename string) (*SiteInfo, error) {
	data, err := os.ReadFile(filename) // #nosec G304 - test helper function, filename from test data
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseSiteInfo(data)
}

// LoadSiteList loads site list from a JSON file (used for testing)
func LoadSiteList(filename string) (map[string]SiteListEntry, error) {
	data, err := os.ReadFile(filename) // #nosec G304 - test helper function, filename from test data
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseSiteList(data)
}
