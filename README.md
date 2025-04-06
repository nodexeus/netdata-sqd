# Netdata SQD Collector

This is a Netdata collector plugin written in Go that monitors SQD worker nodes. It collects various metrics from both Prometheus and GraphQL endpoints and follows the official netdata collector architecture for optimal integration.

## Features

- Implements the standard netdata collector architecture
- Monitors multiple SQD worker nodes
- Collects comprehensive metrics from Prometheus endpoints
- Collects additional metrics from GraphQL endpoints
- Configurable update interval and worker endpoints
- Official netdata collector file structure

## Directory Structure

```
netdata-sqd/
├── build.sh                 # Build and installation script
├── collectors/
│   └── sqd/                 # SQD collector module
│       ├── charts.go        # Chart definitions
│       ├── collect.go       # Data collection logic
│       ├── config.go        # Configuration structure and validation
│       ├── config_schema.json # JSON schema for configuration
│       ├── init.go          # Module registration
│       ├── metadata.yaml    # Dashboard and metrics metadata
│       ├── README.md        # Module documentation
│       └── sqd.go           # Core collector implementation
├── cmd/
│   └── go.d.plugin/         # Main plugin entry point
│       └── main.go
└── go.mod                   # Dependencies
```

## Building and Installation

The project includes a build script that simplifies building and installation:

```bash
# Just build the collector
./build.sh

# Build with debug symbols
./build.sh --debug

# Build and install to netdata
./build.sh --install
```

The build script will:
1. Build the collector with optimized settings (or debug symbols if requested)
2. Optionally install the collector to the netdata plugins directory
3. Create a default configuration file in the netdata config directory

### Integration with netdata

This collector follows the structure of the official netdata Go collectors but is designed to be used as a standalone plugin initially. Full integration with the netdata go.d.plugin system requires additional setup:

1. Clone the netdata go.d.plugin repository
2. Copy the sqd collector files to the appropriate directories in the go.d.plugin codebase
3. Register the module in the go.d.plugin module registry
4. Build the go.d.plugin with the SQD collector included

Refer to the [netdata go.d.plugin documentation](https://github.com/netdata/go.d.plugin) for details on integrating custom collectors.

## Configuration

The collector is configured through a YAML configuration file located at `/etc/netdata/go.d/sqd.conf`.

### Configuration Format

```yaml
# SQD collector configuration
update_every: 1

workers:
  - name: default
    prometheus_url: http://localhost:9090
    graphql_url: http://localhost:8080
    port: 9090
  
  - name: worker2
    prometheus_url: http://remote-server:9090
    graphql_url: http://remote-server:8080
    port: 9090
```

### Configuration Options

| Option | Description | Default Value |
|--------|-------------|---------------|
| update_every | Data collection interval in seconds | 1 |
| name | A unique name for the worker | worker |
| prometheus_url | URL of the Prometheus metrics endpoint | http://localhost:9090 |
| graphql_url | URL of the GraphQL API endpoint | http://localhost:8080 |
| port | Port number of the worker | 9090 |

## Metrics

The collector gathers the following metrics for each configured worker:

### Worker Status
- Worker online status
- Worker jailed status

### Worker Connections
- Active connections count

### Worker Queries
- Ongoing queries
- Running queries

### Worker Storage
- Used storage bytes

### Worker Chunks
- Available chunks
- Downloading chunks
- Pending chunks

### Worker Performance
- Uptime over 24 hours (percentage)
- Uptime over 90 days (percentage)

### Worker Rewards
- APR (Annual Percentage Rate)
- Staker APR

### Worker Delegation
- Delegation count

## Requirements

- Go 1.18 or later
- Netdata
- SQD worker nodes with:
  - Prometheus metrics endpoint
  - GraphQL API endpoint
  - `worker_id` label in Prometheus metrics

## Development

The collector follows the official netdata collector architecture:

- **Module Registration**: Each collector registers itself in `init.go`
- **Configuration**: Configuration is defined and validated in `config.go`
- **Charts**: Chart definitions are in `charts.go`
- **Data Collection**: The collection logic is in `collect.go`

To extend the collector:
1. Add new metrics to the appropriate structs in `sqd.go`
2. Add metric collection logic in `collect.go`
3. Update the chart definitions in `charts.go`
4. Update `metadata.yaml` with information about the new metrics

## Troubleshooting

If you encounter issues with the collector, you can enable debug mode by editing the `go.d.conf` file:

```yaml
modules:
  sqd: yes
  debug: yes
```

This will provide more detailed logs about the collector's operation.

## License

MIT 