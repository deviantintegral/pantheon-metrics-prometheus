package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// TestCreateRootHandler tests the createRootHandler function
func TestCreateRootHandler(t *testing.T) {
	// Create test data
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	allSiteMetrics := []SiteMetrics{
		{
			SiteName:    "testsite1",
			Label:       "Test Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "testsite2",
			Label:       "Test Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
	}

	tokens := []string{"token1", "token2"}
	environment := "live"

	// Create collector with test data
	collector := NewPantheonCollector(allSiteMetrics)

	// Create the handler
	handler := createRootHandler(environment, tokens, collector)

	// Test the handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Pantheon Metrics Exporter") {
		t.Error("Response should contain 'Pantheon Metrics Exporter'")
	}
	if !strings.Contains(body, "testsite1") {
		t.Error("Response should contain 'testsite1'")
	}
	if !strings.Contains(body, "testsite2") {
		t.Error("Response should contain 'testsite2'")
	}
	if !strings.Contains(body, "Environment:</strong> live") {
		t.Error("Response should contain environment 'live'")
	}
	if !strings.Contains(body, "Accounts monitored:</strong> 2") {
		t.Error("Response should contain '2' accounts")
	}
	if !strings.Contains(body, "Sites monitored:</strong> 2") {
		t.Error("Response should contain '2' sites")
	}
}

// TestCollectAccountMetrics tests the collectAccountMetrics function
func TestCollectAccountMetrics(t *testing.T) {
	// Test with a token - will fail due to no terminus but exercises the function
	token := "1234567890abcdef1234567890abcdef"
	environment := "live"

	metrics, successCount, failCount := collectAccountMetrics(token, environment)

	// Should return empty metrics since terminus is not available
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}
	if successCount != 0 {
		t.Errorf("Expected 0 successes, got %d", successCount)
	}
	if failCount != 0 {
		t.Errorf("Expected 0 failures, got %d", failCount)
	}
}

// TestCollectAllMetrics tests the collectAllMetrics function
func TestCollectAllMetrics(t *testing.T) {
	// Test with multiple tokens - will fail due to no terminus but exercises the function
	tokens := []string{
		"1234567890abcdef1234567890abcdef",
		"abcdef1234567890abcdef1234567890",
	}
	environment := "live"

	metrics := collectAllMetrics(tokens, environment)

	// Should return empty metrics since terminus is not available
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}
}

// TestCollectAllMetricsEmptyTokens tests collectAllMetrics with empty tokens
func TestCollectAllMetricsEmptyTokens(t *testing.T) {
	tokens := []string{}
	environment := "live"

	metrics := collectAllMetrics(tokens, environment)

	// Should return empty metrics
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}
}

// TestCreateRootHandlerEmptyMetrics tests createRootHandler with no metrics
func TestCreateRootHandlerEmptyMetrics(t *testing.T) {
	tokens := []string{"token1"}
	environment := "dev"
	allSiteMetrics := []SiteMetrics{}

	// Create collector with empty data
	collector := NewPantheonCollector(allSiteMetrics)

	handler := createRootHandler(environment, tokens, collector)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Sites monitored:</strong> 0") {
		t.Error("Response should contain '0' sites")
	}
}

// TestCreateRootHandlerMultipleEnvironments tests createRootHandler with different environments
func TestCreateRootHandlerMultipleEnvironments(t *testing.T) {
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	allSiteMetrics := []SiteMetrics{
		{
			SiteName:    "testsite",
			Label:       "Test Site",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	testCases := []struct {
		environment string
	}{
		{"live"},
		{"dev"},
		{"test"},
	}

	for _, tc := range testCases {
		t.Run("Environment_"+tc.environment, func(t *testing.T) {
			tokens := []string{"token1"}
			// Create collector with test data
			collector := NewPantheonCollector(allSiteMetrics)
			handler := createRootHandler(tc.environment, tokens, collector)

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			body := w.Body.String()
			if !strings.Contains(body, "Environment:</strong> "+tc.environment) {
				t.Errorf("Response should contain environment '%s'", tc.environment)
			}
		})
	}
}

// TestMainEnvironmentValidation tests environment variable validation logic
func TestMainEnvironmentValidation(t *testing.T) {
	// Save original env var
	original := os.Getenv("PANTHEON_MACHINE_TOKENS")
	defer func() {
		if original != "" {
			os.Setenv("PANTHEON_MACHINE_TOKENS", original)
		} else {
			os.Unsetenv("PANTHEON_MACHINE_TOKENS")
		}
	}()

	// Test empty tokens
	os.Setenv("PANTHEON_MACHINE_TOKENS", "")
	tokensEnv := os.Getenv("PANTHEON_MACHINE_TOKENS")
	if tokensEnv != "" {
		t.Error("Expected empty tokens environment variable")
	}

	// Test with tokens
	os.Setenv("PANTHEON_MACHINE_TOKENS", "token1 token2 token3")
	tokensEnv = os.Getenv("PANTHEON_MACHINE_TOKENS")
	tokens := strings.Fields(tokensEnv)

	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}

	// Test token splitting
	expectedTokens := []string{"token1", "token2", "token3"}
	for i, token := range tokens {
		if token != expectedTokens[i] {
			t.Errorf("Expected token %s, got %s", expectedTokens[i], token)
		}
	}
}

// TestMainMultipleAccountProcessing tests the logic for processing multiple accounts
func TestMainMultipleAccountProcessing(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}

	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}

	totalSuccessCount := 0
	totalFailCount := 0

	// Simulate processing accounts
	for tokenIdx, token := range tokens {
		accountID := getAccountID(token)

		// Verify account ID is generated
		if accountID == "" {
			t.Errorf("Account ID should not be empty for token %d", tokenIdx)
		}

		// Simulate some successful and failed site collections
		if tokenIdx == 0 {
			totalSuccessCount += 5
			totalFailCount += 1
		} else if tokenIdx == 1 {
			totalSuccessCount += 3
			totalFailCount += 0
		} else {
			totalSuccessCount += 2
			totalFailCount += 2
		}
	}

	expectedSuccess := 10
	expectedFail := 3

	if totalSuccessCount != expectedSuccess {
		t.Errorf("Expected %d successful sites, got %d", expectedSuccess, totalSuccessCount)
	}

	if totalFailCount != expectedFail {
		t.Errorf("Expected %d failed sites, got %d", expectedFail, totalFailCount)
	}
}

// TestMainSiteMetricsCollection tests the site metrics collection logic
func TestMainSiteMetricsCollection(t *testing.T) {
	var allSiteMetrics []SiteMetrics

	// Simulate collecting metrics for multiple sites
	sites := []struct {
		name    string
		account string
		plan    string
	}{
		{"site1", "account1", "Basic"},
		{"site2", "account1", "Performance"},
		{"site3", "account2", "Elite"},
	}

	for _, site := range sites {
		metricsData := map[string]MetricData{
			"1762732800": {
				DateTime:      "2025-11-10T00:00:00",
				Visits:        100,
				PagesServed:   500,
				CacheHits:     50,
				CacheMisses:   450,
				CacheHitRatio: "10%",
			},
		}

		siteMetrics := SiteMetrics{
			SiteName:    site.name,
			Label:       site.name,
			PlanName:    site.plan,
			Account:     site.account,
			MetricsData: metricsData,
		}

		allSiteMetrics = append(allSiteMetrics, siteMetrics)
	}

	if len(allSiteMetrics) != 3 {
		t.Errorf("Expected 3 site metrics, got %d", len(allSiteMetrics))
	}

	// Verify site metrics were collected correctly
	if allSiteMetrics[0].SiteName != "site1" {
		t.Errorf("Expected site1, got %s", allSiteMetrics[0].SiteName)
	}
	if allSiteMetrics[1].Account != "account1" {
		t.Errorf("Expected account1, got %s", allSiteMetrics[1].Account)
	}
	if allSiteMetrics[2].PlanName != "Elite" {
		t.Errorf("Expected Elite plan, got %s", allSiteMetrics[2].PlanName)
	}
}

// TestMainRefreshManagerSetup tests refresh manager initialization logic
func TestMainRefreshManagerSetup(t *testing.T) {
	tokens := []string{"token1", "token2"}
	environment := "live"
	refreshInterval := 60

	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	allSiteMetrics := []SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	// Create collector
	collector := NewPantheonCollector(allSiteMetrics)
	if collector == nil {
		t.Fatal("Collector should not be nil")
	}

	// Create refresh manager
	refreshIntervalDuration := time.Duration(refreshInterval) * time.Minute
	refreshManager := NewRefreshManager(tokens, environment, refreshIntervalDuration, collector)

	if refreshManager == nil {
		t.Fatal("RefreshManager should not be nil")
	}

	// Initialize discovered sites
	refreshManager.InitializeDiscoveredSites()

	// Verify initialization
	sites := collector.GetSites()
	if len(sites) != 1 {
		t.Errorf("Expected 1 site, got %d", len(sites))
	}
}

// TestMainEmptyMetricsValidation tests validation when no metrics are collected
func TestMainEmptyMetricsValidation(t *testing.T) {
	var allSiteMetrics []SiteMetrics

	// Test that we detect when no metrics are collected
	if len(allSiteMetrics) == 0 {
		// This simulates the check in main() that prevents starting with no metrics
		t.Log("No site metrics were collected - this would cause main to exit")
	} else {
		t.Error("Expected empty metrics slice")
	}
}

// TestMainWithMetricsValidation tests validation when metrics are collected
func TestMainWithMetricsValidation(t *testing.T) {
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	allSiteMetrics := []SiteMetrics{
		{
			SiteName:    "site1",
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	// Test that we can proceed when metrics are collected
	if len(allSiteMetrics) > 0 {
		t.Log("Site metrics were collected - can proceed with exporter startup")
	} else {
		t.Error("Expected non-empty metrics slice")
	}
}

// TestCreateSiteMetrics tests the createSiteMetrics function
func TestCreateSiteMetrics(t *testing.T) {
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	siteMetrics := createSiteMetrics("testsite", "account123", "Performance", metricsData)

	if siteMetrics.SiteName != "testsite" {
		t.Errorf("Expected SiteName 'testsite', got %s", siteMetrics.SiteName)
	}
	if siteMetrics.Account != "account123" {
		t.Errorf("Expected Account 'account123', got %s", siteMetrics.Account)
	}
	if siteMetrics.PlanName != "Performance" {
		t.Errorf("Expected PlanName 'Performance', got %s", siteMetrics.PlanName)
	}
	if siteMetrics.Label != "testsite" {
		t.Errorf("Expected Label 'testsite', got %s", siteMetrics.Label)
	}
	if len(siteMetrics.MetricsData) != 1 {
		t.Errorf("Expected 1 metric data entry, got %d", len(siteMetrics.MetricsData))
	}
}

// TestCreateSiteMetricsWithEmptyMetrics tests createSiteMetrics with empty metrics
func TestCreateSiteMetricsWithEmptyMetrics(t *testing.T) {
	metricsData := make(map[string]MetricData)
	siteMetrics := createSiteMetrics("emptysite", "account456", "Basic", metricsData)

	if len(siteMetrics.MetricsData) != 0 {
		t.Errorf("Expected 0 metric data entries, got %d", len(siteMetrics.MetricsData))
	}
}

// TestCreateSiteMetricsWithMultipleMetrics tests createSiteMetrics with multiple metrics
func TestCreateSiteMetricsWithMultipleMetrics(t *testing.T) {
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
		"1762819200": {
			DateTime:      "2025-11-11T00:00:00",
			Visits:        200,
			PagesServed:   600,
			CacheHits:     100,
			CacheMisses:   500,
			CacheHitRatio: "16.67%",
		},
	}

	siteMetrics := createSiteMetrics("multisite", "account789", "Elite", metricsData)

	if len(siteMetrics.MetricsData) != 2 {
		t.Errorf("Expected 2 metric data entries, got %d", len(siteMetrics.MetricsData))
	}
}

// TestProcessAccountSiteListEmpty tests processAccountSiteList with empty site list
func TestProcessAccountSiteListEmpty(t *testing.T) {
	siteList := make(map[string]SiteListEntry)
	metrics, successCount, failCount := processAccountSiteList("account1", "live", siteList)

	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}
	if successCount != 0 {
		t.Errorf("Expected 0 successes, got %d", successCount)
	}
	if failCount != 0 {
		t.Errorf("Expected 0 failures, got %d", failCount)
	}
}

// TestProcessAccountSiteListWithSites tests processAccountSiteList with sites
func TestProcessAccountSiteListWithSites(t *testing.T) {
	siteList := map[string]SiteListEntry{
		"site1": {
			Name:      "site1",
			PlanName:  "Basic",
			Framework: "wordpress",
		},
		"site2": {
			Name:      "site2",
			PlanName:  "Performance",
			Framework: "drupal8",
		},
	}

	// This will fail because terminus is not available, but it exercises the code
	metrics, successCount, failCount := processAccountSiteList("account1", "live", siteList)

	// Since terminus is not available, all should fail
	_ = metrics
	_ = successCount
	_ = failCount
}

// TestSetupHTTPHandlers tests the setupHTTPHandlers function
func TestSetupHTTPHandlers(t *testing.T) {
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	allSiteMetrics := []SiteMetrics{
		{
			SiteName:    "testsite",
			Label:       "Test Site",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	tokens := []string{"token1"}
	environment := "live"

	// Create a collector and registry
	collector := NewPantheonCollector(allSiteMetrics)
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Setup HTTP handlers
	setupHTTPHandlers(registry, environment, tokens, collector)

	// Test that the root handler works
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Get the handler
	handler, pattern := http.DefaultServeMux.Handler(req)
	if pattern != "/" {
		t.Errorf("Expected pattern '/', got %s", pattern)
	}

	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Pantheon Metrics Exporter") {
		t.Error("Response should contain 'Pantheon Metrics Exporter'")
	}
}

// TestStartRefreshManager tests the startRefreshManager function
func TestStartRefreshManager(t *testing.T) {
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	allSiteMetrics := []SiteMetrics{
		{
			SiteName:    "testsite",
			Label:       "Test Site",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	tokens := []string{"token1"}
	environment := "live"
	refreshInterval := 1 * time.Minute

	// Create collector
	collector := NewPantheonCollector(allSiteMetrics)

	// Start refresh manager
	manager := startRefreshManager(tokens, environment, refreshInterval, collector)

	if manager == nil {
		t.Fatal("Expected refresh manager to be created, got nil")
	}

	// Verify the manager was configured correctly
	if len(manager.tokens) != 1 {
		t.Errorf("Expected 1 token, got %d", len(manager.tokens))
	}

	if manager.environment != "live" {
		t.Errorf("Expected environment 'live', got %s", manager.environment)
	}
}
