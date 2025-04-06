package module

// ChartType represents the type of chart
type ChartType string

// Chart types
const (
	LineChart    ChartType = "line"
	AreaChart    ChartType = "area"
	StackedChart ChartType = "stacked"
)

// Charts is a collection of Chart objects
type Charts struct {
	charts []*Chart
}

// Chart represents a netdata chart
type Chart struct {
	ID       string
	Title    string
	Units    string
	Family   string
	Context  string
	Type     ChartType
	Priority int

	Dimensions []*Dimension
}

// Dimension represents a netdata chart dimension
type Dimension struct {
	ID        string
	Name      string
	Algorithm string
	Multiplier int
	Divisor    int
	Hidden    bool
}

// Add adds a chart to the collection
func (c *Charts) Add(chart *Chart) {
	c.charts = append(c.charts, chart)
}

// Get returns a chart by ID
func (c *Charts) Get(id string) *Chart {
	for _, chart := range c.charts {
		if chart.ID == id {
			return chart
		}
	}
	return nil
}

// Remove removes a chart by ID
func (c *Charts) Remove(id string) {
	for i, chart := range c.charts {
		if chart.ID == id {
			c.charts = append(c.charts[:i], c.charts[i+1:]...)
			return
		}
	}
}

// All returns all charts
func (c *Charts) All() []*Chart {
	return c.charts
}
