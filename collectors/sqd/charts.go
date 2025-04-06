package sqd

import (
	"fmt"
	"strings"
)

type chartType string

// Chart types
const (
	chartTypeArea   chartType = "area"
	chartTypeLine   chartType = "line"
	chartTypeStacked chartType = "stacked"
)

// Chart contexts
const (
	chartContextWorkerStatus     = "sqd.worker_status"
	chartContextWorkerConnections = "sqd.worker_connections"
	chartContextWorkerQueries     = "sqd.worker_queries"
	chartContextWorkerStorage     = "sqd.worker_storage"
	chartContextWorkerChunks      = "sqd.worker_chunks"
	chartContextWorkerPerformance = "sqd.worker_performance"
	chartContextWorkerRewards     = "sqd.worker_rewards"
	chartContextWorkerDelegation  = "sqd.worker_delegation"
)

// Chart families
const (
	chartFamilySQD = "sqd"
)

// Chart and Dimension types are defined in sqd.go

// initCharts initializes charts for the SQD collector
func (s *SQD) initCharts() {
	// Create charts for each worker
	for _, worker := range s.Workers {
		workerID := sanitizeID(worker.Name)
		
		// Worker status chart
		statusChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_status", workerID),
			Title:    fmt.Sprintf("Worker %s Status", worker.Name),
			Units:    "status",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 1,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_online", workerID), Name: "Online", Algo: "absolute"},
				{ID: fmt.Sprintf("worker_%s_jailed", workerID), Name: "Jailed", Algo: "absolute"},
			},
		}
		s.charts[statusChart.ID] = statusChart
		
		// Connections chart
		connectionsChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_connections", workerID),
			Title:    fmt.Sprintf("Worker %s Connections", worker.Name),
			Units:    "connections",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 2,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_active_connections", workerID), Name: "Active", Algo: "absolute"},
			},
		}
		s.charts[connectionsChart.ID] = connectionsChart
		
		// Queries chart
		queriesChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_queries", workerID),
			Title:    fmt.Sprintf("Worker %s Queries", worker.Name),
			Units:    "queries",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 3,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_ongoing_queries", workerID), Name: "Ongoing", Algo: "absolute"},
				{ID: fmt.Sprintf("worker_%s_running_queries", workerID), Name: "Running", Algo: "absolute"},
			},
		}
		s.charts[queriesChart.ID] = queriesChart
		
		// Storage chart
		storageChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_storage", workerID),
			Title:    fmt.Sprintf("Worker %s Storage", worker.Name),
			Units:    "bytes",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 4,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_storage_bytes", workerID), Name: "Used", Algo: "absolute"},
			},
		}
		s.charts[storageChart.ID] = storageChart
		
		// Chunks chart
		chunksChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_chunks", workerID),
			Title:    fmt.Sprintf("Worker %s Chunks", worker.Name),
			Units:    "chunks",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 5,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_chunks_available", workerID), Name: "Available", Algo: "absolute"},
				{ID: fmt.Sprintf("worker_%s_chunks_downloading", workerID), Name: "Downloading", Algo: "absolute"},
				{ID: fmt.Sprintf("worker_%s_chunks_pending", workerID), Name: "Pending", Algo: "absolute"},
			},
		}
		s.charts[chunksChart.ID] = chunksChart
		
		// Performance chart
		perfChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_performance", workerID),
			Title:    fmt.Sprintf("Worker %s Performance", worker.Name),
			Units:    "percentage",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 6,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_uptime_24h", workerID), Name: "Uptime 24h", Algo: "absolute", Mul: 1, Div: 10},
				{ID: fmt.Sprintf("worker_%s_uptime_90d", workerID), Name: "Uptime 90d", Algo: "absolute", Mul: 1, Div: 10},
			},
		}
		s.charts[perfChart.ID] = perfChart
		
		// APR chart
		aprChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_rewards", workerID),
			Title:    fmt.Sprintf("Worker %s Rewards", worker.Name),
			Units:    "percentage",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 7,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_apr", workerID), Name: "APR", Algo: "absolute", Mul: 1, Div: 10},
				{ID: fmt.Sprintf("worker_%s_staker_apr", workerID), Name: "Staker APR", Algo: "absolute", Mul: 1, Div: 10},
			},
		}
		s.charts[aprChart.ID] = aprChart
		
		// Delegation chart
		delegationChart := &Chart{
			ID:       fmt.Sprintf("worker_%s_delegation", workerID),
			Title:    fmt.Sprintf("Worker %s Delegation", worker.Name),
			Units:    "count",
			Family:   chartFamilySQD,
			Type:     string(chartTypeLine),
			Priority: 8,
			Dimensions: []*Dimension{
				{ID: fmt.Sprintf("worker_%s_delegation_count", workerID), Name: "Count", Algo: "absolute"},
			},
		}
		s.charts[delegationChart.ID] = delegationChart
	}
}

// Helper function to create safe chart IDs
func sanitizeID(id string) string {
	// Replace unsafe characters with underscores
	return strings.ReplaceAll(strings.ReplaceAll(id, ":", "_"), ".", "_")
}
