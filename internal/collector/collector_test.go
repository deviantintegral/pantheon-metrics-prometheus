package collector

import (
	"testing"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	testCollectorSite1 = "site1"
)

func TestNewPantheonCollector(t *testing.T) {
	// Test creating a new collector with multiple sites
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
			SiteName:    testCollectorSite1,
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

	collector := NewPantheonCollector(sites)
	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if len(collector.sites) != 2 {
		t.Errorf("Expected 2 sites in collector, got %d", len(collector.sites))
	}
}

func TestDescribe(t *testing.T) {
	// Test Describe method sends all metric descriptors
	sites := []pantheon.SiteMetrics{}
	collector := NewPantheonCollector(sites)

	ch := make(chan *prometheus.Desc, 5)
	collector.Describe(ch)
	close(ch)

	// Count the descriptors sent
	count := 0
	for range ch {
		count++
	}

	// Should have 5 metric descriptors (visits, pages_served, cache_hits, cache_misses, cache_hit_ratio)
	if count != 5 {
		t.Errorf("Expected 5 metric descriptors, got %d", count)
	}
}

func TestCollect(t *testing.T) {
	// Test Collect method collects metrics from all sites
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        837,
			PagesServed:   3081,
			CacheHits:     119,
			CacheMisses:   2962,
			CacheHitRatio: "3.86%",
		},
		"1762819200": {
			DateTime:      "2025-11-11T00:00:00",
			Visits:        824,
			PagesServed:   2950,
			CacheHits:     151,
			CacheMisses:   2799,
			CacheHitRatio: "5.12%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 20)
	collector.Collect(ch)
	close(ch)

	// Count the metrics sent
	count := 0
	for range ch {
		count++
	}

	// Should have 10 metrics (5 metric types × 2 timestamps)
	if count != 10 {
		t.Errorf("Expected 10 metrics, got %d", count)
	}
}

func TestCollectWithMultipleSites(t *testing.T) {
	// Test Collect with multiple sites
	metricsData1 := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        837,
			PagesServed:   3081,
			CacheHits:     119,
			CacheMisses:   2962,
			CacheHitRatio: "3.86%",
		},
	}

	metricsData2 := map[string]pantheon.MetricData{
		"1762819200": {
			DateTime:      "2025-11-11T00:00:00",
			Visits:        500,
			PagesServed:   2000,
			CacheHits:     100,
			CacheMisses:   1900,
			CacheHitRatio: "5.00%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData1,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: metricsData2,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 20)
	collector.Collect(ch)
	close(ch)

	// Count the metrics sent
	count := 0
	for range ch {
		count++
	}

	// Should have 10 metrics (5 metric types × 2 sites with 1 timestamp each)
	if count != 10 {
		t.Errorf("Expected 10 metrics, got %d", count)
	}
}

func TestCollectWithInvalidTimestamp(t *testing.T) {
	// Test Collect handles invalid timestamps gracefully
	metricsData := map[string]pantheon.MetricData{
		"invalid_timestamp": {
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
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should have 0 metrics due to invalid timestamp
	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 metrics due to invalid timestamp, got %d", count)
	}
}

func TestCollectWithInvalidCacheHitRatio(t *testing.T) {
	// Test Collect handles invalid cache hit ratio gracefully
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        837,
			PagesServed:   3081,
			CacheHits:     119,
			CacheMisses:   2962,
			CacheHitRatio: "invalid%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should still collect metrics, but cache hit ratio will be 0
	count := 0
	for range ch {
		count++
	}

	// Should have 5 metrics (one for each metric type)
	if count != 5 {
		t.Errorf("Expected 5 metrics, got %d", count)
	}
}

func TestUpdateSites(t *testing.T) {
	// Test UpdateSites method
	initialMetrics := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	initialSites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: initialMetrics,
		},
	}

	collector := NewPantheonCollector(initialSites)

	// Create new sites to update
	newMetrics := map[string]pantheon.MetricData{
		"1762819200": {
			DateTime:      "2025-11-11T00:00:00",
			Visits:        200,
			PagesServed:   1000,
			CacheHits:     100,
			CacheMisses:   900,
			CacheHitRatio: "10%",
		},
	}

	newSites := []pantheon.SiteMetrics{
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: newMetrics,
		},
		{
			SiteName:    "site3",
			Label:       "Site 3",
			PlanName:    "Elite",
			Account:     "account3",
			MetricsData: newMetrics,
		},
	}

	// Update sites
	collector.UpdateSites(newSites)

	// Get sites to verify update
	sites := collector.GetSites()
	if len(sites) != 2 {
		t.Errorf("Expected 2 sites after update, got %d", len(sites))
	}

	if sites[0].SiteName != "site2" {
		t.Errorf("Expected first site to be site2, got %s", sites[0].SiteName)
	}

	if sites[1].SiteName != "site3" {
		t.Errorf("Expected second site to be site3, got %s", sites[1].SiteName)
	}
}

func TestGetSites(t *testing.T) {
	// Test GetSites returns a copy of sites
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
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	// Get sites
	retrievedSites := collector.GetSites()

	// Verify we get a copy
	if len(retrievedSites) != 1 {
		t.Errorf("Expected 1 site, got %d", len(retrievedSites))
	}

	if retrievedSites[0].SiteName != testCollectorSite1 {
		t.Errorf("Expected site name 'site1', got %s", retrievedSites[0].SiteName)
	}

	// Modify the retrieved copy - should not affect original
	retrievedSites[0].SiteName = "modified"

	// Get sites again and verify original is unchanged
	retrievedSites2 := collector.GetSites()
	if retrievedSites2[0].SiteName != testCollectorSite1 {
		t.Errorf("Expected original site name 'site1' to be unchanged, got %s", retrievedSites2[0].SiteName)
	}
}

func TestUpdateSiteMetrics(t *testing.T) {
	// Test UpdateSiteMetrics updates metrics for a specific site
	initialMetrics := map[string]pantheon.MetricData{
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
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: initialMetrics,
		},
		{
			SiteName:    "site2",
			Label:       "Site 2",
			PlanName:    "Performance",
			Account:     "account2",
			MetricsData: initialMetrics,
		},
	}

	collector := NewPantheonCollector(sites)

	// Create new metrics to update
	newMetrics := map[string]pantheon.MetricData{
		"1762819200": {
			DateTime:      "2025-11-11T00:00:00",
			Visits:        200,
			PagesServed:   1000,
			CacheHits:     100,
			CacheMisses:   900,
			CacheHitRatio: "10%",
		},
	}

	// Update metrics for site1
	collector.UpdateSiteMetrics("account1", testCollectorSite1, newMetrics)

	// Get sites and verify only site1 was updated
	updatedSites := collector.GetSites()

	if len(updatedSites[0].MetricsData) != 1 {
		t.Errorf("Expected 1 metrics entry for site1, got %d", len(updatedSites[0].MetricsData))
	}

	if updatedSites[0].MetricsData["1762819200"].Visits != 200 {
		t.Errorf("Expected site1 visits to be 200, got %d", updatedSites[0].MetricsData["1762819200"].Visits)
	}

	// Verify site2 was not updated
	if updatedSites[1].MetricsData["1762732800"].Visits != 100 {
		t.Errorf("Expected site2 visits to remain 100, got %d", updatedSites[1].MetricsData["1762732800"].Visits)
	}
}

func TestUpdateSiteMetricsNonExistent(t *testing.T) {
	// Test UpdateSiteMetrics with non-existent site
	initialMetrics := map[string]pantheon.MetricData{
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
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: initialMetrics,
		},
	}

	collector := NewPantheonCollector(sites)

	// Try to update metrics for non-existent site
	newMetrics := map[string]pantheon.MetricData{
		"1762819200": {
			DateTime:      "2025-11-11T00:00:00",
			Visits:        200,
			PagesServed:   1000,
			CacheHits:     100,
			CacheMisses:   900,
			CacheHitRatio: "10%",
		},
	}

	// This should not crash or cause errors, just no-op
	collector.UpdateSiteMetrics("account999", "nonexistent", newMetrics)

	// Verify original site was not affected
	updatedSites := collector.GetSites()
	if len(updatedSites) != 1 {
		t.Errorf("Expected 1 site, got %d", len(updatedSites))
	}

	if updatedSites[0].MetricsData["1762732800"].Visits != 100 {
		t.Errorf("Expected site1 visits to remain 100, got %d", updatedSites[0].MetricsData["1762732800"].Visits)
	}
}

func TestCollectWithEmptyMetricsData(t *testing.T) {
	// Test Collect with sites that have empty metrics data
	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: map[string]pantheon.MetricData{}, // Empty metrics
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should have 0 metrics since metrics data is empty
	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("Expected 0 metrics with empty metrics data, got %d", count)
	}
}

func TestCollectWithNoDataCacheHitRatio(t *testing.T) {
	// Test Collect with cache hit ratio "--" (terminus-golang uses this when pages_served is 0,
	// matching Terminus CLI behavior; Pantheon API doesn't return cache_hit_ratio directly)
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        0,
			PagesServed:   0,
			CacheHits:     0,
			CacheMisses:   0,
			CacheHitRatio: "--", // No data indicator from terminus-golang
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should still collect metrics with cache hit ratio defaulting to 0
	count := 0
	for range ch {
		count++
	}

	// Should have 5 metrics (one for each metric type)
	if count != 5 {
		t.Errorf("Expected 5 metrics, got %d", count)
	}
}

func TestCollectWithNoCacheHitRatioPercentSign(t *testing.T) {
	// Test Collect with cache hit ratio that has no % sign
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        837,
			PagesServed:   3081,
			CacheHits:     119,
			CacheMisses:   2962,
			CacheHitRatio: "3.86", // No % sign
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should still collect metrics
	count := 0
	for range ch {
		count++
	}

	// Should have 5 metrics (one for each metric type)
	if count != 5 {
		t.Errorf("Expected 5 metrics, got %d", count)
	}
}

func TestNewPantheonCollectorWithEmptySites(t *testing.T) {
	// Test creating collector with empty sites array
	sites := []pantheon.SiteMetrics{}
	collector := NewPantheonCollector(sites)

	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if len(collector.sites) != 0 {
		t.Errorf("Expected 0 sites, got %d", len(collector.sites))
	}

	// Verify descriptors are still created
	ch := make(chan *prometheus.Desc, 5)
	collector.Describe(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	if count != 5 {
		t.Errorf("Expected 5 descriptors even with empty sites, got %d", count)
	}
}

func TestUpdateSitesWithEmptyArray(t *testing.T) {
	// Test UpdateSites with empty array
	initialMetrics := map[string]pantheon.MetricData{
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
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: initialMetrics,
		},
	}

	collector := NewPantheonCollector(sites)

	// Update with empty array
	collector.UpdateSites([]pantheon.SiteMetrics{})

	// Get sites and verify they are empty
	updatedSites := collector.GetSites()
	if len(updatedSites) != 0 {
		t.Errorf("Expected 0 sites after updating with empty array, got %d", len(updatedSites))
	}
}

func TestCollectWithZeroValues(t *testing.T) {
	// Test Collect with zero values in metrics
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        0,
			PagesServed:   0,
			CacheHits:     0,
			CacheMisses:   0,
			CacheHitRatio: "0%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should still collect metrics even with zero values
	count := 0
	for range ch {
		count++
	}

	if count != 5 {
		t.Errorf("Expected 5 metrics with zero values, got %d", count)
	}
}

func TestCollectWithLargeNumbers(t *testing.T) {
	// Test Collect with very large numbers
	metricsData := map[string]pantheon.MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        9999999,
			PagesServed:   88888888,
			CacheHits:     7777777,
			CacheMisses:   66666666,
			CacheHitRatio: "99.99%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	if count != 5 {
		t.Errorf("Expected 5 metrics, got %d", count)
	}
}

func TestCollectWithNegativeTimestamp(t *testing.T) {
	// Test Collect with negative timestamp (edge case)
	metricsData := map[string]pantheon.MetricData{
		"-100": {
			DateTime:      "1960-01-01T00:00:00",
			Visits:        100,
			PagesServed:   500,
			CacheHits:     50,
			CacheMisses:   450,
			CacheHitRatio: "10%",
		},
	}

	sites := []pantheon.SiteMetrics{
		{
			SiteName:    testCollectorSite1,
			Label:       "Site 1",
			PlanName:    "Basic",
			Account:     "account1",
			MetricsData: metricsData,
		},
	}

	collector := NewPantheonCollector(sites)

	ch := make(chan prometheus.Metric, 10)
	collector.Collect(ch)
	close(ch)

	// Should collect metrics even with negative timestamp
	count := 0
	for range ch {
		count++
	}

	if count != 5 {
		t.Errorf("Expected 5 metrics, got %d", count)
	}
}
