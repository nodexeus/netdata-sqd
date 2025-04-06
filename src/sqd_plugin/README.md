# Netdata SQD Plugin

This is a Netdata plugin written in Go that monitors SQD worker nodes. It collects various metrics from both Prometheus and GraphQL endpoints.

## Features

- Monitors multiple SQD worker nodes
- Collects metrics from Prometheus endpoints
- Collects additional metrics from GraphQL endpoints
- Real-time monitoring with 1-second update interval
- Configurable ports and endpoints

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

The plugin can be configured through a JSON configuration file or environment variables:

### JSON Configuration
Create a `config.json` file with the following structure:
```json
{
  "workers": [
    {
      "prometheus_url": "http://localhost:9090",
      "graphql_url": "http://localhost:8080",
      "port": 9090
    }
  ]
}
```

### Environment Variables
- `NETDATA_SQD_WORKERS`: JSON string containing worker configurations
- `NETDATA_SQD_UPDATE_EVERY`: Update interval in seconds (default: 1)

## Metrics

The plugin collects the following metrics:

### From Prometheus
- CPU usage (percentage)
- Memory usage (MB)
- Network traffic (bytes/s)
- Query latency (ms)
- Indexing rate
- Running and queued tasks

### From GraphQL
- Indexed blocks count
- Query statistics
- Cache hit rate
- Indexing progress
- Performance metrics

## Requirements

- Go 1.16 or later
- Netdata
- SQD worker nodes with:
  - Prometheus metrics endpoint
  - GraphQL API endpoint
  - `peer_id` label in Prometheus metrics

## License

MIT 