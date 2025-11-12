package main

import (
	"testing"
)

func TestNewPantheonCollector(t *testing.T) {
	// Test creating a new collector with multiple sites
	metricsData := map[string]MetricData{
		"1762732800": {
			DateTime:      "2025-11-10T00:00:00",
			Visits:        837,
			PagesServed:   3081,
			CacheHits:     119,
			CacheMisses:   2962,
			CacheHitRatio: "3.86%",
		},
	}

	sites := []SiteMetrics{
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

	collector := NewPantheonCollector(sites)
	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	if len(collector.sites) != 2 {
		t.Errorf("Expected 2 sites in collector, got %d", len(collector.sites))
	}
}
