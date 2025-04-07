package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	// Import the collector package
	"github.com/netdata/sqd_plugin/collectors/sqd"
)

var (
	version = "development"
	commit  = "unknown"
)

// Configuration represents the collector configuration
type Configuration struct {
	UpdateEvery int                `json:"update_every"`
	Workers     []sqd.WorkerConfig `json:"workers"`
}

func main() {
	// Check if we're being called by netdata for discovery
	if len(os.Args) > 1 && os.Args[1] == "discover" {
		// Output plugin metadata for netdata
		fmt.Println("PLUGIN_VERSION=1.0.0")
		fmt.Println("PLUGIN_TYPE=collector")
		fmt.Println("PLUGIN_CAPABILITIES=charts")
		return
	}

	// Check if we're being called by netdata for configuration
	if len(os.Args) > 1 && os.Args[1] == "config" {
		// Output default configuration
		fmt.Println("update_every: 60")
		fmt.Println("workers:")
		fmt.Println("  - name: default")
		fmt.Println("    prometheus_url: http://localhost:9090")
		fmt.Println("    graphql_url: https://subsquid.squids.live/subsquid-network-mainnet/graphql")
		return
	}

	fmt.Printf("SQD Collector %s (commit: %s)\n", version, commit)

	// Try to find config file
	configPaths := []string{
		"./sqd.conf",
		"/etc/netdata/go.d/sqd.conf", // Updated path to match installation
		"/etc/netdata/sqd.conf",
		filepath.Join(os.Getenv("HOME"), ".config/netdata/sqd.conf"),
	}

	// Load configuration
	config := Configuration{
		UpdateEvery: 1,
		Workers: []sqd.WorkerConfig{
			{
				Name:          "default",
				PrometheusURL: "http://localhost:9090",
				GraphQLURL:    "https://subsquid.squids.live/subsquid-network-mainnet/graphql",
			},
		},
	}

	// Try to load configuration from file
	for _, path := range configPaths {
		if data, err := ioutil.ReadFile(path); err == nil {
			fmt.Printf("Loading configuration from %s\n", path)
			if err := json.Unmarshal(data, &config); err != nil {
				log.Fatalf("Error parsing configuration: %v\n", err)
			}
			break
		}
	}

	// Create collector
	collector := sqd.New()
	collector.UpdateEvery = config.UpdateEvery
	collector.Workers = config.Workers

	// Initialize collector
	collector.Init()
	// Perform a sanity check to make sure collection works
	if !collector.Check() {
		log.Fatalf("Failed to initialize collector: Check() failed")
	}

	fmt.Printf("Collector initialized with update interval: %d seconds\n", config.UpdateEvery)
	for _, worker := range config.Workers {
		fmt.Printf("Monitoring worker: %s (Prometheus: %s, GraphQL: %s)\n",
			worker.Name, worker.PrometheusURL, worker.GraphQLURL)
	}

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Collection loop
	ticker := time.NewTicker(time.Duration(config.UpdateEvery) * time.Second)
	defer ticker.Stop()

	fmt.Println("Collector started. Press Ctrl+C to exit.")

	for {
		select {
		case <-ticker.C:
			// Collect metrics
			metrics := collector.Collect()

			// First, print chart definitions if needed (first run)
			static := make(map[string]bool)

			// Print metrics in netdata format
			for key, value := range metrics {
				// Only print chart definition once
				if !static[key] {
					fmt.Printf("CHART sqd.%s '' 'SQD Worker Metric: %s' '%s' 'sqd' 'subsquid' 'line' 1 %d\n",
						key, key, determineUnit(key), config.UpdateEvery)
					fmt.Printf("DIMENSION value '' absolute 1 1\n")
					static[key] = true
				}

				// Print the actual data
				fmt.Printf("BEGIN sqd.%s\n", key)

				// Use appropriate precision for different types of metrics
				if strings.Contains(key, "apr") || strings.Contains(key, "uptime") || strings.Contains(key, "traffic_weight") {
					// Use more precision for percentage values
					fmt.Printf("SET value = %.6f\n", value)
				} else {
					// Use regular precision for other values
					fmt.Printf("SET value = %f\n", value)
				}

				fmt.Printf("END\n")
			}

		case sig := <-sigCh:
			fmt.Printf("Received signal: %v, shutting down\n", sig)
			return
		}
	}
}

// determineUnit returns the appropriate unit for a given metric
func determineUnit(metricName string) string {
	if strings.Contains(metricName, "bytes") {
		return "bytes"
	} else if strings.Contains(metricName, "apr") {
		return "percentage"
	} else if strings.Contains(metricName, "uptime") {
		return "percentage"
	} else if strings.Contains(metricName, "traffic_weight") {
		return "percentage"
	} else if strings.Contains(metricName, "online") || strings.Contains(metricName, "jailed") {
		return "boolean"
	} else if strings.Contains(metricName, "count") {
		return "count"
	} else if strings.Contains(metricName, "connections") {
		return "connections"
	} else if strings.Contains(metricName, "queries") {
		return "queries"
	} else if strings.Contains(metricName, "chunks") {
		return "chunks"
	} else {
		return "value"
	}
}
