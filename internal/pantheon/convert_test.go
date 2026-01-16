package pantheon

import (
	"testing"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

func TestConvertSite(t *testing.T) {
	site := &models.Site{
		Name:               "test-site",
		ID:                 "site-uuid-1234",
		PlanName:           "Performance Large",
		Framework:          "drupal10",
		PreferredZoneLabel: "US East",
		Owner:              "user-uuid-5678",
		Created:            1705276800, // Unix timestamp
		Frozen:             false,
		IsFrozen:           false,
	}

	result := ConvertSite(site)

	if result.Name != "test-site" {
		t.Errorf("Expected Name='test-site', got '%s'", result.Name)
	}
	if result.ID != "site-uuid-1234" {
		t.Errorf("Expected ID='site-uuid-1234', got '%s'", result.ID)
	}
	if result.PlanName != "Performance Large" {
		t.Errorf("Expected PlanName='Performance Large', got '%s'", result.PlanName)
	}
	if result.Framework != "drupal10" {
		t.Errorf("Expected Framework='drupal10', got '%s'", result.Framework)
	}
	if result.Region != "US East" {
		t.Errorf("Expected Region='US East', got '%s'", result.Region)
	}
	if result.Owner != "user-uuid-5678" {
		t.Errorf("Expected Owner='user-uuid-5678', got '%s'", result.Owner)
	}
	if result.Created != 1705276800 {
		t.Errorf("Expected Created=1705276800, got %d", result.Created)
	}
	if result.Frozen {
		t.Error("Expected Frozen=false")
	}
}

func TestConvertSiteFrozen(t *testing.T) {
	// Test with Frozen=true
	site := &models.Site{
		Name:   "frozen-site",
		ID:     "frozen-id",
		Frozen: true,
	}

	result := ConvertSite(site)
	if !result.Frozen {
		t.Error("Expected Frozen=true when site.Frozen is true")
	}

	// Test with IsFrozen=true
	site2 := &models.Site{
		Name:     "is-frozen-site",
		ID:       "is-frozen-id",
		IsFrozen: true,
	}

	result2 := ConvertSite(site2)
	if !result2.Frozen {
		t.Error("Expected Frozen=true when site.IsFrozen is true")
	}
}

func TestConvertSiteListItem(t *testing.T) {
	site := &models.SiteListItem{
		Name:               "list-site",
		ID:                 "list-uuid-1234",
		PlanName:           "Basic",
		Framework:          "wordpress",
		PreferredZoneLabel: "EU West",
		Owner:              "owner-uuid",
		Created:            1685577600, // Unix timestamp
		Frozen:             true,
		IsFrozen:           false,
	}

	result := ConvertSiteListItem(site)

	if result.Name != "list-site" {
		t.Errorf("Expected Name='list-site', got '%s'", result.Name)
	}
	if result.ID != "list-uuid-1234" {
		t.Errorf("Expected ID='list-uuid-1234', got '%s'", result.ID)
	}
	if result.PlanName != "Basic" {
		t.Errorf("Expected PlanName='Basic', got '%s'", result.PlanName)
	}
	if result.Framework != "wordpress" {
		t.Errorf("Expected Framework='wordpress', got '%s'", result.Framework)
	}
	if result.Region != "EU West" {
		t.Errorf("Expected Region='EU West', got '%s'", result.Region)
	}
	if !result.Frozen {
		t.Error("Expected Frozen=true")
	}
}

func TestConvertMetrics(t *testing.T) {
	metrics := &models.Metrics{
		Datetime:      "2024-01-15T00:00:00Z",
		Timestamp:     1705276800,
		Visits:        1500,
		PagesServed:   5000,
		CacheHits:     4000,
		CacheMisses:   1000,
		CacheHitRatio: "80.0%",
	}

	result := ConvertMetrics(metrics)

	if result.DateTime != "2024-01-15T00:00:00Z" {
		t.Errorf("Expected DateTime='2024-01-15T00:00:00Z', got '%s'", result.DateTime)
	}
	if result.Visits != 1500 {
		t.Errorf("Expected Visits=1500, got %d", result.Visits)
	}
	if result.PagesServed != 5000 {
		t.Errorf("Expected PagesServed=5000, got %d", result.PagesServed)
	}
	if result.CacheHits != 4000 {
		t.Errorf("Expected CacheHits=4000, got %d", result.CacheHits)
	}
	if result.CacheMisses != 1000 {
		t.Errorf("Expected CacheMisses=1000, got %d", result.CacheMisses)
	}
	if result.CacheHitRatio != "80.0%" {
		t.Errorf("Expected CacheHitRatio='80.0%%', got '%s'", result.CacheHitRatio)
	}
}

func TestConvertMetricsToMap(t *testing.T) {
	metrics := []*models.Metrics{
		{
			Datetime:      "2024-01-15T00:00:00Z",
			Timestamp:     1705276800,
			Visits:        100,
			PagesServed:   500,
			CacheHits:     400,
			CacheMisses:   100,
			CacheHitRatio: "80.0%",
		},
		{
			Datetime:      "2024-01-16T00:00:00Z",
			Timestamp:     1705363200,
			Visits:        200,
			PagesServed:   1000,
			CacheHits:     900,
			CacheMisses:   100,
			CacheHitRatio: "90.0%",
		},
	}

	result := ConvertMetricsToMap(metrics)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}

	entry1, ok := result["1705276800"]
	if !ok {
		t.Fatal("Expected entry for timestamp 1705276800")
	}
	if entry1.Visits != 100 {
		t.Errorf("Expected Visits=100, got %d", entry1.Visits)
	}

	entry2, ok := result["1705363200"]
	if !ok {
		t.Fatal("Expected entry for timestamp 1705363200")
	}
	if entry2.Visits != 200 {
		t.Errorf("Expected Visits=200, got %d", entry2.Visits)
	}
}

func TestConvertMetricsToMapEmpty(t *testing.T) {
	metrics := []*models.Metrics{}
	result := ConvertMetricsToMap(metrics)

	if len(result) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(result))
	}
}

func TestConvertSitesToMap(t *testing.T) {
	sites := []*models.Site{
		{
			Name:     "site-one",
			ID:       "uuid-1",
			PlanName: "Basic",
		},
		{
			Name:     "site-two",
			ID:       "uuid-2",
			PlanName: "Performance Small",
		},
	}

	result := ConvertSitesToMap(sites)

	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}

	entry1, ok := result["uuid-1"]
	if !ok {
		t.Fatal("Expected entry for uuid-1")
	}
	if entry1.Name != "site-one" {
		t.Errorf("Expected Name='site-one', got '%s'", entry1.Name)
	}
	if entry1.PlanName != "Basic" {
		t.Errorf("Expected PlanName='Basic', got '%s'", entry1.PlanName)
	}

	entry2, ok := result["uuid-2"]
	if !ok {
		t.Fatal("Expected entry for uuid-2")
	}
	if entry2.Name != "site-two" {
		t.Errorf("Expected Name='site-two', got '%s'", entry2.Name)
	}
}

func TestConvertSitesToMapEmpty(t *testing.T) {
	sites := []*models.Site{}
	result := ConvertSitesToMap(sites)

	if len(result) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(result))
	}
}

func TestConvertSiteListItemIsFrozen(t *testing.T) {
	// Test with IsFrozen=true only
	site := &models.SiteListItem{
		Name:     "is-frozen-site",
		ID:       "is-frozen-id",
		IsFrozen: true,
		Frozen:   false,
	}

	result := ConvertSiteListItem(site)
	if !result.Frozen {
		t.Error("Expected Frozen=true when site.IsFrozen is true")
	}
}

func TestConvertMetricsZeroValues(t *testing.T) {
	// Test with zero values
	metrics := &models.Metrics{
		Datetime:      "",
		Timestamp:     0,
		Visits:        0,
		PagesServed:   0,
		CacheHits:     0,
		CacheMisses:   0,
		CacheHitRatio: "0%",
	}

	result := ConvertMetrics(metrics)

	if result.DateTime != "" {
		t.Errorf("Expected DateTime='', got '%s'", result.DateTime)
	}
	if result.Visits != 0 {
		t.Errorf("Expected Visits=0, got %d", result.Visits)
	}
	if result.CacheHitRatio != "0%" {
		t.Errorf("Expected CacheHitRatio='0%%', got '%s'", result.CacheHitRatio)
	}
}
