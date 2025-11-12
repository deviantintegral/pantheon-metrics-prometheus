package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricData represents a single metric entry from Terminus
type MetricData struct {
	DateTime      string `json:"datetime"`
	Visits        int    `json:"visits"`
	PagesServed   int    `json:"pages_served"`
	CacheHits     int    `json:"cache_hits"`
	CacheMisses   int    `json:"cache_misses"`
	CacheHitRatio string `json:"cache_hit_ratio"`
}

// SiteConfig represents the site configuration
type SiteConfig struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	PlanName string `json:"plan_name"`
}

// PantheonCollector collects Pantheon metrics
type PantheonCollector struct {
	metricsData map[string]MetricData
	siteName    string
	siteLabel   string
	sitePlan    string

	visits      *prometheus.Desc
	pagesServed *prometheus.Desc
	cacheHits   *prometheus.Desc
	cacheMisses *prometheus.Desc
	cacheHitRatio *prometheus.Desc
}

// NewPantheonCollector creates a new Pantheon metrics collector
func NewPantheonCollector(metricsData map[string]MetricData, siteName, siteLabel, sitePlan string) *PantheonCollector {
	return &PantheonCollector{
		metricsData: metricsData,
		siteName:    siteName,
		siteLabel:   siteLabel,
		sitePlan:    sitePlan,
		visits: prometheus.NewDesc(
			"pantheon_visits",
			"Number of visits",
			[]string{"name", "label", "plan"},
			nil,
		),
		pagesServed: prometheus.NewDesc(
			"pantheon_pages_served",
			"Number of pages served",
			[]string{"name", "label", "plan"},
			nil,
		),
		cacheHits: prometheus.NewDesc(
			"pantheon_cache_hits",
			"Number of cache hits",
			[]string{"name", "label", "plan"},
			nil,
		),
		cacheMisses: prometheus.NewDesc(
			"pantheon_cache_misses",
			"Number of cache misses",
			[]string{"name", "label", "plan"},
			nil,
		),
		cacheHitRatio: prometheus.NewDesc(
			"pantheon_cache_hit_ratio",
			"Cache hit ratio as percentage",
			[]string{"name", "label", "plan"},
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
	for timestampStr, data := range c.metricsData {
		// Convert Unix timestamp string to time.Time
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			log.Printf("Error parsing timestamp %s: %v", timestampStr, err)
			continue
		}
		ts := time.Unix(timestamp, 0)

		// Parse cache hit ratio (remove % sign and convert to float)
		cacheHitRatioStr := strings.TrimSuffix(data.CacheHitRatio, "%")
		cacheHitRatioVal, err := strconv.ParseFloat(cacheHitRatioStr, 64)
		if err != nil {
			log.Printf("Error parsing cache hit ratio %s: %v", data.CacheHitRatio, err)
			cacheHitRatioVal = 0
		}

		// Create metrics with labels and timestamps
		ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
			c.visits,
			prometheus.GaugeValue,
			float64(data.Visits),
			c.siteName, c.siteLabel, c.sitePlan,
		))

		ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
			c.pagesServed,
			prometheus.GaugeValue,
			float64(data.PagesServed),
			c.siteName, c.siteLabel, c.sitePlan,
		))

		ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
			c.cacheHits,
			prometheus.GaugeValue,
			float64(data.CacheHits),
			c.siteName, c.siteLabel, c.sitePlan,
		))

		ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
			c.cacheMisses,
			prometheus.GaugeValue,
			float64(data.CacheMisses),
			c.siteName, c.siteLabel, c.sitePlan,
		))

		ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
			c.cacheHitRatio,
			prometheus.GaugeValue,
			cacheHitRatioVal,
			c.siteName, c.siteLabel, c.sitePlan,
		))
	}
}

func loadMetricsData(filename string) (map[string]MetricData, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var metricsData map[string]MetricData
	if err := json.Unmarshal(data, &metricsData); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return metricsData, nil
}

func loadSiteConfig(filename string) (*SiteConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var config SiteConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &config, nil
}

func main() {
	// Load site configuration
	siteConfig, err := loadSiteConfig("site-config.json")
	if err != nil {
		log.Fatalf("Failed to load site config: %v", err)
	}

	// Load metrics data from example file
	metricsData, err := loadMetricsData("example-metrics.json")
	if err != nil {
		log.Fatalf("Failed to load metrics data: %v", err)
	}

	// Create collector with labels from config
	collector := NewPantheonCollector(
		metricsData,
		siteConfig.Name,
		siteConfig.Label,
		siteConfig.PlanName,
	)

	// Register the collector
	registry := prometheus.NewRegistry()
	registry.MustRegister(collector)

	// Create HTTP handler
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// Root handler with instructions
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
<html>
<head><title>Pantheon Metrics Exporter</title></head>
<body>
<h1>Pantheon Metrics Exporter</h1>
<p>Metrics are available at <a href="/metrics">/metrics</a></p>
</body>
</html>
`)
	})

	// Start server
	port := ":8080"
	log.Printf("Starting Pantheon metrics exporter on %s", port)
	log.Printf("Metrics available at http://localhost%s/metrics", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
