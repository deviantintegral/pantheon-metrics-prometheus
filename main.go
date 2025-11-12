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

// SiteListEntry represents a single site from terminus site:list
type SiteListEntry struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	PlanName    string `json:"plan_name"`
	Framework   string `json:"framework"`
	Region      string `json:"region"`
	Owner       string `json:"owner"`
	Created     int64  `json:"created"`
	Memberships string `json:"memberships"`
	Frozen      bool   `json:"frozen"`
}

// SiteMetrics holds metrics data for a specific site
type SiteMetrics struct {
	SiteName    string
	Label       string
	PlanName    string
	MetricsData map[string]MetricData
}

// PantheonCollector collects Pantheon metrics for multiple sites
type PantheonCollector struct {
	sites []SiteMetrics

	visits        *prometheus.Desc
	pagesServed   *prometheus.Desc
	cacheHits     *prometheus.Desc
	cacheMisses   *prometheus.Desc
	cacheHitRatio *prometheus.Desc
}

// NewPantheonCollector creates a new Pantheon metrics collector
func NewPantheonCollector(sites []SiteMetrics) *PantheonCollector {
	return &PantheonCollector{
		sites: sites,
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
				site.SiteName, site.Label, site.PlanName,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.pagesServed,
				prometheus.GaugeValue,
				float64(data.PagesServed),
				site.SiteName, site.Label, site.PlanName,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.cacheHits,
				prometheus.GaugeValue,
				float64(data.CacheHits),
				site.SiteName, site.Label, site.PlanName,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.cacheMisses,
				prometheus.GaugeValue,
				float64(data.CacheMisses),
				site.SiteName, site.Label, site.PlanName,
			))

			ch <- prometheus.NewMetricWithTimestamp(ts, prometheus.MustNewConstMetric(
				c.cacheHitRatio,
				prometheus.GaugeValue,
				cacheHitRatioVal,
				site.SiteName, site.Label, site.PlanName,
			))
		}
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

func parseSiteList(data []byte) (map[string]SiteListEntry, error) {
	var siteList map[string]SiteListEntry
	if err := json.Unmarshal(data, &siteList); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return siteList, nil
}

func loadSiteList(filename string) (map[string]SiteListEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseSiteList(data)
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

func fetchAllSites() (map[string]SiteListEntry, error) {
	log.Printf("Fetching all sites from Terminus...")
	output, err := executeTerminusCommand("site:list", "--format=json")
	if err != nil {
		return nil, err
	}

	return parseSiteList(output)
}

func fetchSiteMetrics(siteName, environment string) (*SiteMetrics, error) {
	// Fetch metrics data
	metricsData, err := fetchMetricsData(siteName, environment)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}

	// For site:list, we don't get the label, so we'll use the site name as the label
	// In a real scenario, you might want to fetch site:info for each site to get the proper label
	return &SiteMetrics{
		SiteName:    siteName,
		Label:       siteName, // Using siteName as label since site:list doesn't provide it
		PlanName:    "",       // Will be updated from site list
		MetricsData: metricsData,
	}, nil
}

func main() {
	// Parse command-line flags
	environment := flag.String("env", "live", "Pantheon environment (default: live)")
	port := flag.String("port", "8080", "HTTP server port (default: 8080)")
	flag.Parse()

	// Fetch all sites from Terminus
	siteList, err := fetchAllSites()
	if err != nil {
		log.Fatalf("Failed to fetch site list: %v", err)
	}

	log.Printf("Found %d sites", len(siteList))

	// Collect metrics for all accessible sites
	var allSiteMetrics []SiteMetrics
	successCount := 0
	failCount := 0

	for _, site := range siteList {
		log.Printf("Processing site: %s (plan: %s)", site.Name, site.PlanName)

		// Fetch metrics for this site
		metricsData, err := fetchMetricsData(site.Name, *environment)
		if err != nil {
			log.Printf("Warning: Failed to fetch metrics for %s: %v", site.Name, err)
			failCount++
			continue
		}

		// Create SiteMetrics entry
		siteMetrics := SiteMetrics{
			SiteName:    site.Name,
			Label:       site.Name, // site:list doesn't provide a label field, using name
			PlanName:    site.PlanName,
			MetricsData: metricsData,
		}

		allSiteMetrics = append(allSiteMetrics, siteMetrics)
		successCount++
		log.Printf("Successfully loaded %d metric entries for %s", len(metricsData), site.Name)
	}

	log.Printf("Metrics collection complete: %d successful, %d failed", successCount, failCount)

	if len(allSiteMetrics) == 0 {
		log.Fatal("No site metrics were collected. Cannot start exporter.")
	}

	// Create collector with all site metrics
	collector := NewPantheonCollector(allSiteMetrics)

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
<p><strong>Environment:</strong> %s</p>
<p><strong>Sites monitored:</strong> %d</p>
<ul>
`, *environment, len(allSiteMetrics))

		for _, site := range allSiteMetrics {
			fmt.Fprintf(w, "<li>%s (plan: %s, %d metrics)</li>\n",
				site.SiteName, site.PlanName, len(site.MetricsData))
		}

		fmt.Fprintf(w, `
</ul>
<p>Metrics are available at <a href="/metrics">/metrics</a></p>
</body>
</html>
`)
	})

	// Start server
	serverAddr := ":" + *port
	log.Printf("Starting Pantheon metrics exporter on %s", serverAddr)
	log.Printf("Exporting metrics for %d sites", len(allSiteMetrics))
	log.Printf("Metrics available at http://localhost%s/metrics", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
