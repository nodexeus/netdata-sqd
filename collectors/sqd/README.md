# SQD Collector

This collector monitors [SQD](https://sqd.io/) worker nodes by collecting metrics from Prometheus and GraphQL endpoints.

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

## Configuration

Edit the `go.d/sqd.conf` configuration file using `edit-config` from the Netdata [config directory](https://github.com/netdata/netdata/blob/master/docs/configure/nodes.md), which is typically at `/etc/netdata`.

```bash
cd /etc/netdata # Replace with your Netdata config directory
sudo ./edit-config go.d/sqd.conf
```

Here's a sample configuration:

```yaml
update_every: 1
workers:
  - name: worker1
    prometheus_url: http://localhost:9090
    graphql_url: http://localhost:8080
    port: 9090
  - name: worker2
    prometheus_url: http://remote-server:9090
    graphql_url: http://remote-server:8080
    port: 9090
```

You can add any number of workers by adding more entries to the `workers` array.

### Configuration Options

The following options can be defined globally or per worker:

| Option | Description | Default Value |
|--------|-------------|---------------|
| update_every | Data collection interval in seconds | 1 |
| name | A unique name for the worker | worker |
| prometheus_url | URL of the Prometheus metrics endpoint | http://localhost:9090 |
| graphql_url | URL of the GraphQL API endpoint | http://localhost:8080 |
| port | Port number of the worker | 9090 |

## Troubleshooting

### Common Issues

- **No metrics collected**: Ensure the worker nodes are running and accessible via the provided URLs.
- **Missing worker metrics**: Check if the worker has the proper `worker_id` label in its Prometheus metrics.
- **GraphQL query failures**: Verify the GraphQL endpoint is correctly configured and responding to queries.

### Debug Mode

To enable debug mode and get more detailed logs, update the `go.d.conf` file:

```yaml
modules:
  sqd: yes
  debug: yes
```

## Requirements

- SQD worker nodes with:
  - Prometheus metrics endpoint
  - GraphQL API endpoint
  - `worker_id` label in Prometheus metrics
