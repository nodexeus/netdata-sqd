package sqd

import (
	"net/http"
	"time"
)

const (
	defaultUpdateEvery = 60
)

// New creates a new SQD collector
func New() *SQD {
	return &SQD{
		Config: Config{
			UpdateEvery: defaultUpdateEvery,
		},
		charts:            make(map[string]*Chart),
		prometheusMetrics: make(map[string]*PrometheusMetrics),
		graphqlMetrics:    make(map[string]*GraphQLMetrics),
		httpClient:        &http.Client{Timeout: 10 * time.Second},
	}
}

// Chart represents a chart definition
type Chart struct {
	ID         string
	Title      string
	Units      string
	Family     string
	Type       string
	Dimensions []*Dimension
	Priority   int
}

// Dimension represents a chart dimension
type Dimension struct {
	ID   string
	Name string
	Algo string
	Mul  int
	Div  int
}

// SQD is the main collector struct
type SQD struct {
	UpdateEvery int
	Workers     []WorkerConfig

	Config            `yaml:",inline" json:",inline"`
	charts            map[string]*Chart
	prometheusMetrics map[string]*PrometheusMetrics
	graphqlMetrics    map[string]*GraphQLMetrics
	httpClient        *http.Client
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

// Cleanup performs cleanup if needed
func (s *SQD) Cleanup() {
	// Nothing to clean up
}

// Check is defined in init.go
