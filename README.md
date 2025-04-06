# Netdata SQD Plugin

This is a Netdata plugin written in Go that monitors SQD worker nodes. It collects various metrics from both Prometheus and GraphQL endpoints and follows the netdata Module interface pattern for optimal integration.

## Features

- Implements the standard netdata Module interface pattern
- Monitors multiple SQD worker nodes
- Collects comprehensive metrics from Prometheus endpoints
- Collects additional metrics from GraphQL endpoints
- Configurable update interval and worker endpoints
- Modular and maintainable code structure

## Architecture

The plugin follows the netdata Module interface pattern with the following components:

- **Main Program**: Initializes the collector, sets up chart definitions, and handles the collection loop
- **SQD Module**: Implements the Module interface with Init(), Check(), Charts(), and Collect() methods
- **Charts**: Defines chart structures with dimensions for visualization in netdata

## Installation

1. Build the plugin:
```bash
go build -o sqd_plugin
```

2. Make the plugin executable:
```bash
chmod +x sqd_plugin
```

3. Move the plugin to Netdata's plugins directory:
```bash
sudo cp sqd_plugin /usr/libexec/netdata/plugins.d/
```

## Configuration

The plugin can be configured through a JSON configuration file:

### JSON Configuration
Create a `sqd.conf` file in one of these locations:
- `./sqd.conf` (current directory)
- `/etc/netdata/sqd.conf`
- `$HOME/.config/netdata/sqd.conf`

Configuration structure:
```json
{
  "update_every": 1,
  "workers": [
    {
      "name": "default",
      "prometheus_url": "http://localhost:9090",
      "graphql_url": "http://localhost:8080",
      "port": 9090
    }
  ]
}
```

Parameters:
- `update_every`: Update interval in seconds (default: 1)
- `workers`: Array of worker configurations
  - `name`: Friendly name for the worker (used in chart identifiers)
  - `prometheus_url`: URL to the Prometheus metrics endpoint
  - `graphql_url`: URL to the GraphQL API endpoint
  - `port`: Port number of the worker

## Metrics

The plugin collects the following metrics for each worker:

### Worker Status Charts
- Online status
- Jailed status

### Connections Chart
- Active connections

### Queries Chart
- Ongoing queries
- Running queries

### Storage Chart
- Used storage bytes

### Chunks Chart
- Available chunks
- Downloading chunks
- Pending chunks

### Performance Chart
- Uptime 24 hours (percentage)
- Uptime 90 days (percentage)

### Rewards Chart
- APR (percentage)
- Staker APR (percentage)

### Delegation Chart
- Delegation count

## Requirements

- Go 1.16 or later
- Netdata
- SQD worker nodes with:
  - Prometheus metrics endpoint
  - GraphQL API endpoint
  - `worker_id` label in Prometheus metrics

## Development

The codebase is organized as follows:

- `main.go`: Entry point that loads configuration and sets up the collector loop
- `module/sqd.go`: Core implementation of the SQD collector module
- `module/charts.go`: Chart definitions and helper functions

To extend the plugin:
1. Add new metrics to the `PrometheusMetrics` or `GraphQLMetrics` structs
2. Add metric collection logic in `getPrometheusMetrics()` or `getGraphQLMetrics()`
3. Update the `initCharts()` method to define new charts
4. Map the metrics to chart dimensions in the `Collect()` method

## License

MIT 