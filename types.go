package main

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
	Account     string // Account identifier (truncated token)
	MetricsData map[string]MetricData
}
