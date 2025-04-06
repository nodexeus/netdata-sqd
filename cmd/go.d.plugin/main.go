package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	UpdateEvery int                 `json:"update_every"`
	Workers     []sqd.WorkerConfig `json:"workers"`
}

func main() {
	fmt.Printf("SQD Collector %s (commit: %s)\n", version, commit)

	// Try to find config file
	configPaths := []string{
		"./sqd.conf",
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
				GraphQLURL:    "http://localhost:8080",
				Port:          9090,
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
			
			// Print metrics in netdata format
			for key, value := range metrics {
				fmt.Printf("BEGIN sqd.%s\n", key)
				fmt.Printf("SET value = %d\n", value)
				fmt.Printf("END\n")
			}
			
		case sig := <-sigCh:
			fmt.Printf("Received signal: %v, shutting down\n", sig)
			return
		}
	}
}
