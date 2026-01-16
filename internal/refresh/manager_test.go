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
	client := pantheon.NewClient()
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
			SiteID:      "site-uuid-1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)

	manager := NewManager(client, tokens, environment, refreshInterval, collector)

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

	if manager.client != client {
		t.Error("Expected client to be set")
	}
}

func TestNewManagerWithMultipleTokens(t *testing.T) {
	client := pantheon.NewClient()
	tokens := []string{"token1", "token2", "token3", "token4"}
	environment := testEnvDev
	refreshInterval := 30 * time.Minute

	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)

	manager := NewManager(client, tokens, environment, refreshInterval, collector)

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
	client := pantheon.NewClient()
	tokens := []string{}
	environment := "test"
	refreshInterval := 15 * time.Minute

	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)

	manager := NewManager(client, tokens, environment, refreshInterval, collector)

	if manager == nil {
		t.Fatal("Expected refresh manager to be created, got nil")
	}

	if len(manager.tokens) != 0 {
		t.Errorf("Expected 0 tokens, got %d", len(manager.tokens))
	}
}

func TestNewManagerWithDifferentIntervals(t *testing.T) {
	client := pantheon.NewClient()
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)

	// Test 5 minutes
	manager1 := NewManager(client, tokens, environment, 5*time.Minute, collector)
	if manager1.refreshInterval != 5*time.Minute {
		t.Errorf("Expected refresh interval 5m, got %v", manager1.refreshInterval)
	}

	// Test 2 hours
	manager2 := NewManager(client, tokens, environment, 120*time.Minute, collector)
	if manager2.refreshInterval != 120*time.Minute {
		t.Errorf("Expected refresh interval 120m, got %v", manager2.refreshInterval)
	}

	// Test 1 minute
	manager3 := NewManager(client, tokens, environment, 1*time.Minute, collector)
	if manager3.refreshInterval != 1*time.Minute {
		t.Errorf("Expected refresh interval 1m, got %v", manager3.refreshInterval)
	}
}

func TestManagerStart(t *testing.T) {
	client := pantheon.NewClient()
	tokens := []string{}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 1*time.Minute, collector)

	// Start should not panic even with empty tokens
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Start() panicked: %v", r)
		}
	}()

	manager.Start()

	// Give goroutines a moment to start
	time.Sleep(10 * time.Millisecond)
}

func TestRefreshSiteMetricsWithInvalidToken(t *testing.T) {
	client := pantheon.NewClient()
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 1*time.Minute, collector)

	// Try to refresh metrics for a non-existent account
	// This should log a warning and return without panicking
	manager.refreshSiteMetrics("nonexistent", "somesite", "site-uuid")

	// If we get here without panic, test passes
}

func TestRefreshAllSiteListsEmptyTokens(t *testing.T) {
	client := pantheon.NewClient()
	tokens := []string{}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 1*time.Minute, collector)

	// This should complete without panic even with no tokens
	manager.refreshAllSiteLists()

	// Verify sites are empty
	currentSites := collector.GetSites()
	if len(currentSites) != 0 {
		t.Errorf("Expected 0 sites with empty tokens, got %d", len(currentSites))
	}
}

func TestManagerWithExistingSites(t *testing.T) {
	client := pantheon.NewClient()
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
			SiteID:      "site-uuid-1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			SiteID:      "site-uuid-2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 30*time.Minute, collector)

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

func TestRefreshMetricsWithQueueTickerFires(t *testing.T) {
	client := pantheon.NewClient()
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
			SiteID:      "site-uuid-1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "token1id",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 3*time.Minute, collector)

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
	client := pantheon.NewClient()
	tokens := []string{"token1"}
	environment := testEnvLive
	sites := []pantheon.SiteMetrics{}
	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 1*time.Minute, collector)

	manager.InitializeDiscoveredSites()

	if len(manager.discoveredSites) != 0 {
		t.Errorf("Expected 0 discovered sites, got %d", len(manager.discoveredSites))
	}
}

func TestInitializeDiscoveredSitesWithSites(t *testing.T) {
	client := pantheon.NewClient()
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
			SiteID:      "site-uuid-1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			SiteID:      "site-uuid-2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site3",
			SiteID:      "site-uuid-3",
			Label:       "Site 3",
			PlanName:    "Elite",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := collector.NewPantheonCollector(sites)
	manager := NewManager(client, tokens, environment, 1*time.Minute, collector)

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

func TestBuildSiteKeyMap(t *testing.T) {
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
			SiteID:      "site-uuid-1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site2",
			SiteID:      "site-uuid-2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
		{
			SiteName:    "site3",
			SiteID:      "site-uuid-3",
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
	sites := []pantheon.SiteMetrics{}
	siteMap := buildSiteKeyMap(sites)

	if len(siteMap) != 0 {
		t.Errorf("Expected empty site map, got %d entries", len(siteMap))
	}
}

func TestFindAddedSites(t *testing.T) {
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

func TestFindRemovedSites(t *testing.T) {
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

func TestDurationConstants(t *testing.T) {
	if RefreshMetricsDuration != "1d" {
		t.Errorf("Expected RefreshMetricsDuration to be '1d', got '%s'", RefreshMetricsDuration)
	}

	if InitialMetricsDuration != "28d" {
		t.Errorf("Expected InitialMetricsDuration to be '28d', got '%s'", InitialMetricsDuration)
	}
}
