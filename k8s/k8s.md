# Top 20 Kubernetes Interview Questions for a Senior DevOps Engineer (5+ Years)

---

## 🏗️ CATEGORY 1: Kubernetes Architecture & Internals

---

### 1. Explain Kubernetes Architecture in Depth — Control Plane to Data Plane

**Expected Answer:**

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

| Component | Role | Key Responsibility |
|-----------|------|-------------------|
| **kube-apiserver** | API Gateway | Validates/processes all REST requests, only component that writes to etcd |
| **etcd** | Distributed KV Store | Stores entire cluster state, uses Raft consensus |
| **kube-scheduler** | Pod Scheduler | Assigns pods to nodes based on resources, affinity, taints |
| **kube-controller-manager** | Control Loop Runner | Runs Node, Replication, Endpoint, ServiceAccount controllers |
| **cloud-controller-manager** | Cloud Integration | Manages LBs, volumes, node lifecycle via cloud APIs |

**Data Plane Components:**

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

**Kubernetes Networking Rules:**
```
1. Every Pod gets its own IP address
2. Pods can communicate with all other pods without NAT
3. Nodes can communicate with all pods without NAT
4. Pod IP is the same from inside and outside the pod
```

**CNI Plugins Comparison:**

| CNI Plugin | Network Model | Network Policy | Encryption | Best For |
|-----------|--------------|----------------|------------|---------|
| **Calico** | BGP / VXLAN | ✅ Advanced | WireGuard | Enterprise, large scale |
| **Flannel** | VXLAN | ❌ Basic | ❌ | Simple clusters |
| **Cilium** | eBPF | ✅ L7 | WireGuard | High performance, observability |
| **Weave** | Mesh / VXLAN | ✅ | ✅ | Multi-cloud |
| **AWS VPC CNI** | Native VPC IPs | Via Calico | VPC | EKS |

**Service Types:**
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

```
❌ Native Kubernetes Secrets → Base64 encoded (NOT encrypted by default)
✅ Production approach → External secret management
```

**Enable etcd Encryption at Rest:**
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
