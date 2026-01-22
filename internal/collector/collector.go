// Package collector provides a Prometheus collector for Pantheon metrics.
package collector

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
	"github.com/prometheus/client_golang/prometheus"
)

// PantheonCollector collects Pantheon metrics for multiple sites
type PantheonCollector struct {
	sites   []pantheon.SiteMetrics
	mu      sync.RWMutex
	nowFunc func() time.Time // Function to get current time (for testing)

	visits        *prometheus.Desc
	pagesServed   *prometheus.Desc
	cacheHits     *prometheus.Desc
	cacheMisses   *prometheus.Desc
	cacheHitRatio *prometheus.Desc
}

// NewPantheonCollector creates a new Pantheon metrics collector
func NewPantheonCollector(sites []pantheon.SiteMetrics) *PantheonCollector {
	return &PantheonCollector{
		sites:   sites,
		nowFunc: time.Now,
		visits: prometheus.NewDesc(
			"pantheon_visits_total",
			"Total number of visits to a Pantheon site",
			[]string{"site_id", "site_name", "plan", "account"},
			nil,
		),
		pagesServed: prometheus.NewDesc(
			"pantheon_pages_served_total",
			"Total number of pages served by a Pantheon site",
			[]string{"site_id", "site_name", "plan", "account"},
			nil,
		),
		cacheHits: prometheus.NewDesc(
			"pantheon_cache_hits_total",
			"Total number of cache hits for a Pantheon site",
			[]string{"site_id", "site_name", "plan", "account"},
			nil,
		),
		cacheMisses: prometheus.NewDesc(
			"pantheon_cache_misses_total",
			"Total number of cache misses for a Pantheon site",
			[]string{"site_id", "site_name", "plan", "account"},
			nil,
		),
		cacheHitRatio: prometheus.NewDesc(
			"pantheon_cache_hit_ratio",
			"Cache hit ratio for a Pantheon site (0-1)",
			[]string{"site_id", "site_name", "plan", "account"},
			nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *PantheonCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.visits
	ch <- c.pagesServed
	ch <- c.cacheHits
	ch <- c.cacheMisses
	ch <- c.cacheHitRatio
}

// Collect implements prometheus.Collector
func (c *PantheonCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, site := range c.sites {
		// Collect and sort all timestamps
		type timestampedData struct {
			timestamp int64
			data      pantheon.MetricData
		}
		var sortedData []timestampedData

		for timestampStr, data := range site.MetricsData {
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				log.Printf("Error parsing timestamp %s: %v", timestampStr, err)
				continue
			}
			sortedData = append(sortedData, timestampedData{timestamp: timestamp, data: data})
		}

		if len(sortedData) == 0 {
			continue
		}

		// Sort by timestamp ascending
		sort.Slice(sortedData, func(i, j int) bool {
			return sortedData[i].timestamp < sortedData[j].timestamp
		})

		now := c.nowFunc()

		// Emit metrics for each timestamp with forward-fill until the next timestamp
		for i, td := range sortedData {
			// Determine the end time for forward-fill
			var endTimestamp int64
			if i < len(sortedData)-1 {
				// Forward-fill until just before the next metric timestamp
				endTimestamp = sortedData[i+1].timestamp
			} else {
				// For the last metric, forward-fill until now
				endTimestamp = now.Unix()
			}

			c.emitMetricsWithForwardFill(ch, &site, td.data, td.timestamp, endTimestamp)
		}
	}
}

// emitMetricsWithForwardFill emits metrics at startTimestamp and then every minute
// until endTimestamp (exclusive)
func (c *PantheonCollector) emitMetricsWithForwardFill(
	ch chan<- prometheus.Metric,
	site *pantheon.SiteMetrics,
	data pantheon.MetricData,
	startTimestamp, endTimestamp int64,
) {
	cacheHitRatioVal := c.parseCacheHitRatio(data.CacheHitRatio)

	// Emit metrics every minute from startTimestamp until endTimestamp
	for ts := startTimestamp; ts < endTimestamp; ts += 60 {
		t := time.Unix(ts, 0)

		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			c.visits,
			prometheus.GaugeValue,
			float64(data.Visits),
			site.SiteName, site.Label, site.PlanName, site.Account,
		))

		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			c.pagesServed,
			prometheus.GaugeValue,
			float64(data.PagesServed),
			site.SiteName, site.Label, site.PlanName, site.Account,
		))

		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			c.cacheHits,
			prometheus.GaugeValue,
			float64(data.CacheHits),
			site.SiteName, site.Label, site.PlanName, site.Account,
		))

		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			c.cacheMisses,
			prometheus.GaugeValue,
			float64(data.CacheMisses),
			site.SiteName, site.Label, site.PlanName, site.Account,
		))

		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			c.cacheHitRatio,
			prometheus.GaugeValue,
			cacheHitRatioVal,
			site.SiteName, site.Label, site.PlanName, site.Account,
		))
	}
}

// parseCacheHitRatio parses cache hit ratio string to float64 ratio (0-1).
// Handles "--" as a special "no data" indicator from terminus-golang
// (Pantheon API doesn't return cache_hit_ratio; it's calculated by the library,
// which uses "--" when pages_served is 0, matching Terminus CLI behavior).
// Input is expected as percentage string (e.g., "50%" or "50"), output is ratio (0-1).
func (c *PantheonCollector) parseCacheHitRatio(ratio string) float64 {
	if ratio == "--" {
		return 0
	}
	cacheHitRatioStr := strings.TrimSuffix(ratio, "%")
	cacheHitRatioVal, err := strconv.ParseFloat(cacheHitRatioStr, 64)
	if err != nil {
		log.Printf("Error parsing cache hit ratio %s: %v", ratio, err)
		return 0
	}
	// Convert percentage (0-100) to ratio (0-1) per Prometheus naming conventions
	return cacheHitRatioVal / 100
}

// UpdateSites updates the sites in the collector (thread-safe)
func (c *PantheonCollector) UpdateSites(sites []pantheon.SiteMetrics) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sites = sites
}

// GetSites returns a copy of the current sites (thread-safe)
func (c *PantheonCollector) GetSites() []pantheon.SiteMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sitesCopy := make([]pantheon.SiteMetrics, len(c.sites))
	copy(sitesCopy, c.sites)
	return sitesCopy
}

// UpdateSiteMetrics updates metrics for a specific site (thread-safe)
func (c *PantheonCollector) UpdateSiteMetrics(accountID, siteName string, metricsData map[string]pantheon.MetricData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.sites {
		if c.sites[i].Account == accountID && c.sites[i].SiteName == siteName {
			c.sites[i].MetricsData = metricsData
			return
		}
	}
}

// SetNowFunc sets the function used to get the current time (for testing)
func (c *PantheonCollector) SetNowFunc(f func() time.Time) {
	c.nowFunc = f
}
