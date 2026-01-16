package pantheon

import (
	"os"
	"testing"
)

const (
	testSiteName     = "site1234"
	testSiteLabel    = "Example Site"
	testSitePlanName = "Performance Small"
)

func TestLoadMetricsData(t *testing.T) {
	// Test loading metrics data from JSON file
	metricsData, err := LoadMetricsData("../../testdata/example-metrics.json")
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
	siteInfo, err := LoadSiteInfo("../../testdata/site-info.json")
	if err != nil {
		t.Fatalf("Failed to load site info: %v", err)
	}

	if siteInfo.Name != testSiteName {
		t.Errorf("Expected name=%s, got %s", testSiteName, siteInfo.Name)
	}

	if siteInfo.Label != testSiteLabel {
		t.Errorf("Expected label=%s, got %s", testSiteLabel, siteInfo.Label)
	}

	if siteInfo.PlanName != testSitePlanName {
		t.Errorf("Expected plan_name=%s, got %s", testSitePlanName, siteInfo.PlanName)
	}
}

func TestLoadSiteConfig(t *testing.T) {
	// Test loading site config (old format for backwards compatibility)
	config, err := LoadSiteConfig("../../testdata/site-config.json")
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
	_, err := LoadMetricsData("../../testdata/nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadSiteInfoError(t *testing.T) {
	// Test error handling for non-existent file
	_, err := LoadSiteInfo("../../testdata/nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestParseMetricsFromTerminus(t *testing.T) {
	// Test parsing metrics from terminus command output
	data, err := os.ReadFile("../../testdata/example-metrics.json")
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
	data, err := os.ReadFile("../../testdata/site-info.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	siteInfo, err := parseSiteInfo(data)
	if err != nil {
		t.Fatalf("Failed to parse site info: %v", err)
	}

	if siteInfo.Name != testSiteName {
		t.Errorf("Expected name=%s, got %s", testSiteName, siteInfo.Name)
	}

	if siteInfo.Label != testSiteLabel {
		t.Errorf("Expected label=%s, got %s", testSiteLabel, siteInfo.Label)
	}

	if siteInfo.PlanName != testSitePlanName {
		t.Errorf("Expected plan_name=%s, got %s", testSitePlanName, siteInfo.PlanName)
	}
}

func TestLoadSiteList(t *testing.T) {
	// Test loading site list from JSON file
	siteList, err := LoadSiteList("../../testdata/site-list.json")
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
	data, err := os.ReadFile("../../testdata/site-list.json")
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
	accountID := GetAccountID(token)
	expected := "90abcdef"
	if accountID != expected {
		t.Errorf("Expected account ID %s, got %s", expected, accountID)
	}

	// Test with short token
	shortToken := "abc"
	accountID = GetAccountID(shortToken)
	if accountID != shortToken {
		t.Errorf("Expected account ID %s, got %s", shortToken, accountID)
	}
}

func TestParseMetricsDataError(t *testing.T) {
	// Test parsing invalid JSON
	invalidJSON := []byte(`{"invalid": "json"`)
	_, err := parseMetricsData(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseSiteInfoError(t *testing.T) {
	// Test parsing invalid JSON
	invalidJSON := []byte(`{"invalid": "json"`)
	_, err := parseSiteInfo(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseSiteListError(t *testing.T) {
	// Test parsing invalid JSON
	invalidJSON := []byte(`{"invalid": "json"`)
	_, err := parseSiteList(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadMetricsDataInvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpfile, err := os.CreateTemp("", "invalid-metrics-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Write invalid JSON
	if _, err := tmpfile.Write([]byte(`{"invalid": "json"`)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpfile.Close()

	// Try to load the invalid file
	_, err = LoadMetricsData(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadSiteConfigInvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpfile, err := os.CreateTemp("", "invalid-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Write invalid JSON
	if _, err := tmpfile.Write([]byte(`{"invalid": "json"`)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpfile.Close()

	// Try to load the invalid file
	_, err = LoadSiteConfig(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadSiteInfoInvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpfile, err := os.CreateTemp("", "invalid-info-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Write invalid JSON
	if _, err := tmpfile.Write([]byte(`{"invalid": "json"`)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpfile.Close()

	// Try to load the invalid file
	_, err = LoadSiteInfo(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadSiteListInvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpfile, err := os.CreateTemp("", "invalid-list-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

	// Write invalid JSON
	if _, err := tmpfile.Write([]byte(`{"invalid": "json"`)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	_ = tmpfile.Close()

	// Try to load the invalid file
	_, err = LoadSiteList(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestLoadSiteConfigError(t *testing.T) {
	// Test error handling for non-existent file
	_, err := LoadSiteConfig("../../testdata/nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadSiteListError(t *testing.T) {
	// Test error handling for non-existent file
	_, err := LoadSiteList("../../testdata/nonexistent.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestNewClient(t *testing.T) {
	// Test that NewClient creates a client with a session manager
	client := NewClient()
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.sessionManager == nil {
		t.Fatal("Expected client to have session manager")
	}
}

func TestGetAccountIDExact8Chars(t *testing.T) {
	// Test with exactly 8 character token
	token := "12345678"
	accountID := GetAccountID(token)
	if accountID != "12345678" {
		t.Errorf("Expected account ID %s, got %s", "12345678", accountID)
	}
}

func TestGetAccountIDEmpty(t *testing.T) {
	// Test with empty token
	token := ""
	accountID := GetAccountID(token)
	if accountID != "" {
		t.Errorf("Expected empty account ID, got %s", accountID)
	}
}

func TestClientInvalidateSession(t *testing.T) {
	// Test that InvalidateSession does not panic and works correctly
	client := NewClient()

	// Should not panic even with non-existent token
	client.InvalidateSession("non-existent-token")

	// Should not panic with empty token
	client.InvalidateSession("")

	// Should work with a normal token
	client.InvalidateSession("some-machine-token")
}
