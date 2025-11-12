package main

import (
	"log"
	"math"
	"sync"
	"time"
)

// RefreshManager manages periodic refresh of site lists and metrics
type RefreshManager struct {
	tokens          []string
	environment     string
	refreshInterval time.Duration
	collector       *PantheonCollector
	mu              sync.Mutex
}

// NewRefreshManager creates a new refresh manager
func NewRefreshManager(tokens []string, environment string, refreshInterval time.Duration, collector *PantheonCollector) *RefreshManager {
	return &RefreshManager{
		tokens:          tokens,
		environment:     environment,
		refreshInterval: refreshInterval,
		collector:       collector,
	}
}

// Start begins the periodic refresh process
func (rm *RefreshManager) Start() {
	// Start site list refresh (every refresh interval)
	go rm.refreshSiteListsPeriodically()

	// Start metrics refresh with queue-based processing
	go rm.refreshMetricsWithQueue()
}

// refreshSiteListsPeriodically refreshes site lists for all accounts
func (rm *RefreshManager) refreshSiteListsPeriodically() {
	ticker := time.NewTicker(rm.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Starting site list refresh...")
		rm.refreshAllSiteLists()
	}
}

// refreshAllSiteLists refreshes the site list for all accounts
func (rm *RefreshManager) refreshAllSiteLists() {
	var allSiteMetrics []SiteMetrics

	for _, token := range rm.tokens {
		accountID := getAccountID(token)
		log.Printf("Refreshing site list for account %s", accountID)

		// Authenticate with this token
		if err := authenticateWithToken(token); err != nil {
			log.Printf("Warning: Failed to authenticate account %s during refresh: %v", accountID, err)
			continue
		}

		// Fetch all sites for this account
		siteList, err := fetchAllSites()
		if err != nil {
			log.Printf("Warning: Failed to fetch site list for account %s during refresh: %v", accountID, err)
			continue
		}

		log.Printf("Account %s: Found %d sites", accountID, len(siteList))

		// Get existing metrics for sites
		existingSites := rm.collector.GetSites()
		existingMetricsMap := make(map[string]map[string]MetricData)
		for _, site := range existingSites {
			key := site.Account + ":" + site.SiteName
			existingMetricsMap[key] = site.MetricsData
		}

		// Create site metrics entries, preserving existing metrics data
		for _, site := range siteList {
			key := accountID + ":" + site.Name
			metricsData := existingMetricsMap[key]
			if metricsData == nil {
				metricsData = make(map[string]MetricData)
			}

			siteMetrics := SiteMetrics{
				SiteName:    site.Name,
				Label:       site.Name,
				PlanName:    site.PlanName,
				Account:     accountID,
				MetricsData: metricsData,
			}
			allSiteMetrics = append(allSiteMetrics, siteMetrics)
		}
	}

	if len(allSiteMetrics) > 0 {
		rm.collector.UpdateSites(allSiteMetrics)
		log.Printf("Site list refresh complete: now monitoring %d sites", len(allSiteMetrics))
	}
}

// refreshMetricsWithQueue processes metrics refresh using a queue to prevent stampedes
func (rm *RefreshManager) refreshMetricsWithQueue() {
	sites := rm.collector.GetSites()
	if len(sites) == 0 {
		log.Printf("No sites to refresh metrics for")
		return
	}

	// Calculate how many sites to process per minute
	totalSites := len(sites)
	refreshMinutes := rm.refreshInterval.Minutes()
	sitesPerMinute := int(math.Ceil(float64(totalSites) / refreshMinutes))

	log.Printf("Metrics refresh: processing %d sites per minute (%d sites total, %.0f minute interval)",
		sitesPerMinute, totalSites, refreshMinutes)

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	siteIndex := 0

	for range ticker.C {
		// Refresh site list to get current sites
		currentSites := rm.collector.GetSites()
		if len(currentSites) == 0 {
			continue
		}

		// Reset index if it exceeds current site count
		if siteIndex >= len(currentSites) {
			siteIndex = 0
		}

		// Process the next batch of sites
		endIndex := siteIndex + sitesPerMinute
		if endIndex > len(currentSites) {
			endIndex = len(currentSites)
		}

		sitesToProcess := currentSites[siteIndex:endIndex]
		log.Printf("Refreshing metrics for %d sites (index %d-%d of %d)",
			len(sitesToProcess), siteIndex, endIndex-1, len(currentSites))

		for _, site := range sitesToProcess {
			go rm.refreshSiteMetrics(site.Account, site.SiteName)
		}

		siteIndex = endIndex
		if siteIndex >= len(currentSites) {
			siteIndex = 0
			log.Printf("Completed full metrics refresh cycle, starting over")
		}
	}
}

// refreshSiteMetrics refreshes metrics for a single site
func (rm *RefreshManager) refreshSiteMetrics(accountID, siteName string) {
	// Find the token for this account
	var token string
	for _, t := range rm.tokens {
		if getAccountID(t) == accountID {
			token = t
			break
		}
	}

	if token == "" {
		log.Printf("Warning: No token found for account %s", accountID)
		return
	}

	// Authenticate with this token
	if err := authenticateWithToken(token); err != nil {
		log.Printf("Warning: Failed to authenticate account %s for metrics refresh: %v", accountID, err)
		return
	}

	// Fetch metrics for this site
	metricsData, err := fetchMetricsData(siteName, rm.environment)
	if err != nil {
		log.Printf("Warning: Failed to refresh metrics for %s.%s: %v", accountID, siteName, err)
		return
	}

	// Update the collector
	rm.collector.UpdateSiteMetrics(accountID, siteName, metricsData)
	log.Printf("Successfully refreshed metrics for %s.%s (%d entries)", accountID, siteName, len(metricsData))
}
