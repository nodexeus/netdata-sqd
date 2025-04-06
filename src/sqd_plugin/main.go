package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/common/expfmt"
)

// WorkerConfig holds the configuration for each worker
type WorkerConfig struct {
	PrometheusURL string `json:"prometheus_url"`
	GraphQLURL    string `json:"graphql_url"`
	Port          int    `json:"port"`
}

// PluginConfig holds the configuration for the plugin
type PluginConfig struct {
	Workers []WorkerConfig `json:"workers"`
}

// PrometheusMetrics represents the metrics we collect from Prometheus
type PrometheusMetrics struct {
	WorkerID            string
	ActiveConnections   float64
	OngoingProbes       float64
	OngoingQueries      float64
	QueueSizes          map[string]float64
	DiscardedMessages   float64
	HeartbeatsPublished float64
	HeartbeatsReceived  float64
	ChunksAvailable     float64
	ChunksDownloading   float64
	ChunksPending       float64
	ChunksDownloaded    float64
	ChunksFailed        float64
	ChunksRemoved       float64
	UsedStorageBytes    float64
	QueriesExecuted     map[string]float64
	QueryResultSize     struct {
		Sum   float64
		Count float64
	}
	NumReadChunks struct {
		Sum   float64
		Count float64
	}
	RunningQueries    float64
	WorkerStatus      string
	GossipsubMessages float64
	IdentifyErrors    float64
	PingRTT           struct {
		Sum   float64
		Count float64
	}
	PingFailures map[string]float64
}

// GraphQLMetrics represents the metrics we collect from GraphQL
type GraphQLMetrics struct {
	APR                    float64
	Bond                   string
	JailReason             *string
	Jailed                 bool
	Name                   string
	Online                 bool
	Queries24Hours         string
	Queries90Days          string
	ScannedData24Hours     string
	ScannedData90Days      string
	ServedData24Hours      string
	ServedData90Days       string
	StakerAPR              float64
	Status                 string
	StoredData             string
	TotalDelegation        string
	TotalDelegationRewards string
	TrafficWeight          float64
	Uptime24Hours          float64
	Uptime90Days           float64
	Version                string
	DelegationCount        int
	ClaimedReward          string
	ClaimableReward        string
	CapedDelegation        string
}

var config PluginConfig

func main() {
	// Read configuration from environment or file
	config.Workers = []WorkerConfig{
		{
			PrometheusURL: "http://localhost:9090",
			GraphQLURL:    "http://localhost:8080",
			Port:          9090,
		},
	}

	// Netdata plugin protocol
	fmt.Println("CHART sqd.workers 'SQD Workers' 'workers' 'workers' 'workers' 'line' 1000")
	fmt.Println("DIMENSION worker_count 'Workers' 'absolute' 1 1")
	fmt.Println("DIMENSION tasks_running 'Running Tasks' 'absolute' 1 1")
	fmt.Println("DIMENSION tasks_queued 'Queued Tasks' 'absolute' 1 1")

	fmt.Println("CHART sqd.cpu 'SQD CPU Usage' 'percentage' 'cpu' 'cpu' 'line' 1000")
	fmt.Println("DIMENSION cpu_usage 'CPU Usage' 'percentage' 1 1")

	fmt.Println("CHART sqd.memory 'SQD Memory Usage' 'MB' 'memory' 'memory' 'line' 1000")
	fmt.Println("DIMENSION memory_usage 'Memory Usage' 'absolute' 1 1")

	fmt.Println("CHART sqd.network 'SQD Network Usage' 'bytes/s' 'network' 'network' 'line' 1000")
	fmt.Println("DIMENSION network_in 'Network In' 'absolute' 1 1")
	fmt.Println("DIMENSION network_out 'Network Out' 'absolute' 1 1")

	fmt.Println("CHART sqd.performance 'SQD Performance' 'ms' 'performance' 'performance' 'line' 1000")
	fmt.Println("DIMENSION query_latency 'Query Latency' 'absolute' 1 1")
	fmt.Println("DIMENSION indexing_rate 'Indexing Rate' 'absolute' 1 1")

	fmt.Println("CHART sqd.indexing 'SQD Indexing' 'blocks' 'indexing' 'indexing' 'line' 1000")
	fmt.Println("DIMENSION indexed_blocks 'Indexed Blocks' 'absolute' 1 1")
	fmt.Println("DIMENSION indexing_progress 'Indexing Progress' 'percentage' 1 1")

	// Main collection loop
	for {
		collectMetrics()
		time.Sleep(1 * time.Second)
	}
}

func collectMetrics() {
	var totalTasksRunning, totalTasksQueued int
	var totalCPU, totalMemory, totalNetworkIn, totalNetworkOut float64
	var totalQueryLatency, totalIndexingRate float64
	var totalIndexingProgress float64
	workerCount := 0

	for _, worker := range config.Workers {
		// Collect Prometheus metrics
		promMetrics, err := getPrometheusMetrics(worker.PrometheusURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting Prometheus metrics from %s: %v\n", worker.PrometheusURL, err)
			continue
		}

		totalTasksRunning += int(promMetrics.RunningQueries)
		totalTasksQueued += int(promMetrics.OngoingQueries)
		totalCPU += promMetrics.ActiveConnections
		totalMemory += promMetrics.UsedStorageBytes
		totalNetworkIn += promMetrics.QueueSizes["in"]
		totalNetworkOut += promMetrics.QueueSizes["out"]
		totalQueryLatency += promMetrics.QueryResultSize.Sum / promMetrics.QueryResultSize.Count
		totalIndexingRate += promMetrics.NumReadChunks.Sum / promMetrics.NumReadChunks.Count
		totalIndexingProgress += 1.0 // Assuming indexing_progress is 100% for simplicity
		workerCount++
	}

	// Output metrics in Netdata format
	fmt.Printf("BEGIN sqd.workers\n")
	fmt.Printf("SET worker_count = %d\n", workerCount)
	fmt.Printf("SET tasks_running = %d\n", totalTasksRunning)
	fmt.Printf("SET tasks_queued = %d\n", totalTasksQueued)
	fmt.Printf("END\n")

	fmt.Printf("BEGIN sqd.cpu\n")
	fmt.Printf("SET cpu_usage = %.2f\n", totalCPU/float64(workerCount))
	fmt.Printf("END\n")

	fmt.Printf("BEGIN sqd.memory\n")
	fmt.Printf("SET memory_usage = %.2f\n", totalMemory/float64(workerCount))
	fmt.Printf("END\n")

	fmt.Printf("BEGIN sqd.network\n")
	fmt.Printf("SET network_in = %.2f\n", totalNetworkIn/float64(workerCount))
	fmt.Printf("SET network_out = %.2f\n", totalNetworkOut/float64(workerCount))
	fmt.Printf("END\n")

	fmt.Printf("BEGIN sqd.performance\n")
	fmt.Printf("SET query_latency = %.2f\n", totalQueryLatency/float64(workerCount))
	fmt.Printf("SET indexing_rate = %.2f\n", totalIndexingRate/float64(workerCount))
	fmt.Printf("END\n")

	fmt.Printf("BEGIN sqd.indexing\n")
	fmt.Printf("SET indexing_progress = %.2f\n", totalIndexingProgress/float64(workerCount))
	fmt.Printf("END\n")
}

func getPrometheusMetrics(url string) (*PrometheusMetrics, error) {
	resp, err := http.Get(url + "/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		return nil, err
	}

	metrics := &PrometheusMetrics{
		QueueSizes:      make(map[string]float64),
		QueriesExecuted: make(map[string]float64),
		PingFailures:    make(map[string]float64),
	}

	// Extract metrics from Prometheus response
	for name, mf := range metricFamilies {
		for _, m := range mf.Metric {
			// Get worker_id from labels
			for _, l := range m.Label {
				if l.GetName() == "worker_id" {
					metrics.WorkerID = l.GetValue()
				}
			}

			// Map Prometheus metrics to our structure
			switch name {
			case "active_connections":
				metrics.ActiveConnections = m.Gauge.GetValue()
			case "ongoing_probes":
				metrics.OngoingProbes = m.Gauge.GetValue()
			case "ongoing_queries":
				metrics.OngoingQueries = m.Gauge.GetValue()
			case "queue_size":
				for _, l := range m.Label {
					if l.GetName() == "queue_name" {
						metrics.QueueSizes[l.GetValue()] = m.Gauge.GetValue()
					}
				}
			case "discarded_messages_total":
				metrics.DiscardedMessages = m.Counter.GetValue()
			case "heartbeats_published_total":
				metrics.HeartbeatsPublished = m.Counter.GetValue()
			case "heartbeats_received_total":
				metrics.HeartbeatsReceived = m.Counter.GetValue()
			case "chunks_available":
				metrics.ChunksAvailable = m.Gauge.GetValue()
			case "chunks_downloading":
				metrics.ChunksDownloading = m.Gauge.GetValue()
			case "chunks_pending":
				metrics.ChunksPending = m.Gauge.GetValue()
			case "chunks_downloaded_total":
				metrics.ChunksDownloaded = m.Counter.GetValue()
			case "chunks_failed_download_total":
				metrics.ChunksFailed = m.Counter.GetValue()
			case "chunks_removed_total":
				metrics.ChunksRemoved = m.Counter.GetValue()
			case "used_storage_bytes":
				metrics.UsedStorageBytes = m.Gauge.GetValue()
			case "num_queries_executed_total":
				for _, l := range m.Label {
					if l.GetName() == "status" {
						metrics.QueriesExecuted[l.GetValue()] = m.Counter.GetValue()
					}
				}
			case "query_result_size_bytes_sum":
				metrics.QueryResultSize.Sum = m.Summary.GetSampleSum()
			case "query_result_size_bytes_count":
				metrics.QueryResultSize.Count = float64(m.Summary.GetSampleCount())
			case "num_read_chunks_sum":
				metrics.NumReadChunks.Sum = m.Summary.GetSampleSum()
			case "num_read_chunks_count":
				metrics.NumReadChunks.Count = float64(m.Summary.GetSampleCount())
			case "running_queries":
				metrics.RunningQueries = m.Gauge.GetValue()
			case "worker_status":
				for _, l := range m.Label {
					if l.GetName() == "worker_status" {
						metrics.WorkerStatus = l.GetValue()
					}
				}
			case "libp2p_gossipsub_messages_total":
				metrics.GossipsubMessages = m.Counter.GetValue()
			case "libp2p_identify_errors_total":
				metrics.IdentifyErrors = m.Counter.GetValue()
			case "libp2p_ping_rtt_seconds_sum":
				metrics.PingRTT.Sum = m.Summary.GetSampleSum()
			case "libp2p_ping_rtt_seconds_count":
				metrics.PingRTT.Count = float64(m.Summary.GetSampleCount())
			case "libp2p_ping_failure_total":
				for _, l := range m.Label {
					if l.GetName() == "reason" {
						metrics.PingFailures[l.GetValue()] = m.Counter.GetValue()
					}
				}
			}
		}
	}

	return metrics, nil
}

func getGraphQLMetrics(url, workerID string) (*GraphQLMetrics, error) {
	query := fmt.Sprintf(`{
		workers(where: {peerId_eq: "%s"}) {
			apr
			bond
			jailReason
			jailed
			name
			online
			queries24Hours
			queries90Days
			scannedData24Hours
			scannedData90Days
			servedData24Hours
			servedData90Days
			stakerApr
			status
			storedData
			totalDelegation
			totalDelegationRewards
			trafficWeight
			uptime24Hours
			uptime90Days
			version
			delegationCount
			claimedReward
			claimableReward
			capedDelegation
		}
	}`, workerID)

	reqBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Workers []GraphQLMetrics `json:"workers"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data.Workers) == 0 {
		return nil, fmt.Errorf("no worker found with ID %s", workerID)
	}

	return &result.Data.Workers[0], nil
}
