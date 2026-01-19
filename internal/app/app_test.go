package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/collector"
	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	testEnvLive = "live"
)

// TestCreateRootHandler tests the createRootHandler function
func TestCreateRootHandler(t *testing.T) {
	// Create test data
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

	allSiteMetrics := []pantheon.SiteMetrics{
		{
			SiteName:    "testsite1",
			SiteID:      "site-uuid-1",
			Label:       "Test Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
		{
			SiteName:    "testsite2",
			SiteID:      "site-uuid-2",
			Label:       "Test Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData,
		},
	}

	tokens := []string{"token1", "token2"}
	environment := testEnvLive

	// Create collector with test data
	c := collector.NewPantheonCollector(allSiteMetrics)

	// Create the handler
	handler := createRootHandler(environment, tokens, c)

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

// TestCreateRootHandlerEmptyMetrics tests createRootHandler with empty metrics
func TestCreateRootHandlerEmptyMetrics(t *testing.T) {
	allSiteMetrics := []pantheon.SiteMetrics{}
	tokens := []string{"token1"}
	environment := testEnvLive

	c := collector.NewPantheonCollector(allSiteMetrics)
	handler := createRootHandler(environment, tokens, c)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Sites monitored:</strong> 0") {
		t.Error("Response should show 0 sites monitored")
	}
}

// TestCreateRootHandlerMultipleEnvironments tests createRootHandler with different environments
func TestCreateRootHandlerMultipleEnvironments(t *testing.T) {
	tests := []struct {
		name string
		env  string
	}{
		{"live environment", "live"},
		{"dev environment", "dev"},
		{"test environment", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := collector.NewPantheonCollector([]pantheon.SiteMetrics{})
			handler := createRootHandler(tt.env, []string{}, c)

			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			body := w.Body.String()
			if !strings.Contains(body, "Environment:</strong> "+tt.env) {
				t.Errorf("Response should contain environment '%s'", tt.env)
			}
		})
	}
}

// TestCreateSiteMetrics tests the createSiteMetrics function
func TestCreateSiteMetrics(t *testing.T) {
	siteName := "testsite"
	siteID := "site-uuid-123"
	accountID := "account123"
	planName := "Basic"
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

	result := createSiteMetrics(siteName, siteID, accountID, planName, metricsData)

	if result.SiteName != siteName {
		t.Errorf("Expected SiteName %s, got %s", siteName, result.SiteName)
	}
	if result.SiteID != siteID {
		t.Errorf("Expected SiteID %s, got %s", siteID, result.SiteID)
	}
	if result.Account != accountID {
		t.Errorf("Expected Account %s, got %s", accountID, result.Account)
	}
	if result.PlanName != planName {
		t.Errorf("Expected PlanName %s, got %s", planName, result.PlanName)
	}
	if result.Label != siteName {
		t.Errorf("Expected Label %s, got %s", siteName, result.Label)
	}
	if len(result.MetricsData) != len(metricsData) {
		t.Errorf("Expected %d metrics entries, got %d", len(metricsData), len(result.MetricsData))
	}
}

// TestCreateSiteMetricsWithEmptyMetrics tests createSiteMetrics with empty metrics
func TestCreateSiteMetricsWithEmptyMetrics(t *testing.T) {
	siteName := "testsite"
	siteID := "site-uuid-123"
	accountID := "account123"
	planName := "Basic"
	metricsData := map[string]pantheon.MetricData{}

	result := createSiteMetrics(siteName, siteID, accountID, planName, metricsData)

	if len(result.MetricsData) != 0 {
		t.Errorf("Expected empty metrics, got %d entries", len(result.MetricsData))
	}
}

// TestCreateSiteMetricsWithMultipleMetrics tests createSiteMetrics with multiple metric entries
func TestCreateSiteMetricsWithMultipleMetrics(t *testing.T) {
	siteName := "testsite"
	siteID := "site-uuid-123"
	accountID := "account123"
	planName := "Performance"
	metricsData := map[string]pantheon.MetricData{
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
			Visits:        150,
			PagesServed:   600,
			CacheHits:     60,
			CacheMisses:   540,
			CacheHitRatio: "10%",
		},
		"1762905600": {
			DateTime:      "2025-11-12T00:00:00",
			Visits:        200,
			PagesServed:   700,
			CacheHits:     70,
			CacheMisses:   630,
			CacheHitRatio: "10%",
		},
	}

	result := createSiteMetrics(siteName, siteID, accountID, planName, metricsData)

	if len(result.MetricsData) != 3 {
		t.Errorf("Expected 3 metrics entries, got %d", len(result.MetricsData))
	}
}

// TestSetupHTTPHandlers tests the SetupHTTPHandlers function
func TestSetupHTTPHandlers(t *testing.T) {
	registry := prometheus.NewRegistry()
	environment := testEnvLive
	tokens := []string{"token1"}
	c := collector.NewPantheonCollector([]pantheon.SiteMetrics{})

	// This should not panic
	SetupHTTPHandlers(registry, environment, tokens, c)
}

// TestStartRefreshManager tests the StartRefreshManager function
func TestStartRefreshManager(t *testing.T) {
	client := pantheon.NewClient(false)
	tokens := []string{"token1"}
	environment := testEnvLive
	refreshInterval := 1 * time.Minute
	c := collector.NewPantheonCollector([]pantheon.SiteMetrics{})

	// This should not panic and should return a manager
	manager := StartRefreshManager(client, tokens, environment, refreshInterval, c, 0, "")

	if manager == nil {
		t.Error("Expected refresh manager to be created, got nil")
	}
}

// TestStartRefreshManagerWithOrgID tests the StartRefreshManager function with org filter
func TestStartRefreshManagerWithOrgID(t *testing.T) {
	client := pantheon.NewClient(false)
	tokens := []string{"token1"}
	environment := testEnvLive
	refreshInterval := 1 * time.Minute
	c := collector.NewPantheonCollector([]pantheon.SiteMetrics{})
	orgID := "org-uuid-12345"

	// This should not panic and should return a manager
	manager := StartRefreshManager(client, tokens, environment, refreshInterval, c, 0, orgID)

	if manager == nil {
		t.Error("Expected refresh manager to be created, got nil")
	}
}

// TestInitialMetricsDurationConstant tests that the constant is set correctly
func TestInitialMetricsDurationConstant(t *testing.T) {
	if InitialMetricsDuration != "28d" {
		t.Errorf("Expected InitialMetricsDuration to be '28d', got '%s'", InitialMetricsDuration)
	}
}

// TestCollectAllSiteListsEmptyTokens tests CollectAllSiteLists with empty tokens
func TestCollectAllSiteListsEmptyTokens(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{}

	result, tokenSiteData := CollectAllSiteLists(ctx, client, tokens, 0, "")

	if len(result) != 0 {
		t.Errorf("Expected 0 sites with empty tokens, got %d", len(result))
	}
	if len(tokenSiteData) != 0 {
		t.Errorf("Expected 0 token site data entries with empty tokens, got %d", len(tokenSiteData))
	}
}

// TestCollectAllSiteListsInvalidTokens tests CollectAllSiteLists with invalid tokens
// The authentication will fail and the function should handle errors gracefully
func TestCollectAllSiteListsInvalidTokens(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{"invalid-token-1", "invalid-token-2"}

	// This should complete without panic, handling auth failures gracefully
	result, tokenSiteData := CollectAllSiteLists(ctx, client, tokens, 0, "")

	// With invalid tokens, we expect 0 sites (auth will fail for all)
	if len(result) != 0 {
		t.Errorf("Expected 0 sites with invalid tokens, got %d", len(result))
	}
	if len(tokenSiteData) != 0 {
		t.Errorf("Expected 0 token site data entries with invalid tokens, got %d", len(tokenSiteData))
	}
}

// TestCollectAllMetricsEmptyTokens tests CollectAllMetrics with empty tokens
func TestCollectAllMetricsEmptyTokens(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{}
	environment := testEnvLive

	result := CollectAllMetrics(ctx, client, tokens, environment, 0, "")

	if len(result) != 0 {
		t.Errorf("Expected 0 sites with empty tokens, got %d", len(result))
	}
}

// TestCollectAllMetricsInvalidTokens tests CollectAllMetrics with invalid tokens
// The authentication will fail and the function should handle errors gracefully
func TestCollectAllMetricsInvalidTokens(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{"invalid-token"}
	environment := testEnvLive

	// This should complete without panic, handling auth failures gracefully
	result := CollectAllMetrics(ctx, client, tokens, environment, 0, "")

	// With invalid tokens, we expect 0 sites
	if len(result) != 0 {
		t.Errorf("Expected 0 sites with invalid tokens, got %d", len(result))
	}
}

// TestCollectAllMetricsWithSitesEmptyTokens tests CollectAllMetricsWithSites with empty tokens
func TestCollectAllMetricsWithSitesEmptyTokens(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{}
	environment := testEnvLive
	preFetchedSites := map[string]AccountSiteData{}

	result := CollectAllMetricsWithSites(ctx, client, tokens, environment, preFetchedSites, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 sites with empty tokens, got %d", len(result))
	}
}

// TestCollectAllMetricsWithSitesMissingToken tests CollectAllMetricsWithSites when token is not in pre-fetched data
func TestCollectAllMetricsWithSitesMissingToken(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{"token1"}
	environment := testEnvLive
	preFetchedSites := map[string]AccountSiteData{} // Empty, no matching token

	result := CollectAllMetricsWithSites(ctx, client, tokens, environment, preFetchedSites, 0)

	if len(result) != 0 {
		t.Errorf("Expected 0 sites with missing token data, got %d", len(result))
	}
}

// TestCollectAllMetricsWithSitesWithData tests CollectAllMetricsWithSites with pre-fetched data
func TestCollectAllMetricsWithSitesWithData(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	token := "test-token"
	tokens := []string{token}
	environment := testEnvLive
	preFetchedSites := map[string]AccountSiteData{
		token: {
			AccountID: "account1",
			Sites: map[string]pantheon.SiteListEntry{
				"site-uuid-1": {
					Name:     "testsite1",
					ID:       "site-uuid-1",
					PlanName: "Basic",
				},
			},
		},
	}

	// This will fail to fetch metrics (invalid token) but should use the pre-fetched data
	result := CollectAllMetricsWithSites(ctx, client, tokens, environment, preFetchedSites, 0)

	// With invalid token, metrics fetch will fail, so result should be empty
	if len(result) != 0 {
		t.Errorf("Expected 0 sites (metrics fetch fails), got %d", len(result))
	}
}

// TestProcessAccountSiteListEmpty tests processAccountSiteList with empty site list
func TestProcessAccountSiteListEmpty(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	token := "test-token"
	accountID := "test-account"
	environment := testEnvLive
	siteList := map[string]pantheon.SiteListEntry{}

	siteMetrics, successCount, failCount := processAccountSiteList(ctx, client, token, accountID, environment, siteList, 0, 0)

	if len(siteMetrics) != 0 {
		t.Errorf("Expected 0 site metrics with empty site list, got %d", len(siteMetrics))
	}
	if successCount != 0 {
		t.Errorf("Expected 0 success count, got %d", successCount)
	}
	if failCount != 0 {
		t.Errorf("Expected 0 fail count with empty site list, got %d", failCount)
	}
}

// TestProcessAccountSiteListWithSites tests processAccountSiteList with sites
// This will fail to fetch metrics but exercises the error handling path
func TestProcessAccountSiteListWithSites(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	token := "test-token"
	accountID := "test-account"
	environment := testEnvLive
	siteList := map[string]pantheon.SiteListEntry{
		"site-uuid-1": {
			Name:     "testsite1",
			ID:       "site-uuid-1",
			PlanName: "Basic",
		},
		"site-uuid-2": {
			Name:     "testsite2",
			ID:       "site-uuid-2",
			PlanName: "Performance",
		},
	}

	// This will fail to fetch metrics (invalid token) but should not panic
	siteMetrics, successCount, failCount := processAccountSiteList(ctx, client, token, accountID, environment, siteList, 0, 0)

	// Expect 0 successful, 2 failed (can't fetch metrics with invalid token)
	if len(siteMetrics) != 0 {
		t.Errorf("Expected 0 site metrics (auth fails), got %d", len(siteMetrics))
	}
	if successCount != 0 {
		t.Errorf("Expected 0 success count, got %d", successCount)
	}
	if failCount != 2 {
		t.Errorf("Expected 2 fail count, got %d", failCount)
	}
}

// TestCollectAccountMetricsInvalidToken tests collectAccountMetrics with invalid token
func TestCollectAccountMetricsInvalidToken(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	token := "invalid-token"
	environment := testEnvLive

	// This should complete without panic, handling auth failure gracefully
	siteMetrics, successCount, failCount := collectAccountMetrics(ctx, client, token, environment, 0, 0, "")

	// With invalid token, we expect 0 metrics (auth will fail)
	if len(siteMetrics) != 0 {
		t.Errorf("Expected 0 site metrics with invalid token, got %d", len(siteMetrics))
	}
	if successCount != 0 {
		t.Errorf("Expected 0 success count, got %d", successCount)
	}
	if failCount != 0 {
		t.Errorf("Expected 0 fail count (auth fails before site list), got %d", failCount)
	}
}

// TestCollectAllSiteListsWithOrgID tests CollectAllSiteLists with organization filter
func TestCollectAllSiteListsWithOrgID(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{"invalid-token"}
	orgID := "org-uuid-12345"

	// This should complete without panic, handling auth failure gracefully
	result, tokenSiteData := CollectAllSiteLists(ctx, client, tokens, 0, orgID)

	// With invalid tokens, we expect 0 sites (auth will fail)
	if len(result) != 0 {
		t.Errorf("Expected 0 sites with invalid tokens and orgID, got %d", len(result))
	}
	if len(tokenSiteData) != 0 {
		t.Errorf("Expected 0 token site data entries with invalid tokens, got %d", len(tokenSiteData))
	}
}

// TestCollectAllMetricsWithOrgID tests CollectAllMetrics with organization filter
func TestCollectAllMetricsWithOrgID(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	tokens := []string{"invalid-token"}
	environment := testEnvLive
	orgID := "org-uuid-12345"

	// This should complete without panic, handling auth failure gracefully
	result := CollectAllMetrics(ctx, client, tokens, environment, 0, orgID)

	// With invalid tokens, we expect 0 sites
	if len(result) != 0 {
		t.Errorf("Expected 0 sites with invalid tokens and orgID, got %d", len(result))
	}
}

// TestCollectAccountMetricsWithOrgID tests collectAccountMetrics with organization filter
func TestCollectAccountMetricsWithOrgID(t *testing.T) {
	client := pantheon.NewClient(false)
	ctx := context.Background()
	token := "invalid-token"
	environment := testEnvLive
	orgID := "org-uuid-12345"

	// This should complete without panic, handling auth failure gracefully
	siteMetrics, successCount, failCount := collectAccountMetrics(ctx, client, token, environment, 0, 0, orgID)

	// With invalid token, we expect 0 metrics (auth will fail)
	if len(siteMetrics) != 0 {
		t.Errorf("Expected 0 site metrics with invalid token and orgID, got %d", len(siteMetrics))
	}
	if successCount != 0 {
		t.Errorf("Expected 0 success count, got %d", successCount)
	}
	if failCount != 0 {
		t.Errorf("Expected 0 fail count, got %d", failCount)
	}
}
