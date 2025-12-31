package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NodeHealingPolicySpec defines the desired state of NodeHealingPolicy
type NodeHealingPolicySpec struct {
	// NodeSelector selects which nodes are covered by this policy.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Thresholds defines the criteria for determining node health.
	Thresholds Thresholds `json:"thresholds,omitempty"`

	// Remediation defines the actions to take when a node is unhealthy.
	Remediation Remediation `json:"remediation,omitempty"`

	// Limits defines safety guardrails for remediation.
	Limits Limits `json:"limits,omitempty"`
}

type Thresholds struct {
	// UnhealthyScore is the health score (0.0 - 1.0) above which a node is considered unhealthy.
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	UnhealthyScore float64 `json:"unhealthyScore"`

	// EvaluationWindow is the duration for which the score must persist before action.
	// +kubebuilder:default="5m"
	EvaluationWindow metav1.Duration `json:"evaluationWindow,omitempty"`
}

type Remediation struct {
	// DrainTimeout is the maximum duration to wait for a node to drain.
	// +kubebuilder:default="10m"
	DrainTimeout metav1.Duration `json:"drainTimeout,omitempty"`

	// Cooldown is the minimum time between remediations on the same node/pool.
	// +kubebuilder:default="30m"
	Cooldown metav1.Duration `json:"cooldown,omitempty"`
}

type Limits struct {
	// MaxConcurrentDrains is the maximum number of nodes that can be draining simultaneously.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	MaxConcurrentDrains int `json:"maxConcurrentDrains,omitempty"`
}

// NodeHealingPolicyStatus defines the observed state of NodeHealingPolicy
type NodeHealingPolicyStatus struct {
	// ActiveRemediations tracks currently ongoing remediation actions.
	// +optional
	ActiveRemediations []string `json:"activeRemediations,omitempty"`

	// LastEvaluated is the timestamp of the last health check.
	// +optional
	LastEvaluated *metav1.Time `json:"lastEvaluated,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NodeHealingPolicy is the Schema for the nodehealingpolicies API
type NodeHealingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeHealingPolicySpec   `json:"spec,omitempty"`
	Status NodeHealingPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeHealingPolicyList contains a list of NodeHealingPolicy
type NodeHealingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeHealingPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeHealingPolicy{}, &NodeHealingPolicyList{})
}
