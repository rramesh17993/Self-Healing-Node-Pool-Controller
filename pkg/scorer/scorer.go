package scorer

// MetricName represents the name of a health signal.
type MetricName string

const (
	MetricDiskIOWait     MetricName = "disk_io_wait"
	MetricNetworkDrops   MetricName = "network_drops"
	MetricKubeletErrors  MetricName = "kubelet_errors"
	MetricMemoryPressure MetricName = "memory_pressure"
	MetricConditionFlaps MetricName = "condition_flaps"
)

// Scorer calculates the health score of a node based on signals and weights.
type Scorer struct {
	Weights map[MetricName]float64
}

// NewScorer creates a new Scorer with the provided weights.
func NewScorer(weights map[MetricName]float64) *Scorer {
	s := &Scorer{
		Weights: weights,
	}
	s.normalizeWeights()
	return s
}

// DefaultScorer returns a scorer with default standard weights.
func DefaultScorer() *Scorer {
	return NewScorer(map[MetricName]float64{
		MetricDiskIOWait:     0.30,
		MetricNetworkDrops:   0.20,
		MetricKubeletErrors:  0.20,
		MetricMemoryPressure: 0.15,
		MetricConditionFlaps: 0.15,
	})
}

// normalizeWeights ensures the weights sum to 1.0.
func (s *Scorer) normalizeWeights() {
	var total float64
	for _, w := range s.Weights {
		total += w
	}
	if total == 0 {
		return
	}
	for k, w := range s.Weights {
		s.Weights[k] = w / total
	}
}

// CalculateScore computes the weighted health score.
// Returns a score between 0.0 (healthy) and 1.0 (unhealthy).
func (s *Scorer) CalculateScore(signals map[MetricName]float64) float64 {
	var totalScore float64

	for metric, weight := range s.Weights {
		if val, ok := signals[metric]; ok {
			// Clamp signal value between 0 and 1
			if val > 1.0 {
				val = 1.0
			}
			if val < 0.0 {
				val = 0.0
			}
			totalScore += val * weight
		}
	}

	return totalScore
}
