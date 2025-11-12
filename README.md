# Pantheon Metrics Prometheus Exporter

A Go application that fetches Pantheon site metrics using the Terminus CLI and exposes them in Prometheus format for scraping. The exporter automatically discovers all sites across multiple Pantheon accounts and collects metrics for each one.

## Features

- **Multi-account support**: Monitor sites across multiple Pantheon accounts simultaneously
- Automatically discovers all sites using `terminus site:list`
- Fetches metrics for all accessible sites using `terminus env:metrics`
- Exposes metrics via HTTP for Prometheus scraping
- Includes historical timestamps for each metric
- Gracefully handles inaccessible sites and accounts
- Supports monitoring multiple sites in a single exporter instance
- Account-based labeling for distinguishing metrics across accounts

## Prerequisites

- Go 1.16 or later
- [Terminus CLI](https://pantheon.io/docs/terminus/install) installed
- One or more Pantheon machine tokens

## Installation

```bash
go build -o pantheon-metrics-exporter
```

## Usage

### Setting Up Machine Tokens

The exporter requires Pantheon machine tokens to authenticate. Set the `PANTHEON_MACHINE_TOKENS` environment variable with one or more space-separated tokens:

```bash
# Single account
export PANTHEON_MACHINE_TOKENS="your-machine-token-here"

# Multiple accounts
export PANTHEON_MACHINE_TOKENS="token1 token2 token3"
```

To create a machine token:
1. Log into your Pantheon Dashboard
2. Go to Account > Machine Tokens
3. Click "Create Token"
4. Copy the generated token

### Running the Exporter

```bash
./pantheon-metrics-exporter [-env=ENVIRONMENT] [-port=PORT] [-refreshInterval=MINUTES]
```

### Command-Line Flags

- `-env` (optional): Pantheon environment to monitor (default: `live`)
- `-port` (optional): HTTP server port (default: `8080`)
- `-refreshInterval` (optional): How often to refresh data in minutes (default: `60`)

### Examples

```bash
# Export metrics for all sites across all configured accounts
export PANTHEON_MACHINE_TOKENS="token1 token2"
./pantheon-metrics-exporter

# Monitor dev environment
export PANTHEON_MACHINE_TOKENS="your-token"
./pantheon-metrics-exporter -env=dev

# Use a custom port
export PANTHEON_MACHINE_TOKENS="your-token"
./pantheon-metrics-exporter -port=9090

# Refresh data every 30 minutes
export PANTHEON_MACHINE_TOKENS="your-token"
./pantheon-metrics-exporter -refreshInterval=30
```

## How It Works

### Initial Startup

1. On startup, the exporter reads machine tokens from the `PANTHEON_MACHINE_TOKENS` environment variable
2. For each token, the exporter:
   - Authenticates using `terminus auth:login --machine-token=TOKEN`
   - Runs `terminus site:list --format=json` to discover all sites for that account
   - Fetches metrics for each site using `terminus env:metrics SITE.ENV --format=json`
   - Labels all metrics with an account identifier (last 8 characters of the token)
3. Sites or accounts that are inaccessible or return errors are logged and skipped
4. Successfully collected metrics from all accounts are aggregated and exposed via the `/metrics` endpoint
5. The exporter shows a summary page at the root URL listing all monitored accounts and sites

### Periodic Refresh

The exporter automatically refreshes data at the interval specified by `-refreshInterval`:

1. **Site List Refresh**: Every refresh interval, the exporter re-runs `terminus site:list` for all accounts to detect added or removed sites
2. **Metrics Refresh**: Metrics are refreshed using a queue-based system to prevent API stampedes:
   - Sites are distributed evenly across the refresh interval
   - For example, with 100 sites and a 60-minute interval: 2 sites are processed every minute (100 / 60 = 1.67, rounded up)
   - This ensures steady API usage rather than bursts of requests
   - The queue automatically cycles through all sites continuously

## Metrics Exposed

The following metrics are exposed for each site:

- `pantheon_visits` - Number of visits
- `pantheon_pages_served` - Number of pages served
- `pantheon_cache_hits` - Number of cache hits
- `pantheon_cache_misses` - Number of cache misses
- `pantheon_cache_hit_ratio` - Cache hit ratio as percentage

Each metric includes the following labels:
- `name` - Site identifier (from `terminus site:list`)
- `label` - Site name (currently same as name, as site:list doesn't provide a separate display name)
- `plan` - Pantheon plan type (from `terminus site:list`)
- `account` - Account identifier (last 8 characters of the machine token for privacy)

## Example Metrics Output

```
# HELP pantheon_visits Number of visits
# TYPE pantheon_visits gauge
pantheon_visits{account="abc12345",label="site1234",name="site1234",plan="Performance Small"} 837 1762732800000
pantheon_visits{account="abc12345",label="site1234",name="site1234",plan="Performance Small"} 824 1762819200000
pantheon_visits{account="def67890",label="site5678",name="site5678",plan="Basic"} 456 1762732800000
pantheon_visits{account="def67890",label="site5678",name="site5678",plan="Basic"} 478 1762819200000

# HELP pantheon_cache_hit_ratio Cache hit ratio as percentage
# TYPE pantheon_cache_hit_ratio gauge
pantheon_cache_hit_ratio{account="abc12345",label="site1234",name="site1234",plan="Performance Small"} 3.86 1762732800000
pantheon_cache_hit_ratio{account="abc12345",label="site1234",name="site1234",plan="Performance Small"} 5.12 1762819200000
```

Note: The timestamps (e.g., 1762732800000) are Unix timestamps in milliseconds, as required by Prometheus for historical metrics.

## Prometheus Configuration

Add the following to your `prometheus.yml` configuration:

```yaml
scrape_configs:
  - job_name: 'pantheon-metrics'
    static_configs:
      - targets: ['localhost:8080']
    # Scrape interval can be frequent since the exporter handles refresh internally
    scrape_interval: 1m
    scrape_timeout: 30s
```

**Note:** With the built-in refresh mechanism, Prometheus can scrape frequently (e.g., every 1 minute) without causing API stampedes. The exporter manages Terminus API calls internally using the queue-based refresh system.

## Error Handling

The exporter handles errors gracefully:
- Sites that fail to return metrics are logged with a warning and skipped
- The exporter will start successfully as long as at least one site returns metrics
- Individual metric parsing errors are logged but don't prevent other metrics from being collected

## Development

### Running Tests

```bash
go test -v
```

### Test Data

Test data is located in the `testdata/` directory:
- `example-metrics.json` - Sample metrics output from `terminus env:metrics`
- `site-info.json` - Sample site info output from `terminus site:info`
- `site-list.json` - Sample site list output from `terminus site:list`
- `site-config.json` - Legacy configuration format (for backwards compatibility)

## Architecture

The application consists of several key components:

- **SiteListEntry**: Represents a site from `terminus site:list`
- **SiteMetrics**: Holds metrics data for a specific site including account identifier
- **PantheonCollector**: Thread-safe Prometheus Collector for multiple sites across multiple accounts
- **RefreshManager**: Manages periodic site list and metrics refresh
- **authenticateWithToken()**: Authenticates with Terminus using a machine token
- **getAccountID()**: Generates an account identifier from a token (last 8 chars)
- **fetchAllSites()**: Retrieves all sites from Terminus for the currently authenticated account
- **fetchMetricsData()**: Fetches metrics for a specific site/environment

The application flow:
1. **Startup**: Authenticate with each token, fetch sites, collect initial metrics
2. **Site List Refresh**: Every refresh interval, update the list of monitored sites
3. **Metrics Queue**: Continuously process metrics updates distributed evenly over the refresh interval
4. **Thread Safety**: All collector updates use mutex locks to prevent race conditions

## Future Enhancements

- Support for fetching `terminus site:info` to get proper site labels
- Concurrent metric fetching for faster startup
- Metrics endpoint to show last update time and collection status per site
- Support for filtering which sites to monitor (include/exclude patterns)
- Configurable queue processing rate (currently fixed at 1-minute intervals)
- Prometheus service discovery integration
- Containerization with Docker
- Health check endpoint

## Troubleshooting

### PANTHEON_MACHINE_TOKENS not set
- Ensure you've set the environment variable before running: `export PANTHEON_MACHINE_TOKENS="token1 token2"`
- Verify it's set: `echo $PANTHEON_MACHINE_TOKENS`

### Authentication failures
- Verify your machine tokens are valid
- Check that tokens haven't expired or been revoked in your Pantheon Dashboard
- Test authentication manually: `terminus auth:login --machine-token=YOUR_TOKEN`

### No sites found for an account
- The application will skip accounts with authentication failures
- Check the application logs for warning messages about specific accounts
- Verify the account has sites: `terminus site:list` (after manual authentication)

### Metrics not appearing for some sites
- Check the application logs for warning messages
- Verify you have permission to view metrics for those sites
- Try running `terminus env:metrics SITE.ENV --format=json` manually
- Some sites may not have metrics available for the specified environment

### Slow startup with multiple accounts
- This is normal when monitoring many accounts and sites
- Each account authentication and each site requires a separate API call to Terminus
- Startup time scales with: (number of accounts) × (average sites per account)
- Consider the concurrent fetching enhancement mentioned in Future Enhancements

## License

This project is open source and available for use and modification.
