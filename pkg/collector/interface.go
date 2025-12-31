package collector

import (
	"context"

	"github.com/example/self-healing-nodepool/pkg/scorer"
)

// NodeSignalCollector defines the interface for gathering health signals from a node.
type NodeSignalCollector interface {
	// CollectSignals fetches the health signals for a specific node.
	CollectSignals(ctx context.Context, nodeName string) (map[scorer.MetricName]float64, error)
}
