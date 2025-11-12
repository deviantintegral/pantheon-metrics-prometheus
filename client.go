package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
)

// executeTerminusCommand executes a terminus CLI command and returns the output
func executeTerminusCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("terminus", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("terminus command failed: %w\nOutput: %s", err, string(output))
	}
	return output, nil
}

// authenticateWithToken authenticates with Terminus using a machine token
func authenticateWithToken(token string) error {
	log.Printf("Authenticating with machine token...")
	_, err := executeTerminusCommand("auth:login", "--machine-token="+token)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}

// getAccountID returns an account identifier from a machine token (last 8 chars)
func getAccountID(token string) string {
	// Return last 8 characters of token for identification
	if len(token) >= 8 {
		return token[len(token)-8:]
	}
	return token
}

// fetchSiteInfo fetches site information from Terminus
func fetchSiteInfo(siteName string) (*SiteInfo, error) {
	log.Printf("Fetching site info for %s...", siteName)
	output, err := executeTerminusCommand("site:info", siteName, "--format=json")
	if err != nil {
		return nil, err
	}

	return parseSiteInfo(output)
}

// fetchMetricsData fetches metrics data for a site from Terminus
func fetchMetricsData(siteName, environment string) (map[string]MetricData, error) {
	log.Printf("Fetching metrics for %s.%s...", siteName, environment)
	output, err := executeTerminusCommand("env:metrics", fmt.Sprintf("%s.%s", siteName, environment), "--format=json")
	if err != nil {
		return nil, err
	}

	return parseMetricsData(output)
}

// fetchAllSites fetches the list of all sites from Terminus
func fetchAllSites() (map[string]SiteListEntry, error) {
	log.Printf("Fetching all sites from Terminus...")
	output, err := executeTerminusCommand("site:list", "--format=json")
	if err != nil {
		return nil, err
	}

	return parseSiteList(output)
}

// fetchSiteMetrics fetches both site info and metrics data
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

// parseSiteInfo parses site info JSON data
func parseSiteInfo(data []byte) (*SiteInfo, error) {
	var siteInfo SiteInfo
	if err := json.Unmarshal(data, &siteInfo); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return &siteInfo, nil
}

// parseMetricsData parses metrics JSON data
func parseMetricsData(data []byte) (map[string]MetricData, error) {
	var metricsData map[string]MetricData
	if err := json.Unmarshal(data, &metricsData); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return metricsData, nil
}

// parseSiteList parses site list JSON data
func parseSiteList(data []byte) (map[string]SiteListEntry, error) {
	var siteList map[string]SiteListEntry
	if err := json.Unmarshal(data, &siteList); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	return siteList, nil
}

// loadMetricsData loads metrics data from a JSON file (used for testing)
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

// loadSiteConfig loads site config from a JSON file (legacy format, used for testing)
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

// loadSiteInfo loads site info from a JSON file (used for testing)
func loadSiteInfo(filename string) (*SiteInfo, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseSiteInfo(data)
}

// loadSiteList loads site list from a JSON file (used for testing)
func loadSiteList(filename string) (map[string]SiteListEntry, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parseSiteList(data)
}
