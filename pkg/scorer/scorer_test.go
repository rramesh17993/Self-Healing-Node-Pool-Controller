package scorer

import (
	"testing"
)

func TestScorer_CalculateScore(t *testing.T) {
	tests := []struct {
		name    string
		weights map[MetricName]float64
		signals map[MetricName]float64
		want    float64
	}{
		{
			name: "All signals present, equal weights",
			weights: map[MetricName]float64{
				"signal1": 0.5,
				"signal2": 0.5,
			},
			signals: map[MetricName]float64{
				"signal1": 1.0,
				"signal2": 0.0,
			},
			want: 0.5,
		},
		{
			name: "Signal missing",
			weights: map[MetricName]float64{
				"signal1": 0.5,
				"signal2": 0.5,
			},
			signals: map[MetricName]float64{
				"signal1": 0.8,
			},
			want: 0.4,
		},
		{
			name: "Weights normalization",
			weights: map[MetricName]float64{
				"signal1": 1.0,
				"signal2": 1.0,
			},
			signals: map[MetricName]float64{
				"signal1": 1.0,
				"signal2": 1.0,
			},
			want: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScorer(tt.weights)
			got := s.CalculateScore(tt.signals)
			if got != tt.want {
				t.Errorf("Scorer.CalculateScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
