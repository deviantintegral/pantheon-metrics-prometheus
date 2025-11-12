# Pantheon Metrics Prometheus Exporter

A Go application that fetches Pantheon site metrics using the Terminus CLI and exposes them in Prometheus format for scraping.

## Features

- Fetches site information using `terminus site:info`
- Fetches metrics using `terminus env:metrics`
- Exposes metrics via HTTP for Prometheus scraping
- Includes historical timestamps for each metric
- Automatically extracts site labels (name, label, plan) from Terminus

## Prerequisites

- Go 1.16 or later
- [Terminus CLI](https://pantheon.io/docs/terminus/install) installed and authenticated
- Pantheon site access with appropriate permissions

## Installation

```bash
go build -o pantheon-metrics-exporter
```

## Usage

The application requires the Terminus CLI to be installed and authenticated. Run:

```bash
./pantheon-metrics-exporter -site=SITENAME [-env=ENVIRONMENT] [-port=PORT]
```

### Command-Line Flags

- `-site` (required): Pantheon site name
- `-env` (optional): Pantheon environment (default: `live`)
- `-port` (optional): HTTP server port (default: `8080`)

### Examples

```bash
# Export metrics for the live environment
./pantheon-metrics-exporter -site=my-site

# Export metrics for the dev environment
./pantheon-metrics-exporter -site=my-site -env=dev

# Use a custom port
./pantheon-metrics-exporter -site=my-site -port=9090
```

## Metrics Exposed

The following metrics are exposed:

- `pantheon_visits` - Number of visits
- `pantheon_pages_served` - Number of pages served
- `pantheon_cache_hits` - Number of cache hits
- `pantheon_cache_misses` - Number of cache misses
- `pantheon_cache_hit_ratio` - Cache hit ratio as percentage

Each metric includes the following labels:
- `name` - Site identifier (from `terminus site:info`)
- `label` - Site display name (from `terminus site:info`)
- `plan` - Pantheon plan type (from `terminus site:info`)

## Example Metrics Output

```
# HELP pantheon_visits Number of visits
# TYPE pantheon_visits gauge
pantheon_visits{label="Example Site",name="site1234",plan="Performance Small"} 837 1762732800000
pantheon_visits{label="Example Site",name="site1234",plan="Performance Small"} 824 1762819200000

# HELP pantheon_cache_hit_ratio Cache hit ratio as percentage
# TYPE pantheon_cache_hit_ratio gauge
pantheon_cache_hit_ratio{label="Example Site",name="site1234",plan="Performance Small"} 3.86 1762732800000
pantheon_cache_hit_ratio{label="Example Site",name="site1234",plan="Performance Small"} 5.12 1762819200000
```

Note: The timestamps (e.g., 1762732800000) are Unix timestamps in milliseconds, as required by Prometheus for historical metrics.

## Prometheus Configuration

Add the following to your `prometheus.yml` configuration:

```yaml
scrape_configs:
  - job_name: 'pantheon-metrics'
    static_configs:
      - targets: ['localhost:8080']
    # Increase scrape interval if you're rate-limited by Terminus
    scrape_interval: 5m
```

## Development

### Running Tests

```bash
go test -v
```

### Test Data

Test data is located in the `testdata/` directory:
- `example-metrics.json` - Sample metrics output from `terminus env:metrics`
- `site-info.json` - Sample site info output from `terminus site:info`
- `site-config.json` - Legacy configuration format (for backwards compatibility)

## Future Enhancements

- Periodic metric refresh with configurable intervals
- Support for multiple sites in a single instance
- Caching to reduce Terminus API calls
- Prometheus service discovery integration
- Containerization with Docker

## License

This project is open source and available for use and modification.
