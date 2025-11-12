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
	discoveredSites map[string]bool // Track sites discovered since app start (account:site format)
}

// NewRefreshManager creates a new refresh manager
func NewRefreshManager(tokens []string, environment string, refreshInterval time.Duration, collector *PantheonCollector) *RefreshManager {
	return &RefreshManager{
		tokens:          tokens,
		environment:     environment,
		refreshInterval: refreshInterval,
		collector:       collector,
		discoveredSites: make(map[string]bool),
	}
}

// InitializeDiscoveredSites populates the discovered sites map with initial sites
func (rm *RefreshManager) InitializeDiscoveredSites() {
	sites := rm.collector.GetSites()
	for _, site := range sites {
		key := site.Account + ":" + site.SiteName
		rm.discoveredSites[key] = true
	}
	log.Printf("Initialized with %d discovered sites", len(rm.discoveredSites))
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

	// Get current sites to track changes
	existingSites := rm.collector.GetSites()
	currentSitesMap := make(map[string]bool)
	for _, site := range existingSites {
		key := site.Account + ":" + site.SiteName
		currentSitesMap[key] = true
	}

	// Track new sites for this refresh
	newSitesMap := make(map[string]bool)
	var addedSites []string
	totalSitesFound := 0

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

		totalSitesFound += len(siteList)

		// Get existing metrics for sites
		existingMetricsMap := make(map[string]map[string]MetricData)
		for _, site := range existingSites {
			key := site.Account + ":" + site.SiteName
			existingMetricsMap[key] = site.MetricsData
		}

		// Create site metrics entries, preserving existing metrics data
		for _, site := range siteList {
			key := accountID + ":" + site.Name
			newSitesMap[key] = true

			metricsData := existingMetricsMap[key]
			if metricsData == nil {
				metricsData = make(map[string]MetricData)
			}

			// Check if this is a newly added site (not in current list)
			if !currentSitesMap[key] {
				// Check if it's newly discovered (never seen before)
				if !rm.discoveredSites[key] {
					addedSites = append(addedSites, key)
					rm.discoveredSites[key] = true
				}
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

	// Find removed sites
	var removedSites []string
	for key := range currentSitesMap {
		if !newSitesMap[key] {
			removedSites = append(removedSites, key)
		}
	}

	// Update collector
	if len(allSiteMetrics) > 0 {
		rm.collector.UpdateSites(allSiteMetrics)
		log.Printf("Site list updated: %d sites found", totalSitesFound)

		if len(addedSites) > 0 {
			log.Printf("Sites added: %v", addedSites)
		}

		if len(removedSites) > 0 {
			log.Printf("Sites removed: %v", removedSites)
		}
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
	log.Printf("Updated metrics for site %s.%s", accountID, siteName)
}
