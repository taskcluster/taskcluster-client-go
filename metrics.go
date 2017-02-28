package taskcluster

import (
	"sort"
	"sync"
	"time"
)

// MetricsCollector is an interface to be implemented by a metrics collector.
type MetricsCollector interface {
	// Increment named counter by given value
	Count(name string, value float64)
	// Measure values for named metric (for percentiles/histogram metrics)
	Measure(name string, values ...float64)
}

// MetricsLogger is a trivial implementation of the MetricsCollector interface.
// Which will log collected metrics to Logger every 30 seconds.
type MetricsLogger struct {
	Logger
	m        sync.Mutex
	timer    *time.Timer
	counters map[string]float64
	measures map[string][]float64
}

// Count increments counter for name by given value.
func (m *MetricsLogger) Count(name string, value float64) {
	m.m.Lock()
	defer m.m.Unlock()

	// Create counters and increment
	if m.counters == nil {
		m.counters = make(map[string]float64)
	}
	m.counters[name] += value

	// Start timer if not already started
	if m.timer == nil {
		m.timer = time.AfterFunc(30*time.Second, m.Flush)
	}
}

// Measure aggregates value for named measure.
func (m *MetricsLogger) Measure(name string, value ...float64) {
	m.m.Lock()
	defer m.m.Unlock()

	// Create measures and add values
	if m.measures == nil {
		m.measures = make(map[string][]float64)
	}
	m.measures[name] = append(m.measures[name], value...)

	// Start timer if not already started
	if m.timer == nil {
		m.timer = time.AfterFunc(30*time.Second, m.Flush)
	}
}

// Flush force prints metrics to the given logger
func (m *MetricsLogger) Flush() {
	m.m.Lock()
	// Take the metrics
	counters := m.counters
	measures := m.measures
	m.counters = nil
	m.measures = nil

	// Clear the timer
	if m.timer != nil {
		m.timer.Stop()
		m.timer = nil
	}
	m.m.Unlock()

	// Print counters
	for name, value := range counters {
		m.Logger.Println("Counter: ", name, " = ", value)
	}

	// Print measures
	for name, values := range measures {
		sort.Float64s(values)
		m.Logger.Println("Measure: ", name, " median=", values[len(values)/2])
	}
}
