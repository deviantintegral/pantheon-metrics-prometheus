package pantheon

import (
	"strconv"

	"github.com/deviantintegral/terminus-golang/pkg/api/models"
)

// ConvertSite converts a library Site to our SiteListEntry.
func ConvertSite(site *models.Site) SiteListEntry {
	return SiteListEntry{
		Name:      site.Name,
		ID:        site.ID,
		PlanName:  site.PlanName,
		Framework: site.Framework,
		Region:    site.PreferredZoneLabel,
		Owner:     site.Owner,
		Created:   site.Created,
		Frozen:    site.Frozen || site.IsFrozen,
		// Memberships field would need formatting from MembershipUserID and MembershipRole
	}
}

// ConvertSiteListItem converts a library SiteListItem to our SiteListEntry.
func ConvertSiteListItem(site *models.SiteListItem) SiteListEntry {
	return SiteListEntry{
		Name:      site.Name,
		ID:        site.ID,
		PlanName:  site.PlanName,
		Framework: site.Framework,
		Region:    site.PreferredZoneLabel,
		Owner:     site.Owner,
		Created:   site.Created,
		Frozen:    site.Frozen || site.IsFrozen,
	}
}

// ConvertMetrics converts library Metrics to our MetricData.
func ConvertMetrics(m *models.Metrics) MetricData {
	return MetricData{
		DateTime:      m.Datetime,
		Visits:        int(m.Visits),
		PagesServed:   int(m.PagesServed),
		CacheHits:     int(m.CacheHits),
		CacheMisses:   int(m.CacheMisses),
		CacheHitRatio: m.CacheHitRatio,
	}
}

// ConvertMetricsToMap converts a slice of library Metrics to our map format.
// The map keys are Unix timestamps as strings.
func ConvertMetricsToMap(metrics []*models.Metrics) map[string]MetricData {
	result := make(map[string]MetricData, len(metrics))
	for _, m := range metrics {
		key := strconv.FormatInt(m.Timestamp, 10)
		result[key] = ConvertMetrics(m)
	}
	return result
}

// ConvertSitesToMap converts a slice of library Sites to our map format.
// The map keys are site IDs.
func ConvertSitesToMap(sites []*models.Site) map[string]SiteListEntry {
	result := make(map[string]SiteListEntry, len(sites))
	for _, site := range sites {
		result[site.ID] = ConvertSite(site)
	}
	return result
}
