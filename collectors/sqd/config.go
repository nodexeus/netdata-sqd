package sqd

import (
	"errors"
	"fmt"
)

// Config is the collector configuration
type Config struct {
	UpdateEvery int            `yaml:"update_every" json:"update_every"`
	Workers     []WorkerConfig `yaml:"workers" json:"workers"`
}

// WorkerConfig holds configuration for each SQD worker
type WorkerConfig struct {
	Name          string `yaml:"name" json:"name"`
	PrometheusURL string `yaml:"prometheus_url" json:"prometheus_url"`
	GraphQLURL    string `yaml:"graphql_url" json:"graphql_url"`
}

// Validate validates the collector configuration
func (c Config) Validate() error {
	if len(c.Workers) == 0 {
		return errors.New("no workers configured")
	}

	for i, worker := range c.Workers {
		if worker.Name == "" {
			return fmt.Errorf("worker[%d]: name cannot be empty", i)
		}
		if worker.PrometheusURL == "" {
			return fmt.Errorf("worker[%d]: prometheus_url cannot be empty", i)
		}
		if worker.GraphQLURL == "" {
			return fmt.Errorf("worker[%d]: graphql_url cannot be empty", i)
		}
	}

	return nil
}

// Init initializes the config with default values
func (c *Config) Init() {
	if c.UpdateEvery <= 0 {
		c.UpdateEvery = 1
	}

	// Provide default worker if none configured
	if len(c.Workers) == 0 {
		c.Workers = []WorkerConfig{
			{
				Name:          "default",
				PrometheusURL: "http://localhost:9090",
				GraphQLURL:    "https://subsquid.squids.live/subsquid-network-mainnet/graphql",
			},
		}
	}

	// Set default worker name if empty
	for i := range c.Workers {
		if c.Workers[i].Name == "" {
			c.Workers[i].Name = fmt.Sprintf("worker_%d", i+1)
		}
	}
}
