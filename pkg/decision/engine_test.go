package decision

import (
	"reflect"
	"testing"
	"time"

	"github.com/example/self-healing-nodepool/pkg/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEngine_Evaluate(t *testing.T) {
	defaultPolicy := &v1alpha1.NodeHealingPolicy{
		Spec: v1alpha1.NodeHealingPolicySpec{
			Thresholds: v1alpha1.Thresholds{
				UnhealthyScore: 0.8,
			},
			Remediation: v1alpha1.Remediation{
				Cooldown: metav1.Duration{Duration: 30 * time.Minute},
			},
		},
	}

	type args struct {
		score               float64
		policy              *v1alpha1.NodeHealingPolicy
		lastRemediationTime time.Time
	}
	tests := []struct {
		name string
		args args
		want Decision
	}{
		{
			name: "Healthy Node",
			args: args{
				score:               0.5,
				policy:              defaultPolicy,
				lastRemediationTime: time.Time{}, // Never
			},
			want: Decision{Action: ActionNone, Reason: "Node is healthy"},
		},
		{
			name: "Unhealthy Node - No Cooldown",
			args: args{
				score:               0.85,
				policy:              defaultPolicy,
				lastRemediationTime: time.Now().Add(-1 * time.Hour), // Long ago
			},
			want: Decision{Action: ActionRemediate, Reason: "Health score 0.85 exceeds threshold 0.80"},
		},
		{
			name: "Unhealthy Node - Within Cooldown",
			args: args{
				score:               0.9,
				policy:              defaultPolicy,
				lastRemediationTime: time.Now().Add(-10 * time.Minute), // Recently
			},
			want: Decision{Action: ActionMonitor, Reason: "Node is unhealthy but within cooldown period"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{}
			if got := e.Evaluate(tt.args.score, tt.args.policy, tt.args.lastRemediationTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Engine.Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
