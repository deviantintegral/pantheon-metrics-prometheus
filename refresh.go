package main

import (
	"log"
	"math"
	"sync/atomic"
	"time"
)

// RefreshManager manages periodic refresh of site lists and metrics
type RefreshManager struct {
	tokens          []string
	environment     string
	refreshInterval time.Duration
	collector       *PantheonCollector
	discoveredSites map[string]bool   // Track sites discovered since app start (account:site format)
	accountTokenMap map[string]string // Map from account email to token
	tickerInterval  time.Duration     // Interval for metrics refresh ticker (defaults to 1 minute)
	tickerFireCount int64             // Counter for ticker fires (for testing)
}

// NewRefreshManager creates a new refresh manager
func NewRefreshManager(tokens []string, environment string, refreshInterval time.Duration, collector *PantheonCollector) *RefreshManager {
	return &RefreshManager{
		tokens:          tokens,
		environment:     environment,
		refreshInterval: refreshInterval,
		collector:       collector,
		discoveredSites: make(map[string]bool),
		accountTokenMap: make(map[string]string),
		tickerInterval:  1 * time.Minute, // Default to 1 minute
	}
}

// SetTickerInterval sets the ticker interval for metrics refresh (useful for testing)
func (rm *RefreshManager) SetTickerInterval(interval time.Duration) {
	rm.tickerInterval = interval
}

// GetTickerFireCount returns the number of times the ticker has fired (useful for testing)
func (rm *RefreshManager) GetTickerFireCount() int64 {
	return atomic.LoadInt64(&rm.tickerFireCount)
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

// buildSiteKeyMap creates a map of site keys from a list of sites
func buildSiteKeyMap(sites []SiteMetrics) map[string]bool {
	siteMap := make(map[string]bool)
	for _, site := range sites {
		key := site.Account + ":" + site.SiteName
		siteMap[key] = true
	}
	return siteMap
}

// findAddedSites returns the list of sites that are in newSites but not in currentSites
func findAddedSites(currentSites, newSites map[string]bool, discoveredSites map[string]bool) []string {
	var added []string
	for key := range newSites {
		if !currentSites[key] {
			// Check if it's newly discovered (never seen before)
			if !discoveredSites[key] {
				added = append(added, key)
			}
		}
	}
	return added
}

// findRemovedSites returns the list of sites that are in currentSites but not in newSites
func findRemovedSites(currentSites, newSites map[string]bool) []string {
	var removed []string
	for key := range currentSites {
		if !newSites[key] {
			removed = append(removed, key)
		}
	}
	return removed
}

// refreshAllSiteLists refreshes the site list for all accounts
func (rm *RefreshManager) refreshAllSiteLists() {
	var allSiteMetrics []SiteMetrics

	// Get current sites to track changes
	existingSites := rm.collector.GetSites()
	currentSitesMap := buildSiteKeyMap(existingSites)

	// Track new sites for this refresh
	newSitesMap := make(map[string]bool)
	totalSitesFound := 0

	for _, token := range rm.tokens {
		// Authenticate with this token
		if err := authenticateWithToken(token); err != nil {
			// Use token suffix as fallback for logging if auth fails
			accountID := getAccountID(token)
			log.Printf("Warning: Failed to authenticate account %s during refresh: %v", accountID, err)
			continue
		}

		// Get the authenticated account email
		accountID, err := getAuthenticatedAccountEmail()
		if err != nil {
			// Use token suffix as fallback if we can't get email
			accountID = getAccountID(token)
			log.Printf("Warning: Failed to get account email during refresh, using token suffix %s: %v", accountID, err)
		}

		// Store the mapping for later use
		rm.accountTokenMap[accountID] = token

		log.Printf("Refreshing site list for account %s", accountID)

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

	// Find added and removed sites
	addedSites := findAddedSites(currentSitesMap, newSitesMap, rm.discoveredSites)
	removedSites := findRemovedSites(currentSitesMap, newSitesMap)

	// Mark newly added sites as discovered
	for _, key := range addedSites {
		rm.discoveredSites[key] = true
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
	ticker := time.NewTicker(rm.tickerInterval)
	defer ticker.Stop()

	siteIndex := 0
	lastTotalSites := 0

	for range ticker.C {
		// Increment ticker fire count for testing
		atomic.AddInt64(&rm.tickerFireCount, 1)
		// Get current sites
		currentSites := rm.collector.GetSites()
		if len(currentSites) == 0 {
			log.Printf("Waiting for sites to be populated before starting metrics refresh...")
			continue
		}

		// If this is the first time we have sites, log the configuration
		if lastTotalSites == 0 && len(currentSites) > 0 {
			totalSites := len(currentSites)
			refreshMinutes := rm.refreshInterval.Minutes()
			sitesPerMinute := int(math.Ceil(float64(totalSites) / refreshMinutes))
			log.Printf("Metrics refresh: processing %d sites per minute (%d sites total, %.0f minute interval)",
				sitesPerMinute, totalSites, refreshMinutes)
		}

		// Recalculate sites per minute in case site count has changed
		totalSites := len(currentSites)
		refreshMinutes := rm.refreshInterval.Minutes()
		sitesPerMinute := int(math.Ceil(float64(totalSites) / refreshMinutes))

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

		lastTotalSites = totalSites
	}
}

// refreshSiteMetrics refreshes metrics for a single site
func (rm *RefreshManager) refreshSiteMetrics(accountID, siteName string) {
	// Find the token for this account from the mapping
	token, ok := rm.accountTokenMap[accountID]
	if !ok {
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
