package refresh

import (
	"testing"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/collector"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
)

const (
	testEnvLive = "live"
	testEnvDev  = "dev"
	testToken32 = "1234567890abcdef1234567890abcdef"
)

func TestNewManager(t *testing.T) {
	// Test creating a new refresh manager
	tokens := []string{"token1", "token2"}
	environment := testEnvLive
	refreshInterval := 60 * time.Minute

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        837,
			PagesServed:   3081,
			CacheHits:     119,
			CacheMisses:   2962,
			CacheHitRatio: "3.86%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)

	manager := NewManager(tokens, environment, refreshInterval, collector)

	if manager == nil {
		t.Fatal("Expected refresh manager to be created, got nil")
	}

	if len(manager.tokens) != 2 {
		t.Errorf("Expected 2 tokens, got %d", len(manager.tokens))
	}

	if manager.environment != testEnvLive {
		t.Errorf("Expected environment 'live', got %s", manager.environment)
	}

	if manager.refreshInterval != 60*time.Minute {
		t.Errorf("Expected refresh interval 60m, got %v", manager.refreshInterval)
	}

	if manager.collector != collector {
		t.Error("Expected collector to be set")
	}
}

func TestNewManagerWithMultipleTokens(t *testing.T) {
	// Test creating a refresh manager with multiple tokens
	tokens := []string{"token1", "token2", "token3", "token4"}
	environment := testEnvDev
	refreshInterval := 30 * time.Minute

	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)

	manager := NewManager(tokens, environment, refreshInterval, collector)

	if len(manager.tokens) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(manager.tokens))
	}

	if manager.environment != testEnvDev {
		t.Errorf("Expected environment 'dev', got %s", manager.environment)
	}

	if manager.refreshInterval != 30*time.Minute {
		t.Errorf("Expected refresh interval 30m, got %v", manager.refreshInterval)
	}
}

func TestNewManagerWithEmptyTokens(t *testing.T) {
	// Test creating a refresh manager with empty tokens
	tokens := []string{}
	environment := "test"
	refreshInterval := 15 * time.Minute

	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)

	manager := NewManager(tokens, environment, refreshInterval, collector)

	if manager == nil {
		t.Fatal("Expected refresh manager to be created, got nil")
	}

	if len(manager.tokens) != 0 {
		t.Errorf("Expected 0 tokens, got %d", len(manager.tokens))
	}
}

func TestNewManagerWithDifferentIntervals(t *testing.T) {
	// Test creating refresh managers with different intervals
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)

	// Test 5 minutes
	manager1 := NewManager(tokens, environment, 5*time.Minute, collector)
	if manager1.refreshInterval != 5*time.Minute {
		t.Errorf("Expected refresh interval 5m, got %v", manager1.refreshInterval)
	}

	// Test 2 hours
	manager2 := NewManager(tokens, environment, 120*time.Minute, collector)
	if manager2.refreshInterval != 120*time.Minute {
		t.Errorf("Expected refresh interval 120m, got %v", manager2.refreshInterval)
	}

	// Test 1 minute
	manager3 := NewManager(tokens, environment, 1*time.Minute, collector)
	if manager3.refreshInterval != 1*time.Minute {
		t.Errorf("Expected refresh interval 1m, got %v", manager3.refreshInterval)
	}
}

func TestManagerStart(t *testing.T) {
	// Test that Start() launches goroutines without panicking
	tokens := []string{}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// Start should not panic even with empty tokens
	// We don't wait for goroutines to complete as they run indefinitely
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Start() panicked: %v", r)
		}
	}()

	// Just verify Start can be called without panic
	// The goroutines will run in background but won't do anything useful without valid tokens
	manager.Start()

	// Give goroutines a moment to start
	time.Sleep(10 * time.Millisecond)
}

func TestRefreshMetricsWithQueueEmptySites(_ *testing.T) {
	// Test refreshMetricsWithQueue with no sites
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{} // Empty sites
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// This should return immediately since there are no sites
	done := make(chan bool, 1)
	go func() {
		manager.refreshMetricsWithQueue()
		done <- true
	}()

	// Wait a short time to see if function returns quickly
	select {
	case <-done:
		// Good, function returned as expected
	case <-time.After(100 * time.Millisecond):
		// Also acceptable, as the function may enter the ticker loop
	}
}

func TestRefreshSiteMetricsWithInvalidToken(_ *testing.T) {
	// Test refreshSiteMetrics with an account that doesn't have a matching token
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// Try to refresh metrics for a non-existent account
	// This should log a warning and return without panicking
	manager.refreshSiteMetrics("nonexistent", "somesite")

	// If we get here without panic, test passes
}

func TestRefreshAllSiteListsEmptyTokens(t *testing.T) {
	// Test refreshAllSiteLists with empty tokens
	tokens := []string{}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// This should complete without panic even with no tokens
	manager.refreshAllSiteLists()

	// Verify sites are empty
	currentSites := collector.GetSites()
	if len(currentSites) != 0 {
		t.Errorf("Expected 0 sites with empty tokens, got %d", len(currentSites))
	}
}

func TestManagerWithExistingSites(t *testing.T) {
	// Test refresh manager behavior with existing sites in collector
	tokens := []string{"token1", "token2"}
	environment := testEnvDev

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 30*time.Minute, collector)

	// Verify manager has access to existing sites through collector
	currentSites := collector.GetSites()
	if len(currentSites) != 2 {
		t.Errorf("Expected 2 sites, got %d", len(currentSites))
	}

	// Verify manager properties
	if manager.environment != testEnvDev {
		t.Errorf("Expected environment 'dev', got %s", manager.environment)
	}
}

func TestRefreshMetricsWithQueueWithSites(_ *testing.T) {
	// Test refreshMetricsWithQueue with actual sites
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "token1id",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// Start the refresh queue in background
	done := make(chan bool, 1)
	go func() {
		// Let it run for a short time
		time.Sleep(50 * time.Millisecond)
		done <- true
	}()

	// Start the refresh
	go manager.refreshMetricsWithQueue()

	// Wait for timeout
	<-done
}

func TestRefreshAllSiteListsWithExistingSites(_ *testing.T) {
	// Test refreshAllSiteLists when collector already has sites
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	// Start with some existing sites
	existingSites := []pantheon.SiteMetrics{
		{
			SiteName:    "oldsite",
			Label:       "Old Site",
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(existingSites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// Call refreshAllSiteLists
	// This will fail when trying to authenticate, but will exercise the code path
	manager.refreshAllSiteLists()

	// The sites should remain unchanged since authentication will fail
	// This test exercises the code but won't successfully update sites
}

func TestRefreshSiteMetricsWithMatchingToken(_ *testing.T) {
	// Test refreshSiteMetrics with a token that matches via pantheon.GetAccountID
	token := testToken32
	accountID := pantheon.GetAccountID(token) // Should return "90abcdef"

	tokens := []string{token}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// Try to refresh metrics - will fail at authentication but exercises the token lookup path
	manager.refreshSiteMetrics(accountID, "somesite")
}

func TestRefreshSiteListsPeriodically(_ *testing.T) {
	// Test refreshSiteListsPeriodically starts and runs
	tokens := []string{}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 50*time.Millisecond, collector)

	// Start the periodic refresh in background
	done := make(chan bool, 1)
	go func() {
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	// Start refreshSiteListsPeriodically - it will run in background
	go manager.refreshSiteListsPeriodically()

	// Wait briefly
	<-done

	// If we get here without panic, test passes
}

func TestRefreshAllSiteListsMultipleTokens(_ *testing.T) {
	// Test refreshAllSiteLists with multiple tokens
	tokens := []string{"token1", "token2", "token3"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	// Start with some existing sites
	existingSites := []pantheon.SiteMetrics{
		{
			SiteName:    "oldsite1",
			Label:       "Old Site 1",
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		},
		{
			SiteName:    "oldsite2",
			Label:       "Old Site 2",
			PlanName:    "Performance",
			Account:     "token2id",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(existingSites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	// Call refreshAllSiteLists with multiple tokens
	// This will fail when trying to authenticate, but will exercise the loop
	manager.refreshAllSiteLists()
}

// testRefreshMetricsWithQueueHelper is a helper function to test refreshMetricsWithQueue
func testRefreshMetricsWithQueueHelper(interval, sleepDuration time.Duration) {
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "token1id",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site3",
			Label:       "Site 3",
			PlanName:    "Elite",
			Account:     "token1id",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, interval, collector)

	done := make(chan bool, 1)
	go func() {
		time.Sleep(sleepDuration)
		done <- true
	}()

	go manager.refreshMetricsWithQueue()

	<-done
}

func TestRefreshMetricsWithQueueLongInterval(_ *testing.T) {
	// Test refreshMetricsWithQueue with a longer interval
	testRefreshMetricsWithQueueHelper(3*time.Minute, 50*time.Millisecond)
}

func TestRefreshMetricsWithQueueManySites(_ *testing.T) {
	// Test refreshMetricsWithQueue with many sites to exercise batching
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	// Create many sites to test the queue batching and cycling
	sites := make([]pantheon.SiteMetrics, 10)
	for i := 0; i < 10; i++ {
		sites[i] = pantheon.SiteMetrics{
			SiteName:    "site" + string(rune('0'+i)),
			Label:       "Site " + string(rune('0'+i)),
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		}
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 5*time.Minute, collector)

	// Start the refresh queue in background
	done := make(chan bool, 1)
	go func() {
		// Let it run for a short time to exercise the logic
		time.Sleep(50 * time.Millisecond)
		done <- true
	}()

	// Start the refresh
	go manager.refreshMetricsWithQueue()

	// Wait for timeout
	<-done
}

func TestRefreshMetricsWithQueueShortInterval(_ *testing.T) {
	// Test refreshMetricsWithQueue with a very short interval to exercise ticker logic
	testRefreshMetricsWithQueueHelper(3*time.Minute, 10*time.Millisecond)
}

func TestRefreshMetricsWithQueueTickerFires(t *testing.T) {
	// Test that the ticker actually fires and processes sites
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 3*time.Minute, collector)

	// Use a short ticker interval for testing (2 seconds)
	manager.SetTickerInterval(2 * time.Second)

	// Start refresh queue
	go manager.refreshMetricsWithQueue()

	// Wait for ticker to fire at least twice (5 seconds should be enough for 2 fires at 2s interval)
	time.Sleep(5 * time.Second)

	// Verify the ticker fired at least twice
	fireCount := manager.GetTickerFireCount()
	if fireCount < 2 {
		t.Errorf("Expected ticker to fire at least 2 times, but it fired %d times", fireCount)
	}

	// Verify the ticker fired but not too many times (should be 2-3 fires in 5 seconds with 2s interval)
	if fireCount > 4 {
		t.Errorf("Expected ticker to fire 2-3 times in 5 seconds, but it fired %d times", fireCount)
	}

	t.Logf("Ticker fired %d times in 5 seconds (expected 2-3)", fireCount)
}

func TestInitializeDiscoveredSites(t *testing.T) {
	// Test InitializeDiscoveredSites with no sites
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	manager.InitializeDiscoveredSites()

	if len(manager.discoveredSites) != 0 {
		t.Errorf("Expected 0 discovered sites, got %d", len(manager.discoveredSites))
	}
}

func TestInitializeDiscoveredSitesWithSites(t *testing.T) {
	// Test InitializeDiscoveredSites with multiple sites
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site3",
			Label:       "Site 3",
			PlanName:    "Elite",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	manager.InitializeDiscoveredSites()

	expectedSites := 3
	if len(manager.discoveredSites) != expectedSites {
		t.Errorf("Expected %d discovered sites, got %d", expectedSites, len(manager.discoveredSites))
	}

	// Verify the site keys are correct
	expectedKeys := map[string]bool{
		"account1:site1": true,
		"account2:site2": true,
		"account1:site3": true,
	}

	for key := range expectedKeys {
		if !manager.discoveredSites[key] {
			t.Errorf("Expected discovered site key %s not found", key)
		}
	}
}

func TestInitializeDiscoveredSitesDuplicateAccounts(t *testing.T) {
	// Test InitializeDiscoveredSites with multiple sites from same account
	tokens := []string{"token1"}
	environment := testEnvLive

	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "sameaccount",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "sameaccount",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(tokens, environment, 1*time.Minute, collector)

	manager.InitializeDiscoveredSites()

	if len(manager.discoveredSites) != 2 {
		t.Errorf("Expected 2 discovered sites, got %d", len(manager.discoveredSites))
	}

	// Verify both sites are tracked
	if !manager.discoveredSites["sameaccount:site1"] {
		t.Error("Expected site1 to be discovered")
	}
	if !manager.discoveredSites["sameaccount:site2"] {
		t.Error("Expected site2 to be discovered")
	}
}

func TestBuildSiteKeyMap(t *testing.T) {
	// Test building a site key map from a list of sites
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site3",
			Label:       "Site 3",
			PlanName:    "Elite",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	siteMap := buildSiteKeyMap(sites)

	expectedKeys := []string{"account1:site1", "account2:site2", "account1:site3"}
	if len(siteMap) != 3 {
		t.Errorf("Expected 3 entries in site map, got %d", len(siteMap))
	}

	for _, key := range expectedKeys {
		if !siteMap[key] {
			t.Errorf("Expected site key %s to be in map", key)
		}
	}
}

func TestBuildSiteKeyMapEmpty(t *testing.T) {
	// Test building a site key map from an empty list
	sites := []pantheon.SiteMetrics{}
	siteMap := buildSiteKeyMap(sites)

	if len(siteMap) != 0 {
		t.Errorf("Expected empty site map, got %d entries", len(siteMap))
	}
}

func TestFindAddedSites(t *testing.T) {
	// Test finding added sites
	currentSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	newSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
		"account1:site3": true,
		"account2:site4": true,
	}

	discoveredSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	addedSites := findAddedSites(currentSites, newSites, discoveredSites)

	expectedAdded := []string{"account1:site3", "account2:site4"}
	if len(addedSites) != 2 {
		t.Errorf("Expected 2 added sites, got %d", len(addedSites))
	}

	for _, key := range expectedAdded {
		found := false
		for _, added := range addedSites {
			if added == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find added site %s", key)
		}
	}
}

func TestFindAddedSitesNone(t *testing.T) {
	// Test when no sites are added
	currentSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	newSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	discoveredSites := map[string]bool{}

	addedSites := findAddedSites(currentSites, newSites, discoveredSites)

	if len(addedSites) != 0 {
		t.Errorf("Expected 0 added sites, got %d", len(addedSites))
	}
}

func TestFindAddedSitesAlreadyDiscovered(t *testing.T) {
	// Test when new sites were already discovered before
	currentSites := map[string]bool{
		"account1:site1": true,
	}

	newSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	discoveredSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true, // Already discovered
	}

	addedSites := findAddedSites(currentSites, newSites, discoveredSites)

	// site2 is new to currentSites but was already in discoveredSites, so it shouldn't be added
	if len(addedSites) != 0 {
		t.Errorf("Expected 0 added sites, got %d", len(addedSites))
	}
}

func TestFindRemovedSites(t *testing.T) {
	// Test finding removed sites
	currentSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
		"account1:site3": true,
	}

	newSites := map[string]bool{
		"account1:site1": true,
	}

	removedSites := findRemovedSites(currentSites, newSites)

	expectedRemoved := []string{"account1:site2", "account1:site3"}
	if len(removedSites) != 2 {
		t.Errorf("Expected 2 removed sites, got %d", len(removedSites))
	}

	for _, key := range expectedRemoved {
		found := false
		for _, removed := range removedSites {
			if removed == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find removed site %s", key)
		}
	}
}

func TestFindRemovedSitesNone(t *testing.T) {
	// Test when no sites are removed
	currentSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	newSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
		"account1:site3": true,
	}

	removedSites := findRemovedSites(currentSites, newSites)

	if len(removedSites) != 0 {
		t.Errorf("Expected 0 removed sites, got %d", len(removedSites))
	}
}

func TestFindRemovedSitesAll(t *testing.T) {
	// Test when all sites are removed
	currentSites := map[string]bool{
		"account1:site1": true,
		"account1:site2": true,
	}

	newSites := map[string]bool{}

	removedSites := findRemovedSites(currentSites, newSites)

	if len(removedSites) != 2 {
		t.Errorf("Expected 2 removed sites, got %d", len(removedSites))
	}
}
