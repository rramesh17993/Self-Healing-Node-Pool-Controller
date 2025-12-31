package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/example/self-healing-nodepool/pkg/apis/v1alpha1"
	"github.com/example/self-healing-nodepool/pkg/collector"
	"github.com/example/self-healing-nodepool/pkg/decision"
	"github.com/example/self-healing-nodepool/pkg/remediation"
	"github.com/example/self-healing-nodepool/pkg/scorer"
)

// NodeHealthReconciler reconciles a Node object
type NodeHealthReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	Collector  collector.NodeSignalCollector
	Scorer     *scorer.Scorer
	Decision   *decision.Engine
	Remediator *remediation.Executor
	Policy     *v1alpha1.NodeHealingPolicy
}

// Reconcile is the main loop.
// For this MVP, we assume the Reconciler is triggered by Node updates or a ticker.
func (r *NodeHealthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("node", req.NamespacedName)

	// 1. Fetch Node
	var node corev1.Node
	if err := r.Get(ctx, req.NamespacedName, &node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Fetch Policy (Simplified: Using a singleton policy passed in struct, or we could fetch CRD)
	// In production, we'd look up the NodeHealingPolicy matching the node labels.
	policy := r.Policy
	// Skip if no policy applies (logic omitted for brevity).

	// 3. Collect Signals
	signals, err := r.Collector.CollectSignals(ctx, node.Name)
	if err != nil {
		log.Error(err, "failed to collect signals")
		return ctrl.Result{}, err // Retry
	}

	// 4. Score
	score := r.Scorer.CalculateScore(signals)
	log.Info("node health scored", "score", score)

	// 5. Decide
	// TODO: Get last remediation time from NodeHealingPolicy Status or Node annotation
	// CURRENT LIMITATION: The LastEvaluated time is not currently persisted to the NodeHealingPolicy Status.
	// This means that if the controller restarts, it might lose track of the last remediation time.
	// Future work: Update NodeHealingPolicy.Status with the last decision time.
	lastRemediation := time.Time{} // Placeholder
	dec := r.Decision.Evaluate(score, policy, lastRemediation)

	// 6. Execute
	switch dec.Action {
	case decision.ActionRemediate:
		log.Info("Remediating node", "reason", dec.Reason)
		if err := r.Remediator.CordonNode(ctx, node.Name); err != nil {
			return ctrl.Result{}, err
		}
		// Async drain? Or sync? simpler to do sync for now or launch go routine (but dangerous in reconciler)
		// Better: set state to Draining, return, and let next reconcile loop handle drain progress.
		// For MVP, simplistic blocking call:
		if err := r.Remediator.DrainNode(ctx, node.Name); err != nil {
			log.Error(err, "failed to drain node")
			return ctrl.Result{}, err
		}
		// Trigger Replacement (omitted)
	case decision.ActionMonitor:
		log.Info("Monitoring node", "reason", dec.Reason)
	}

	// Requeue to ensure continuous monitoring even if no events
	return ctrl.Result{RequeueAfter: policy.Spec.Thresholds.EvaluationWindow.Duration}, nil
}

func (r *NodeHealthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
