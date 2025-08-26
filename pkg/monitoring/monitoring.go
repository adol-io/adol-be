package monitoring

import (
	"runtime"
	"sync"
	"time"

	"github.com/nicklaros/adol/pkg/logger"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Name         string                 `json:"name"`
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	ResponseTime time.Duration          `json:"response_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// HealthChecker manages health checks
type HealthChecker struct {
	checks map[string]func() HealthCheck
	logger logger.EnhancedLogger
	mutex  sync.RWMutex
}

// MetricType represents different types of metrics
type MetricType string

const (
	CounterMetric MetricType = "counter"
	GaugeMetric   MetricType = "gauge"
	TimerMetric   MetricType = "timer"
)

// Metric represents a single metric
type Metric struct {
	Name      string                 `json:"name"`
	Type      MetricType             `json:"type"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Counter represents a counter metric
type Counter struct {
	value  int64
	labels map[string]string
	mutex  sync.RWMutex
}

// Gauge represents a gauge metric
type Gauge struct {
	value  float64
	labels map[string]string
	mutex  sync.RWMutex
}

// Timer represents a timer metric
type Timer struct {
	durations []time.Duration
	sum       time.Duration
	count     int64
	labels    map[string]string
	mutex     sync.RWMutex
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	counters map[string]*Counter
	gauges   map[string]*Gauge
	timers   map[string]*Timer
	logger   logger.EnhancedLogger
	mutex    sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger logger.EnhancedLogger) *MetricsCollector {
	return &MetricsCollector{
		counters: make(map[string]*Counter),
		gauges:   make(map[string]*Gauge),
		timers:   make(map[string]*Timer),
		logger:   logger,
	}
}

// Counter methods
func (mc *MetricsCollector) Counter(name string, labels map[string]string) *Counter {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.buildKey(name, labels)
	if counter, exists := mc.counters[key]; exists {
		return counter
	}

	counter := &Counter{labels: labels}
	mc.counters[key] = counter
	return counter
}

func (c *Counter) Inc() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value++
}

func (c *Counter) Value() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

// Gauge methods
func (mc *MetricsCollector) Gauge(name string, labels map[string]string) *Gauge {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.buildKey(name, labels)
	if gauge, exists := mc.gauges[key]; exists {
		return gauge
	}

	gauge := &Gauge{labels: labels}
	mc.gauges[key] = gauge
	return gauge
}

func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value = value
}

func (g *Gauge) Value() float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.value
}

// Timer methods
func (mc *MetricsCollector) Timer(name string, labels map[string]string) *Timer {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.buildKey(name, labels)
	if timer, exists := mc.timers[key]; exists {
		return timer
	}

	timer := &Timer{
		durations: make([]time.Duration, 0),
		labels:    labels,
	}
	mc.timers[key] = timer
	return timer
}

func (t *Timer) Record(duration time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.durations = append(t.durations, duration)
	t.sum += duration
	t.count++
}

func (t *Timer) Average() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()
	if t.count == 0 {
		return 0
	}
	return time.Duration(int64(t.sum) / t.count)
}

// Metrics collection
func (mc *MetricsCollector) GetAllMetrics() []Metric {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	var metrics []Metric
	now := time.Now()

	for name, counter := range mc.counters {
		metrics = append(metrics, Metric{
			Name:      name,
			Type:      CounterMetric,
			Value:     float64(counter.Value()),
			Labels:    counter.labels,
			Timestamp: now,
		})
	}

	for name, gauge := range mc.gauges {
		metrics = append(metrics, Metric{
			Name:      name,
			Type:      GaugeMetric,
			Value:     gauge.Value(),
			Labels:    gauge.labels,
			Timestamp: now,
		})
	}

	for name, timer := range mc.timers {
		metrics = append(metrics, Metric{
			Name:      name,
			Type:      TimerMetric,
			Value:     float64(timer.Average().Milliseconds()),
			Labels:    timer.labels,
			Timestamp: now,
		})
	}

	return metrics
}

func (mc *MetricsCollector) buildKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += ":" + k + "=" + v
	}
	return key
}

func (mc *MetricsCollector) RecordSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mc.Gauge("system_memory_alloc_bytes", nil).Set(float64(m.Alloc))
	mc.Gauge("system_goroutines", nil).Set(float64(runtime.NumGoroutine()))
}

// Health checker methods
func NewHealthChecker(logger logger.EnhancedLogger) *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]func() HealthCheck),
		logger: logger,
	}
}

func (hc *HealthChecker) RegisterCheck(name string, checkFunc func() HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = checkFunc
}

func (hc *HealthChecker) RunChecks() map[string]HealthCheck {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	results := make(map[string]HealthCheck)
	for name, checkFunc := range hc.checks {
		start := time.Now()
		result := checkFunc()
		result.ResponseTime = time.Since(start)
		result.Timestamp = time.Now()
		results[name] = result

		hc.logger.LogHealthCheck(name, result.Status == HealthStatusHealthy, result.ResponseTime, result.Message)
	}

	return results
}

func (hc *HealthChecker) GetOverallHealth() HealthStatus {
	checks := hc.RunChecks()

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}