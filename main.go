package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// createSiteMetrics creates a SiteMetrics struct from site list entry and metrics data
func createSiteMetrics(siteName, accountID, planName string, metricsData map[string]MetricData) SiteMetrics {
	return SiteMetrics{
		SiteName:    siteName,
		Label:       siteName, // site:list doesn't provide a label field, using name
		PlanName:    planName,
		Account:     accountID,
		MetricsData: metricsData,
	}
}

// processAccountSiteList processes a list of sites for an account and collects metrics
func processAccountSiteList(accountID, environment string, siteList map[string]SiteListEntry) ([]SiteMetrics, int, int) {
	var siteMetrics []SiteMetrics
	successCount := 0
	failCount := 0

	for _, site := range siteList {
		log.Printf("Account %s: Processing site %s (plan: %s)", accountID, site.Name, site.PlanName)

		// Fetch metrics for this site
		metricsData, err := fetchMetricsData(site.Name, environment)
		if err != nil {
			log.Printf("Warning: Failed to fetch metrics for %s.%s: %v", accountID, site.Name, err)
			failCount++
			continue
		}

		// Create SiteMetrics entry with account label
		metrics := createSiteMetrics(site.Name, accountID, site.PlanName, metricsData)
		siteMetrics = append(siteMetrics, metrics)
		successCount++
		log.Printf("Account %s: Successfully loaded %d metric entries for %s", accountID, len(metricsData), site.Name)
	}

	return siteMetrics, successCount, failCount
}

// collectAccountMetrics collects metrics for a single account
func collectAccountMetrics(token, environment string) ([]SiteMetrics, int, int) {
	var siteMetrics []SiteMetrics
	successCount := 0
	failCount := 0

	// Authenticate with this token
	if err := authenticateWithToken(token); err != nil {
		// Use token suffix as fallback for logging if auth fails
		accountID := getAccountID(token)
		log.Printf("Warning: Failed to authenticate account %s: %v", accountID, err)
		return siteMetrics, successCount, failCount
	}

	// Get the authenticated account email
	accountID, err := getAuthenticatedAccountEmail()
	if err != nil {
		// Use token suffix as fallback if we can't get email
		accountID = getAccountID(token)
		log.Printf("Warning: Failed to get account email, using token suffix %s: %v", accountID, err)
	}

	// Fetch all sites for this account
	siteList, err := fetchAllSites()
	if err != nil {
		log.Printf("Warning: Failed to fetch site list for account %s: %v", accountID, err)
		return siteMetrics, successCount, failCount
	}

	log.Printf("Account %s: Found %d sites", accountID, len(siteList))

	// Process all sites
	siteMetrics, successCount, failCount = processAccountSiteList(accountID, environment, siteList)

	log.Printf("Account %s: Metrics collection complete: %d successful, %d failed", accountID, successCount, failCount)
	return siteMetrics, successCount, failCount
}

// collectAllSiteLists collects site lists for all accounts without fetching metrics
func collectAllSiteLists(tokens []string) []SiteMetrics {
	var allSiteMetrics []SiteMetrics

	for tokenIdx, token := range tokens {
		log.Printf("Loading site list for account %d/%d", tokenIdx+1, len(tokens))

		// Authenticate with this token
		if err := authenticateWithToken(token); err != nil {
			// Use token suffix as fallback for logging if auth fails
			accountID := getAccountID(token)
			log.Printf("Warning: Failed to authenticate account %s: %v", accountID, err)
			continue
		}

		// Get the authenticated account email
		accountID, err := getAuthenticatedAccountEmail()
		if err != nil {
			// Use token suffix as fallback if we can't get email
			accountID = getAccountID(token)
			log.Printf("Warning: Failed to get account email, using token suffix %s: %v", accountID, err)
		}

		// Fetch all sites for this account
		siteList, err := fetchAllSites()
		if err != nil {
			log.Printf("Warning: Failed to fetch site list for account %s: %v", accountID, err)
			continue
		}

		log.Printf("Account %s: Found %d sites", accountID, len(siteList))

		// Create site metrics entries with empty metrics data
		for _, site := range siteList {
			siteMetrics := SiteMetrics{
				SiteName:    site.Name,
				Label:       site.Name,
				PlanName:    site.PlanName,
				Account:     accountID,
				MetricsData: make(map[string]MetricData),
			}
			allSiteMetrics = append(allSiteMetrics, siteMetrics)
		}
	}

	log.Printf("Site list collection complete: %d sites found across %d accounts", len(allSiteMetrics), len(tokens))
	return allSiteMetrics
}

// collectAllMetrics collects metrics for all accounts
func collectAllMetrics(tokens []string, environment string) []SiteMetrics {
	var allSiteMetrics []SiteMetrics
	totalSuccessCount := 0
	totalFailCount := 0

	for tokenIdx, token := range tokens {
		log.Printf("Processing account %d/%d", tokenIdx+1, len(tokens))

		siteMetrics, successCount, failCount := collectAccountMetrics(token, environment)
		allSiteMetrics = append(allSiteMetrics, siteMetrics...)
		totalSuccessCount += successCount
		totalFailCount += failCount
	}

	log.Printf("Overall metrics collection complete: %d successful, %d failed across %d accounts", totalSuccessCount, totalFailCount, len(tokens))
	return allSiteMetrics
}

// createRootHandler creates the HTTP handler for the root path
func createRootHandler(environment string, tokens []string, collector *PantheonCollector) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allSiteMetrics := collector.GetSites()

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
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
			fmt.Fprintf(w, "<li>[%s] %s (plan: %s, %d metrics)</li>\n",
				site.Account, site.SiteName, site.PlanName, len(site.MetricsData))
		}

		fmt.Fprintf(w, `
</ul>
<p>Metrics are available at <a href="/metrics">/metrics</a></p>
</body>
</html>
`)
	}
}

// setupHTTPHandlers sets up HTTP routes for the metrics exporter
func setupHTTPHandlers(registry *prometheus.Registry, environment string, tokens []string, collector *PantheonCollector) {
	// Create HTTP handler for metrics
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Root handler with instructions
	http.HandleFunc("/", createRootHandler(environment, tokens, collector))
}

// startRefreshManager creates and starts the refresh manager
func startRefreshManager(tokens []string, environment string, refreshInterval time.Duration, collector *PantheonCollector) *RefreshManager {
	refreshManager := NewRefreshManager(tokens, environment, refreshInterval, collector)
	refreshManager.Start()
	return refreshManager
}

func main() {
	// Parse command-line flags
	environment := flag.String("env", "live", "Pantheon environment (default: live)")
	port := flag.String("port", "8080", "HTTP server port (default: 8080)")
	refreshInterval := flag.Int("refreshInterval", 60, "Refresh interval in minutes (default: 60)")
	flag.Parse()

	// Read machine tokens from environment variable
	tokensEnv := os.Getenv("PANTHEON_MACHINE_TOKENS")
	if tokensEnv == "" {
		log.Fatal("PANTHEON_MACHINE_TOKENS environment variable is not set")
	}

	// Split tokens by space
	tokens := strings.Fields(tokensEnv)
	if len(tokens) == 0 {
		log.Fatal("No tokens found in PANTHEON_MACHINE_TOKENS")
	}

	log.Printf("Found %d Pantheon account(s) to process", len(tokens))

	// Collect site lists first (fast - no metrics)
	log.Printf("Loading site lists...")
	allSites := collectAllSiteLists(tokens)

	// Create collector with sites (empty metrics initially)
	collector := NewPantheonCollector(allSites)

	// Register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Setup HTTP handlers
	setupHTTPHandlers(registry, *environment, tokens, collector)

	// Start refresh manager
	refreshIntervalDuration := time.Duration(*refreshInterval) * time.Minute
	refreshManager := startRefreshManager(tokens, *environment, refreshIntervalDuration, collector)
	refreshManager.InitializeDiscoveredSites()
	log.Printf("Refresh manager started (interval: %d minutes)", *refreshInterval)

	// Collect initial metrics in background goroutine
	go func() {
		log.Printf("Starting initial metrics collection in background...")
		allSiteMetrics := collectAllMetrics(tokens, *environment)

		if len(allSiteMetrics) > 0 {
			log.Printf("Initial metrics collection complete: %d sites with metrics", len(allSiteMetrics))
			collector.UpdateSites(allSiteMetrics)
		}
	}()

	// Start server
	serverAddr := ":" + *port
	log.Printf("Starting Pantheon metrics exporter on %s", serverAddr)
	log.Printf("Metrics available at http://localhost%s/metrics", serverAddr)
	log.Printf("Server is ready to serve requests (metrics collection running in background)")
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
