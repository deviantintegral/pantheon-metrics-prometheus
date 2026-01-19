// Package app provides the main application logic for the Pantheon metrics exporter.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/collector"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/refresh"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InitialMetricsDuration is used for the first metrics fetch (28 days of history).
const InitialMetricsDuration = "28d"

// AccountSiteData holds pre-fetched site data for an account
type AccountSiteData struct {
	AccountID string
	Sites     map[string]pantheon.SiteListEntry
}

// createSiteMetrics creates a SiteMetrics struct from site list entry and metrics data
func createSiteMetrics(siteName, siteID, accountID, planName string, metricsData map[string]pantheon.MetricData) pantheon.SiteMetrics {
	return pantheon.SiteMetrics{
		SiteName:    siteName,
		SiteID:      siteID,
		Label:       siteName, // site:list doesn't provide a label field, using name
		PlanName:    planName,
		Account:     accountID,
		MetricsData: metricsData,
	}
}

// processAccountSiteList processes a list of sites for an account and collects metrics
// siteLimit and currentCount are used to limit the total number of sites processed globally.
func processAccountSiteList(ctx context.Context, client *pantheon.Client, token, accountID, environment string, siteList map[string]pantheon.SiteListEntry, siteLimit, currentCount int) ([]pantheon.SiteMetrics, int, int) {
	siteMetrics := make([]pantheon.SiteMetrics, 0, len(siteList))
	successCount := 0
	failCount := 0

	for siteID, site := range siteList {
		// Check if we've reached the global site limit
		if siteLimit > 0 && (currentCount+len(siteMetrics)) >= siteLimit {
			log.Printf("Site limit reached (%d sites), stopping metrics collection", siteLimit)
			break
		}

		log.Printf("Account %s: Processing site %s (plan: %s)", accountID, site.Name, site.PlanName)

		// Fetch metrics for this site (use 28d for initial fetch)
		metricsData, err := client.FetchMetricsData(ctx, token, siteID, environment, InitialMetricsDuration)
		if err != nil {
			log.Printf("Warning: Failed to fetch metrics for %s.%s: %v", accountID, site.Name, err)
			failCount++
			continue
		}

		// Create SiteMetrics entry with account label
		metrics := createSiteMetrics(site.Name, siteID, accountID, site.PlanName, metricsData)
		siteMetrics = append(siteMetrics, metrics)
		successCount++
		log.Printf("Account %s: Successfully loaded %d metric entries for %s", accountID, len(metricsData), site.Name)
	}

	return siteMetrics, successCount, failCount
}

// collectAccountMetrics collects metrics for a single account
// siteLimit and currentCount are used to limit the total number of sites processed globally.
func collectAccountMetrics(ctx context.Context, client *pantheon.Client, token, environment string, siteLimit, currentCount int) ([]pantheon.SiteMetrics, int, int) {
	var siteMetrics []pantheon.SiteMetrics
	successCount := 0
	failCount := 0

	// Authenticate with this token
	accountID, err := client.Authenticate(ctx, token)
	if err != nil {
		// Use token suffix as fallback for logging if auth fails
		accountID = pantheon.GetAccountID(token)
		log.Printf("Warning: Failed to authenticate account %s: %v", accountID, err)
		return siteMetrics, successCount, failCount
	}

	// Fetch all sites for this account
	siteList, err := client.FetchAllSites(ctx, token)
	if err != nil {
		log.Printf("Warning: Failed to fetch site list for account %s: %v", accountID, err)
		return siteMetrics, successCount, failCount
	}

	log.Printf("Account %s: Found %d sites", accountID, len(siteList))

	// Process all sites
	siteMetrics, successCount, failCount = processAccountSiteList(ctx, client, token, accountID, environment, siteList, siteLimit, currentCount)

	log.Printf("Account %s: Metrics collection complete: %d successful, %d failed", accountID, successCount, failCount)
	return siteMetrics, successCount, failCount
}

// CollectAllSiteLists collects site lists for all accounts without fetching metrics.
// Returns the site metrics for the collector and a map of token -> AccountSiteData for later use.
// If siteLimit > 0, only the first siteLimit sites are returned.
func CollectAllSiteLists(ctx context.Context, client *pantheon.Client, tokens []string, siteLimit int) ([]pantheon.SiteMetrics, map[string]AccountSiteData) {
	var allSiteMetrics []pantheon.SiteMetrics
	tokenSiteData := make(map[string]AccountSiteData)

	for tokenIdx, token := range tokens {
		log.Printf("Loading site list for account %d/%d", tokenIdx+1, len(tokens))

		// Authenticate with this token
		accountID, err := client.Authenticate(ctx, token)
		if err != nil {
			// Use token suffix as fallback for logging if auth fails
			accountID = pantheon.GetAccountID(token)
			log.Printf("Warning: Failed to authenticate account %s: %v", accountID, err)
			continue
		}

		// Fetch all sites for this account
		siteList, err := client.FetchAllSites(ctx, token)
		if err != nil {
			log.Printf("Warning: Failed to fetch site list for account %s: %v", accountID, err)
			continue
		}

		log.Printf("Account %s: Found %d sites", accountID, len(siteList))

		// Store the fetched data for later use
		tokenSiteData[token] = AccountSiteData{
			AccountID: accountID,
			Sites:     siteList,
		}

		// Create site metrics entries with empty metrics data
		for siteID, site := range siteList {
			siteMetrics := pantheon.SiteMetrics{
				SiteName:    site.Name,
				SiteID:      siteID,
				Label:       site.Name,
				PlanName:    site.PlanName,
				Account:     accountID,
				MetricsData: make(map[string]pantheon.MetricData),
			}
			allSiteMetrics = append(allSiteMetrics, siteMetrics)

			// Apply site limit if set
			if siteLimit > 0 && len(allSiteMetrics) >= siteLimit {
				log.Printf("Site limit reached (%d sites), stopping collection", siteLimit)
				break
			}
		}

		// Check if limit reached after processing account
		if siteLimit > 0 && len(allSiteMetrics) >= siteLimit {
			break
		}
	}

	log.Printf("Site list collection complete: %d sites found across %d accounts", len(allSiteMetrics), len(tokens))
	return allSiteMetrics, tokenSiteData
}

// CollectAllMetrics collects metrics for all accounts (fetches site lists fresh)
// If siteLimit > 0, only the first siteLimit sites are processed.
func CollectAllMetrics(ctx context.Context, client *pantheon.Client, tokens []string, environment string, siteLimit int) []pantheon.SiteMetrics {
	var allSiteMetrics []pantheon.SiteMetrics
	totalSuccessCount := 0
	totalFailCount := 0

	for tokenIdx, token := range tokens {
		log.Printf("Processing account %d/%d", tokenIdx+1, len(tokens))

		siteMetrics, successCount, failCount := collectAccountMetrics(ctx, client, token, environment, siteLimit, len(allSiteMetrics))
		allSiteMetrics = append(allSiteMetrics, siteMetrics...)
		totalSuccessCount += successCount
		totalFailCount += failCount

		// Check if limit reached after processing account
		if siteLimit > 0 && len(allSiteMetrics) >= siteLimit {
			break
		}
	}

	log.Printf("Overall metrics collection complete: %d successful, %d failed across %d accounts", totalSuccessCount, totalFailCount, len(tokens))
	return allSiteMetrics
}

// CollectAllMetricsWithSites collects metrics using pre-fetched site data (avoids duplicate site fetch)
// If siteLimit > 0, only the first siteLimit sites are processed.
func CollectAllMetricsWithSites(ctx context.Context, client *pantheon.Client, tokens []string, environment string, preFetchedSites map[string]AccountSiteData, siteLimit int) []pantheon.SiteMetrics {
	var allSiteMetrics []pantheon.SiteMetrics
	totalSuccessCount := 0
	totalFailCount := 0

	for tokenIdx, token := range tokens {
		log.Printf("Processing account %d/%d", tokenIdx+1, len(tokens))

		siteData, ok := preFetchedSites[token]
		if !ok {
			log.Printf("Warning: No pre-fetched site data for account %d, skipping", tokenIdx+1)
			continue
		}

		// Process sites using the pre-fetched data
		siteMetrics, successCount, failCount := processAccountSiteList(ctx, client, token, siteData.AccountID, environment, siteData.Sites, siteLimit, len(allSiteMetrics))
		allSiteMetrics = append(allSiteMetrics, siteMetrics...)
		totalSuccessCount += successCount
		totalFailCount += failCount

		log.Printf("Account %s: Metrics collection complete: %d successful, %d failed", siteData.AccountID, successCount, failCount)

		// Check if limit reached after processing account
		if siteLimit > 0 && len(allSiteMetrics) >= siteLimit {
			break
		}
	}

	log.Printf("Overall metrics collection complete: %d successful, %d failed across %d accounts", totalSuccessCount, totalFailCount, len(tokens))
	return allSiteMetrics
}

// createRootHandler creates the HTTP handler for the root path
func createRootHandler(environment string, tokens []string, c *collector.PantheonCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		allSiteMetrics := c.GetSites()

		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, `
<html>
<head><title>Pantheon Metrics Exporter</title></head>
<body>
<h1>Pantheon Metrics Exporter</h1>
<p><strong>Environment:</strong> %s</p>
<p><strong>Accounts monitored:</strong> %d</p>
<p><strong>Sites monitored:</strong> %d</p>
<ul>
`, environment, len(tokens), len(allSiteMetrics))

		for _, site := range allSiteMetrics {
			_, _ = fmt.Fprintf(w, "<li>[%s] %s (plan: %s, %d metrics)</li>\n",
				site.Account, site.SiteName, site.PlanName, len(site.MetricsData))
		}

		_, _ = fmt.Fprintf(w, `
</ul>
<p>Metrics are available at <a href="/metrics">/metrics</a></p>
</body>
</html>
`)
	}
}

// SetupHTTPHandlers sets up HTTP routes for the metrics exporter
func SetupHTTPHandlers(registry *prometheus.Registry, environment string, tokens []string, c *collector.PantheonCollector) {
	// Create HTTP handler for metrics
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Root handler with instructions
	http.HandleFunc("/", createRootHandler(environment, tokens, c))
}

// StartRefreshManager creates and starts the refresh manager
func StartRefreshManager(client *pantheon.Client, tokens []string, environment string, refreshInterval time.Duration, c *collector.PantheonCollector, siteLimit int) *refresh.Manager {
	refreshManager := refresh.NewManager(client, tokens, environment, refreshInterval, c, siteLimit)
	refreshManager.Start()
	return refreshManager
}
