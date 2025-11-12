package main

import (
	"os"
	"testing"
)

func TestLoadMetricsData(t *testing.T) {
	// Test loading metrics data from JSON file
	metricsData, err := loadMetricsData("testdata/example-metrics.json")
	if err != nil {
		t.Fatalf("Failed to load metrics data: %v", err)
	}

	// Verify we have the expected number of entries
	if len(metricsData) != 2 {
		t.Errorf("Expected 2 metric entries, got %d", len(metricsData))
	}

	// Verify first entry
	data1, ok := metricsData["1762732800"]
	if !ok {
		t.Fatal("Expected entry for timestamp 1762732800")
	}

	if data1.Visits != 837 {
		t.Errorf("Expected visits=837, got %d", data1.Visits)
	}

	if data1.PagesServed != 3081 {
		t.Errorf("Expected pages_served=3081, got %d", data1.PagesServed)
	}

	if data1.CacheHits != 119 {
		t.Errorf("Expected cache_hits=119, got %d", data1.CacheHits)
	}

	if data1.CacheMisses != 2962 {
		t.Errorf("Expected cache_misses=2962, got %d", data1.CacheMisses)
	}

	if data1.CacheHitRatio != "3.86%" {
		t.Errorf("Expected cache_hit_ratio=3.86%%, got %s", data1.CacheHitRatio)
	}

	// Verify second entry
	data2, ok := metricsData["1762819200"]
	if !ok {
		t.Fatal("Expected entry for timestamp 1762819200")
	}

	if data2.Visits != 824 {
		t.Errorf("Expected visits=824, got %d", data2.Visits)
	}
}

func TestLoadSiteInfo(t *testing.T) {
	// Test loading site info from JSON file
	siteInfo, err := loadSiteInfo("testdata/site-info.json")
	if err != nil {
		t.Fatalf("Failed to load site info: %v", err)
	}

	if siteInfo.Name != "site1234" {
		t.Errorf("Expected name=site1234, got %s", siteInfo.Name)
	}

	if siteInfo.Label != "Example Site" {
		t.Errorf("Expected label=Example Site, got %s", siteInfo.Label)
	}

	if siteInfo.PlanName != "Performance Small" {
		t.Errorf("Expected plan_name=Performance Small, got %s", siteInfo.PlanName)
	}
}

func TestLoadSiteConfig(t *testing.T) {
	// Test loading site config (old format for backwards compatibility)
	config, err := loadSiteConfig("testdata/site-config.json")
	if err != nil {
		t.Fatalf("Failed to load site config: %v", err)
	}

	if config.Name != "site1234" {
		t.Errorf("Expected name=site1234, got %s", config.Name)
	}

	if config.Label != "Example Site" {
		t.Errorf("Expected label=Example Site, got %s", config.Label)
	}

	if config.PlanName != "Performance Small" {
		t.Errorf("Expected plan_name=Performance Small, got %s", config.PlanName)
	}
}

func TestLoadMetricsDataError(t *testing.T) {
	// Test error handling for non-existent file
	_, err := loadMetricsData("testdata/nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadSiteInfoError(t *testing.T) {
	// Test error handling for non-existent file
	_, err := loadSiteInfo("testdata/nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestParseMetricsFromTerminus(t *testing.T) {
	// Test parsing metrics from terminus command output
	data, err := os.ReadFile("testdata/example-metrics.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	metricsData, err := parseMetricsData(data)
	if err != nil {
		t.Fatalf("Failed to parse metrics data: %v", err)
	}

	if len(metricsData) != 2 {
		t.Errorf("Expected 2 metric entries, got %d", len(metricsData))
	}
}

func TestParseSiteInfoFromTerminus(t *testing.T) {
	// Test parsing site info from terminus command output
	data, err := os.ReadFile("testdata/site-info.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	siteInfo, err := parseSiteInfo(data)
	if err != nil {
		t.Fatalf("Failed to parse site info: %v", err)
	}

	if siteInfo.Name != "site1234" {
		t.Errorf("Expected name=site1234, got %s", siteInfo.Name)
	}

	if siteInfo.Label != "Example Site" {
		t.Errorf("Expected label=Example Site, got %s", siteInfo.Label)
	}

	if siteInfo.PlanName != "Performance Small" {
		t.Errorf("Expected plan_name=Performance Small, got %s", siteInfo.PlanName)
	}
}

func TestLoadSiteList(t *testing.T) {
	// Test loading site list from JSON file
	siteList, err := loadSiteList("testdata/site-list.json")
	if err != nil {
		t.Fatalf("Failed to load site list: %v", err)
	}

	// Verify we have the expected number of entries
	if len(siteList) != 2 {
		t.Errorf("Expected 2 site entries, got %d", len(siteList))
	}

	// Check for first site
	foundSite1 := false
	foundSite2 := false

	for _, site := range siteList {
		if site.Name == "site1234" {
			foundSite1 = true
			if site.PlanName != "Performance Small" {
				t.Errorf("Expected plan_name=Performance Small for site1234, got %s", site.PlanName)
			}
			if site.Framework != "drupal8" {
				t.Errorf("Expected framework=drupal8 for site1234, got %s", site.Framework)
			}
		}

		if site.Name == "site5678" {
			foundSite2 = true
			if site.PlanName != "Basic" {
				t.Errorf("Expected plan_name=Basic for site5678, got %s", site.PlanName)
			}
			if site.Framework != "wordpress" {
				t.Errorf("Expected framework=wordpress for site5678, got %s", site.Framework)
			}
		}
	}

	if !foundSite1 {
		t.Error("Expected to find site1234 in site list")
	}

	if !foundSite2 {
		t.Error("Expected to find site5678 in site list")
	}
}

func TestParseSiteList(t *testing.T) {
	// Test parsing site list from terminus command output
	data, err := os.ReadFile("testdata/site-list.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	siteList, err := parseSiteList(data)
	if err != nil {
		t.Fatalf("Failed to parse site list: %v", err)
	}

	if len(siteList) != 2 {
		t.Errorf("Expected 2 site entries, got %d", len(siteList))
	}
}

func TestGetAccountID(t *testing.T) {
	// Test with full-length token
	token := "1234567890abcdef1234567890abcdef"
	accountID := getAccountID(token)
	expected := "90abcdef"
	if accountID != expected {
		t.Errorf("Expected account ID %s, got %s", expected, accountID)
	}

	// Test with short token
	shortToken := "abc"
	accountID = getAccountID(shortToken)
	if accountID != shortToken {
		t.Errorf("Expected account ID %s, got %s", shortToken, accountID)
	}
}
