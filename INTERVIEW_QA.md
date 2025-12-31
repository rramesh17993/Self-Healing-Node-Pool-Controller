# Engineering Interview Guide: Self-Healing Controller

## Architectural Design Patterns

### Q1: Justification for Custom Operator vs. Standard Autoscaling?
**Technical Response**:
"Standard autoscalers (Cluster Autoscaler/Karpenter) rely primarily on Resource requests or Heartbeat failures. They are blind to **application-layer degradation**. Our controller implements a **Custom Metric Controller** pattern. By ingesting application-specific signals (e.g., 'Database Connection Latency' or 'Disk IOPS Saturation'), we can preemptively drain a node that is technically 'Ready' but functionally useless, maintaining a higher SLO for stateful workloads."

### Q2: Control Loop Hysteresis & Stabilization?
**Technical Response**:
"To prevent **flapping** (rapid oscillation between Healthy/Unhealthy states), we implemented a dual-layer stabilization mechanism:
1.  **Sliding Window Evaluation**: Metrics are aggregated over a 5-minute window to smooth out transient spikes.
2.  **State Machine Cooldowns**: The remediation state machine enforces a strict cooldown (default 30m) post-eviction. This acts as a circuit breaker, preventing a cascading failure scenario where the controller inadvertently depletes the entire pool due to a cluster-wide metric anomaly."

### Q3: Failure Modes & Resiliency?
**Technical Response**:
"The system is designed to **Fail Open**. If the metric pipeline fails (e.g., Prometheus unreachable), the Collector returns a 'Zero-Value' or Error. The Decision Engine ignores incomplete data rather than defaulting to 'Unhealthy'. This ensures that observability outages never cause infrastructure outages. Additionally, we enforce a `MaxConcurrentDrains` semaphore to strictly limit blast radius."

## Kubernetes Internals & Implementation

### Q4: Deep Dive: The Reconcile Loop?
**Technical Response**:
"I implemented a **Level-Triggered** reconciliation loop using `controller-runtime`.
1.  **Watch**: The Informer cache triggers the loop on Node updates.
2.  **Sync**: We fetch the live state (Metrics) and compare it to the desired state (Policy thresholds).
3.  **Idempotency**: The logic is stateless. If the controller restarts, it re-evaluates the world based on current metrics. State persistence (last transition time) is managed via the Node's ObjectMeta or CRD Status, ensuring resilience across controller restarts."

### Q5: Safe Eviction Strategy?
**Technical Response**:
"We interface directly with the **Eviction API** (`policy/v1`), which is safer than raw Pod deletion because it respects **Pod Disruption Budgets (PDBs)**.
The `DrainNode` routine acts as a sophisticated orchestrator:
1.  It filters DaemonSets (which are node-bound).
2.  It creates Eviction sub-resources for each pod.
3.  It wraps the operation in a `Context` with a deadline to prevent the controller from hanging on 'stuck' pods (e.g., those caught in finalizers)."

### Q6: Security Posture?
**Technical Response**:
"We adhere to the principle of **Least Privilege**.
1.  **Distroless Base Image**: The container image is built From `gcr.io/distroless/static`. It lacks a shell, shell utilities, or package managers, effectively neutralizing a vast class of RCE exploits.
2.  **RBAC Scoping**: The ClusterRole is strictly scoped to `patch nodes` and `create evictions`. It cannot read Secrets or ConfigMaps unrelated to its function."

## Behavior & Scenario Questions

### Q7: You found a bug in the random score generator. How did you fix it?
**Answer**:
"During the audit, I noticed the mock collector was returning `rand.Float64()`. This made tests non-deterministic and could lead to accidental node deletions during development. I replaced it with a deterministic 'Safe' value (0.0) and added explicit warning logs, ensuring that 'Self-Healing' only triggers when we explicitly inject faults."

### Q8: How would you scale this for a cluster with 5000 nodes?
**Answer**:
"Right now, it's a single controller processes. For scale:
1.  **Sharding**: We could run multiple replicas of the controller, each handling a subset of nodes (using consistent hashing or label selectors).
2.  **Caching**: The controller-runtime client uses an Informer cache, so we heavily rely on local cache reads rather than hitting the API server for every `Get`."

## Coding Questions (Go specific)

### Q9: Why use interfaces for `Provider` and `Collector`?
**Answer**:
"To make the code testable and modular. The `Provider` interface allows me to swap out the real AWS implementation for a Mock implementation during unit tests (like I did in `executor_test.go`). It also means adding Azure support later is just adding a new struct that satisfies the interface, without changing the core loop."
