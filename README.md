# Self-Healing Kubernetes Node Pool Controller

[![CI](https://github.com/rramesh17993/Self-Healing-Node-Pool-Controller/actions/workflows/ci.yaml/badge.svg)](https://github.com/rramesh17993/Self-Healing-Node-Pool-Controller/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rramesh17993/Self-Healing-Node-Pool-Controller)](https://goreportcard.com/report/github.com/rramesh17993/Self-Healing-Node-Pool-Controller)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Built and Maintained by [Rajesh Ramesh](https://github.com/rramesh17993)**

A production-grade Kubernetes Controller designed to close the observability gap between "Node Ready" and "Application Healthy". It proactively detects silent node degradation (e.g., slow disk I/O, network packet drops, zombie processes) and automates the remediation lifecycle.

## Architecture

The controller operates as a closed-loop control system independent of the cloud provider's control plane.

```mermaid
graph TD
    subgraph "Observation Loop"
        Prometheus[Prometheus / Metrics Adapter] -->|Raw Signals| Collector
        Kubelet[Kubelet] -->|Node Conditions| Collector
    end

    subgraph "Control Loop"
        Collector -->|Normalized Metrics| Scorer
        Scorer -->|Health Score (0.0-1.0)| DecisionEngine
        DecisionEngine -->|Action Plan| Executor
    end

    subgraph "Remediation"
        Executor -->|Cordon & Drain| K8sAPI
        Executor -->|Terminate Instance| CloudProvider
    end
```

## Core Features

- **Signal Normalization**: Ingests disparate metrics (latency, error rates, saturation) and computes a unified `HealthScore`.
- **Hysteresis & Stabilization**: Implements evaluation windows and cooldown periods to prevent remediation flapping during transient spikes.
- **Safety First**:
    - **Gradual Drains**: Respects PDBs (Pod Disruption Budgets) with configurable timeouts.
    - **Blast Radius Containment**: Rate limits concurrent remediations to prevent service availability drops.

## Getting Started

### Prerequisites
- Go 1.23+
- Kubernetes 1.25+
- Helm 3.x

### Quick Start

1.  **Clone the Repository**
    ```bash
    git clone https://github.com/rramesh17993/Self-Healing-Node-Pool-Controller.git
    cd Self-Healing-Node-Pool-Controller
    ```

2.  **Deploy via Helm**
    ```bash
    helm install node-healer ./deploy/helm \
      --set policy.unhealthyScore=0.7 \
      --set policy.maxConcurrentDrains=1
    ```

3.  **Verify Operation**
    ```bash
    kubectl get pods -l app.kubernetes.io/name=self-healing-nodepool
    ```

## Configuration

Control the healing sensitivity via `deploy/helm/values.yaml`:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `policy.unhealthyScore` | Threshold (0.0-1.0) to trigger remediation. Lower is more sensitive. | `0.6` |
| `policy.evaluationWindow` | Duration to observe signals before acting. | `5m` |
| `policy.cooldown` | Minimum time between remediations on the same node. | `30m` |

## Contributing

Engineering improvements and PRs are welcome. Please ensure all commits pass the CI pipeline (`make test`).

## License

MIT License. Copyright (c) 2025 Rajesh Ramesh.
