// Package pantheon provides types and client functions for interacting with Pantheon via the terminus-golang library.
package pantheon

import "context"

// ClientInterface defines the interface for Pantheon API operations.
// This allows for mocking in tests.
type ClientInterface interface {
	// Authenticate authenticates with a machine token and returns the account email.
	Authenticate(ctx context.Context, machineToken string) (string, error)

	// GetEmail returns the email for the given machine token.
	GetEmail(ctx context.Context, machineToken string) (string, error)

	// FetchAllSites fetches the list of all sites for the authenticated user.
	// If orgID is non-empty, only sites from that organization will be returned.
	FetchAllSites(ctx context.Context, machineToken string, orgID string) (map[string]SiteListEntry, error)

	// FetchMetricsData fetches metrics data for a site.
	FetchMetricsData(ctx context.Context, machineToken, siteID, environment, duration string) (map[string]MetricData, error)

	// InvalidateSession removes a session, forcing re-authentication on next use.
	InvalidateSession(machineToken string)
}

// Ensure Client implements ClientInterface
var _ ClientInterface = (*Client)(nil)
