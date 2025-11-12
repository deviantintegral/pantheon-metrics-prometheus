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

	// Collect metrics for all accounts and sites
	var allSiteMetrics []SiteMetrics
	totalSuccessCount := 0
	totalFailCount := 0

	for tokenIdx, token := range tokens {
		accountID := getAccountID(token)
		log.Printf("Processing account %d/%d (ID: %s)", tokenIdx+1, len(tokens), accountID)

		// Authenticate with this token
		if err := authenticateWithToken(token); err != nil {
			log.Printf("Warning: Failed to authenticate account %s: %v", accountID, err)
			continue
		}

		// Fetch all sites for this account
		siteList, err := fetchAllSites()
		if err != nil {
			log.Printf("Warning: Failed to fetch site list for account %s: %v", accountID, err)
			continue
		}

		log.Printf("Account %s: Found %d sites", accountID, len(siteList))

		// Collect metrics for all accessible sites in this account
		successCount := 0
		failCount := 0

		for _, site := range siteList {
			log.Printf("Account %s: Processing site %s (plan: %s)", accountID, site.Name, site.PlanName)

			// Fetch metrics for this site
			metricsData, err := fetchMetricsData(site.Name, *environment)
			if err != nil {
				log.Printf("Warning: Failed to fetch metrics for %s.%s: %v", accountID, site.Name, err)
				failCount++
				continue
			}

			// Create SiteMetrics entry with account label
			siteMetrics := SiteMetrics{
				SiteName:    site.Name,
				Label:       site.Name, // site:list doesn't provide a label field, using name
				PlanName:    site.PlanName,
				Account:     accountID,
				MetricsData: metricsData,
			}

			allSiteMetrics = append(allSiteMetrics, siteMetrics)
			successCount++
			log.Printf("Account %s: Successfully loaded %d metric entries for %s", accountID, len(metricsData), site.Name)
		}

		log.Printf("Account %s: Metrics collection complete: %d successful, %d failed", accountID, successCount, failCount)
		totalSuccessCount += successCount
		totalFailCount += failCount
	}

	log.Printf("Overall metrics collection complete: %d successful, %d failed across %d accounts", totalSuccessCount, totalFailCount, len(tokens))

	if len(allSiteMetrics) == 0 {
		log.Fatal("No site metrics were collected. Cannot start exporter.")
	}

	// Create collector with all site metrics
	collector := NewPantheonCollector(allSiteMetrics)

	// Register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Create HTTP handler
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Root handler with instructions
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
`, *environment, len(tokens), len(allSiteMetrics))

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
	})

	// Start refresh manager
	refreshIntervalDuration := time.Duration(*refreshInterval) * time.Minute
	refreshManager := NewRefreshManager(tokens, *environment, refreshIntervalDuration, collector)
	refreshManager.InitializeDiscoveredSites()
	refreshManager.Start()
	log.Printf("Refresh manager started (interval: %d minutes)", *refreshInterval)

	// Start server
	serverAddr := ":" + *port
	log.Printf("Starting Pantheon metrics exporter on %s", serverAddr)
	log.Printf("Exporting metrics for %d sites", len(allSiteMetrics))
	log.Printf("Metrics available at http://localhost%s/metrics", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
