package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
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

// SiteConfig represents the site configuration (legacy format)
type SiteConfig struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	PlanName string `json:"plan_name"`
}

// SiteInfo represents site information from terminus site:info
type SiteInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Created      string `json:"created"`
	Framework    string `json:"framework"`
	Organization string `json:"organization"`
	ServiceLevel string `json:"service_level"`
	Upstream     string `json:"upstream"`
	PHPVersion   string `json:"php_version"`
	HolderType   string `json:"holder_type"`
	HolderID     string `json:"holder_id"`
	Owner        string `json:"owner"`
	Frozen       bool   `json:"frozen"`
	PlanName     string `json:"plan_name"`
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

func loadSiteInfo(filename string) (*SiteInfo, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseSiteInfo(data)
}

func parseSiteInfo(data []byte) (*SiteInfo, error) {
	var siteInfo SiteInfo
	if err := json.Unmarshal(data, &siteInfo); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &siteInfo, nil
}

func parseMetricsData(data []byte) (map[string]MetricData, error) {
	var metricsData map[string]MetricData
	if err := json.Unmarshal(data, &metricsData); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return metricsData, nil
}

func executeTerminusCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("terminus", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("terminus command failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}

func fetchSiteInfo(siteName string) (*SiteInfo, error) {
	log.Printf("Fetching site info for %s...", siteName)
	output, err := executeTerminusCommand("site:info", siteName, "--format=json")
	if err != nil {
		return nil, err
	}

	return parseSiteInfo(output)
}

func fetchMetricsData(siteName, environment string) (map[string]MetricData, error) {
	log.Printf("Fetching metrics for %s.%s...", siteName, environment)
	output, err := executeTerminusCommand("env:metrics", fmt.Sprintf("%s.%s", siteName, environment), "--format=json")
	if err != nil {
		return nil, err
	}

	return parseMetricsData(output)
}

func main() {
	// Parse command-line flags
	siteName := flag.String("site", "", "Pantheon site name (required)")
	environment := flag.String("env", "live", "Pantheon environment (default: live)")
	port := flag.String("port", "8080", "HTTP server port (default: 8080)")
	flag.Parse()

	if *siteName == "" {
		log.Fatal("Error: -site flag is required")
	}

	// Fetch site information from Terminus
	siteInfo, err := fetchSiteInfo(*siteName)
	if err != nil {
		log.Fatalf("Failed to fetch site info: %v", err)
	}

	log.Printf("Site: %s (%s)", siteInfo.Name, siteInfo.Label)
	log.Printf("Plan: %s", siteInfo.PlanName)

	// Fetch metrics data from Terminus
	metricsData, err := fetchMetricsData(*siteName, *environment)
	if err != nil {
		log.Fatalf("Failed to fetch metrics data: %v", err)
	}

	log.Printf("Loaded %d metric entries", len(metricsData))

	// Create collector with data from Terminus
	collector := NewPantheonCollector(
		metricsData,
		siteInfo.Name,
		siteInfo.Label,
		siteInfo.PlanName,
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
<p><strong>Site:</strong> %s (%s)</p>
<p><strong>Environment:</strong> %s</p>
<p><strong>Plan:</strong> %s</p>
<p>Metrics are available at <a href="/metrics">/metrics</a></p>
</body>
</html>
`, siteInfo.Label, siteInfo.Name, *environment, siteInfo.PlanName)
	})

	// Start server
	serverAddr := ":" + *port
	log.Printf("Starting Pantheon metrics exporter on %s", serverAddr)
	log.Printf("Metrics available at http://localhost%s/metrics", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
