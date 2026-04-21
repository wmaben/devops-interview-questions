# Top 20 Docker Interview Questions for a Senior DevOps Engineer (5+ Years)

---

## 🏗️ CATEGORY 1: Docker Architecture & Internals

---

### 1. Explain Docker's Architecture in Depth. How Does It Work Under the Hood?

**Expected Answer:**

Docker follows a **client-server architecture** with these core components:

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Client (CLI)                   │
│              docker build / run / pull                   │
└────────────────────────┬────────────────────────────────┘
                         │ REST API
┌────────────────────────▼────────────────────────────────┐
│                   Docker Daemon (dockerd)                │
│   ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  │
│   │  Container  │  │    Image     │  │   Network    │  │
│   │  Manager    │  │   Manager    │  │   Manager    │  │
│   └─────────────┘  └──────────────┘  └──────────────┘  │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                  containerd (Runtime)                    │
│              ┌──────────────────────┐                   │
│              │      runc (OCI)      │                   │
│              └──────────────────────┘                   │
└─────────────────────────────────────────────────────────┘
```

**Key Components:**
- **Docker Client** — CLI that sends commands via REST API
- **dockerd** — Background daemon managing containers, images, volumes, networks
- **containerd** — High-level runtime managing container lifecycle
- **runc** — Low-level OCI-compliant runtime that actually creates containers
- **Docker Registry** — Stores and distributes images (Docker Hub, ECR, GCR)

**Linux Kernel Features Used:**
| Feature | Purpose |
|---------|---------|
| **Namespaces** | Process isolation (PID, NET, MNT, UTS, IPC, USER) |
| **cgroups** | Resource limiting (CPU, Memory, I/O) |
| **Union File System** | Layered image filesystem (overlay2) |
| **seccomp** | System call filtering |
| **capabilities** | Fine-grained privilege control |

---

### 2. What is the Difference Between Docker Image Layers and How Does Copy-on-Write (CoW) Work?

**Expected Answer:**

Docker images are built in **read-only layers**. Each instruction in a Dockerfile creates a new layer.

```
┌────────────────────────────┐  ← Container Layer (R/W)
├────────────────────────────┤  ← COPY app/ /app (Layer 4)
├────────────────────────────┤  ← RUN pip install (Layer 3)
├────────────────────────────┤  ← RUN apt-get update (Layer 2)
├────────────────────────────┤  ← FROM ubuntu:22.04 (Layer 1 - Base)
└────────────────────────────┘
```

**Copy-on-Write (CoW) Mechanism:**
- All image layers are **read-only**
- When a container starts, a thin **writable layer** is added on top
- When a file from a lower layer is modified, it is **copied up** to the writable layer first, then modified
- This means **multiple containers share the same base image layers** — saving disk space

```bash
# Check image layers
docker image inspect nginx --format '{{json .RootFS.Layers}}'

# Check overlay2 filesystem layers
docker inspect <container_id> --format '{{json .GraphDriver}}'
```

**Storage Drivers:**
| Driver | Use Case |
|--------|---------|
| **overlay2** | Default, recommended for most Linux distros |
| **devicemapper** | Older RHEL/CentOS systems |
| **btrfs** | Advanced filesystem features |
| **vfs** | Testing only, no CoW benefits |

---

### 3. What are Linux Namespaces and cgroups? How Does Docker Use Them?

**Expected Answer:**

**Namespaces** — provide isolation for containers:

```bash
# View namespaces of a running container
docker inspect <container_id> --format '{{.State.Pid}}'
ls -la /proc/<PID>/ns/

# Types of namespaces Docker uses:
# pid    → Isolates process IDs
# net    → Isolates network interfaces
# mnt    → Isolates mount points
# uts    → Isolates hostname and domain name
# ipc    → Isolates inter-process communication
# user   → Isolates user and group IDs
```

**cgroups** — provide resource limits:

```bash
# View cgroup resource limits for a container
cat /sys/fs/cgroup/memory/docker/<container_id>/memory.limit_in_bytes

# Set resource limits when running a container
docker run -d \
  --memory="512m" \
  --memory-swap="1g" \
  --cpus="1.5" \
  --cpu-shares=512 \
  --pids-limit=100 \
  nginx
```

**cgroup v1 vs v2:**
- **cgroup v1** — Each resource type has its own hierarchy
- **cgroup v2** — Unified hierarchy, better resource distribution, used by modern kernels

---

## 🐳 CATEGORY 2: Dockerfile Best Practices

---

### 4. What Are Multi-Stage Builds and Why Are They Critical in Production?

**Expected Answer:**

Multi-stage builds allow you to use **multiple FROM statements** in a single Dockerfile, keeping only what's needed in the final image.

```dockerfile
# ─────────────────────────────────────────
# Stage 1: Builder
# ─────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy and download dependencies first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server .

# ─────────────────────────────────────────
# Stage 2: Security Scanner (optional)
# ─────────────────────────────────────────
FROM aquasec/trivy:latest AS scanner
COPY --from=builder /app/server /app/server
RUN trivy filesystem --exit-code 1 --severity HIGH,CRITICAL /app/server

# ─────────────────────────────────────────
# Stage 3: Final Minimal Image
# ─────────────────────────────────────────
FROM scratch AS final

# Copy only the binary and necessary certs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/server /server

EXPOSE 8080
USER 1001
ENTRYPOINT ["/server"]
```

**Size Comparison:**
| Approach | Image Size |
|----------|-----------|
| Single stage (golang:1.21) | ~900 MB |
| Multi-stage (alpine) | ~15 MB |
| Multi-stage (scratch) | ~6 MB |

---

### 5. How Do You Optimize a Dockerfile for Faster Builds and Smaller Images?

**Expected Answer:**

**❌ Unoptimized Dockerfile:**
```dockerfile
FROM node:18
WORKDIR /app
COPY . .
RUN npm install
RUN npm run build
EXPOSE 3000
CMD ["node", "server.js"]
```

**✅ Fully Optimized Dockerfile:**
```dockerfile
# Use specific version tags — never 'latest' in production
FROM node:18.19-alpine3.19 AS base

# Install only production dependencies
FROM base AS deps
WORKDIR /app
COPY package*.json ./
# Use ci for reproducible installs
RUN npm ci --only=production && npm cache clean --force

# Build stage
FROM base AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Final minimal image
FROM node:18.19-alpine3.19 AS runner
WORKDIR /app

# Create non-root user
RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs

# Copy only what's needed
COPY --from=deps    --chown=nextjs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nextjs:nodejs /app/dist ./dist
COPY --from=builder --chown=nextjs:nodejs /app/package.json ./

# Use non-root user
USER nextjs

EXPOSE 3000
ENV NODE_ENV=production

# Use exec form to prevent signal issues
CMD ["node", "dist/server.js"]
```

**Key Optimization Rules:**
```
✅ Order instructions from LEAST to MOST frequently changing
✅ Combine RUN commands to reduce layers
✅ Use .dockerignore aggressively
✅ Use specific base image tags
✅ Remove package manager caches in same RUN step
✅ Use multi-stage builds
✅ Never run as root
✅ Use COPY over ADD (unless extracting archives)
```

**Essential .dockerignore:**
```
node_modules
.git
.gitignore
*.md
*.log
.env*
dist
coverage
.nyc_output
Dockerfile*
docker-compose*
```

---

### 6. What is the Difference Between CMD, ENTRYPOINT, and RUN?

**Expected Answer:**

```dockerfile
# RUN — Executes during IMAGE BUILD, creates a new layer
RUN apt-get update && apt-get install -y curl

# CMD — Default command when container STARTS, easily overridden
CMD ["nginx", "-g", "daemon off;"]         # Exec form (preferred)
CMD nginx -g "daemon off;"                 # Shell form (avoid)

# ENTRYPOINT — Main process, NOT easily overridden
ENTRYPOINT ["docker-entrypoint.sh"]        # Exec form
```

**Interaction Between CMD and ENTRYPOINT:**
```dockerfile
ENTRYPOINT ["python", "app.py"]
CMD ["--port", "8080"]   # Acts as default arguments to ENTRYPOINT

# docker run myimage           → python app.py --port 8080
# docker run myimage --port 9090 → python app.py --port 9090
# docker run --entrypoint /bin/sh myimage → overrides ENTRYPOINT
```

| | RUN | CMD | ENTRYPOINT |
|--|-----|-----|------------|
| **When** | Build time | Runtime | Runtime |
| **Creates Layer** | Yes | No | No |
| **Overridable** | N/A | Yes (easy) | Yes (--entrypoint flag) |
| **Purpose** | Install software | Default args | Main process |

---

## 🌐 CATEGORY 3: Networking

---

### 7. Explain Docker Networking in Depth. What Are the Different Network Drivers?

**Expected Answer:**

```bash
# List all networks
docker network ls

# Inspect a network
docker network inspect bridge
```

**Network Drivers:**

```
┌──────────────────────────────────────────────────────────┐
│                    Docker Networking                      │
│                                                          │
│  ┌─────────┐  ┌─────────┐  ┌────────┐  ┌────────────┐  │
│  │ bridge  │  │  host   │  │  none  │  │  overlay   │  │
│  │(default)│  │         │  │        │  │  (swarm)   │  │
│  └─────────┘  └─────────┘  └────────┘  └────────────┘  │
│                                                          │
│  ┌──────────┐  ┌──────────────────────────────────────┐ │
│  │  macvlan │  │           ipvlan                     │ │
│  │          │  │                                      │ │
│  └──────────┘  └──────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

| Driver | Use Case | Key Characteristic |
|--------|---------|-------------------|
| **bridge** | Default single-host | NAT, container-to-container via IP/name |
| **host** | High performance | Shares host network stack, no isolation |
| **none** | Maximum isolation | No networking at all |
| **overlay** | Multi-host (Swarm/K8s) | Encrypted VXLAN tunnels across hosts |
| **macvlan** | Legacy app integration | Container gets its own MAC address |
| **ipvlan** | High-scale environments | L2/L3 isolation modes |

**User-Defined Bridge vs Default Bridge:**
```bash
# ❌ Default bridge — containers communicate only via IP
docker run -d --name app1 nginx
docker run -d --name app2 nginx
# app2 CANNOT reach app1 by name

# ✅ User-defined bridge — automatic DNS resolution by name
docker network create my-network
docker run -d --name app1 --network my-network nginx
docker run -d --name app2 --network my-network nginx
# app2 CAN reach app1 via: curl http://app1
```

**Container Port Mapping Internals:**
```bash
# Host port 8080 → Container port 80
docker run -p 8080:80 nginx

# Docker uses iptables rules under the hood
iptables -t nat -L DOCKER
```

---

### 8. How Do You Secure Docker Container Networking in Production?

**Expected Answer:**

```bash
# 1. Disable inter-container communication on default bridge
dockerd --icc=false

# 2. Use user-defined networks with explicit connections
docker network create --driver bridge \
  --opt com.docker.network.bridge.enable_icc=false \
  --subnet 172.20.0.0/16 \
  secure-network

# 3. Enable encrypted overlay networks in Swarm
docker network create \
  --driver overlay \
  --opt encrypted \
  --attachable \
  secure-overlay

# 4. Restrict published ports to localhost only
docker run -p 127.0.0.1:8080:80 nginx

# 5. Use network policies via docker-compose
```

```yaml
# docker-compose.yml — Network segmentation
version: "3.9"
services:
  frontend:
    image: nginx
    networks:
      - frontend-net

  backend:
    image: myapp
    networks:
      - frontend-net   # Can talk to frontend
      - backend-net    # Can talk to database

  database:
    image: postgres
    networks:
      - backend-net    # Isolated — no frontend access

networks:
  frontend-net:
    driver: bridge
  backend-net:
    driver: bridge
    internal: true     # No external internet access
```

---

## 💾 CATEGORY 4: Storage & Volumes

---

### 9. What is the Difference Between Volumes, Bind Mounts, and tmpfs? When Do You Use Each?

**Expected Answer:**

```
┌──────────────────────────────────────────────────────┐
│                  Container                           │
│    /app/data ─────────┬──────────────────────────── │
└───────────────────────┼──────────────────────────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
   ┌─────▼──────┐ ┌────▼──────┐ ┌────▼─────┐
   │   Volume   │ │Bind Mount │ │  tmpfs   │
   │/var/lib/   │ │ Any host  │ │  Memory  │
   │docker/vols │ │   path    │ │  only    │
   └────────────┘ └───────────┘ └──────────┘
```

**Volumes:**
```bash
# Create and manage independently of containers
docker volume create mydata

docker run -d \
  --mount type=volume,source=mydata,target=/app/data \
  myapp

# Best for: Databases, persistent app data, sharing between containers
```

**Bind Mounts:**
```bash
# Map host directory directly into container
docker run -d \
  --mount type=bind,source=/host/config,target=/app/config,readonly \
  myapp

# Or shorthand
docker run -v /host/config:/app/config:ro myapp

# Best for: Development hot-reload, injecting config files
```

**tmpfs:**
```bash
# In-memory only — data lost when container stops
docker run -d \
  --mount type=tmpfs,target=/app/temp,tmpfs-size=100m \
  myapp

# Best for: Sensitive data (tokens, secrets), temp processing files
```

| | Volume | Bind Mount | tmpfs |
|--|--------|-----------|-------|
| **Managed by** | Docker | Host OS | Docker |
| **Location** | /var/lib/docker/volumes | Any host path | Memory |
| **Persists** | Yes | Yes | No |
| **Performance** | High | Medium | Highest |
| **Portable** | Yes | No | Yes |
| **Best For** | Production data | Development | Secrets/Temp |

---

## 🔒 CATEGORY 5: Security

---

### 10. What Are the Docker Security Best Practices for Production Deployments?

**Expected Answer:**

**1. Run as Non-Root User:**
```dockerfile
RUN addgroup --system appgroup && \
    adduser --system --ingroup appgroup appuser

USER appuser
```

**2. Use Read-Only Filesystem:**
```bash
docker run --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=100m \
  myapp
```

**3. Drop Linux Capabilities:**
```bash
docker run \
  --cap-drop ALL \
  --cap-add NET_BIND_SERVICE \
  myapp
```

**4. Enable seccomp Profiles:**
```bash
docker run \
  --security-opt seccomp=/path/to/seccomp-profile.json \
  myapp
```

**5. Use AppArmor/SELinux:**
```bash
docker run \
  --security-opt apparmor=docker-default \
  myapp
```

**6. Scan Images for Vulnerabilities:**
```bash
# Using Trivy
trivy image myapp:latest

# Using Docker Scout
docker scout cves myapp:latest

# Using Snyk
snyk container test myapp:latest
```

**7. Limit Resources (Prevent DoS):**
```bash
docker run \
  --memory="512m" \
  --cpus="1.0" \
  --pids-limit=100 \
  --ulimit nofile=1024:1024 \
  myapp
```

**8. Use Docker Content Trust:**
```bash
export DOCKER_CONTENT_TRUST=1
docker pull nginx  # Will only pull signed images
```

**9. Secrets Management:**
```bash
# Never use environment variables for secrets in production

# Docker Swarm Secrets
echo "mysecretpassword" | docker secret create db_password -
docker service create \
  --secret db_password \
  myapp

# Access in container at: /run/secrets/db_password
```

---

### 11. How Do You Prevent Container Breakouts and Privilege Escalation?

**Expected Answer:**

```bash
# 1. Never run privileged containers
docker run --privileged myapp  # ❌ NEVER in production

# 2. Use no-new-privileges flag
docker run --security-opt no-new-privileges:true myapp

# 3. Use rootless Docker mode
# Install and run Docker as non-root user
dockerd-rootless-setuptool.sh install

# 4. Use User Namespaces
# /etc/docker/daemon.json
{
  "userns-remap": "default"
}

# 5. Restrict /proc and /sys
docker run \
  --read-only \
  --tmpfs /proc/keys \
  --security-opt no-new-privileges:true \
  myapp
```

**daemon.json Hardening:**
```json
{
  "icc": false,
  "userns-remap": "default",
  "no-new-privileges": true,
  "live-restore": true,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  },
  "default-ulimits": {
    "nofile": {
      "Name": "nofile",
      "Hard": 64000,
      "Soft": 64000
    }
  }
}
```

---

## ⚙️ CATEGORY 6: Docker Compose & Orchestration

---

### 12. Explain Docker Compose — Advanced Features for Production

**Expected Answer:**

```yaml
version: "3.9"

# Reusable YAML anchors
x-common-env: &common-env
  TZ: UTC
  LOG_LEVEL: info

x-healthcheck: &default-healthcheck
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s

services:
  app:
    image: myapp:${APP_VERSION:-latest}
    build:
      context: .
      dockerfile: Dockerfile
      target: runner           # Multi-stage build target
      cache_from:
        - myapp:latest         # Use previous image as cache
      args:
        BUILD_ENV: production
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: "0.5"
          memory: 512M
        reservations:
          cpus: "0.25"
          memory: 256M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
    environment:
      <<: *common-env
      DATABASE_URL: postgresql://db:5432/mydb
    env_file:
      - .env.production
    secrets:
      - db_password
      - api_key
    networks:
      - frontend
      - backend
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
    healthcheck:
      <<: *default-healthcheck
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
    secrets:
      - db_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    networks:
      - backend
    healthcheck:
      <<: *default-healthcheck
      test: ["CMD-SHELL", "pg_isready -U postgres"]

secrets:
  db_password:
    file: ./secrets/db_password.txt
  api_key:
    external: true            # Managed externally (Vault, AWS SM)

networks:
  frontend:
    driver: bridge
  backend:
    driver: bridge
    internal: true

volumes:
  postgres_data:
    driver: local
    driver_opts:
      type: nfs
      o: addr=10.0.0.1,rw
      device: ":/path/to/nfs"
```

```bash
# Useful Compose commands for production
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d
docker compose config          # Validate and view merged config
docker compose top             # View running processes
docker compose events          # Stream real-time events
```

---

### 13. How Do You Handle Zero-Downtime Deployments with Docker?

**Expected Answer:**

```bash
# Strategy 1: Rolling Update with Docker Swarm
docker service update \
  --image myapp:v2.0 \
  --update-parallelism 1 \
  --update-delay 30s \
  --update-failure-action rollback \
  --rollback-parallelism 1 \
  --health-cmd "curl -f http://localhost/health" \
  --health-interval 10s \
  --health-retries 3 \
  myservice

# Strategy 2: Blue-Green Deployment
# Step 1: Deploy new (green) version alongside old (blue)
docker compose -f docker-compose.green.yml up -d

# Step 2: Health check green environment
curl http://green-app/health

# Step 3: Switch load balancer traffic
# (Update nginx/traefik config)

# Step 4: Stop blue environment
docker compose -f docker-compose.blue.yml down

# Strategy 3: Canary Deployment with Labels
docker service create \
  --name myapp-canary \
  --replicas 1 \
  --label traefik.weight=10 \
  myapp:v2.0

docker service update \
  --replicas 9 \
  myapp-stable
```

---

## 📊 CATEGORY 7: Monitoring & Debugging

---

### 14. How Do You Debug a Docker Container in Production?

**Expected Answer:**

```bash
# 1. View real-time logs with timestamps
docker logs -f --timestamps --tail=100 <container_id>

# 2. Check container resource usage
docker stats --no-stream --format \
  "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"

# 3. Execute commands inside running container
docker exec -it <container_id> /bin/sh

# 4. Inspect container details
docker inspect <container_id> | jq '.[0].State'

# 5. Check container events
docker events --filter container=<container_id>

# 6. Copy files to/from container
docker cp <container_id>:/app/logs/error.log ./local-error.log

# 7. Attach to container (careful — CTRL+C kills container)
docker attach --sig-proxy=false <container_id>

# 8. Debug a stopped container
docker commit <stopped_container_id> debug-image
docker run -it --entrypoint /bin/sh debug-image

# 9. Use nsenter to debug container namespaces from host
PID=$(docker inspect --format '{{.State.Pid}}' <container_id>)
nsenter --target $PID --mount --uts --ipc --net --pid

# 10. Use docker diff to see filesystem changes
docker diff <container_id>
# A = Added, C = Changed, D = Deleted

# 11. Check OOM kills
dmesg | grep -i "out of memory"
docker events --filter event=oom
```

---

### 15. How Do You Monitor Docker Containers at Scale?

**Expected Answer:**

**Full Observability Stack:**

```yaml
# Prometheus + Grafana + cAdvisor Stack
version: "3.9"
services:
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:v0.47.2
    privileged: true
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:ro
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
    ports:
      - "8080:8080"

  prometheus:
    image: prom/prometheus:v2.47.2
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=15d'
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:10.2.0
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./dashboards:/etc/grafana/provisioning/dashboards
    ports:
      - "3000:3000"

  loki:
    image: grafana/loki:2.9.2
    ports:
      - "3100:3100"

  promtail:
    image: grafana/promtail:2.9.2
    volumes:
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - /var/log:/var/log:ro
```

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
    scrape_interval: 15s

  - job_name: 'docker'
    static_configs:
      - targets: ['dockerd:9323']  # Docker daemon metrics
```

**Key Metrics to Monitor:**
```
Container CPU usage      → container_cpu_usage_seconds_total
Container Memory usage   → container_memory_usage_bytes
Container Network I/O    → container_network_transmit_bytes_total
Container Restart count  → container_last_seen (track restarts)
OOM events               → container_oom_events_total
Image pull duration      → engine_daemon_image_pull_duration
```

---

## 🚀 CATEGORY 8: Performance & Advanced Topics

---

### 16. How Do You Optimize Docker for Production Performance?

**Expected Answer:**

**1. Image Pull Optimization:**
```bash
# Use image pull policy wisely
# Pre-pull images on all nodes
docker pull myapp:v2.0

# Use image mirrors for Docker Hub rate limits
# /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://mirror.gcr.io",
    "https://registry-1.docker.io"
  ]
}
```

**2. Build Cache Optimization:**
```bash
# Use BuildKit for parallel builds
export DOCKER_BUILDKIT=1

# Use cache mounts (BuildKit feature)
RUN --mount=type=cache,target=/var/cache/apt \
    apt-get update && apt-get install -y curl

RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline
```

**3. Container Runtime Optimization:**
```bash
# Use host networking for latency-sensitive workloads
docker run --network=host myapp

# Tune kernel parameters for high-traffic containers
docker run \
  --sysctl net.core.somaxconn=65535 \
  --sysctl net.ipv4.tcp_tw_reuse=1 \
  nginx
```

**4. Storage Performance:**
```bash
# Use volume for database — never bind mount
docker volume create --driver local \
  --opt type=tmpfs \
  --opt device=tmpfs \
  --opt o=size=1g,uid=1000 \
  fast_storage

# Use direct-lvm for devicemapper (production)
# overlay2 with SSD is the fastest standard option
```

---

### 17. Explain Docker BuildKit and Its Advanced Features

**Expected Answer:**

```bash
# Enable BuildKit
export DOCKER_BUILDKIT=1
# Or permanently in daemon.json:
{ "features": { "buildkit": true } }
```

```dockerfile
# syntax=docker/dockerfile:1.6

FROM ubuntu:22.04

# 1. Cache Mounts — persist across builds
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y python3

# 2. Secret Mounts — never stored in image layers
RUN --mount=type=secret,id=github_token \
    git clone https://$(cat /run/secrets/github_token)@github.com/org/repo.git

# 3. SSH Mounts — for private Git repos
RUN --mount=type=ssh \
    git clone git@github.com:org/private-repo.git

# 4. Bind Mounts — access host files without copying
RUN --mount=type=bind,source=requirements.txt,target=/tmp/req.txt \
    pip install -r /tmp/req.txt

# 5. Here documents (BuildKit 1.4+)
RUN <<EOF
    set -e
    apt-get update
    apt-get install -y curl wget
    rm -rf /var/lib/apt/lists/*
EOF
```

```bash
# Build with secrets
docker build \
  --secret id=github_token,src=~/.github_token \
  --ssh default=$SSH_AUTH_SOCK \
  --progress=plain \
  -t myapp:latest .

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --push \
  -t myrepo/myapp:latest .
```

---

### 18. How Do You Implement Docker in a CI/CD Pipeline?

**Expected Answer:**

```yaml
# GitHub Actions Pipeline — Production Grade
name: Docker CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
        with:
          driver-opts: |
            image=moby/buildkit:latest

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=semver,pattern={{version}}
            type=sha,prefix=sha-

      - name: Build and push with cache
        uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: ${{ github.event_name != 'pull_request' }}
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            BUILD_DATE=${{ github.event.repository.updated_at }}
            VCS_REF=${{ github.sha }}

      - name: Scan for vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:sha-${{ github.sha }}
          format: sarif
          output: trivy-results.sarif
          severity: CRITICAL,HIGH
          exit-code: 1

      - name: Sign image with Cosign
        uses: sigstore/cosign-installer@v3
      - run: |
          cosign sign --yes \
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:sha-${{ github.sha }}
```

---

### 19. How Do You Manage Docker at Scale — Docker Swarm vs Kubernetes?

**Expected Answer:**

**Docker Swarm Setup:**
```bash
# Initialize Swarm
docker swarm init --advertise-addr <MANAGER-IP>

# Add worker nodes
docker swarm join --token <TOKEN> <MANAGER-IP>:2377

# Deploy a stack
docker stack deploy -c docker-compose.yml mystack

# Scale service
docker service scale mystack_app=5

# Rolling update with health check
docker service update \
  --image myapp:v2.0 \
  --update-parallelism 2 \
  --update-delay 10s \
  --update-failure-action rollback \
  --health-cmd "curl -f http://localhost/health || exit 1" \
  mystack_app

# View service state
docker service ps mystack_app --no-trunc
```

**Comparison:**

| Feature | Docker Swarm | Kubernetes |
|---------|-------------|------------|
| **Setup Complexity** | Simple | Complex |
| **Learning Curve** | Low | High |
| **Auto-Scaling** | Manual | HPA/VPA/KEDA |
| **Self-Healing** | Basic | Advanced |
| **Networking** | Overlay/ingress | CNI plugins |
| **Storage** | Volumes | PV/PVC/StorageClass |
| **Config Mgmt** | Configs/Secrets | ConfigMaps/Secrets |
| **Ecosystem** | Limited | Massive |
| **Production Scale** | Medium (< 100 nodes) | Large (1000s of nodes) |
| **Ideal For** | Small teams, simpler workloads | Enterprise, complex workloads |

---

### 20. You Have a Container That Keeps Crashing in Production. What is Your Debugging Methodology?

**Expected Answer (Systematic Approach):**

```bash
# ─────────────────────────────────────
# STEP 1: Identify the problem
# ─────────────────────────────────────
# Check container status and restart count
docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.ID}}"

# Check events for crash signals
docker events --filter container=<name> --since 1h

# ─────────────────────────────────────
# STEP 2: Examine logs
# ─────────────────────────────────────
# View recent logs with timestamps
docker logs --timestamps --tail=500 <container_id>

# Check previous container logs (before restart)
docker logs --previous <container_id>

# ─────────────────────────────────────
# STEP 3: Check resource constraints
# ─────────────────────────────────────
docker stats <container_id> --no-stream

# Check if OOM killed
docker inspect <container_id> \
  --format '{{.State.OOMKilled}} | ExitCode: {{.State.ExitCode}}'

# Decode exit codes
# Exit 0   = Clean exit
# Exit 1   = App error
# Exit 137 = OOM or SIGKILL (128+9)
# Exit 139 = Segfault (128+11)
# Exit 143 = SIGTERM not handled (128+15)

# ─────────────────────────────────────
# STEP 4: Inspect configuration
# ─────────────────────────────────────
docker inspect <container_id> | jq '.[0] | {
  State: .State,
  Mounts: .Mounts,
  Env: .Config.Env,
  Resources: .HostConfig
}'

# ─────────────────────────────────────
# STEP 5: Check health checks
# ─────────────────────────────────────
docker inspect <container_id> \
  --format '{{json .State.Health}}' | jq .

# ─────────────────────────────────────
# STEP 6: Reproduce with debug image
# ─────────────────────────────────────
# Override entrypoint to get shell
docker run -it \
  --entrypoint /bin/sh \
  --env-file .env.production \
  myapp:latest

# ─────────────────────────────────────
# STEP 7: Check host-level issues
# ─────────────────────────────────────
# Disk space (common issue)
df -h && docker system df

# Inode exhaustion
df -i

# Network connectivity
docker exec -it <container_id> nslookup google.com
docker exec -it <container_id> curl -v telnet://db:5432

# ─────────────────────────────────────
# STEP 8: Cleanup and Recover
# ─────────────────────────────────────
# Free up space if disk is full
docker system prune -af --volumes

# Force remove stuck container
docker rm -f <container_id>

# Restart with increased resources
docker run -d \
  --memory="1g" \
  --restart=on-failure:5 \
  myapp:latest
```

---

> 💡 **Pro Interview Tips for Senior Candidates:**
> - Always mention **tradeoffs** — no silver bullets
> - Talk about **real incidents** you've solved
> - Discuss **security implications** proactively
> - Know when **NOT** to use Docker
> - Show familiarity with **container ecosystem** (containerd, Podman, BuildKit)
> - Mention **compliance** concerns (SOC2, PCI-DSS) when relevant
