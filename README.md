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

Currently, the application reads metrics from `example-metrics.json` in the same directory. The labels are hardcoded in the main function:

- `name`: "example1234"
- `label`: "Example.com"
- `plan`: "Performance Small"

## Example Metrics Output

```
# HELP pantheon_visits Number of visits
# TYPE pantheon_visits gauge
pantheon_visits{label="Example.com",name="example1234",plan="Performance Small"} 837 1762732800000
pantheon_visits{label="Example.com",name="example1234",plan="Performance Small"} 824 1762819200000
```

Note: The timestamps (e.g., 1762732800000) are Unix timestamps in milliseconds, as required by Prometheus for historical metrics.

## Future Enhancements

- Direct integration with `terminus env:metrics --format=json` command
- Configuration file for labels and settings
- Support for multiple sites
- Automated metric refresh intervals
