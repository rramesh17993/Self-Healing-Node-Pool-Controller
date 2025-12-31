package main

import (
	"flag"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/example/self-healing-nodepool/pkg/apis/v1alpha1"
	"github.com/example/self-healing-nodepool/pkg/collector"
	"github.com/example/self-healing-nodepool/pkg/controller"
	"github.com/example/self-healing-nodepool/pkg/decision"
	"github.com/example/self-healing-nodepool/pkg/remediation"
	"github.com/example/self-healing-nodepool/pkg/scorer"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: ":8081",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Dependencies
	promCollector := collector.NewPrometheusCollector("http://prometheus-service:9090")
	defaultScorer := scorer.DefaultScorer()
	decisionEngine := decision.NewEngine()

	// Initialize Clientset for Eviction API
	kubeClient, err := kubernetes.NewForConfig(mgr.GetConfig())
	if err != nil {
		setupLog.Error(err, "unable to create kubeClient")
		os.Exit(1)
	}

	remediator := &remediation.Executor{
		Client:     mgr.GetClient(),
		KubeClient: kubeClient,
	}

	// Default Policy (Hardcoded for MVP)
	// IN PRODUCTION: This should be removed. The controller should watch for NodeHealingPolicy CRs
	// created by the user and apply them dynamically to matching nodes.
	// For this scaffold, we use a single hardcoded policy for simplicity.
	policy := &v1alpha1.NodeHealingPolicy{
		Spec: v1alpha1.NodeHealingPolicySpec{
			Thresholds: v1alpha1.Thresholds{
				UnhealthyScore:   0.6,
				EvaluationWindow: metav1.Duration{Duration: 5 * time.Minute},
			},
			Remediation: v1alpha1.Remediation{
				Cooldown: metav1.Duration{Duration: 30 * time.Minute},
			},
		},
	}

	if err = (&controller.NodeHealthReconciler{
		Client:     mgr.GetClient(),
		Log:        ctrl.Log.WithName("controllers").WithName("NodeHealth"),
		Scheme:     mgr.GetScheme(),
		Collector:  promCollector,
		Scorer:     defaultScorer,
		Decision:   decisionEngine,
		Remediator: remediator,
		Policy:     policy,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeHealth")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
