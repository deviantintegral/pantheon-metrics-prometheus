# Pantheon Metrics Prometheus Exporter

A Go application that exposes Pantheon site metrics from the `terminus env:metrics` command in a format that can be scraped by Prometheus.

## Features

- Parses Pantheon metrics JSON data
- Exposes metrics via HTTP for Prometheus scraping
- Includes historical timestamps for each metric
- Supports custom labels for site identification

## Metrics Exposed

The following metrics are exposed:

- `pantheon_visits` - Number of visits
- `pantheon_pages_served` - Number of pages served
- `pantheon_cache_hits` - Number of cache hits
- `pantheon_cache_misses` - Number of cache misses
- `pantheon_cache_hit_ratio` - Cache hit ratio as percentage

Each metric includes the following labels:
- `name` - Site identifier
- `label` - Site name/label
- `plan` - Pantheon plan type

## Building

```bash
go build -o pantheon-metrics-exporter
```

## Running

```bash
./pantheon-metrics-exporter
```

The server will start on port 8080. Metrics are available at: `http://localhost:8080/metrics`

## Configuration

The application requires two JSON files in the same directory:

1. **site-config.json** - Contains site labels and identification:
   ```json
   {
       "name": "site1234",
       "label": "Example Site",
       "plan_name": "Performance Small"
   }
   ```

2. **example-metrics.json** - Contains metrics data from `terminus env:metrics --format=json`

## Example Metrics Output

```
# HELP pantheon_visits Number of visits
# TYPE pantheon_visits gauge
pantheon_visits{label="Example Site",name="site1234",plan="Performance Small"} 837 1762732800000
pantheon_visits{label="Example Site",name="site1234",plan="Performance Small"} 824 1762819200000
```

Note: The timestamps (e.g., 1762732800000) are Unix timestamps in milliseconds, as required by Prometheus for historical metrics.

## Future Enhancements

- Direct integration with `terminus env:metrics --format=json` command
- Support for multiple sites
- Automated metric refresh intervals
- Command-line flags for custom config file paths
