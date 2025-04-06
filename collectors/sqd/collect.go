package sqd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/prometheus/common/expfmt"
)

// Collect gathers metrics from SQD workers
func (s *SQD) Collect() map[string]int64 {
	metrics := make(map[string]int64)
	
	for _, worker := range s.Workers {
		workerID := s.safeID(fmt.Sprintf("%s_%d", worker.Name, worker.Port))
		
		// Get Prometheus metrics
		promMetrics, err := s.getPrometheusMetrics(worker.PrometheusURL, workerID)
		if err != nil {
			fmt.Printf("error collecting Prometheus metrics from %s: %v\n", worker.PrometheusURL, err)
		} else {
			s.prometheusMetrics[workerID] = promMetrics
		}
		
		// Get GraphQL metrics
		gqlMetrics, err := s.getGraphQLMetrics(worker.GraphQLURL, workerID)
		if err != nil {
			fmt.Printf("error collecting GraphQL metrics from %s: %v\n", worker.GraphQLURL, err)
		} else {
			s.graphqlMetrics[workerID] = gqlMetrics
		}
		
		// Map metrics to netdata format
		if promMetrics != nil {
			metrics[fmt.Sprintf("worker_%s_active_connections", workerID)] = int64(promMetrics.ActiveConnections)
			metrics[fmt.Sprintf("worker_%s_ongoing_probes", workerID)] = int64(promMetrics.OngoingProbes)
			metrics[fmt.Sprintf("worker_%s_ongoing_queries", workerID)] = int64(promMetrics.OngoingQueries)
			metrics[fmt.Sprintf("worker_%s_heartbeats_published", workerID)] = int64(promMetrics.HeartbeatsPublished)
			metrics[fmt.Sprintf("worker_%s_heartbeats_received", workerID)] = int64(promMetrics.HeartbeatsReceived)
			metrics[fmt.Sprintf("worker_%s_chunks_available", workerID)] = int64(promMetrics.ChunksAvailable)
			metrics[fmt.Sprintf("worker_%s_chunks_downloading", workerID)] = int64(promMetrics.ChunksDownloading)
			metrics[fmt.Sprintf("worker_%s_chunks_pending", workerID)] = int64(promMetrics.ChunksPending)
			metrics[fmt.Sprintf("worker_%s_chunks_downloaded", workerID)] = int64(promMetrics.ChunksDownloaded)
			metrics[fmt.Sprintf("worker_%s_chunks_failed", workerID)] = int64(promMetrics.ChunksFailed)
			metrics[fmt.Sprintf("worker_%s_chunks_removed", workerID)] = int64(promMetrics.ChunksRemoved)
			metrics[fmt.Sprintf("worker_%s_storage_bytes", workerID)] = int64(promMetrics.UsedStorageBytes)
			metrics[fmt.Sprintf("worker_%s_running_queries", workerID)] = int64(promMetrics.RunningQueries)
		}
		
		if gqlMetrics != nil {
			metrics[fmt.Sprintf("worker_%s_online", workerID)] = boolToInt64(gqlMetrics.Online)
			metrics[fmt.Sprintf("worker_%s_jailed", workerID)] = boolToInt64(gqlMetrics.Jailed)
			metrics[fmt.Sprintf("worker_%s_apr", workerID)] = int64(gqlMetrics.APR * 1000)
			metrics[fmt.Sprintf("worker_%s_staker_apr", workerID)] = int64(gqlMetrics.StakerAPR * 1000)
			metrics[fmt.Sprintf("worker_%s_uptime_24h", workerID)] = int64(gqlMetrics.Uptime24Hours * 1000)
			metrics[fmt.Sprintf("worker_%s_uptime_90d", workerID)] = int64(gqlMetrics.Uptime90Days * 1000)
			metrics[fmt.Sprintf("worker_%s_traffic_weight", workerID)] = int64(gqlMetrics.TrafficWeight * 1000)
			metrics[fmt.Sprintf("worker_%s_delegation_count", workerID)] = int64(gqlMetrics.DelegationCount)
		}
	}
	
	return metrics
}

// getPrometheusMetrics fetches metrics from Prometheus
func (s *SQD) getPrometheusMetrics(url, workerID string) (*PrometheusMetrics, error) {
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
func (s *SQD) getGraphQLMetrics(url, workerID string) (*GraphQLMetrics, error) {
	// Extract the actual worker ID from the combined ID (name_port)
	parts := strings.Split(workerID, "_")
	queryID := parts[len(parts)-1] // Use the last part as the query ID

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
	}`, queryID)

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
		return nil, fmt.Errorf("no worker found with ID %s", queryID)
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
