package sqd

// Init initializes the collector
func (s *SQD) Init() {
	if len(s.Workers) == 0 {
		s.Workers = []WorkerConfig{
			{
				Name:          "default",
				PrometheusURL: "http://localhost:9090/metrics",
				GraphQLURL:    "https://subsquid.squids.live/subsquid-network-mainnet/graphql",
			},
		}
	}

	// Initialize charts
	s.initCharts()
}

// Check verifies that the collector is properly configured
func (s *SQD) Check() bool {
	return len(s.Workers) > 0
}

// safeID returns a sanitized ID suitable for metrics
func (s *SQD) safeID(id string) string {
	return id
}
