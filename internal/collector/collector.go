// Package collector provides a Prometheus collector for Pantheon metrics.
package collector

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deviantintegral/pantheon-metrics-prometheus/internal/pantheon"
	"github.com/prometheus/client_golang/prometheus"
)

// PantheonCollector collects Pantheon metrics for multiple sites
type PantheonCollector struct {
	sites []pantheon.SiteMetrics
	mu    sync.RWMutex

	visits        *prometheus.Desc
	pagesServed   *prometheus.Desc
	cacheHits     *prometheus.Desc
	cacheMisses   *prometheus.Desc
	cacheHitRatio *prometheus.Desc
}

// NewPantheonCollector creates a new Pantheon metrics collector
func NewPantheonCollector(sites []pantheon.SiteMetrics) *PantheonCollector {
	return &PantheonCollector{
		sites: sites,
		visits: prometheus.NewDesc(
			"pantheon_visits",
			"Number of visits",
			[]string{"name", "label", "plan", "account"},
			nil,
		),
		pagesServed: prometheus.NewDesc(
			"pantheon_pages_served",
			"Number of pages served",
			[]string{"name", "label", "plan", "account"},
			nil,
		),
		cacheHits: prometheus.NewDesc(
			"pantheon_cache_hits",
			"Number of cache hits",
			[]string{"name", "label", "plan", "account"},
			nil,
		),
		cacheMisses: prometheus.NewDesc(
			"pantheon_cache_misses",
			"Number of cache misses",
			[]string{"name", "label", "plan", "account"},
			nil,
		),
		cacheHitRatio: prometheus.NewDesc(
			"pantheon_cache_hit_ratio",
			"Cache hit ratio as percentage",
			[]string{"name", "label", "plan", "account"},
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
		for timestampStr, data := range site.MetricsData {
			// Convert Unix timestamp string to time.Time
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				log.Printf("Error parsing timestamp %s: %v", timestampStr, err)
				continue
			}
			ts := time.Unix(timestamp, 0)

			// Parse cache hit ratio (remove % sign and convert to float)
			// Handle "--" as a special "no data" indicator from Pantheon API
			var cacheHitRatioVal float64
			if data.CacheHitRatio == "--" {
				cacheHitRatioVal = 0
			} else {
				cacheHitRatioStr := strings.TrimSuffix(data.CacheHitRatio, "%")
				var err error
				cacheHitRatioVal, err = strconv.ParseFloat(cacheHitRatioStr, 64)
				if err != nil {
					log.Printf("Error parsing cache hit ratio %s: %v", data.CacheHitRatio, err)
					cacheHitRatioVal = 0
				}
			}

			// Create metrics with labels and timestamps
			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.visits,
				prometheus.GaugeValue,
				float64(data.Visits),
				site.SiteName, site.Label, site.PlanName, site.Account,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.pagesServed,
				prometheus.GaugeValue,
				float64(data.PagesServed),
				site.SiteName, site.Label, site.PlanName, site.Account,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.cacheHits,
				prometheus.GaugeValue,
				float64(data.CacheHits),
				site.SiteName, site.Label, site.PlanName, site.Account,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.cacheMisses,
				prometheus.GaugeValue,
				float64(data.CacheMisses),
				site.SiteName, site.Label, site.PlanName, site.Account,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.cacheHitRatio,
				prometheus.GaugeValue,
				cacheHitRatioVal,
				site.SiteName, site.Label, site.PlanName, site.Account,
			))
		}
	}
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
