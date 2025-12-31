package decision

import (
	"fmt"
	"time"

	"github.com/example/self-healing-nodepool/pkg/apis/v1alpha1"
)

type ActionType string

const (
	ActionNone      ActionType = "None"
	ActionMonitor   ActionType = "Monitor"
	ActionRemediate ActionType = "Remediate"
)

type Decision struct {
	Action ActionType
	Reason string
}

// Engine is responsible for making remediation decisions based on health scores and policies.
// It is a stateless component that takes inputs (score, policy, history) and returns a Decision.
type Engine struct{}

// NewEngine creates a new decision engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Evaluate determines the next action based on the score and policy.
func (e *Engine) Evaluate(score float64, policy *v1alpha1.NodeHealingPolicy, lastRemediationTime time.Time) Decision {
	threshold := policy.Spec.Thresholds.UnhealthyScore

	if score < threshold {
		return Decision{Action: ActionNone, Reason: "Node is healthy"}
	}

	// Score >= threshold. Check cooldown.
	cooldown := policy.Spec.Remediation.Cooldown.Duration
	if time.Since(lastRemediationTime) < cooldown {
		return Decision{
			Action: ActionMonitor,
			Reason: "Node is unhealthy but within cooldown period",
		}
	}

	return Decision{
		Action: ActionRemediate,
		Reason: fmt.Sprintf("Health score %.2f exceeds threshold %.2f", score, threshold),
	}
}
