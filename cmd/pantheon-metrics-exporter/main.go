// Package main is the entry point for the Pantheon metrics exporter.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/app"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/collector"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	// Parse command-line flags
	environment := flag.String("env", "live", "Pantheon environment (default: live)")
	port := flag.String("port", "8080", "HTTP server port (default: 8080)")
	refreshInterval := flag.Int("refreshInterval", 60, "Refresh interval in minutes (default: 60)")
	debug := flag.Bool("debug", false, "Enable debug logging of HTTP requests and responses to stderr")
	siteLimit := flag.Int("siteLimit", 0, "Maximum number of sites to query (0 = no limit)")
	orgID := flag.String("orgID", "", "Limit metrics to sites from this organization ID (optional)")
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

	// Create the Pantheon API client with debug logging if enabled
	client := pantheon.NewClient(*debug)
	ctx := context.Background()

	// Log organization filter if specified
	if *orgID != "" {
		log.Printf("Filtering sites to organization: %s", *orgID)
	}

	// Collect site lists first (fast - no metrics)
	log.Printf("Loading site lists...")
	allSites, preFetchedSites := app.CollectAllSiteLists(ctx, client, tokens, *siteLimit, *orgID)

	// Create collector with sites (empty metrics initially)
	pantheonCollector := collector.NewPantheonCollector(allSites)

	// Register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(pantheonCollector)

	// Setup HTTP handlers
	app.SetupHTTPHandlers(registry, *environment, tokens, pantheonCollector)

	// Start refresh manager
	refreshIntervalDuration := time.Duration(*refreshInterval) * time.Minute
	refreshManager := app.StartRefreshManager(client, tokens, *environment, refreshIntervalDuration, pantheonCollector, *siteLimit, *orgID)
	refreshManager.InitializeDiscoveredSites()
	refreshManager.InitializeAccountTokenMap()
	log.Printf("Refresh manager started (interval: %d minutes)", *refreshInterval)

	// Collect initial metrics in background goroutine (using pre-fetched site lists)
	// Metrics are updated incrementally as each site is processed
	go func() {
		log.Printf("Starting initial metrics collection in background...")
		// Update collector incrementally as each site's metrics are fetched
		onMetricsFetched := func(accountID, siteName string, metricsData map[string]pantheon.MetricData) {
			pantheonCollector.UpdateSiteMetrics(accountID, siteName, metricsData)
		}
		allSiteMetrics := app.CollectAllMetricsWithSites(ctx, client, tokens, *environment, preFetchedSites, *siteLimit, onMetricsFetched)

		log.Printf("Initial metrics collection complete: %d sites with metrics", len(allSiteMetrics))
	}()

	// Start server with timeouts
	serverAddr := ":" + *port
	log.Printf("Starting Pantheon metrics exporter on %s", serverAddr)
	log.Printf("Metrics available at http://localhost%s/metrics", serverAddr)
	log.Printf("Server is ready to serve requests (metrics collection running in background)")

	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
