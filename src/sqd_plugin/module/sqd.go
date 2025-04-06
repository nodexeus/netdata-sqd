package module

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/common/expfmt"
)

// SQDCollector represents the SQD netdata collector module
type SQDCollector struct {
	Workers []WorkerConfig
	charts  *Charts

	// Metrics storage
	prometheusMetrics map[string]*PrometheusMetrics
	graphqlMetrics    map[string]*GraphQLMetrics
}

// WorkerConfig holds configuration for each SQD worker
type WorkerConfig struct {
	PrometheusURL string `yaml:"prometheus_url"`
	GraphQLURL    string `yaml:"graphql_url"`
	Port          int    `yaml:"port"`
	Name          string `yaml:"name"`
}

// PrometheusMetrics represents metrics collected from Prometheus
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

// GraphQLMetrics represents metrics collected from GraphQL
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

// New creates a new SQD collector instance
func New() *SQDCollector {
	return &SQDCollector{
		prometheusMetrics: make(map[string]*PrometheusMetrics),
		graphqlMetrics:    make(map[string]*GraphQLMetrics),
	}
}

// Init initializes the SQD collector
func (sc *SQDCollector) Init() bool {
	// Initialize with default values if none provided
	if len(sc.Workers) == 0 {
		sc.Workers = []WorkerConfig{
			{
				Name:          "default",
				PrometheusURL: "http://localhost:9090",
				GraphQLURL:    "http://localhost:8080",
				Port:          9090,
			},
		}
	}

	// Initialize metrics maps
	for _, worker := range sc.Workers {
		workerId := fmt.Sprintf("%s:%d", worker.Name, worker.Port)
		
		sc.prometheusMetrics[workerId] = &PrometheusMetrics{
			WorkerID:        workerId,
			QueueSizes:      make(map[string]float64),
			QueriesExecuted: make(map[string]float64),
			PingFailures:    make(map[string]float64),
		}
		
		sc.graphqlMetrics[workerId] = &GraphQLMetrics{}
	}

	if err := sc.initCharts(); err != nil {
		return false
	}

	return true
}

// Check verifies the collector is properly configured
func (sc *SQDCollector) Check() bool {
	return len(sc.Workers) > 0
}

// Charts returns the charts for this collector
func (sc *SQDCollector) Charts() *Charts {
	return sc.charts
}

// Cleanup performs cleanup if needed
func (sc *SQDCollector) Cleanup() {}

// Collect gathers metrics from SQD workers
func (sc *SQDCollector) Collect() map[string]int64 {
	metrics := make(map[string]int64)
	
	for _, worker := range sc.Workers {
		workerId := fmt.Sprintf("%s:%d", worker.Name, worker.Port)
		
		// Get Prometheus metrics
		promMetrics, err := sc.getPrometheusMetrics(worker.PrometheusURL, workerId)
		if err == nil && promMetrics != nil {
			sc.prometheusMetrics[workerId] = promMetrics
		}
		
		// Get GraphQL metrics
		gqlMetrics, err := sc.getGraphQLMetrics(worker.GraphQLURL, workerId)
		if err == nil && gqlMetrics != nil {
			sc.graphqlMetrics[workerId] = gqlMetrics
		}
		
		// Map metrics to netdata format
		if promMetrics != nil {
			metrics[fmt.Sprintf("worker_%s_active_connections", workerId)] = int64(promMetrics.ActiveConnections)
			metrics[fmt.Sprintf("worker_%s_ongoing_probes", workerId)] = int64(promMetrics.OngoingProbes)
			metrics[fmt.Sprintf("worker_%s_ongoing_queries", workerId)] = int64(promMetrics.OngoingQueries)
			metrics[fmt.Sprintf("worker_%s_heartbeats_published", workerId)] = int64(promMetrics.HeartbeatsPublished)
			metrics[fmt.Sprintf("worker_%s_heartbeats_received", workerId)] = int64(promMetrics.HeartbeatsReceived)
			metrics[fmt.Sprintf("worker_%s_chunks_available", workerId)] = int64(promMetrics.ChunksAvailable)
			metrics[fmt.Sprintf("worker_%s_chunks_downloading", workerId)] = int64(promMetrics.ChunksDownloading)
			metrics[fmt.Sprintf("worker_%s_chunks_pending", workerId)] = int64(promMetrics.ChunksPending)
			metrics[fmt.Sprintf("worker_%s_chunks_downloaded", workerId)] = int64(promMetrics.ChunksDownloaded)
			metrics[fmt.Sprintf("worker_%s_chunks_failed", workerId)] = int64(promMetrics.ChunksFailed)
			metrics[fmt.Sprintf("worker_%s_chunks_removed", workerId)] = int64(promMetrics.ChunksRemoved)
			metrics[fmt.Sprintf("worker_%s_storage_bytes", workerId)] = int64(promMetrics.UsedStorageBytes)
			metrics[fmt.Sprintf("worker_%s_running_queries", workerId)] = int64(promMetrics.RunningQueries)
		}
		
		if gqlMetrics != nil {
			metrics[fmt.Sprintf("worker_%s_online", workerId)] = boolToInt64(gqlMetrics.Online)
			metrics[fmt.Sprintf("worker_%s_jailed", workerId)] = boolToInt64(gqlMetrics.Jailed)
			metrics[fmt.Sprintf("worker_%s_apr", workerId)] = int64(gqlMetrics.APR * 1000)
			metrics[fmt.Sprintf("worker_%s_staker_apr", workerId)] = int64(gqlMetrics.StakerAPR * 1000)
			metrics[fmt.Sprintf("worker_%s_uptime_24h", workerId)] = int64(gqlMetrics.Uptime24Hours * 1000)
			metrics[fmt.Sprintf("worker_%s_uptime_90d", workerId)] = int64(gqlMetrics.Uptime90Days * 1000)
			metrics[fmt.Sprintf("worker_%s_traffic_weight", workerId)] = int64(gqlMetrics.TrafficWeight * 1000)
			metrics[fmt.Sprintf("worker_%s_delegation_count", workerId)] = int64(gqlMetrics.DelegationCount)
		}
	}
	
	return metrics
}

// getPrometheusMetrics fetches metrics from Prometheus
func (sc *SQDCollector) getPrometheusMetrics(url, workerID string) (*PrometheusMetrics, error) {
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
		WorkerID:        workerID,
		QueueSizes:      make(map[string]float64),
		QueriesExecuted: make(map[string]float64),
		PingFailures:    make(map[string]float64),
	}

	for name, mf := range metricFamilies {
		for _, m := range mf.Metric {
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

// getGraphQLMetrics fetches metrics from GraphQL
func (sc *SQDCollector) getGraphQLMetrics(url, workerID string) (*GraphQLMetrics, error) {
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

// Helper function to convert bool to int64
func boolToInt64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// initCharts initializes the collector charts
func (sc *SQDCollector) initCharts() error {
	sc.charts = &Charts{}
	
	// Create charts for each worker
	for _, worker := range sc.Workers {
		workerId := fmt.Sprintf("%s:%d", worker.Name, worker.Port)
		safeId := strings.ReplaceAll(workerId, ":", "_")
		
		// Worker status chart
		statusChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_status", safeId),
			Title:    fmt.Sprintf("Worker %s Status", workerId),
			Units:    "status",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 1,
		}
		statusChart.Dimensions = append(statusChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_online", workerId), Name: "Online", Algorithm: "absolute"},
			&Dimension{ID: fmt.Sprintf("worker_%s_jailed", workerId), Name: "Jailed", Algorithm: "absolute"},
		)
		sc.charts.Add(statusChart)
		
		// Connections chart
		connectionsChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_connections", safeId),
			Title:    fmt.Sprintf("Worker %s Connections", workerId),
			Units:    "connections",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 2,
		}
		connectionsChart.Dimensions = append(connectionsChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_active_connections", workerId), Name: "Active", Algorithm: "absolute"},
		)
		sc.charts.Add(connectionsChart)
		
		// Queries chart
		queriesChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_queries", safeId),
			Title:    fmt.Sprintf("Worker %s Queries", workerId),
			Units:    "queries",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 3,
		}
		queriesChart.Dimensions = append(queriesChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_ongoing_queries", workerId), Name: "Ongoing", Algorithm: "absolute"},
			&Dimension{ID: fmt.Sprintf("worker_%s_running_queries", workerId), Name: "Running", Algorithm: "absolute"},
		)
		sc.charts.Add(queriesChart)
		
		// Storage chart
		storageChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_storage", safeId),
			Title:    fmt.Sprintf("Worker %s Storage", workerId),
			Units:    "bytes",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 4,
		}
		storageChart.Dimensions = append(storageChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_storage_bytes", workerId), Name: "Used", Algorithm: "absolute"},
		)
		sc.charts.Add(storageChart)
		
		// Chunks chart
		chunksChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_chunks", safeId),
			Title:    fmt.Sprintf("Worker %s Chunks", workerId),
			Units:    "chunks",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 5,
		}
		chunksChart.Dimensions = append(chunksChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_chunks_available", workerId), Name: "Available", Algorithm: "absolute"},
			&Dimension{ID: fmt.Sprintf("worker_%s_chunks_downloading", workerId), Name: "Downloading", Algorithm: "absolute"},
			&Dimension{ID: fmt.Sprintf("worker_%s_chunks_pending", workerId), Name: "Pending", Algorithm: "absolute"},
		)
		sc.charts.Add(chunksChart)
		
		// Performance chart
		perfChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_performance", safeId),
			Title:    fmt.Sprintf("Worker %s Performance", workerId),
			Units:    "percentage",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 6,
		}
		perfChart.Dimensions = append(perfChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_uptime_24h", workerId), Name: "Uptime 24h", Multiplier: 1, Divisor: 10},
			&Dimension{ID: fmt.Sprintf("worker_%s_uptime_90d", workerId), Name: "Uptime 90d", Multiplier: 1, Divisor: 10},
		)
		sc.charts.Add(perfChart)
		
		// APR chart
		aprChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_rewards", safeId),
			Title:    fmt.Sprintf("Worker %s Rewards", workerId),
			Units:    "percentage",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 7,
		}
		aprChart.Dimensions = append(aprChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_apr", workerId), Name: "APR", Multiplier: 1, Divisor: 10},
			&Dimension{ID: fmt.Sprintf("worker_%s_staker_apr", workerId), Name: "Staker APR", Multiplier: 1, Divisor: 10},
		)
		sc.charts.Add(aprChart)
		
		// Delegation chart
		delegationChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_delegation", safeId),
			Title:    fmt.Sprintf("Worker %s Delegation", workerId),
			Units:    "count",
			Family:   "sqd",
			Type:     LineChart,
			Priority: 8,
		}
		delegationChart.Dimensions = append(delegationChart.Dimensions, 
			&Dimension{ID: fmt.Sprintf("worker_%s_delegation_count", workerId), Name: "Count", Algorithm: "absolute"},
		)
		sc.charts.Add(delegationChart)
	}
	
	return nil
}
