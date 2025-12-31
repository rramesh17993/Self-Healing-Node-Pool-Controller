# System Internals: Self-Healing Node Pool Controller

## 1. Engineering Motivation
Managed Kubernetes services (EKS, GKE, AKS) provide robust handling for hard failures (Node NotReady). However, they often lack visibility into **gray failures**—scenarios where a node is compliant with the Kubernetes API contract but is functionally degraded for workloads. Examples include:
- **EBS Volume Latency**: I/O wait times spiking >500ms due to neighbor noise.
- **Network Interface Saturation**: Packet drop rates exceeding 0.1% without link failure.
- **Kernel Deadlocks**: Specific syscalls hanging while the Kubelet heartbeat remains active.

This controller addresses these gaps by ingesting granular, application-centric signals and executing automated remediation workflows.

## 2. Architecture Specification

### Control Plane Design
The system implements a non-blocking, level-triggered control loop.

#### A. Signal Collection Layer (`pkg/collector`)
- **Responsibility**: Abstraction of upstream telemetry systems.
- **Implementation**: The `PrometheusCollector` queries vector metrics (e.g., `rate(node_disk_io_time_seconds_total[5m])`).
- **Extensibility**: The `NodeSignalCollector` interface allows strictly typed injection of signals from alternate sources (e.g., Datadog, CloudWatch, or eBPF probes).

#### B. Scoring Engine (`pkg/scorer`)
- **Mechanism**: Normalized Weighted Average.
- **Formula**: $\text{Score} = \sum_{i=1}^{n} (\text{Signal}_i \times \text{Weight}_i)$
- **Calibration**: Weights are configured relative to workload sensitivity. High-throughput database pools may weight I/O Wait at 0.5, while compute grids weight CPU Steal higher.

#### C. Decision Matrix (`pkg/decision`)
- **Deterministic Evaluation**: The engine operates purely on the calculated HealthScore and the active `NodeHealingPolicy`.
- **Hysteresis**: To prevent oscillation ("flapping"), the engine enforces:
    - **Remediation Threshold**: `Score > 0.6` (Strict cutoff).
    - **Cooldown Period**: A configurable window (Default: 30m) post-remediation where the node is immune to further action, allowing for self-recovery or cluster stabilization.

#### D. Remediation Execution (`pkg/remediation`)
- **Workflow**:
    1.  **Isolation (Cordon)**: Patch Node `spec.unschedulable=true`. Immediate cessation of new pod scheduling.
    2.  **Evacuation (Drain)**: Iterate through Pods, respecting `PodDisruptionBudgets`.
    3.  **Sanitization**: Check for DaemonSets (ignored) and local storage constraints.
    4.  **Replacement**: (Cloud Provider Hook) Trigger ASG/VMSS instance termination.

## 3. Key Technical Decisions

### 1. Weighted Scoring vs. Binary Choice
**Decision**: Use a float score (0.0-1.0) instead of just "Healthy/Unhealthy".
**Reasoning**: "Gray failures" are rarely binary. A node might be 70% degraded. A scoring system allows us to tune sensitivity without rewriting logic.

### 2. Safety First (Cooldowns & Timeouts)
**Decision**: Implemented `DrainTimeout` (10m) and `RemediationCooldown` (30m).
**Reasoning**:
- **Drain Timeout**: Prevents the controller from getting stuck forever if a Pod refuses to terminate.
- **Cooldown**: Prevents a runaway loop where the controller kills all nodes if a global metric spikes (e.g., a region-wide network issue).

### 3. Stateless Design
**Decision**: The logic is functional. State (like "last evaluated time") is intended to be stored in the CRD `Status`.
**Reasoning**: If the controller crashes and restarts, it should pick up exactly where it left off. (Note: The current MVP relies on in-memory state or re-evaluation, which is a noted area for production hardening).

## 4. Current Limitations (MVP Status)
- **Hardcoded Policy**: The remediation thresholds are currently defined in `main.go`. In a real-world scenario, these would be dynamic CRDs (`NodeHealingPolicy`).
- **Mocked Cloud Provider**: The `ReplaceNode` call just logs a message. Integration with AWS SDK / GCP Client is needed for the final step.

## 5. How to Run
```bash
# 1. Build
make build

# 2. Deploy (requires generic kubectl access)
helm install controller ./deploy/helm

# 3. Verify
kubectl get pods
kubectl logs -l app.kubernetes.io/name=self-healing-nodepool
```
