package collector

import (
	"context"

	"github.com/example/self-healing-nodepool/pkg/scorer"
)

type PrometheusCollector struct {
	// In a real impl, this would hold the PromAPI client
}

func NewPrometheusCollector(url string) *PrometheusCollector {
	return &PrometheusCollector{}
}

func (c *PrometheusCollector) CollectSignals(ctx context.Context, nodeName string) (map[scorer.MetricName]float64, error) {
	// MOCK IMPLEMENTATION
	// In reality, query Prometheus for node_disk_io_time_seconds_total etc.
	// CRITICAL: Returning 0.0 to avoid accidental self-healing triggers during development/demo.
	return map[scorer.MetricName]float64{
		scorer.MetricDiskIOWait:    0.0,
		scorer.MetricNetworkDrops:  0.0,
		scorer.MetricKubeletErrors: 0.0,
	}, nil
}
