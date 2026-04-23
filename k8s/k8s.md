# Top 20 Kubernetes Interview Questions for a Senior DevOps Engineer (5+ Years)

---

## 🏗️ CATEGORY 1: Kubernetes Architecture & Internals

---

### 1. Explain Kubernetes Architecture in Depth — Control Plane to Data Plane

**Expected Answer:**

Kubernetes follows a **master-worker (control plane / data plane)** architecture designed around a declarative, API-driven model. The **control plane** is the brain of the cluster — it makes all scheduling decisions, stores cluster state, and runs controller loops that continuously reconcile the actual state with the desired state. It should never run user workloads. The **data plane** (worker nodes) is where your actual application containers run. Each worker node runs a set of agents that receive instructions from the control plane and manage the pod lifecycle locally.

The communication between these two planes flows exclusively through the **kube-apiserver** — no component talks directly to another. This single point of coordination is what makes Kubernetes extensible and auditable. The control plane is typically deployed across multiple nodes (minimum 3) for high availability, while the data plane can be scaled horizontally by adding more worker nodes.

```
┌─────────────────────────────────────────────────────────────────────┐
│                        CONTROL PLANE                                │
│                                                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────────┐   │
│  │  API Server │  │  Scheduler  │  │  Controller Manager       │   │
│  │  (kube-     │  │  (kube-     │  │  (kube-controller-manager)│   │
│  │  apiserver) │  │  scheduler) │  │                           │   │
│  └──────┬──────┘  └──────┬──────┘  └──────────┬───────────────┘   │
│         │                │                     │                   │
│  ┌──────▼─────────────────▼─────────────────────▼───────────────┐  │
│  │                    etcd (Cluster State Store)                 │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │             Cloud Controller Manager (Optional)             │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
                              │ REST API
┌─────────────────────────────▼───────────────────────────────────────┐
│                          DATA PLANE (Worker Nodes)                  │
│                                                                     │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │                        Worker Node                             │ │
│  │                                                                │ │
│  │  ┌──────────┐  ┌────────────┐  ┌──────────────────────────┐  │ │
│  │  │  kubelet │  │ kube-proxy │  │  Container Runtime       │  │ │
│  │  │          │  │            │  │  (containerd / CRI-O)    │  │ │
│  │  └──────────┘  └────────────┘  └──────────────────────────┘  │ │
│  │                                                                │ │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐           │ │
│  │  │    Pod      │  │    Pod      │  │    Pod      │           │ │
│  │  │ ┌─────────┐ │  │ ┌─────────┐ │  │ ┌─────────┐ │           │ │
│  │  │ │Container│ │  │ │Container│ │  │ │Container│ │           │ │
│  │  │ └─────────┘ │  │ └─────────┘ │  │ └─────────┘ │           │ │
│  │  └─────────────┘  └─────────────┘  └─────────────┘           │ │
│  └────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

**Control Plane Components:**

The control plane consists of several components that collectively manage the entire cluster. Each component has a single responsibility and communicates via the API server. In production, these are distributed across multiple master nodes for fault tolerance.

| Component | Role | Key Responsibility |
|-----------|------|-------------------|
| **kube-apiserver** | API Gateway | Validates/processes all REST requests, only component that writes to etcd |
| **etcd** | Distributed KV Store | Stores entire cluster state, uses Raft consensus |
| **kube-scheduler** | Pod Scheduler | Assigns pods to nodes based on resources, affinity, taints |
| **kube-controller-manager** | Control Loop Runner | Runs Node, Replication, Endpoint, ServiceAccount controllers |
| **cloud-controller-manager** | Cloud Integration | Manages LBs, volumes, node lifecycle via cloud APIs |

**Data Plane Components:**

The data plane consists of agents running on every worker node. These agents are responsible for receiving pod specifications from the control plane, pulling container images, starting/stopping containers, and reporting health status back. If the control plane goes down temporarily, existing workloads on the data plane continue to run — they just can't be modified or rescheduled.

| Component | Role | Key Responsibility |
|-----------|------|-------------------|
| **kubelet** | Node Agent | Ensures containers in pods are running and healthy |
| **kube-proxy** | Network Proxy | Maintains iptables/ipvs rules for service routing |
| **Container Runtime** | Container Execution | Runs containers via CRI (containerd, CRI-O) |

```bash
# Check control plane component health
kubectl get componentstatuses

# Check node status
kubectl get nodes -o wide

# View API server configuration
kubectl -n kube-system describe pod kube-apiserver-<node>
```

---

### 2. Explain the Kubernetes API Request Lifecycle — From kubectl to Pod Running

**Expected Answer:**

When you run `kubectl apply -f pod.yaml`, the request doesn't just "create a pod" — it goes through a carefully ordered pipeline of **authentication**, **authorization**, **admission control**, **persistence**, **scheduling**, and **execution**. Understanding this lifecycle is critical because every security policy, resource quota, mutation webhook, and scheduling constraint is enforced at a specific stage. If any stage rejects the request, the pod is never created.

The lifecycle follows a **sequential gate model** — each stage must pass before the next one runs. This is also where custom logic (via webhooks) can be injected to enforce organizational policies.

```
kubectl apply -f pod.yaml
       │
       ▼
┌─────────────────┐
│  Authentication │ ← Certificates, Bearer Tokens, OIDC
└────────┬────────┘
         │
┌────────▼────────┐
│  Authorization  │ ← RBAC (Role, ClusterRole, Bindings)
└────────┬────────┘
         │
┌────────▼──────────────┐
│  Admission Controllers│ ← Mutating → Validating Webhooks
│  (MutatingWebhook,    │   PodSecurityAdmission, LimitRanger
│   ValidatingWebhook)  │   ResourceQuota, etc.
└────────┬──────────────┘
         │
┌────────▼────────┐
│  Write to etcd  │ ← Pod object persisted
└────────┬────────┘
         │
┌────────▼────────────┐
│  kube-scheduler     │ ← Watches unscheduled pods
│  Filtering Phase    │   Node selectors, taints, resources
│  Scoring Phase      │   Best fit node selected
└────────┬────────────┘
         │
┌────────▼────────────┐
│  kubelet (on node)  │ ← Watches scheduled pods
│  Pulls image        │
│  Creates containers │
│  Reports status     │
└────────┬────────────┘
         │
┌────────▼────────────┐
│  Container Runtime  │ ← containerd/CRI-O creates container
│  (via CRI)          │
└─────────────────────┘
```

**Admission Controllers Deep Dive:**

Admission controllers are plugins that intercept API requests **after** authentication and authorization but **before** the object is persisted to etcd. They come in two types: **Mutating** admission controllers can modify the incoming object (e.g., injecting sidecar containers, adding default resource limits), while **Validating** admission controllers can only accept or reject the request. Mutating controllers always run first, then validating controllers run on the final mutated object. This is a powerful extensibility point — organizations use custom webhooks here to enforce naming conventions, inject security contexts, or block non-compliant workloads.

```bash
# View enabled admission controllers
kube-apiserver --help | grep admission-plugins

# Common critical admission controllers:
# - NamespaceLifecycle     → Prevents ops in terminating namespaces
# - LimitRanger            → Enforces resource limits
# - ServiceAccount         → Auto-assigns service accounts
# - ResourceQuota          → Enforces namespace quotas
# - PodSecurityAdmission   → Enforces Pod Security Standards
# - MutatingAdmissionWebhook  → Custom mutation logic
# - ValidatingAdmissionWebhook → Custom validation logic
```

---

### 3. How Does etcd Work in Kubernetes and What Happens If It Goes Down?

**Expected Answer:**

**etcd** is a distributed, strongly consistent key-value store that serves as the **single source of truth** for the entire Kubernetes cluster. Every piece of cluster state — pods, services, secrets, configmaps, RBAC policies, custom resources — is stored as a key-value pair in etcd. Only the kube-apiserver reads from and writes to etcd; no other component communicates with it directly.

etcd uses the **Raft consensus algorithm** to maintain consistency across its cluster members. Raft requires a **quorum** (a majority of nodes) to agree before any write is committed. This means an etcd cluster with `n` nodes can tolerate `(n-1)/2` node failures and still function. This is why etcd clusters always have an **odd number** of members — 3 nodes tolerate 1 failure, 5 nodes tolerate 2 failures. Running more than 7 nodes is not recommended because the write latency increases as Raft must replicate to more members.

Because etcd holds the entire cluster state, **backing it up regularly is non-negotiable** in production. A corrupted or lost etcd means the entire cluster configuration is gone — even though running containers continue executing, you lose the ability to manage them.

```bash
# etcd uses Raft consensus — needs (n/2)+1 nodes for quorum
# 3-node etcd = can tolerate 1 failure
# 5-node etcd = can tolerate 2 failures

# Check etcd cluster health
ETCDCTL_API=3 etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key \
  endpoint health --cluster

# Check etcd cluster members
ETCDCTL_API=3 etcdctl member list

# Check etcd database size
ETCDCTL_API=3 etcdctl endpoint status --write-out=table

# Defragment etcd (reduce disk usage)
ETCDCTL_API=3 etcdctl defrag --endpoints=https://127.0.0.1:2379

# Backup etcd snapshot
ETCDCTL_API=3 etcdctl snapshot save /backup/etcd-snapshot-$(date +%Y%m%d).db

# Verify backup
ETCDCTL_API=3 etcdctl snapshot status /backup/etcd-snapshot.db --write-out=table

# Restore etcd from snapshot
ETCDCTL_API=3 etcdctl snapshot restore /backup/etcd-snapshot.db \
  --name=master \
  --initial-cluster=master=https://127.0.0.1:2380 \
  --initial-advertise-peer-urls=https://127.0.0.1:2380 \
  --data-dir=/var/lib/etcd-restored
```

**What Happens When etcd Goes Down:**

The impact of an etcd outage is nuanced — the data plane continues operating independently, but the control plane is completely paralyzed. This is because kubelet and kube-proxy cache their last-known configuration locally and continue enforcing it. However, any operation that requires reading from or writing to cluster state (scheduling, scaling, config changes, API calls) will fail immediately.

```
✅ Running pods/containers → Continue running (kubelet manages locally)
✅ kube-proxy rules       → Continue working (iptables persists)
❌ New pod scheduling     → BLOCKED
❌ ConfigMap/Secret reads → BLOCKED for new mounts
❌ kubectl commands       → ALL FAIL
❌ Service discovery      → New endpoints BLOCKED
❌ HPA scaling            → BLOCKED
❌ Controller loops       → ALL STOP
```

---

## 🔄 CATEGORY 2: Workloads & Scheduling

---

### 4. Explain Pod Scheduling — Affinity, Anti-Affinity, Taints, Tolerations, and Priority

**Expected Answer:**

The **kube-scheduler** is responsible for deciding which node a newly created pod should run on. It follows a two-phase process: **Filtering** (eliminates nodes that don't meet hard constraints like resource availability, node selectors, or taints) and **Scoring** (ranks remaining nodes by soft preferences like affinity weights, resource balance, and topology spread). The highest-scoring node wins.

Kubernetes provides several mechanisms to influence scheduling:

- **Node Affinity**: Tells the scheduler to prefer or require specific nodes based on node labels (e.g., "only run on SSD nodes" or "prefer us-east-1a zone"). It replaces the older `nodeSelector` with richer expression syntax.
- **Pod Affinity/Anti-Affinity**: Controls co-location relative to *other pods* rather than nodes. Pod affinity says "place me near pods with label X" (useful for latency-sensitive communication), while pod anti-affinity says "spread me away from pods with label X" (useful for high availability — ensuring replicas land on different nodes/zones).
- **Taints and Tolerations**: Work in the opposite direction from affinity. A **taint** on a node says "reject all pods unless they explicitly tolerate me." A **toleration** on a pod says "I can run on tainted nodes." This is commonly used to dedicate nodes for special workloads (GPU, high-memory, or system-only nodes).
- **Priority Classes**: Define preemption behavior — higher-priority pods can evict lower-priority pods when resources are scarce. This ensures critical workloads always get scheduled.
- **Topology Spread Constraints**: Provide fine-grained control over how pods are distributed across failure domains (zones, nodes, racks), ensuring even distribution for resilience.

**Node Affinity:**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app-pod
spec:
  affinity:
    nodeAffinity:
      # Hard requirement — pod will NOT schedule if not met
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
          - matchExpressions:
              - key: topology.kubernetes.io/zone
                operator: In
                values:
                  - us-east-1a
                  - us-east-1b
              - key: node-type
                operator: In
                values:
                  - high-memory

      # Soft preference — scheduler tries to honor, but not required
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 80
          preference:
            matchExpressions:
              - key: disk-type
                operator: In
                values:
                  - ssd
        - weight: 20
          preference:
            matchExpressions:
              - key: instance-type
                operator: In
                values:
                  - m5.xlarge

    # Pod Anti-Affinity — spread pods across nodes/zones
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        - labelSelector:
            matchExpressions:
              - key: app
                operator: In
                values:
                  - myapp
          topologyKey: kubernetes.io/hostname  # One pod per node

    # Pod Affinity — co-locate with cache pods
    podAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchLabels:
                app: redis-cache
            topologyKey: kubernetes.io/hostname

  containers:
    - name: app
      image: myapp:latest
```

**Taints and Tolerations:**

Taints and tolerations work as a **node-centric** admission control for scheduling. Unlike affinity (which is pod-centric — the pod says where it wants to go), taints are applied to nodes to repel pods. Only pods with a matching toleration can schedule on a tainted node. This creates a "dedicated node" pattern commonly used for GPU workloads, system components, or tenant isolation.

There are three taint effects:
- **NoSchedule**: New pods without a toleration are blocked, but existing pods remain.
- **PreferNoSchedule**: Soft version — the scheduler avoids the node but doesn't guarantee it.
- **NoExecute**: Most aggressive — existing pods without a toleration are **evicted**, and new pods are blocked. The `tolerationSeconds` field controls how long an existing pod can stay before eviction.

```bash
# Taint a node (dedicated GPU node)
kubectl taint nodes gpu-node-1 \
  dedicated=gpu:NoSchedule

# Taint effects:
# NoSchedule         → New pods without toleration won't schedule
# PreferNoSchedule   → Scheduler avoids node but doesn't guarantee
# NoExecute          → Existing pods evicted + new pods blocked
```

```yaml
# Pod must tolerate the taint to schedule on GPU node
spec:
  tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "gpu"
      effect: "NoSchedule"

    # Tolerate node not-ready for 5 minutes before eviction
    - key: "node.kubernetes.io/not-ready"
      operator: "Exists"
      effect: "NoExecute"
      tolerationSeconds: 300
```

**Priority Classes:**

Priority classes let you assign a **numeric priority value** to pods. When the cluster runs out of resources, the scheduler can **preempt** (evict) lower-priority pods to make room for higher-priority ones. This is critical for ensuring that essential workloads (monitoring, security agents, core services) always run, even during resource contention. The `preemptionPolicy` field controls whether a pod can trigger preemption — setting it to `Never` means the pod waits in the queue without evicting others.

```yaml
# Define priority classes
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: critical-priority
value: 1000000
globalDefault: false
preemptionPolicy: PreemptLowerPriority
description: "Critical system components"

---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: high-priority
value: 100
preemptionPolicy: PreemptLowerPriority

---
# Use in pod
spec:
  priorityClassName: critical-priority
```

**Topology Spread Constraints:**

Topology spread constraints provide granular control over pod distribution across topology domains (zones, nodes, racks, or any custom topology key). Unlike pod anti-affinity which is binary (same node or not), topology spread uses `maxSkew` to define the maximum allowed imbalance between domains. For example, `maxSkew: 1` means no zone should have more than 1 extra pod compared to the zone with the fewest. `whenUnsatisfiable: DoNotSchedule` makes it a hard constraint; `ScheduleAnyway` makes it a soft preference. Multiple constraints can be combined for multi-level spreading (e.g., spread across zones AND across nodes within each zone).

```yaml
spec:
  topologySpreadConstraints:
    - maxSkew: 1                              # Max difference between zones
      topologyKey: topology.kubernetes.io/zone
      whenUnsatisfiable: DoNotSchedule       # Hard (or ScheduleAnyway for soft)
      labelSelector:
        matchLabels:
          app: myapp
    - maxSkew: 1
      topologyKey: kubernetes.io/hostname    # Also spread across nodes
      whenUnsatisfiable: ScheduleAnyway
      labelSelector:
        matchLabels:
          app: myapp
```

---

### 5. Explain Deployment Strategies in Kubernetes — RollingUpdate, Blue-Green, Canary

**Expected Answer:**

Deployment strategies determine how new versions of your application replace old ones. The right choice depends on your tolerance for **downtime**, **risk**, and **infrastructure cost**:

- **Rolling Update** (built-in, default): Gradually replaces old pods with new ones. You control the pace with `maxSurge` (how many extra pods can exist during rollout) and `maxUnavailable` (how many pods can be down simultaneously). This is the simplest approach and works well for stateless services. Rollback is automatic if the readiness probe fails.
- **Blue-Green**: Runs two identical environments — "blue" (current) and "green" (new). Traffic switches instantly by changing the Service selector. This provides **zero-downtime** and **instant rollback** but requires **double the infrastructure** during the transition.
- **Canary**: Sends a small percentage of traffic to the new version while the majority stays on the old version. If the canary performs well (measured by error rates, latency, etc.), traffic is gradually shifted. This is the safest strategy for critical services but requires traffic management tooling (Argo Rollouts, Istio, or NGINX). It catches issues that unit tests and staging environments miss — production-specific failures.

**Rolling Update (Default):**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 10
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 3          # Max pods ABOVE desired count during update
      maxUnavailable: 1    # Max pods BELOW desired count during update
  minReadySeconds: 30      # Wait before marking pod as available
  progressDeadlineSeconds: 600  # Fail if not complete in 10 min
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: app
          image: myapp:v2.0
          readinessProbe:         # Critical for zero-downtime
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 3
```

```bash
# Monitor rollout
kubectl rollout status deployment/myapp

# Pause rollout (canary-like manual approach)
kubectl rollout pause deployment/myapp

# Resume rollout
kubectl rollout resume deployment/myapp

# Rollback to previous version
kubectl rollout undo deployment/myapp

# Rollback to specific revision
kubectl rollout undo deployment/myapp --to-revision=3

# Check rollout history
kubectl rollout history deployment/myapp
kubectl rollout history deployment/myapp --revision=3
```

**Blue-Green Deployment:**

In a blue-green deployment, you maintain two complete environments running simultaneously. The "blue" environment serves all production traffic while "green" is deployed and validated with the new version. Once validated, a single Service selector change switches all traffic instantly to "green." If anything goes wrong, switching the selector back to "blue" provides an immediate rollback. The trade-off is cost — you need double the resources during the transition window. After confirming the green deployment is stable, the old blue deployment is torn down (or kept as a rollback safety net).

```yaml
# Blue (current production)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-blue
spec:
  replicas: 10
  selector:
    matchLabels:
      app: myapp
      version: blue
  template:
    metadata:
      labels:
        app: myapp
        version: blue
    spec:
      containers:
        - name: app
          image: myapp:v1.0
---
# Green (new version)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-green
spec:
  replicas: 10
  selector:
    matchLabels:
      app: myapp
      version: green
  template:
    metadata:
      labels:
        app: myapp
        version: green
    spec:
      containers:
        - name: app
          image: myapp:v2.0
---
# Service — switch traffic by changing selector
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
spec:
  selector:
    app: myapp
    version: blue   # ← Change to 'green' to switch traffic instantly
  ports:
    - port: 80
      targetPort: 8080
```

```bash
# Switch traffic to green
kubectl patch service myapp-service \
  -p '{"spec":{"selector":{"version":"green"}}}'

# Instant rollback if issues
kubectl patch service myapp-service \
  -p '{"spec":{"selector":{"version":"blue"}}}'
```

**Canary Deployment with Argo Rollouts:**

Argo Rollouts extends Kubernetes with a `Rollout` resource that provides advanced deployment strategies natively. For canary deployments, it progressively shifts traffic percentages (e.g., 10% → 30% → 60% → 100%) with configurable pauses and automated analysis between each step. The **analysis** step can run Prometheus queries, check error rates, or call custom webhooks to determine if the canary is healthy. If analysis fails, the rollout automatically aborts and rolls back — no human intervention required. This makes canary deployments safe even for high-traffic production services.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: myapp-rollout
spec:
  replicas: 10
  strategy:
    canary:
      canaryService: myapp-canary-svc
      stableService: myapp-stable-svc
      trafficRouting:
        nginx:
          stableIngress: myapp-ingress
      steps:
        - setWeight: 10        # 10% canary traffic
        - pause:
            duration: 5m       # Wait 5 minutes
        - analysis:            # Run automated analysis
            templates:
              - templateName: success-rate
        - setWeight: 30        # 30% canary traffic
        - pause:
            duration: 10m
        - setWeight: 60
        - pause:
            duration: 10m
        - setWeight: 100       # Full rollout
      analysis:
        successCondition: "result[0] >= 0.95"
        failureLimit: 3
```

---

### 6. How Do Kubernetes Controllers Work? Explain the Reconciliation Loop

**Expected Answer:**

Controllers are the **core automation engine** of Kubernetes. They implement the **declarative model** — you tell Kubernetes *what* you want (desired state), and controllers continuously work to *make it happen* (reconcile actual state to match desired state).

Every controller follows the same pattern called the **reconciliation loop** (also called the "control loop"):

1. **Watch**: The controller uses an **Informer** to receive events from the API server whenever a relevant resource changes (created, updated, deleted). The Informer also maintains a local **cache** to avoid hammering the API server with redundant reads.
2. **Queue**: Events are placed into a **work queue** that handles deduplication, rate-limiting, and retry logic. This ensures the controller processes changes in order and doesn't overwhelm the system.
3. **Reconcile**: The controller's reconcile function compares the **desired state** (from the spec in etcd) with the **actual state** (from the cluster). If they differ, it takes corrective action — creating pods, deleting resources, updating configurations, etc.
4. **Update Status**: After reconciliation, the controller updates the resource's **status subresource** to reflect the current actual state (e.g., `readyReplicas: 3`).

Kubernetes ships with many built-in controllers (Deployment, ReplicaSet, StatefulSet, Node, Job, CronJob, etc.), all running inside the `kube-controller-manager`. You can also build **custom controllers** (operators) to manage your own application-specific resources using frameworks like **Kubebuilder** or **Operator SDK**.

```
┌─────────────────────────────────────────────────────────┐
│              Controller Reconciliation Loop              │
│                                                         │
│   ┌─────────┐    Watch      ┌──────────────────────┐   │
│   │  etcd   │◄──────────────│   Informer/Cache     │   │
│   │(desired │               └──────────┬───────────┘   │
│   │  state) │                          │ Event         │
│   └─────────┘               ┌──────────▼───────────┐   │
│                              │    Work Queue        │   │
│                              └──────────┬───────────┘   │
│                                         │               │
│                              ┌──────────▼───────────┐   │
│                              │  Reconcile Function  │   │
│                              │                      │   │
│                              │  current ≠ desired ? │   │
│                              │  → Take action       │   │
│                              │  → Update status     │   │
│                              └──────────────────────┘   │
│                                                         │
│            Built-in Controllers:                        │
│            - Deployment Controller                      │
│            - ReplicaSet Controller                      │
│            - StatefulSet Controller                     │
│            - Node Controller                            │
│            - Job Controller                             │
│            - CronJob Controller                         │
└─────────────────────────────────────────────────────────┘
```

**Custom Controller with Kubebuilder:**

Kubebuilder is a framework for building Kubernetes operators in Go. It scaffolds the boilerplate — CRD generation, RBAC manifests, controller wiring — and lets you focus on writing the reconcile function. The reconcile function receives a request (containing the namespace/name of the changed object), fetches the current state, compares it to desired state, takes corrective action, and updates the status. The `ctrl.Result` return value controls re-queue behavior — returning `RequeueAfter: 5m` means the controller will check again in 5 minutes even without new events, ensuring eventual consistency.

```go
// Reconcile function — called whenever object changes
func (r *MyAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // 1. Fetch the current state of the object
    myApp := &appsv1.MyApp{}
    if err := r.Get(ctx, req.NamespacedName, myApp); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil // Object deleted
        }
        return ctrl.Result{}, err
    }

    // 2. Get current state of managed resources
    deployment := &appsv1.Deployment{}
    err := r.Get(ctx, types.NamespacedName{
        Name:      myApp.Name,
        Namespace: myApp.Namespace,
    }, deployment)

    // 3. Reconcile — bring current state to desired state
    if errors.IsNotFound(err) {
        // Create deployment
        dep := r.deploymentForMyApp(myApp)
        if err = r.Create(ctx, dep); err != nil {
            return ctrl.Result{}, err
        }
    } else if err != nil {
        return ctrl.Result{}, err
    } else {
        // Update if needed
        if *deployment.Spec.Replicas != myApp.Spec.Replicas {
            deployment.Spec.Replicas = &myApp.Spec.Replicas
            if err = r.Update(ctx, deployment); err != nil {
                return ctrl.Result{}, err
            }
        }
    }

    // 4. Update status subresource
    myApp.Status.ReadyReplicas = deployment.Status.ReadyReplicas
    if err := r.Status().Update(ctx, myApp); err != nil {
        return ctrl.Result{}, err
    }

    // Requeue after 5 minutes for periodic reconciliation
    return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}
```

---

## 🌐 CATEGORY 3: Networking

---

### 7. Explain Kubernetes Networking — CNI, Services, Ingress, and Network Policies

**Expected Answer:**

Kubernetes networking is built on a **flat network model** with strict guarantees. Unlike traditional infrastructure where containers share a host IP and use port mapping, Kubernetes gives every pod its own unique IP address. This simplifies application design — services can communicate using pod IPs directly without worrying about port conflicts or NAT translation.

The actual implementation of this network model is delegated to **CNI (Container Network Interface)** plugins. CNI is a specification that defines how network interfaces are set up for containers. Different plugins implement this differently — some use overlay networks (VXLAN tunnels encapsulating packets), some use BGP routing (advertising pod IPs directly), and some use eBPF (bypassing the kernel networking stack for better performance). The choice of CNI plugin significantly impacts performance, scalability, and available features like network policies and encryption.

**Kubernetes Services** solve the problem of pod ephemerality — pods are created and destroyed constantly, so their IPs are unreliable. A Service provides a **stable virtual IP (ClusterIP)** that routes traffic to a dynamic set of backend pods matched by label selectors. kube-proxy programs the node's networking stack (iptables, IPVS, or eBPF) to forward traffic from the Service IP to healthy pod endpoints.

**Ingress** provides HTTP/HTTPS routing from outside the cluster to internal Services. It acts as a reverse proxy with support for path-based routing, virtual hosts, TLS termination, and rate limiting. An Ingress controller (NGINX, Traefik, HAProxy, or cloud-specific) must be deployed to process Ingress resources.

**Network Policies** implement a **zero-trust network model** by defining firewall rules at the pod level. By default, all pods can communicate with all other pods. Once a NetworkPolicy selects a pod, all traffic not explicitly allowed is denied. This is essential for compliance and security — ensuring that only frontend pods can talk to backend pods, or that production namespaces are isolated from staging.

**Kubernetes Networking Rules:**
```
1. Every Pod gets its own IP address
2. Pods can communicate with all other pods without NAT
3. Nodes can communicate with all pods without NAT
4. Pod IP is the same from inside and outside the pod
```

**CNI Plugins Comparison:**

The CNI plugin you choose defines how pods get their IPs, how traffic is routed between nodes, whether network policies are supported, and whether encryption is available. Here's how the most popular CNI plugins compare:

| CNI Plugin | Network Model | Network Policy | Encryption | Best For |
|-----------|--------------|----------------|------------|---------|
| **Calico** | BGP / VXLAN | ✅ Advanced | WireGuard | Enterprise, large scale |
| **Flannel** | VXLAN | ❌ Basic | ❌ | Simple clusters |
| **Cilium** | eBPF | ✅ L7 | WireGuard | High performance, observability |
| **Weave** | Mesh / VXLAN | ✅ | ✅ | Multi-cloud |
| **AWS VPC CNI** | Native VPC IPs | Via Calico | VPC | EKS |

**Service Types:**

Kubernetes provides four service types, each exposing pods at a different scope:

- **ClusterIP** (default): Accessible only within the cluster. Used for internal service-to-service communication.
- **NodePort**: Extends ClusterIP by opening a static port (30000-32767) on every node's IP. External traffic can reach the service via `<NodeIP>:<NodePort>`. Simple but not production-grade.
- **LoadBalancer**: Extends NodePort by provisioning an external cloud load balancer (AWS ALB/NLB, GCP LB, etc.) that forwards traffic to the NodePorts. This is the standard way to expose services to the internet in cloud environments.
- **ExternalName**: Creates a DNS CNAME record pointing to an external service. No proxying occurs — it's purely DNS-level redirection, useful for referencing external databases or APIs.
- **Headless Service** (`clusterIP: None`): Returns individual pod IPs instead of a virtual IP. Used with StatefulSets where clients need to connect to specific pods (e.g., Kafka brokers, database replicas).

```yaml
# ClusterIP — Internal only (default)
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  type: ClusterIP
  selector:
    app: backend
  ports:
    - port: 80
      targetPort: 8080

---
# NodePort — External via node IP:port (30000-32767)
spec:
  type: NodePort
  ports:
    - port: 80
      targetPort: 8080
      nodePort: 31000   # Optional — auto-assigned if omitted

---
# LoadBalancer — Cloud provider LB
spec:
  type: LoadBalancer
  loadBalancerIP: 10.0.0.1          # Request specific IP (if supported)
  externalTrafficPolicy: Local       # Preserve client IP, avoid extra hop
  ports:
    - port: 443
      targetPort: 8443

---
# ExternalName — DNS CNAME alias
spec:
  type: ExternalName
  externalName: mydb.prod.rds.amazonaws.com

---
# Headless Service — Direct pod DNS (for StatefulSets)
spec:
  clusterIP: None
  selector:
    app: kafka
```

**Network Policy (Zero-Trust):**

Network policies act as pod-level firewalls, implementing a **zero-trust security model**. By default, Kubernetes allows unrestricted pod-to-pod communication. Once a NetworkPolicy targets a pod (via `podSelector`), it switches to **default deny** for the specified `policyTypes` — only explicitly allowed ingress/egress traffic is permitted.

Key concepts:
- **podSelector**: Selects which pods the policy applies to.
- **ingress rules**: Define what sources can send traffic TO the selected pods.
- **egress rules**: Define what destinations the selected pods can send traffic TO.
- Each rule can filter by pod labels (`podSelector`), namespace labels (`namespaceSelector`), IP blocks (`ipBlock`), and ports.
- Always include DNS egress (port 53 UDP/TCP) in egress rules, otherwise pods can't resolve service names.
- Network policies require a CNI plugin that supports them (Calico, Cilium, Weave). Flannel does **not** support network policies.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-policy
  namespace: production
spec:
  podSelector:
    matchLabels:
      app: backend

  policyTypes:
    - Ingress
    - Egress

  ingress:
    # Allow ONLY from frontend pods in same namespace
    - from:
        - podSelector:
            matchLabels:
              app: frontend
        - namespaceSelector:
            matchLabels:
              environment: production
      ports:
        - protocol: TCP
          port: 8080

  egress:
    # Allow to database
    - to:
        - podSelector:
            matchLabels:
              app: postgres
      ports:
        - protocol: TCP
          port: 5432
    # Allow DNS resolution
    - to: []
      ports:
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 53
```

**Ingress with TLS:**

An Ingress resource defines HTTP/HTTPS routing rules that map external hostnames and URL paths to internal Services. Unlike LoadBalancer services (which operate at L4/TCP), Ingress operates at **L7/HTTP** and supports host-based routing (multiple domains on one IP), path-based routing, TLS termination, and middleware like rate limiting and URL rewriting.

An **Ingress controller** must be deployed to process these resources — the Ingress resource itself is just a configuration object. Popular controllers include NGINX Ingress Controller, Traefik, and cloud-native options (AWS ALB Ingress Controller). **cert-manager** can be integrated to automatically provision and renew TLS certificates from Let's Encrypt.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  tls:
    - hosts:
        - api.myapp.com
      secretName: api-tls-cert
  rules:
    - host: api.myapp.com
      http:
        paths:
          - path: /v1
            pathType: Prefix
            backend:
              service:
                name: api-v1-service
                port:
                  number: 80
          - path: /v2
            pathType: Prefix
            backend:
              service:
                name: api-v2-service
                port:
                  number: 80
```

---

### 8. How Does kube-proxy Work? iptables vs IPVS vs eBPF?

**Expected Answer:**

**kube-proxy** is a network component that runs on every node and implements the **Service abstraction**. When you create a Kubernetes Service, kube-proxy programs the node's networking stack to forward traffic destined for the Service's virtual ClusterIP to one of the backend pods. It watches the API server for Service and Endpoint changes and updates the forwarding rules accordingly.

kube-proxy supports three modes, each with different performance characteristics:

- **iptables mode** (legacy default): Creates iptables NAT rules for each Service/endpoint pair. Simple and reliable, but iptables rules are evaluated **linearly** (O(n)), meaning performance degrades significantly with thousands of services. Rule updates are also atomic — the entire chain is replaced on every change.
- **IPVS mode** (recommended for scale): Uses the Linux kernel's IP Virtual Server framework, which is purpose-built for load balancing. It uses hash tables for O(1) lookups, supports multiple load-balancing algorithms (round-robin, least connections, weighted), and handles 10,000+ services efficiently.
- **eBPF mode** (via Cilium — kube-proxy replacement): Cilium can replace kube-proxy entirely by programming eBPF programs directly in the kernel. This bypasses iptables and conntrack entirely, providing the lowest latency and highest throughput. It also enables L7-aware load balancing and rich observability. Requires kernel 4.8+.

```bash
# Check kube-proxy mode
kubectl -n kube-system get configmap kube-proxy -o yaml | grep mode

# Check IPVS rules
ipvsadm -Ln

# Check iptables rules created by kube-proxy
iptables -t nat -L KUBE-SERVICES -n --line-numbers
```

**Comparison:**

| Feature | iptables | IPVS | eBPF (Cilium) |
|---------|---------|------|--------------|
| **Lookup** | Linear O(n) | Hash O(1) | Hash O(1) |
| **Scale** | Degrades at 1000+ services | Handles 10,000+ | Handles 100,000+ |
| **Load Balancing** | Round-robin only | RR, LC, WRR, SH, DH | Full L4/L7 |
| **Connection Tracking** | conntrack | conntrack | Bypasses conntrack |
| **Latency** | Higher at scale | Low | Lowest |
| **Kernel Version** | Any | 4.1+ | 4.8+ |

```yaml
# Configure IPVS mode
apiVersion: kubeproxy.config.k8s.io/v1alpha1
kind: KubeProxyConfiguration
mode: "ipvs"
ipvs:
  scheduler: "lc"          # Least connection
  syncPeriod: "30s"
  minSyncPeriod: "2s"
  strictARP: true          # Required for MetalLB
iptables:
  masqueradeAll: false
  masqueradeBit: 14
  minSyncPeriod: "0s"
  syncPeriod: "30s"
```

---

## 💾 CATEGORY 4: Storage

---

### 9. Explain PersistentVolumes, PVCs, StorageClasses, and CSI Drivers

**Expected Answer:**

Kubernetes separates **storage provisioning** from **storage consumption** using a three-layer abstraction:

- **PersistentVolume (PV)**: A cluster-level resource representing a piece of physical storage (an EBS volume, NFS share, or local disk). It defines capacity, access modes, and the underlying storage driver. PVs exist independently of any pod — they persist across pod restarts and rescheduling.
- **PersistentVolumeClaim (PVC)**: A namespace-scoped request for storage by a user/pod. A PVC specifies the desired size, access mode, and StorageClass. Kubernetes matches the PVC to an available PV (or dynamically provisions one) and **binds** them together. Pods reference PVCs in their volume spec.
- **StorageClass**: Defines a "class" of storage with specific provisioner, parameters, and reclaim policy. It enables **dynamic provisioning** — when a PVC references a StorageClass, Kubernetes automatically creates the underlying PV and cloud volume without manual intervention.

**Access Modes** define how many nodes/pods can mount the volume simultaneously:
- **RWO (ReadWriteOnce)**: Single node can mount read-write. Used by most block storage (EBS, Azure Disk).
- **RWX (ReadWriteMany)**: Multiple nodes can mount read-write simultaneously. Requires shared filesystems (NFS, EFS, CephFS).
- **ROX (ReadOnlyMany)**: Multiple nodes can mount read-only. Useful for shared config/data.
- **RWOP (ReadWriteOncePod)**: Single pod only (K8s 1.22+). Strictest access for sensitive data.

**CSI (Container Storage Interface)** is the standard plugin interface that allows storage vendors (AWS EBS, GCP PD, Ceph, NetApp, etc.) to expose their storage systems to Kubernetes without modifying core Kubernetes code. CSI drivers run as pods in the cluster and handle volume creation, attachment, mounting, and snapshotting.

**Reclaim Policies** control what happens to the PV when its PVC is deleted:
- **Retain**: PV is kept (data preserved) but not available for new claims. Admin must manually clean up.
- **Delete**: PV and the underlying storage are deleted automatically. Use with caution in production.

```
┌─────────────────────────────────────────────────────────────────┐
│                     Kubernetes Storage Flow                      │
│                                                                  │
│  StorageClass          PersistentVolume        Actual Storage   │
│  ┌──────────┐         ┌──────────────┐         ┌────────────┐  │
│  │  AWS EBS │──────── │  pv-001      │──────── │  EBS Vol   │  │
│  │  GP3     │  Static │  10Gi RWO    │         │  vol-xxx   │  │
│  └──────────┘         └──────────────┘         └────────────┘  │
│        │                     │                                  │
│        │ Dynamic              │ Bound                           │
│        │ Provisioning         │                                 │
│  ┌─────▼──────┐        ┌──────▼──────┐                         │
│  │    PVC     │─Claim──│     PVC     │                         │
│  │  10Gi RWO  │        │   Bound     │                         │
│  └────────────┘        └─────────────┘                         │
│        │                                                        │
│  ┌─────▼──────┐                                                 │
│  │    Pod     │ ← Mounts PVC as volume                         │
│  └────────────┘                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**StorageClass:**

A StorageClass enables **dynamic provisioning** — instead of an admin pre-creating PVs, Kubernetes automatically provisions the underlying storage when a PVC is created. Key fields:

- **provisioner**: The CSI driver that creates the volume (e.g., `ebs.csi.aws.com` for AWS EBS).
- **parameters**: Provider-specific settings (volume type, IOPS, encryption, KMS key).
- **volumeBindingMode**: `Immediate` creates the volume as soon as the PVC is created. `WaitForFirstConsumer` delays provisioning until a pod uses the PVC — this is critical in multi-AZ clusters to ensure the volume is created in the same availability zone as the pod.
- **reclaimPolicy**: `Retain` or `Delete` (as described above).
- **allowVolumeExpansion**: When `true`, PVCs can be resized by editing their `spec.resources.requests.storage` field.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
  annotations:
    storageclass.kubernetes.io/is-default-class: "false"
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  iops: "3000"
  throughput: "125"
  encrypted: "true"
  kmsKeyId: arn:aws:kms:us-east-1:123:key/xxx
volumeBindingMode: WaitForFirstConsumer   # Avoid cross-AZ issues
reclaimPolicy: Retain                     # Don't auto-delete PV
allowVolumeExpansion: true
allowedTopologies:
  - matchLabelExpressions:
      - key: topology.kubernetes.io/zone
        values:
          - us-east-1a
          - us-east-1b
```

**PersistentVolume (Static):**

Static provisioning means an admin manually creates PVs that map to existing storage volumes. This is used when you have pre-existing storage (a migrated database volume, a shared NFS export) or when you need fine-grained control over volume placement. The PV spec includes `nodeAffinity` to constrain which nodes the volume can be accessed from — essential for zone-specific block storage like EBS volumes.

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: pv-database
  labels:
    type: ssd
    app: postgres
spec:
  capacity:
    storage: 100Gi
  accessModes:
    - ReadWriteOnce       # RWO — single node
    # ReadWriteMany       # RWX — multiple nodes (NFS, EFS)
    # ReadOnlyMany        # ROX — multiple nodes read-only
    # ReadWriteOncePod    # RWOP — single pod (K8s 1.22+)
  storageClassName: fast-ssd
  persistentVolumeReclaimPolicy: Retain
  csi:
    driver: ebs.csi.aws.com
    volumeHandle: vol-0a1b2c3d4e5f
    fsType: ext4
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: topology.kubernetes.io/zone
              operator: In
              values:
                - us-east-1a
```

**PVC:**

A PVC is the "storage request" that a pod makes. It specifies the desired storage size, access mode, and optionally a StorageClass and label selector. Kubernetes either matches it to an existing PV (for static provisioning) or dynamically creates a new PV (if a StorageClass is specified). Once bound, the PVC-PV binding is exclusive — no other PVC can claim the same PV.

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: production
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 100Gi
  selector:              # For static PV binding
    matchLabels:
      app: postgres
      type: ssd
```

**StatefulSet with VolumeClaimTemplates:**

StatefulSets are designed for **stateful applications** (databases, message queues, distributed systems) that need stable network identities and persistent storage. Unlike Deployments, each StatefulSet pod gets a **unique, stable hostname** (e.g., `postgres-0`, `postgres-1`, `postgres-2`) and its own **dedicated PVC** created from the `volumeClaimTemplates`. These PVCs persist even if the pod is deleted and rescheduled — the new pod reattaches to the same volume, preserving data. The `serviceName` field must reference a headless Service that provides DNS entries for each pod (`postgres-0.postgres-headless.production.svc.cluster.local`).

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: postgres-headless
  replicas: 3
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
        - name: postgres
          image: postgres:15-alpine
          env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgres-secret
                  key: password
          volumeMounts:
            - name: data
              mountPath: /var/lib/postgresql/data
            - name: config
              mountPath: /etc/postgresql/postgresql.conf
              subPath: postgresql.conf
  volumeClaimTemplates:              # Each pod gets its own PVC
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: fast-ssd
        resources:
          requests:
            storage: 100Gi
```

---

## 🔒 CATEGORY 5: Security

---

### 10. Explain Kubernetes RBAC — Roles, ClusterRoles, Bindings, and Service Accounts

**Expected Answer:**

**RBAC (Role-Based Access Control)** is the authorization mechanism in Kubernetes that controls **who** (subject) can perform **what actions** (verbs) on **which resources** (objects). It follows the principle of least privilege — by default, no permissions are granted, and every permission must be explicitly defined.

RBAC has four key building blocks:

- **Role**: Defines a set of permissions (rules) within a **single namespace**. Each rule specifies API groups, resources (pods, deployments, secrets), and verbs (get, list, create, update, delete, patch, watch).
- **ClusterRole**: Same as Role but applies **cluster-wide** (across all namespaces). Also used for non-namespaced resources (nodes, PVs, namespaces themselves) and non-resource URLs (`/healthz`, `/metrics`).
- **RoleBinding**: Binds a Role to one or more subjects (Users, Groups, ServiceAccounts) within a specific namespace. The subject gets the permissions defined in the Role, but only in that namespace.
- **ClusterRoleBinding**: Binds a ClusterRole to subjects across the entire cluster.

**Subjects** can be:
- **Users**: External identities authenticated via certificates, OIDC, or tokens. Kubernetes does not manage user accounts internally — they come from external identity providers.
- **Groups**: Logical groupings of users (e.g., `dev-team`, `ops-team`). Defined by the authentication layer.
- **ServiceAccounts**: Kubernetes-managed identities for pods. Every pod runs under a ServiceAccount (default if not specified). In cloud environments, ServiceAccounts can be mapped to cloud IAM roles (e.g., AWS IRSA, GCP Workload Identity) for secure access to cloud resources.

A common pattern is: create a ClusterRole with the desired permissions, then use RoleBindings in specific namespaces to grant those permissions to different teams. This avoids duplicating Role definitions across namespaces.

```
┌─────────────────────────────────────────────────────┐
│                   RBAC Model                        │
│                                                     │
│  Subject          RoleBinding          Role         │
│  ┌──────────┐     ┌──────────────┐    ┌──────────┐ │
│  │   User   │────►│              │───►│  Role    │ │
│  │  Group   │     │ RoleBinding  │    │(Namespace│ │
│  │ Service  │     │    or        │    │  -scoped)│ │
│  │ Account  │     │  Cluster     │    └──────────┘ │
│  └──────────┘     │ RoleBinding  │    ┌──────────┐ │
│                   │              │───►│ Cluster  │ │
│                   └──────────────┘    │  Role    │ │
│                                       │(Cluster  │ │
│                                       │  -scoped)│ │
│                                       └──────────┘ │
└─────────────────────────────────────────────────────┘
```

```yaml
# Role — Namespace-scoped permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer-role
  namespace: staging
rules:
  - apiGroups: [""]                    # Core API group
    resources: ["pods", "pods/log", "pods/exec"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments", "replicasets"]
    verbs: ["get", "list", "watch", "create", "update", "patch"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: []                          # No secret access

---
# ClusterRole — Cluster-wide permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: node-reader
rules:
  - apiGroups: [""]
    resources: ["nodes", "nodes/status"]
    verbs: ["get", "list", "watch"]
  - nonResourceURLs: ["/metrics", "/healthz"]
    verbs: ["get"]

---
# RoleBinding — Bind role to user/group/SA in namespace
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: dev-binding
  namespace: staging
subjects:
  - kind: User
    name: john@company.com
    apiGroup: rbac.authorization.k8s.io
  - kind: Group
    name: dev-team
    apiGroup: rbac.authorization.k8s.io
  - kind: ServiceAccount
    name: ci-service-account
    namespace: ci-cd
roleRef:
  kind: Role
  name: developer-role
  apiGroup: rbac.authorization.k8s.io

---
# Service Account with IRSA (AWS EKS)
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-service-account
  namespace: production
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789:role/app-role
```

```bash
# Test RBAC permissions
kubectl auth can-i create deployments --namespace=staging
kubectl auth can-i get secrets --namespace=production
kubectl auth can-i '*' '*'   # Check if cluster-admin

# Check as specific user
kubectl auth can-i list pods \
  --namespace=staging \
  --as=john@company.com

# Check as service account
kubectl auth can-i list pods \
  --namespace=staging \
  --as=system:serviceaccount:staging:my-sa

# View all RBAC for a user
kubectl get rolebindings,clusterrolebindings \
  --all-namespaces \
  -o json | jq '.items[] | select(.subjects[]?.name == "john@company.com")'
```

---

### 11. Explain Pod Security — Pod Security Standards, Seccomp, and OPA/Gatekeeper

**Expected Answer:**

Pod security in Kubernetes operates at multiple layers:

**Pod Security Standards (PSS)** are built-in, namespace-level policies that replaced the old PodSecurityPolicy (removed in K8s 1.25). They define three progressively restrictive profiles that control what security-sensitive fields pods can use:
- **Privileged**: No restrictions at all. Used only for system-level workloads (CNI plugins, monitoring agents).
- **Baseline**: Prevents known privilege escalation vectors — blocks `hostNetwork`, `hostPID`, privileged containers, and dangerous capabilities. Suitable for most general-purpose workloads.
- **Restricted**: Enforces current hardening best practices — requires non-root execution, drops all capabilities, enforces read-only root filesystems, and mandates seccomp profiles. This is the target for production application pods.

PSS is enforced by labeling namespaces with the desired mode: `enforce` (reject violations), `audit` (log violations), or `warn` (show warnings). You can use all three simultaneously to roll out policies gradually.

**Seccomp (Secure Computing)** is a Linux kernel feature that restricts which **system calls** a container can make. Kubernetes supports seccomp profiles at the pod and container level. The `RuntimeDefault` profile uses the container runtime's built-in seccomp profile, which blocks ~44 dangerous syscalls while allowing normal application operation. Custom profiles can further restrict syscalls for security-sensitive workloads.

**OPA (Open Policy Agent) Gatekeeper** provides **policy-as-code** for Kubernetes. While PSS covers pod security fields, Gatekeeper can enforce arbitrary policies — required labels, allowed image registries, naming conventions, resource limit ranges, and more. Policies are written in **Rego** (OPA's policy language) and deployed as `ConstraintTemplates` (reusable policy definitions) and `Constraints` (instances that apply templates to specific resources/namespaces).

**Pod Security Standards (PSS) — Replaced PodSecurityPolicy in K8s 1.25:**
```yaml
# Apply Pod Security Standards to namespace
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    # Enforce: Reject non-compliant pods
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: v1.28

    # Audit: Log non-compliant pods (don't reject)
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/audit-version: v1.28

    # Warn: Show warning for non-compliant pods
    pod-security.kubernetes.io/warn: restricted
    pod-security.kubernetes.io/warn-version: v1.28
```

**PSS Levels:**
```
privileged  → No restrictions (for system components)
baseline    → Prevents known privilege escalations
restricted  → Heavily restricted, follows security best practices
```

**Secure Pod Spec (Restricted PSS Compliant):**

The following pod spec demonstrates production security best practices that comply with the `restricted` Pod Security Standard. Each field serves a specific security purpose:

- `runAsNonRoot: true` — Prevents the container from running as root (UID 0), reducing the blast radius of a container escape.
- `readOnlyRootFilesystem: true` — Makes the container filesystem immutable, preventing attackers from writing malicious binaries.
- `allowPrivilegeEscalation: false` — Blocks the `setuid` bit and prevents processes from gaining more privileges than their parent.
- `capabilities.drop: ["ALL"]` — Removes all Linux capabilities (network admin, raw sockets, etc.). Only add back specific capabilities if absolutely needed.
- `seccompProfile.type: RuntimeDefault` — Restricts dangerous system calls using the container runtime's default seccomp profile.
- `automountServiceAccountToken: false` — Prevents the ServiceAccount token from being mounted in the pod, reducing the attack surface if the pod is compromised. Only enable when the pod needs to call the Kubernetes API.
- Writable directories (`/tmp`, `/app/cache`) are provided via `emptyDir` volumes since the root filesystem is read-only.

```yaml
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1001
    runAsGroup: 1001
    fsGroup: 1001
    seccompProfile:
      type: RuntimeDefault     # Use container runtime's default seccomp
    supplementalGroups: [1001]

  containers:
    - name: app
      image: myapp:latest
      securityContext:
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
        capabilities:
          drop: ["ALL"]
          add: ["NET_BIND_SERVICE"]   # Only if binding port < 1024
        seccompProfile:
          type: RuntimeDefault
      resources:
        requests:
          memory: "128Mi"
          cpu: "250m"
        limits:
          memory: "256Mi"
          cpu: "500m"
      volumeMounts:
        - name: tmp
          mountPath: /tmp
        - name: cache
          mountPath: /app/cache

  volumes:
    - name: tmp
      emptyDir: {}
    - name: cache
      emptyDir:
        sizeLimit: "100Mi"

  automountServiceAccountToken: false   # Disable if not needed
```

**OPA Gatekeeper — Policy as Code:**

OPA Gatekeeper extends Kubernetes admission control with a flexible policy engine. It works through two resources:

1. **ConstraintTemplate**: Defines reusable policy logic in Rego (OPA's policy language). The template includes a CRD schema (what parameters the policy accepts) and the Rego code that evaluates incoming resources. Think of it as a "policy blueprint."
2. **Constraint**: Instantiates a template with specific parameters and scope (which resources and namespaces to target). Think of it as "apply this blueprint with these settings."

Gatekeeper runs as a **ValidatingAdmissionWebhook** — every API request is sent to Gatekeeper, which evaluates it against all active constraints. Violations are rejected with descriptive error messages. This makes policy enforcement **shift-left** — non-compliant resources are blocked at admission time, not discovered later via auditing.

```yaml
# ConstraintTemplate — Define the policy
apiVersion: templates.gatekeeper.sh/v1
kind: ConstraintTemplate
metadata:
  name: k8srequiredlabels
spec:
  crd:
    spec:
      names:
        kind: K8sRequiredLabels
      validation:
        openAPIV3Schema:
          type: object
          properties:
            labels:
              type: array
              items:
                type: string
  targets:
    - target: admission.k8s.gatekeeper.sh
      rego: |
        package k8srequiredlabels
        
        violation[{"msg": msg}] {
          provided := {label | input.review.object.metadata.labels[label]}
          required := {label | label := input.parameters.labels[_]}
          missing := required - provided
          count(missing) > 0
          msg := sprintf("Missing required labels: %v", [missing])
        }

---
# Constraint — Enforce the policy
apiVersion: constraints.gatekeeper.sh/v1beta1
kind: K8sRequiredLabels
metadata:
  name: require-team-labels
spec:
  match:
    kinds:
      - apiGroups: ["apps"]
        kinds: ["Deployment"]
    namespaces: ["production", "staging"]
  parameters:
    labels:
      - app
      - team
      - version
      - cost-center
```

---

## 📊 CATEGORY 6: Observability & Reliability

---

### 12. Explain Resource Management — Requests, Limits, QoS Classes, and LimitRange

**Expected Answer:**

Resource management is fundamental to Kubernetes cluster stability. Every container can specify **resource requests** and **resource limits** for CPU and memory:

- **Requests**: The **guaranteed minimum** resources a container needs. The scheduler uses requests to decide which node has enough capacity to place the pod. Requests are also used for eviction decisions — pods using less than their request are safe from eviction.
- **Limits**: The **maximum** resources a container can consume. For memory, exceeding the limit causes the container to be **OOM-killed** (exit code 137). For CPU, exceeding the limit causes the container to be **throttled** (slowed down, not killed).

Based on how requests and limits are set, Kubernetes automatically assigns one of three **QoS (Quality of Service) classes** to each pod. QoS determines **eviction priority** when a node runs low on resources — BestEffort pods are killed first, then Burstable, and Guaranteed pods are killed last.

**LimitRange** is a namespace-level policy that sets **default** requests/limits for containers that don't specify them, and enforces **min/max** bounds. This prevents developers from deploying pods without resource definitions (which would be BestEffort) or requesting excessively large resources.

**ResourceQuota** is a namespace-level policy that caps the **total** resource consumption across all pods in a namespace. It prevents any single team/namespace from consuming an outsized share of cluster resources, and can also limit object counts (number of pods, services, PVCs, etc.).

**QoS Classes (Kubernetes assigns automatically):**
```
┌─────────────────────────────────────────────────┐
│              QoS Priority (High → Low)          │
│                                                 │
│  Guaranteed  →  request == limit (both set)     │
│  Burstable   →  request < limit (or only limit) │
│  BestEffort  →  no requests or limits set       │
│                                                 │
│  Eviction order: BestEffort → Burstable         │
│                  → Guaranteed (last)            │
└─────────────────────────────────────────────────┘
```

```yaml
# Guaranteed QoS — production critical workloads
containers:
  - name: app
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "512Mi"    # Same as request
        cpu: "500m"        # Same as request

---
# Burstable QoS — most workloads
containers:
  - name: app
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"    # Can burst to 2x
        cpu: "1000m"

---
# LimitRange — Default limits for namespace
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: production
spec:
  limits:
    - type: Container
      default:             # Applied if no limits specified
        cpu: "500m"
        memory: "256Mi"
      defaultRequest:      # Applied if no requests specified
        cpu: "100m"
        memory: "128Mi"
      max:                 # Max allowed limits
        cpu: "4"
        memory: "4Gi"
      min:                 # Min allowed requests
        cpu: "50m"
        memory: "64Mi"
    - type: Pod
      max:
        cpu: "8"
        memory: "8Gi"
    - type: PersistentVolumeClaim
      max:
        storage: "50Gi"
      min:
        storage: "1Gi"

---
# ResourceQuota — Namespace-level hard limits
apiVersion: v1
kind: ResourceQuota
metadata:
  name: namespace-quota
  namespace: staging
spec:
  hard:
    # Compute resources
    requests.cpu: "10"
    requests.memory: "20Gi"
    limits.cpu: "20"
    limits.memory: "40Gi"
    # Object count limits
    pods: "50"
    services: "20"
    persistentvolumeclaims: "10"
    secrets: "30"
    configmaps: "30"
    # Storage
    requests.storage: "200Gi"
    fast-ssd.storageclass.storage.k8s.io/requests.storage: "100Gi"
```

---

### 13. How Does Horizontal Pod Autoscaling Work? Explain HPA, VPA, and KEDA

**Expected Answer:**

Kubernetes provides three complementary autoscaling strategies that address different scaling dimensions:

**HPA (Horizontal Pod Autoscaler)** scales the **number of pod replicas** based on observed metrics. The HPA controller queries the Metrics Server (or custom metrics adapters) every 15 seconds (default), computes the desired replica count using the formula `desiredReplicas = ceil(currentReplicas × (currentMetric / targetMetric))`, and scales the Deployment/StatefulSet accordingly. HPA supports CPU utilization, memory, custom metrics (from Prometheus via a metrics adapter), and external metrics (like SQS queue depth).

The `behavior` section controls scaling velocity — `stabilizationWindowSeconds` prevents flapping by waiting before scaling again, and `policies` limit how many pods can be added/removed per time period. Typically, you want to scale up aggressively (handle traffic spikes) and scale down conservatively (avoid premature scale-down during temporary dips).

**VPA (Vertical Pod Autoscaler)** adjusts the **resource requests and limits** of individual containers based on historical usage patterns. It monitors actual CPU/memory consumption and recommends (or automatically applies) right-sized resource values. VPA is essential for workloads where you don't know the right resource values upfront. It operates in four modes: `Off` (recommendations only), `Initial` (set resources only at pod creation), `Recreate` (restart pods to apply new resources), and `Auto` (apply updates whenever possible).

**Important**: HPA and VPA **should not target the same metric** on the same Deployment, as they'll fight each other. A common pattern is using VPA for resource right-sizing and HPA for replica scaling.

**KEDA (Kubernetes Event-Driven Autoscaler)** extends HPA with event-driven scaling triggered by external sources — message queue depth (Kafka lag, SQS messages), Prometheus queries, cron schedules, or custom metrics. KEDA's killer feature is **scale-to-zero** — it can completely remove all replicas when idle and spin them back up when events arrive. This dramatically reduces costs for workloads with intermittent traffic patterns.

**HPA — CPU/Memory based:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  minReplicas: 3
  maxReplicas: 50
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 30    # Wait 30s before scaling up again
      policies:
        - type: Pods
          value: 5
          periodSeconds: 60             # Add max 5 pods per minute
        - type: Percent
          value: 100
          periodSeconds: 60             # Or double the pods per minute
      selectPolicy: Max                 # Use the policy that scales MORE
    scaleDown:
      stabilizationWindowSeconds: 300   # Wait 5 min before scaling down
      policies:
        - type: Pods
          value: 2
          periodSeconds: 60             # Remove max 2 pods per minute
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70        # Scale when avg CPU > 70%
    - type: Resource
      resource:
        name: memory
        target:
          type: AverageValue
          averageValue: 400Mi
    # Custom metric from Prometheus
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "1000"
    # External metric (e.g., SQS queue depth)
    - type: External
      external:
        metric:
          name: sqs_messages_visible
          selector:
            matchLabels:
              queue: my-queue
        target:
          type: AverageValue
          averageValue: "100"
```

**VPA — Right-size resource requests:**

VPA consists of three components: the **Recommender** (analyzes historical resource usage and generates recommendations), the **Updater** (evicts pods that need resource updates), and the **Admission Controller** (sets resource requests on new pods based on recommendations). The `resourcePolicy` section lets you set floor (`minAllowed`) and ceiling (`maxAllowed`) values to prevent VPA from setting unreasonable resource values.

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: app-vpa
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: myapp
  updatePolicy:
    updateMode: "Auto"        # Off | Initial | Recreate | Auto
  resourcePolicy:
    containerPolicies:
      - containerName: "*"
        minAllowed:
          cpu: 100m
          memory: 128Mi
        maxAllowed:
          cpu: 4
          memory: 4Gi
        controlledResources: ["cpu", "memory"]
        controlledValues: RequestsAndLimits
```

**KEDA — Event-driven autoscaling:**

KEDA works by deploying a **ScaledObject** that connects a Deployment to one or more **triggers** (event sources). KEDA monitors these triggers using the `pollingInterval` and calculates the desired replica count. When all triggers report zero activity, KEDA scales the workload to `minReplicaCount` (which can be 0). The `cooldownPeriod` prevents rapid scale-down after the last event. KEDA supports 60+ trigger types including Kafka, AWS SQS, RabbitMQ, Prometheus, Cron, MySQL queries, and more.

```yaml
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  name: app-scaler
spec:
  scaleTargetRef:
    name: myapp
  pollingInterval: 15
  cooldownPeriod: 300
  minReplicaCount: 0         # Scale to ZERO when idle
  maxReplicaCount: 100
  triggers:
    # Scale on Kafka consumer lag
    - type: kafka
      metadata:
        bootstrapServers: kafka:9092
        consumerGroup: my-group
        topic: events
        lagThreshold: "50"

    # Scale on SQS queue
    - type: aws-sqs-queue
      authenticationRef:
        name: aws-credentials
      metadata:
        queueURL: https://sqs.us-east-1.amazonaws.com/123/my-queue
        queueLength: "10"
        awsRegion: us-east-1

    # Scale on Prometheus metric
    - type: prometheus
      metadata:
        serverAddress: http://prometheus:9090
        metricName: http_requests_total
        query: |
          sum(rate(http_requests_total{job="myapp"}[2m]))
        threshold: "1000"
```

---

### 14. How Do You Set Up Full Observability in Kubernetes?

**Expected Answer:**

Full observability in Kubernetes rests on **three pillars**:

1. **Metrics** (what is happening right now): Numeric measurements sampled over time — CPU usage, request rates, error rates, latency percentiles. **Prometheus** is the de facto standard, using a pull-based model where it scrapes `/metrics` endpoints from applications and infrastructure. **Grafana** provides dashboards for visualization.

2. **Logs** (what happened): Structured or unstructured text output from containers. A log aggregation pipeline collects logs from all pods/nodes, indexes them, and makes them searchable. **Loki** (lightweight, label-based log aggregation by Grafana Labs) or the **ELK/EFK stack** (Elasticsearch + Fluentd/Filebeat + Kibana) are common choices. **Promtail** or **Fluentd** runs as a DaemonSet on every node, tailing container log files and shipping them to the aggregation backend.

3. **Traces** (how did a request flow): Distributed traces follow a single request as it flows across multiple microservices, showing latency at each hop. **OpenTelemetry** is the vendor-neutral instrumentation standard, and backends like **Jaeger** or **Tempo** store and query traces.

In Kubernetes, the **kube-prometheus-stack** Helm chart deploys the entire monitoring pipeline in one shot: Prometheus, Grafana, Alertmanager, node-exporter (host metrics), kube-state-metrics (Kubernetes object metrics), and pre-built dashboards. **ServiceMonitors** and **PodMonitors** (CRDs from the Prometheus Operator) declaratively configure what Prometheus scrapes. **PrometheusRules** define alerting conditions using PromQL expressions.

**The Three Pillars:**
```
Metrics → Prometheus + Grafana
Logs    → Loki + Promtail (or ELK/EFK Stack)
Traces  → Jaeger or Tempo + OpenTelemetry
```

**Prometheus Stack via Helm:**
```bash
# Install kube-prometheus-stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --set prometheus.prometheusSpec.retention=30d \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.storageClassName=fast-ssd \
  --set prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage=100Gi \
  --set grafana.adminPassword=securepassword \
  --set alertmanager.alertmanagerSpec.storage.volumeClaimTemplate.spec.resources.requests.storage=10Gi
```

**Custom ServiceMonitor:**

A **ServiceMonitor** is a Prometheus Operator CRD that tells Prometheus which Services to scrape for metrics. Instead of editing Prometheus configuration files manually, you create a ServiceMonitor resource that matches your application's Service by label selector. The Prometheus Operator automatically generates the scrape configuration. The `release: monitoring` label is critical — it must match the Prometheus instance's `serviceMonitorSelector`, otherwise the ServiceMonitor is ignored. The `endpoints` section defines the port, path, interval, and TLS settings for the scrape target.

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: app-monitor
  namespace: monitoring
  labels:
    release: monitoring            # Must match Prometheus selector
spec:
  selector:
    matchLabels:
      app: myapp
  namespaceSelector:
    matchNames:
      - production
  endpoints:
    - port: metrics
      interval: 15s
      path: /metrics
      scheme: http
      tlsConfig:
        insecureSkipVerify: false
```

**PrometheusRule — Alerting:**

A **PrometheusRule** is a CRD that defines alerting rules using PromQL expressions. Each rule specifies:
- **expr**: A PromQL query that returns a value. If the result is non-empty, the alert fires.
- **for**: How long the condition must be true before the alert fires. This prevents alerting on brief spikes.
- **labels**: Metadata used for routing (severity, team) — Alertmanager uses these to determine notification channels.
- **annotations**: Human-readable context (summary, description, runbook link) included in the notification.

Alertmanager receives fired alerts and handles **grouping** (batch related alerts), **silencing** (suppress known issues), **inhibition** (suppress lower-severity alerts when higher-severity ones fire), and **routing** (send to the right team's Slack/PagerDuty based on labels).

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: app-alerts
  namespace: monitoring
spec:
  groups:
    - name: app.rules
      interval: 30s
      rules:
        - alert: HighErrorRate
          expr: |
            rate(http_requests_total{status=~"5.."}[5m]) /
            rate(http_requests_total[5m]) > 0.05
          for: 5m
          labels:
            severity: critical
            team: backend
          annotations:
            summary: "High error rate on {{ $labels.service }}"
            description: "Error rate is {{ $value | humanizePercentage }}"
            runbook: "https://wiki/runbooks/high-error-rate"

        - alert: PodCrashLooping
          expr: |
            increase(kube_pod_container_status_restarts_total[15m]) > 3
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Pod {{ $labels.pod }} is crash looping"

        - alert: NodeMemoryPressure
          expr: |
            (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) < 0.10
          for: 2m
          labels:
            severity: critical
```

---

## 🚀 CATEGORY 7: Advanced Topics

---

### 15. How Do You Manage Secrets in Kubernetes Securely at Scale?

**Expected Answer:**

Kubernetes native Secrets are fundamentally insecure for production use — they are only **Base64 encoded** (not encrypted), stored in plain text in etcd by default, and visible to anyone with RBAC access to the namespace. Managing secrets at scale requires a layered approach:

1. **Encrypt etcd at rest**: Configure the API server to encrypt Secret objects before writing them to etcd using AES-CBC or AES-GCM encryption. This protects against someone gaining raw access to the etcd data directory or backups.

2. **External Secret Management**: Store secrets in a dedicated vault (AWS Secrets Manager, HashiCorp Vault, Azure Key Vault, GCP Secret Manager) and sync them into Kubernetes using operators like **External Secrets Operator (ESO)** or inject them directly using **Vault Agent Injector**. This provides centralized secret management, automatic rotation, audit logging, and fine-grained access control that Kubernetes native Secrets lack.

3. **Minimize RBAC access**: Restrict who can `get`, `list`, or `watch` Secrets via RBAC. Even with encryption at rest, anyone who can read the Secret object gets the decoded value.

4. **Avoid environment variables for secrets**: Mount secrets as files instead, since environment variables can leak through process listings, crash dumps, and logs.

```
❌ Native Kubernetes Secrets → Base64 encoded (NOT encrypted by default)
✅ Production approach → External secret management
```

**Enable etcd Encryption at Rest:**

etcd encryption at rest is configured by creating an `EncryptionConfiguration` file and passing it to the API server via the `--encryption-provider-config` flag. The `providers` list is evaluated in order — the first provider is used for new writes, while subsequent providers are used to decrypt existing data. This allows key rotation: add a new key as the first provider, re-encrypt all secrets with `kubectl get secrets --all-namespaces -o json | kubectl replace -f -`, then remove the old key.

```yaml
# /etc/kubernetes/encryption-config.yaml
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
      - configmaps
    providers:
      - aescbc:               # AES-CBC encryption
          keys:
            - name: key1
              secret: <base64-encoded-32-byte-key>
      - identity: {}          # Fallback for unencrypted data
```

**External Secrets Operator (ESO) with AWS Secrets Manager:**

The **External Secrets Operator** is a Kubernetes operator that synchronizes secrets from external vaults into native Kubernetes Secrets. It works through two resources:

- **ClusterSecretStore / SecretStore**: Configures the connection to the external vault (authentication, region, service endpoint). `ClusterSecretStore` is cluster-scoped (shared across namespaces), while `SecretStore` is namespace-scoped.
- **ExternalSecret**: Declares which external secret keys to fetch, how often to sync (`refreshInterval`), and how to map them to Kubernetes Secret fields. The `template` section allows constructing composite values (e.g., a full DATABASE_URL from individual fields).

ESO continuously reconciles — if a secret is rotated in the external vault, ESO automatically updates the Kubernetes Secret within the refresh interval. When combined with Reloader or stakater/Reloader, pods can be automatically restarted when their mounted secrets change.

```yaml
# SecretStore — Connection to secret backend
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: aws-secret-store
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets-sa
            namespace: external-secrets

---
# ExternalSecret — Sync secret from AWS to K8s
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: database-credentials
  namespace: production
spec:
  refreshInterval: 1h           # Re-sync every hour
  secretStoreRef:
    name: aws-secret-store
    kind: ClusterSecretStore
  target:
    name: database-credentials  # K8s Secret name
    creationPolicy: Owner
    template:
      engineVersion: v2
      data:
        DATABASE_URL: "postgresql://{{ .username }}:{{ .password }}@{{ .host }}:5432/{{ .dbname }}"
  data:
    - secretKey: username
      remoteRef:
        key: prod/database
        property: username
    - secretKey: password
      remoteRef:
        key: prod/database
        property: password
  dataFrom:
    - extract:
        key: prod/app-secrets    # Sync all keys from this secret
```

**HashiCorp Vault with Agent Injector:**

The **Vault Agent Injector** takes a different approach — instead of syncing secrets into Kubernetes Secret objects, it injects a Vault Agent sidecar container into your pods. The sidecar authenticates to Vault (using Kubernetes ServiceAccount tokens), retrieves secrets, and writes them to a shared volume at `/vault/secrets/`. Your application reads secrets from files, never from Kubernetes Secrets.

This approach has advantages: secrets never exist as Kubernetes objects (reducing exposure), Vault handles dynamic secrets (e.g., short-lived database credentials), and automatic renewal/rotation happens within the pod lifecycle. The trade-off is added complexity — an extra container per pod, Vault infrastructure to manage, and debugging becomes harder.

The injector is configured entirely through pod annotations — no changes to the application container spec are needed.

```yaml
spec:
  template:
    metadata:
      annotations:
        vault.hashicorp.com/agent-inject: "true"
        vault.hashicorp.com/agent-inject-status: "update"
        vault.hashicorp.com/role: "myapp-role"
        # Inject as file at /vault/secrets/db-creds
        vault.hashicorp.com/agent-inject-secret-db-creds: "secret/data/myapp/database"
        vault.hashicorp.com/agent-inject-template-db-creds: |
          {{- with secret "secret/data/myapp/database" -}}
          export DB_USER="{{ .Data.data.username }}"
          export DB_PASS="{{ .Data.data.password }}"
          {{- end -}}
```

---

### 16. Explain Kubernetes Operators — When and How to Build One?

**Expected Answer:**

A Kubernetes **Operator** is a custom controller that extends the Kubernetes API to manage complex, stateful applications. It encodes operational knowledge (installation, configuration, scaling, backup, recovery, upgrades) into software, replacing manual runbooks with automated reconciliation.

The operator pattern combines two Kubernetes primitives:

- **CRD (Custom Resource Definition)**: Extends the Kubernetes API with a new resource type. For example, you might create a `PostgresCluster` resource that lets users declare a database cluster with `replicas: 3, version: "15", storage: "100Gi"`. The CRD defines the schema (what fields are valid, their types, and constraints).
- **Custom Controller**: Watches instances of the CRD (Custom Resources) and takes action to reconcile the desired state with actual state. When a user creates a `PostgresCluster`, the controller creates Deployments, Services, PVCs, ConfigMaps, backup CronJobs — everything needed to run a production Postgres cluster.

**When to build vs. buy an Operator:**
- **Use existing operators** for well-known software (databases, message queues, monitoring). The Kubernetes community has battle-tested operators for PostgreSQL, Kafka (Strimzi), Prometheus, cert-manager, etc.
- **Build a custom operator** when you have application-specific operational logic that can't be expressed with standard Kubernetes resources — e.g., custom health checks, application-aware scaling, automated data migrations, or compliance workflows.

**Day-1 operations** (what operators automate at installation): installation, initial configuration, dependency setup.
**Day-2 operations** (ongoing management): scaling, failover, backup/restore, version upgrades, certificate rotation, monitoring integration.

Popular frameworks for building operators: **Kubebuilder** (Go, official CNCF), **Operator SDK** (Go/Ansible/Helm, Red Hat), **Metacontroller** (any language via webhooks), **KUDO** (declarative operators).

```
┌─────────────────────────────────────────────────────────┐
│                  Operator Pattern                       │
│                                                         │
│  CRD (Custom Resource Definition)                       │
│  → Extends Kubernetes API with custom objects           │
│                                                         │
│  Custom Resource (CR)                                   │
│  → Instance of your CRD (like a Pod is of Pod spec)     │
│                                                         │
│  Controller (Operator Logic)                            │
│  → Watches CRs, reconciles desired → actual state       │
│  → Encodes Day-1 AND Day-2 operational knowledge        │
└─────────────────────────────────────────────────────────┘
```

```yaml
# CRD — Define custom resource schema
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: postgresclusters.db.company.com
spec:
  group: db.company.com
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: ["replicas", "version", "storage"]
              properties:
                replicas:
                  type: integer
                  minimum: 1
                  maximum: 10
                version:
                  type: string
                  enum: ["14", "15", "16"]
                storage:
                  type: string
                  pattern: '^[0-9]+(Gi|Ti)$'
                backup:
                  type: object
                  properties:
                    enabled:
                      type: boolean
                    schedule:
                      type: string
            status:
              type: object
              properties:
                phase:
                  type: string
                readyReplicas:
                  type: integer
      subresources:
        status: {}
  scope: Namespaced
  names:
    plural: postgresclusters
    singular: postgrescluster
    kind: PostgresCluster
    shortNames:
      - pgc

---
# Custom Resource usage
apiVersion: db.company.com/v1alpha1
kind: PostgresCluster
metadata:
  name: prod-postgres
  namespace: production
spec:
  replicas: 3
  version: "15"
  storage: "100Gi"
  backup:
    enabled: true
    schedule: "0 2 * * *"
```

```bash
# Popular production-ready operators:
# - cert-manager       → TLS certificate management
# - prometheus-operator → Monitoring stack
# - postgres-operator  → PostgreSQL cluster management
# - strimzi            → Kafka on Kubernetes
# - argo-cd            → GitOps continuous delivery
# - crossplane         → Infrastructure as Kubernetes resources
# - external-secrets   → External secret management
```

---

### 17. How Do You Implement GitOps with ArgoCD in Kubernetes?

**Expected Answer:**

**GitOps** is an operational framework where **Git is the single source of truth** for both infrastructure and application configuration. Instead of running `kubectl apply` manually or through CI pipelines, you commit desired state to a Git repository, and a GitOps controller (ArgoCD, Flux) continuously reconciles the cluster to match that state.

Core GitOps principles:
1. **Declarative**: The entire system is described declaratively (Kubernetes manifests, Helm charts, Kustomize overlays).
2. **Versioned and immutable**: Git provides a complete audit trail — every change has a commit, author, timestamp, and review (via pull requests).
3. **Pulled automatically**: The GitOps controller pulls changes from Git (pull model), rather than CI pushing to the cluster (push model). This is more secure because the cluster doesn't need to expose credentials to CI systems.
4. **Continuously reconciled**: If someone manually changes the cluster (kubectl edit, drift), the controller detects the divergence and reverts it to match Git (self-healing).

**ArgoCD** is the most popular GitOps controller for Kubernetes. It watches Git repositories, compares the desired state (in Git) with the live state (in the cluster), and syncs differences. Key features:
- **App of Apps pattern**: A parent Application that manages child Applications, enabling hierarchical management of multi-service deployments.
- **Multi-cluster support**: A single ArgoCD instance can manage multiple Kubernetes clusters.
- **Sync waves and hooks**: Control the order of resource deployment (e.g., create namespace before deploying apps).
- **SSO integration**: OIDC/SAML authentication with fine-grained project-level RBAC.

**ArgoCD Projects** provide security boundaries — they restrict which Git repos, cluster namespaces, and resource types an Application can access. This prevents a team's ArgoCD Application from deploying to another team's namespace.

```
┌────────────────────────────────────────────────────────────────┐
│                      GitOps Flow                               │
│                                                                │
│  Developer          Git Repo            ArgoCD         K8s    │
│     │                  │                  │              │     │
│     │──git push──►     │                  │              │     │
│     │                  │◄──── Watches ────│              │     │
│     │                  │                  │              │     │
│     │                  │──── Diff ───────►│              │     │
│     │                  │  (desired vs     │              │     │
│     │                  │   actual state)  │              │     │
│     │                  │                  │──── Sync ───►│     │
│     │                  │                  │  (apply      │     │
│     │                  │                  │   manifests) │     │
│     │                  │                  │              │     │
│     │                  │                  │◄── Status ───│     │
└────────────────────────────────────────────────────────────────┘
```

```yaml
# ArgoCD Application — App of Apps pattern
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: production-apps
  namespace: argocd
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: production
  source:
    repoURL: https://github.com/company/k8s-manifests
    targetRevision: main
    path: environments/production
    # Helm support
    helm:
      valueFiles:
        - values-prod.yaml
      parameters:
        - name: image.tag
          value: v1.2.3
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true         # Delete resources removed from Git
      selfHeal: true      # Revert manual cluster changes
      allowEmpty: false
    syncOptions:
      - CreateNamespace=true
      - PrunePropagationPolicy=foreground
      - ApplyOutOfSyncOnly=true
      - ServerSideApply=true
    retry:
      limit: 5
      backoff:
        duration: 5s
        factor: 2
        maxDuration: 3m
  ignoreDifferences:
    - group: apps
      kind: Deployment
      jsonPointers:
        - /spec/replicas    # Ignore HPA-managed replica count

---
# ArgoCD Project — RBAC and restrictions
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: production
  namespace: argocd
spec:
  description: "Production workloads"
  sourceRepos:
    - "https://github.com/company/*"
    - "https://charts.bitnami.com/*"
  destinations:
    - namespace: production
      server: https://kubernetes.default.svc
    - namespace: monitoring
      server: https://kubernetes.default.svc
  clusterResourceWhitelist:
    - group: ""
      kind: Namespace
  namespaceResourceBlacklist:
    - group: ""
      kind: ResourceQuota    # Prevent changing quotas
  roles:
    - name: developer
      description: Read-only for developers
      policies:
        - "p, proj:production:developer, applications, get, production/*, allow"
        - "p, proj:production:developer, applications, sync, production/*, allow"
      groups:
        - dev-team
```

---

### 18. How Do You Perform Cluster Upgrades with Zero Downtime?

**Expected Answer:**

Kubernetes cluster upgrades must be carefully planned because they involve upgrading the control plane and every worker node. Key principles:

1. **One minor version at a time**: Kubernetes supports upgrading only one minor version per step (e.g., 1.27 → 1.28 → 1.29, never 1.27 → 1.29). This is because API deprecations and feature gates change between versions.
2. **Control plane first, then workers**: Always upgrade control plane nodes before worker nodes. Kubernetes guarantees that the kubelet is backward-compatible with the API server up to 2 minor versions behind, but not forward-compatible.
3. **Rolling worker node upgrades**: Upgrade one worker node at a time using the **cordon → drain → upgrade → uncordon** sequence. This ensures workloads are safely rescheduled to healthy nodes before each node goes offline.
4. **PodDisruptionBudgets (PDBs)**: Define the minimum number of pods that must remain available during voluntary disruptions (node drain, upgrades). The `kubectl drain` command respects PDBs and will wait if draining a node would violate the budget. Without PDBs, draining a node could take down all replicas of a service simultaneously.
5. **Test in staging first**: Always upgrade a staging/dev cluster first to catch deprecated API usage, incompatible webhooks, or CNI/CSI driver issues.

**Deprecated API detection** is the most common upgrade blocker. When an API version is removed in a new Kubernetes release (e.g., `extensions/v1beta1` Ingress), all manifests using the old API must be updated before the upgrade. Tools like **Pluto** scan the cluster for deprecated API usage.

```bash
# ─────────────────────────────────────────────
# Pre-Upgrade Checklist
# ─────────────────────────────────────────────

# 1. Check current version and available upgrades
kubectl version --short
kubeadm upgrade plan

# 2. Check deprecated APIs (CRITICAL)
kubectl get all --all-namespaces -o yaml | \
  grep "apiVersion" | sort | uniq -c | sort -rn

# Use Pluto to detect deprecated APIs
pluto detect-all-in-cluster --target-versions k8s=v1.28.0

# 3. Backup etcd
ETCDCTL_API=3 etcdctl snapshot save /backup/pre-upgrade-$(date +%Y%m%d).db

# 4. Review release notes and changelog
# https://kubernetes.io/releases/

# ─────────────────────────────────────────────
# Upgrade Control Plane (one minor version at a time!)
# ─────────────────────────────────────────────

# On first control plane node
apt-get update
apt-get install -y kubeadm=1.28.0-00

kubeadm upgrade plan v1.28.0
kubeadm upgrade apply v1.28.0

# Upgrade kubelet and kubectl
kubectl drain <control-plane-node> \
  --ignore-daemonsets \
  --delete-emptydir-data
  
apt-get install -y kubelet=1.28.0-00 kubectl=1.28.0-00
systemctl daemon-reload && systemctl restart kubelet

kubectl uncordon <control-plane-node>

# Repeat for additional control plane nodes
kubeadm upgrade node  # (not apply)

# ─────────────────────────────────────────────
# Upgrade Worker Nodes (Rolling)
# ─────────────────────────────────────────────

# For each worker node:
# 1. Cordon — prevent new pod scheduling
kubectl cordon <worker-node>

# 2. Drain — evict existing pods (respects PodDisruptionBudgets)
kubectl drain <worker-node> \
  --ignore-daemonsets \
  --delete-emptydir-data \
  --force \
  --grace-period=120 \
  --timeout=300s

# 3. Upgrade packages on the node
ssh worker-node
apt-get install -y kubeadm=1.28.0-00 kubelet=1.28.0-00 kubectl=1.28.0-00
kubeadm upgrade node
systemctl daemon-reload && systemctl restart kubelet

# 4. Uncordon — allow scheduling again
kubectl uncordon <worker-node>

# 5. Verify node is healthy before continuing
kubectl get node <worker-node>
kubectl get pods -A --field-selector spec.nodeName=<worker-node>

# ─────────────────────────────────────────────
# PodDisruptionBudget — Ensure availability during drain
# ─────────────────────────────────────────────
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: app-pdb
spec:
  minAvailable: 2      # At least 2 pods always available
  # OR
  maxUnavailable: 1    # At most 1 pod unavailable at a time
  selector:
    matchLabels:
      app: myapp
```

---

### 19. How Do You Troubleshoot a Production Kubernetes Issue Systematically?

**Expected Answer:**

Effective Kubernetes troubleshooting follows a **top-down, systematic approach** rather than random debugging. Start broad (cluster level), narrow down (namespace, workload, pod), and get specific (container logs, events, exec into pod). The key is understanding **which layer** the problem is at:

- **Scheduling layer**: Pod can't be placed on a node (insufficient resources, taints, affinity mismatch). Look at: `kubectl describe pod` events, `kubectl get events`.
- **Image layer**: Container image can't be pulled (wrong tag, missing imagePullSecret, registry down). Look at: pod events showing `ImagePullBackOff` or `ErrImagePull`.
- **Application layer**: Container starts but crashes or fails health checks (application bug, missing config, OOM). Look at: `kubectl logs`, exit codes, readiness/liveness probe configuration.
- **Network layer**: Service connectivity issues (wrong selector, missing endpoints, NetworkPolicy blocking). Look at: endpoints, DNS resolution, kube-proxy rules.
- **Node layer**: Node unhealthy or under pressure (disk, memory, PID exhaustion). Look at: `kubectl describe node` conditions, `kubectl top nodes`, kubelet logs.

**Exit code interpretation** is critical for understanding container crashes:
- **0**: Clean exit (application finished normally).
- **1**: Generic application error.
- **137** (128 + 9): OOM killed — the container exceeded its memory limit and was killed by the kernel with SIGKILL. Fix: increase memory limits or fix the memory leak.
- **139** (128 + 11): Segmentation fault — the application accessed invalid memory.
- **143** (128 + 15): Graceful termination — the container received SIGTERM and exited. This is normal during rolling updates or scaling down.

**Ephemeral debug containers** (`kubectl debug`) are invaluable for production troubleshooting — they attach a temporary container (with networking tools, shells, etc.) to a running pod without modifying the pod spec or restarting it.

```bash
# ─────────────────────────────────────────────
# SCENARIO: Pods are not starting
# ─────────────────────────────────────────────

# Step 1: Check pod status
kubectl get pods -n production -o wide
kubectl get events -n production \
  --sort-by='.lastTimestamp' | tail -30

# Step 2: Describe the pod
kubectl describe pod <pod-name> -n production
# Look for: Events, Conditions, Init Containers, Resource limits

# Step 3: Check logs
kubectl logs <pod-name> -n production --previous
kubectl logs <pod-name> -n production -c <container-name>
kubectl logs <pod-name> -n production --timestamps --tail=100

# Step 4: Decode exit codes
# 0   = Clean exit
# 1   = Application error
# 137 = OOM killed (128 + 9 SIGKILL)
# 139 = Segfault (128 + 11)
# 143 = Graceful termination (128 + 15 SIGTERM)

# ─────────────────────────────────────────────
# SCENARIO: Service not reachable
# ─────────────────────────────────────────────

# Check endpoints — are pods matched by service selector?
kubectl get endpoints <service-name> -n production
kubectl describe service <service-name> -n production

# Test DNS resolution from inside cluster
kubectl run debug-pod \
  --image=busybox:latest \
  --rm -it --restart=Never \
  -- nslookup myapp-service.production.svc.cluster.local

# Test connectivity
kubectl run debug-pod \
  --image=curlimages/curl \
  --rm -it --restart=Never \
  -- curl -v http://myapp-service.production.svc.cluster.local/health

# Check NetworkPolicy is not blocking
kubectl get networkpolicies -n production
kubectl describe networkpolicy <policy-name>

# ─────────────────────────────────────────────
# SCENARIO: Node issues
# ─────────────────────────────────────────────

# Check node conditions
kubectl describe node <node-name> | grep -A 10 "Conditions:"

# Node Conditions:
# Ready         = Node is healthy
# MemoryPressure = Node is low on memory
# DiskPressure  = Node is low on disk
# PIDPressure   = Too many processes

# Check resource usage on node
kubectl top nodes
kubectl top pods -n production --sort-by=memory

# Check pods on specific node
kubectl get pods -A --field-selector spec.nodeName=<node>

# SSH into node for deeper investigation
ssh <node>
journalctl -u kubelet -f
systemctl status containerd

# ─────────────────────────────────────────────
# SCENARIO: etcd issues / API server slow
# ─────────────────────────────────────────────

# Check API server latency
kubectl get --raw /metrics | grep apiserver_request_duration

# Check etcd health
ETCDCTL_API=3 etcdctl endpoint health --cluster
ETCDCTL_API=3 etcdctl endpoint status --write-out=table

# Check for etcd defragmentation need
# DB size > 80% of quota = defragment needed

# ─────────────────────────────────────────────
# SCENARIO: Image pull failures
# ─────────────────────────────────────────────

# Check imagePullSecret is correct
kubectl get secret regcred -n production -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d

# Verify image exists and tag is correct
docker manifest inspect myregistry.com/myapp:v1.0

# ─────────────────────────────────────────────
# Golden kubectl debugging toolkit
# ─────────────────────────────────────────────

# Swiss-army debug pod (ephemeral container)
kubectl debug -it <pod-name> \
  --image=nicolaka/netshoot \
  --target=<container-name>

# Run debug pod on specific node
kubectl debug node/<node-name> \
  -it \
  --image=ubuntu \
  -- bash

# Port-forward for local testing
kubectl port-forward svc/myapp-service 8080:80 -n production

# Watch pod events in real time
kubectl get events -n production -w \
  --field-selector involvedObject.name=<pod-name>
```

---

### 20. Design a Highly Available, Production-Grade Kubernetes Cluster — Architecture Discussion

**Expected Answer:**

A production-grade Kubernetes cluster must be designed for **high availability (HA)**, **fault tolerance**, **security**, and **operational efficiency**. The key architectural decisions are:

**High Availability**: The control plane must span at least **3 availability zones** with an odd number of etcd members (3 or 5) for Raft quorum. A load balancer distributes API server traffic across control plane nodes. Worker nodes are spread across AZs to survive zone failures. PodDisruptionBudgets and topology spread constraints ensure application availability during node failures.

**Node Pool Strategy**: Use multiple node pools with different instance types for different workload profiles — general-purpose (web servers, APIs), memory-optimized (caches, databases), GPU (ML inference), and spot/preemptible instances (batch processing, dev workloads). Taints on specialized node pools prevent general workloads from landing on expensive GPU nodes.

**Security in Depth**: RBAC for authorization, Pod Security Standards for container hardening, Network Policies for micro-segmentation, OPA Gatekeeper for policy enforcement, External Secrets Operator for secret management, image signing (cosign) for supply chain security, and audit logging for compliance.

**Observability**: The Prometheus/Grafana/Alertmanager stack for metrics and alerting, Loki for logs, Tempo/Jaeger for traces, and Kubecost for cost visibility. Alert fatigue is the enemy — design alerts around **SLOs (Service Level Objectives)** with clear runbooks.

**Disaster Recovery**: Automated etcd backups (every 30 minutes), Velero for cluster state and PV snapshot backups, documented and tested recovery procedures with clear RTO/RPO targets, and optionally multi-region active-passive clusters for the most critical workloads.

**GitOps**: ArgoCD with the app-of-apps pattern ensures all cluster configuration is version-controlled, reviewed via pull requests, and automatically reconciled. This provides audit trails, easy rollbacks, and prevents configuration drift.

```
┌─────────────────────────────────────────────────────────────────────┐
│              Production HA Kubernetes Architecture                   │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                     Global Load Balancer                     │   │
│  │                  (Route53 / Cloudflare)                      │   │
│  └─────────────────────────┬────────────────────────────────────┘   │
│                            │                                        │
│  ┌─────────────────────────▼────────────────────────────────────┐   │
│  │             Regional Load Balancer (AWS ALB/NLB)             │   │
│  └─────┬─────────────────────────────────────┬──────────────────┘   │
│        │                                     │                      │
│  ┌─────▼──────┐  ┌─────────────┐  ┌──────────▼────┐                │
│  │  Control   │  │  Control    │  │   Control     │                │
│  │  Plane 1   │  │  Plane 2   │  │   Plane 3    │ ← 3 AZs        │
│  │  (AZ-a)    │  │  (AZ-b)    │  │   (AZ-c)     │                │
│  └──────┬─────┘  └──────┬──────┘  └───────┬───────┘                │
│         │               │                 │                        │
│  ┌──────▼───────────────▼─────────────────▼────────────────────┐   │
│  │                  etcd Cluster (3 or 5 nodes)                 │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  Node Pool   │  │  Node Pool   │  │  Node Pool   │              │
│  │  General     │  │  High Mem    │  │  GPU         │              │
│  │  (AZ-a,b,c)  │  │  (AZ-a,b,c)  │  │  (AZ-a)      │              │
│  └──────────────┘  └──────────────┘  └──────────────┘              │
└─────────────────────────────────────────────────────────────────────┘
```

**Production Cluster Checklist:**

```yaml
# ─── CONTROL PLANE ───────────────────────────────────────
Control Plane:
  replicas: 3                    # Odd number for quorum
  zones: [us-east-1a, us-east-1b, us-east-1c]
  apiServer:
    auditLogging: enabled        # Required for compliance
    oidcProvider: enabled        # SSO integration
    admissionControllers:
      - PodSecurityAdmission
      - OPA-Gatekeeper
      - ResourceQuota
  etcd:
    replicas: 3
    encryption: true             # Encrypt secrets at rest
    backup:
      schedule: "*/30 * * * *"  # Every 30 minutes
      retention: 7d
      destination: s3://etcd-backups

# ─── WORKER NODES ─────────────────────────────────────────
NodePools:
  - name: general-purpose
    instanceType: m5.2xlarge
    minNodes: 3
    maxNodes: 50
    zones: [us-east-1a, us-east-1b, us-east-1c]
    autoScaling: true
    spotEnabled: true            # Cost optimization (with on-demand fallback)
    spotPercentage: 70

  - name: memory-optimized
    instanceType: r5.4xlarge
    minNodes: 0
    maxNodes: 20
    taints:
      - key: workload-type
        value: memory-intensive
        effect: NoSchedule

# ─── NETWORKING ───────────────────────────────────────────
Networking:
  cni: cilium
  encryption: WireGuard          # Encrypt pod-to-pod traffic
  loadBalancer: aws-load-balancer-controller
  ingressController: nginx
  serviceMesh: istio             # mTLS, traffic management, observability
  networkPolicies: enforced

# ─── SECURITY ─────────────────────────────────────────────
Security:
  rbac: enabled
  podSecurityStandards: restricted
  imageScanning: trivy
  admissionControl: opa-gatekeeper
  secretManagement: external-secrets-operator
  secretBackend: aws-secrets-manager
  containerRuntime: containerd
  imageSigning: cosign
  auditLogs: cloudwatch

# ─── OBSERVABILITY ────────────────────────────────────────
Observability:
  metrics: kube-prometheus-stack
  logging: loki-stack
  tracing: tempo + opentelemetry
  dashboards: grafana
  alerting: alertmanager + pagerduty
  costMonitoring: kubecost

# ─── DISASTER RECOVERY ────────────────────────────────────
DisasterRecovery:
  etcdBackup: every-30-min
  veleroBackup: daily            # Cluster state + PV snapshots
  rto: 30min                    # Recovery Time Objective
  rpo: 30min                    # Recovery Point Objective
  multiRegion: active-passive

# ─── GITOPS ───────────────────────────────────────────────
GitOps:
  tool: argocd
  strategy: app-of-apps
  autoSync: true
  selfHeal: true
  imageUpdater: enabled          # Auto-update image tags from registry
```

```bash
# Cluster health validation script
#!/bin/bash
echo "=== Node Status ===" && kubectl get nodes -o wide
echo "=== Control Plane ===" && kubectl get pods -n kube-system
echo "=== etcd Health ===" && kubectl -n kube-system exec etcd-master -- \
  etcdctl endpoint health --cluster
echo "=== PodDisruptionBudgets ===" && kubectl get pdb --all-namespaces
echo "=== ResourceQuotas ===" && kubectl describe resourcequota --all-namespaces
echo "=== Top Nodes ===" && kubectl top nodes
echo "=== Failing Pods ===" && kubectl get pods --all-namespaces \
  --field-selector=status.phase!=Running,status.phase!=Succeeded
```

---

> 💡 **Senior Engineer Interview Tips:**
> - Always discuss **trade-offs** (e.g., IPVS vs iptables, Cilium vs Calico)
> - Reference **real incidents** — "We had a situation where etcd hit quota..."
> - Show knowledge of **ecosystem tools** — ArgoCD, Cilium, KEDA, Crossplane
> - Understand **cloud-specific nuances** — EKS, GKE, AKS differences
> - Discuss **cost optimization** — Spot instances, right-sizing, Kubecost
> - Mention **compliance requirements** — SOC2, PCI-DSS audit logging, RBAC policies
> - Know **when NOT to use Kubernetes** — small teams, simple apps may not need it
