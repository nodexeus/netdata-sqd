package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/netdata/sqd_plugin/module"
)

// Config holds the plugin configuration
type Config struct {
	UpdateEvery int                 `json:"update_every"`
	Workers     []module.WorkerConfig `json:"workers"`
}

func loadConfig() (*Config, error) {
	// Default configuration
	config := &Config{
		UpdateEvery: 1,
		Workers: []module.WorkerConfig{
			{
				Name:          "default",
				PrometheusURL: "http://localhost:9090",
				GraphQLURL:    "http://localhost:8080",
				Port:          9090,
			},
		},
	}

	// Try to read from config file if exists
	configPath := ""
	
	// Check for config in various locations
	candidates := []string{
		"./sqd.conf",
		"/etc/netdata/sqd.conf",
		filepath.Join(os.Getenv("HOME"), ".config/netdata/sqd.conf"),
	}
	
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}
	
	if configPath != "" {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			return nil, err
		}
		
		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func main() {
	// Load configuration
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create collector
	collector := module.New()
	
	// Set collector configuration from loaded config
	collector.Workers = config.Workers

	// Initialize the collector
	if !collector.Init() {
		fmt.Fprintf(os.Stderr, "Failed to initialize SQD collector\n")
		os.Exit(1)
	}

	// Check if collector is properly configured
	if !collector.Check() {
		fmt.Fprintf(os.Stderr, "SQD collector check failed\n")
		os.Exit(1)
	}

	// Get charts
	charts := collector.Charts()
	if charts == nil {
		fmt.Fprintf(os.Stderr, "SQD collector returned no charts\n")
		os.Exit(1)
	}

	// Header
	fmt.Println("CHART netdata.plugin_sqd_update_every 'SQD update frequency' 'seconds' 'sqd' 'sqd.update_every' line 1000000000 1")
	fmt.Printf("DIMENSION update_every 'update every' absolute 1 1\n")
	fmt.Printf("BEGIN netdata.plugin_sqd_update_every\n")
	fmt.Printf("SET update_every = %d\n", config.UpdateEvery)
	fmt.Printf("END\n")

	// Define charts
	for _, chart := range charts.All() {
		fmt.Printf("CHART %s '%s' '%s' '%s' '%s' '%s' %d\n", 
			chart.ID, chart.Title, chart.Units, chart.Family, 
			chart.Context, chart.Type, chart.Priority)
		
		for _, dim := range chart.Dimensions {
			algo := dim.Algorithm
			if algo == "" {
				algo = "absolute"
			}
			
			mult := 1
			div := 1
			if dim.Multiplier != 0 {
				mult = dim.Multiplier
			}
			if dim.Divisor != 0 {
				div = dim.Divisor
			}
			
			hidden := ""
			if dim.Hidden {
				hidden = " hidden"
			}
			
			fmt.Printf("DIMENSION %s '%s' '%s' %d %d%s\n", 
				dim.ID, dim.Name, algo, mult, div, hidden)
		}
	}

	// Data collection loop
	ticker := time.NewTicker(time.Duration(config.UpdateEvery) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		metrics := collector.Collect()
		
		// Output metrics for each chart
		for _, chart := range charts.All() {
			fmt.Printf("BEGIN %s\n", chart.ID)
			
			for _, dim := range chart.Dimensions {
				if value, ok := metrics[dim.ID]; ok {
					fmt.Printf("SET %s = %d\n", dim.ID, value)
				} else {
					fmt.Printf("SET %s = 0\n", dim.ID)
				}
			}
			
			fmt.Printf("END\n")
		}
	}
}


