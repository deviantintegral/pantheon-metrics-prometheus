// Package refresh provides periodic refresh management for site lists and metrics.
package refresh

import (
	"context"
	"log"
	"math"
	"sync/atomic"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/collector"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
)

// RefreshMetricsDuration is used for subsequent metrics refresh (1 day to minimize duplicate data).
const RefreshMetricsDuration = "1d"

// InitialMetricsDuration is used for the first metrics fetch for new sites (28 days of history).
const InitialMetricsDuration = "28d"

// Manager manages periodic refresh of site lists and metrics
type Manager struct {
	client          *pantheon.Client
	tokens          []string
	environment     string
	refreshInterval time.Duration
	collector       *collector.PantheonCollector
	discoveredSites map[string]bool   // Track sites discovered since app start (account:site format)
	accountTokenMap map[string]string // Map from account email to token
	tickerInterval  time.Duration     // Interval for metrics refresh ticker (defaults to 1 minute)
	tickerFireCount int64             // Counter for ticker fires (for testing)
	siteLimit       int               // Maximum number of sites to query (0 = no limit)
	orgID           string            // Organization ID to filter sites (empty for all sites)
}

// NewManager creates a new refresh manager
func NewManager(client *pantheon.Client, tokens []string, environment string, refreshInterval time.Duration, c *collector.PantheonCollector, siteLimit int, orgID string) *Manager {
	return &Manager{
		client:          client,
		tokens:          tokens,
		environment:     environment,
		refreshInterval: refreshInterval,
		collector:       c,
		discoveredSites: make(map[string]bool),
		accountTokenMap: make(map[string]string),
		tickerInterval:  1 * time.Minute, // Default to 1 minute
		siteLimit:       siteLimit,
		orgID:           orgID,
	}
}

// SetTickerInterval sets the ticker interval for metrics refresh (useful for testing)
func (rm *Manager) SetTickerInterval(interval time.Duration) {
	rm.tickerInterval = interval
}

// GetTickerFireCount returns the number of times the ticker has fired (useful for testing)
func (rm *Manager) GetTickerFireCount() int64 {
	return atomic.LoadInt64(&rm.tickerFireCount)
}

// InitializeDiscoveredSites populates the discovered sites map with initial sites
func (rm *Manager) InitializeDiscoveredSites() {
	sites := rm.collector.GetSites()
	for _, site := range sites {
		key := site.Account + ":" + site.SiteName
		rm.discoveredSites[key] = true
	}
	log.Printf("Initialized with %d discovered sites", len(rm.discoveredSites))
}

// InitializeAccountTokenMap authenticates all tokens and populates the account-to-token mapping.
// This must be called before Start() to ensure tokens are available for metrics refresh.
func (rm *Manager) InitializeAccountTokenMap() {
	ctx := context.Background()
	for _, token := range rm.tokens {
		accountID, err := rm.client.Authenticate(ctx, token)
		if err != nil {
			accountID = pantheon.GetAccountID(token)
			log.Printf("Warning: Failed to authenticate account %s during token map initialization: %v", accountID, err)
			continue
		}
		rm.accountTokenMap[accountID] = token
	}
	log.Printf("Initialized account token map with %d accounts", len(rm.accountTokenMap))
}

// Start begins the periodic refresh process
func (rm *Manager) Start() {
	// Start site list refresh (every refresh interval)
	go rm.refreshSiteListsPeriodically()

	// Start metrics refresh with queue-based processing
	go rm.refreshMetricsWithQueue()
}

// refreshSiteListsPeriodically refreshes site lists for all accounts
func (rm *Manager) refreshSiteListsPeriodically() {
	ticker := time.NewTicker(rm.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Starting site list refresh...")
		rm.refreshAllSiteLists()
	}
}

// buildSiteKeyMap creates a map of site keys from a list of sites
func buildSiteKeyMap(sites []pantheon.SiteMetrics) map[string]bool {
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
func (rm *Manager) refreshAllSiteLists() {
	ctx := context.Background()
	var allSiteMetrics []pantheon.SiteMetrics

	// Get current sites to track changes
	existingSites := rm.collector.GetSites()
	currentSitesMap := buildSiteKeyMap(existingSites)

	// Track new sites for this refresh
	newSitesMap := make(map[string]bool)
	totalSitesFound := 0

	// Get existing metrics for sites (do this once outside the loop)
	existingMetricsMap := make(map[string]map[string]pantheon.MetricData)
	for _, site := range existingSites {
		key := site.Account + ":" + site.SiteName
		existingMetricsMap[key] = site.MetricsData
	}

	for _, token := range rm.tokens {
		// Check if we've reached the site limit
		if rm.siteLimit > 0 && len(allSiteMetrics) >= rm.siteLimit {
			break
		}

		// Authenticate with this token
		accountID, err := rm.client.Authenticate(ctx, token)
		if err != nil {
			// Use token suffix as fallback for logging if auth fails
			accountID = pantheon.GetAccountID(token)
			log.Printf("Warning: Failed to authenticate account %s during refresh: %v", accountID, err)
			continue
		}

		// Store the mapping for later use
		rm.accountTokenMap[accountID] = token

		log.Printf("Refreshing site list for account %s", accountID)

		// Fetch all sites for this account (filtered by orgID if provided)
		siteList, err := rm.client.FetchAllSites(ctx, token, rm.orgID)
		if err != nil {
			log.Printf("Warning: Failed to fetch site list for account %s during refresh: %v", accountID, err)
			continue
		}

		totalSitesFound += len(siteList)

		// Create site metrics entries, preserving existing metrics data
		for siteID, site := range siteList {
			// Check if we've reached the site limit
			if rm.siteLimit > 0 && len(allSiteMetrics) >= rm.siteLimit {
				log.Printf("Site limit reached (%d sites), stopping refresh", rm.siteLimit)
				break
			}

			key := accountID + ":" + site.Name
			newSitesMap[key] = true

			metricsData := existingMetricsMap[key]
			if metricsData == nil {
				metricsData = make(map[string]pantheon.MetricData)
			}

			siteMetrics := pantheon.SiteMetrics{
				SiteName:    site.Name,
				SiteID:      siteID,
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
func (rm *Manager) refreshMetricsWithQueue() {
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
		log.Printf("Refreshing metrics for %d sites (sites %d-%d of %d)",
			len(sitesToProcess), siteIndex+1, endIndex, len(currentSites))

		for _, site := range sitesToProcess {
			go rm.refreshSiteMetrics(site.Account, site.SiteName, site.SiteID)
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
func (rm *Manager) refreshSiteMetrics(accountID, siteName, siteID string) {
	ctx := context.Background()

	// Find the token for this account from the mapping
	token, ok := rm.accountTokenMap[accountID]
	if !ok {
		log.Printf("Warning: No token found for account %s", accountID)
		return
	}

	// Determine duration based on whether this site has been fetched before
	duration := RefreshMetricsDuration
	key := accountID + ":" + siteName
	if !rm.discoveredSites[key] {
		// First time fetching this site, use longer duration
		duration = InitialMetricsDuration
		rm.discoveredSites[key] = true
	}

	// Fetch metrics for this site
	metricsData, err := rm.client.FetchMetricsData(ctx, token, siteID, rm.environment, duration)
	if err != nil {
		log.Printf("Warning: Failed to refresh metrics for %s.%s: %v", accountID, siteName, err)
		return
	}

	// Update the collector
	rm.collector.UpdateSiteMetrics(accountID, siteName, metricsData)
	log.Printf("Updated metrics for site %s.%s", accountID, siteName)
}
