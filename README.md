# lru_cache

A small distributed cache written in Go to explore core system design concepts such as peer routing, request coalescing, and cache eviction strategies.

The project focuses on architectural clarity and minimal dependencies so the system can run locally without external infrastructure while still supporting optional service discovery for more realistic distributed setups.

---

## Overview

`lru_cache` implements a lightweight distributed cache with local LRU / LRU-2 storage and peer-to-peer communication.

Inspired by systems like **groupcache**, the goal is to make the internal mechanics of distributed caching explicit and easy to reason about rather than hiding them behind heavy abstractions.

Key capabilities include:

- Namespace-based caching via groups  
- Consistent hashing for peer routing  
- Duplicate suppression using singleflight  
- Optional etcd-backed service discovery  

This project prioritizes transparency and learnability over production completeness.

---

## Features

- Local in-memory cache with **LRU** and **LRU-2** eviction  
- Group namespace with getter-driven loading  
- Consistent hashing to distribute keys across nodes  
- singleflight to prevent thundering-herd cache misses  
- gRPC peer communication using a JSON codec (no protoc required)  
- Optional TTL expiration  
- Optional etcd service discovery  

---

## Design Choices

- **LRU-2 alongside LRU**  
  Reduces cache pollution under bursty or scan-heavy workloads.

- **Consistent hashing instead of centralized routing**  
  Avoids introducing a coordination bottleneck.

- **JSON over protobuf for gRPC**  
  Keeps the development workflow lightweight and dependency-free.

- **singleflight for request coalescing**  
  Ensures only one load occurs when multiple requests hit the same missing key.

- **Optional etcd integration**  
  Enables realistic service discovery without making external infrastructure mandatory.

---

## Tradeoffs and Limitations

This project favors simplicity and architectural visibility over production hardening.

Current limitations include:

- No cross-node replication  
- Minimal failure handling  
- Limited observability (metrics / tracing not included)  
- No authentication or transport security  
- Not optimized for extreme scale  

The intent is to clearly expose distributed caching mechanics rather than provide a production-ready cache.

---

## Minimal Setup (No External Dependencies)

Run a two-node cache cluster locally:

```bash
go run ./example -addr 127.0.0.1:9001 -peers 127.0.0.1:9001,127.0.0.1:9002
go run ./example -addr 127.0.0.1:9002 -peers 127.0.0.1:9001,127.0.0.1:9002
```

Then issue requests from a client to observe peer routing and cache hits.

This setup is intentionally dependency-free to keep the development feedback loop fast.

---

## Service Discovery (Optional)

For a more realistic distributed environment, etcd can be used for dynamic node discovery.

### Run etcd locally (Docker)

```bash
docker run -d --name etcd \
  -p 2379:2379 -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.12 \
  /usr/local/bin/etcd \
  --name etcd0 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-cluster etcd0=http://0.0.0.0:2380 \
  --initial-cluster-state new \
  --initial-cluster-token etcd-cluster
```

### Start cache nodes with discovery enabled

```bash
go run ./example -addr 127.0.0.1:9001 -etcd 127.0.0.1:2379
go run ./example -addr 127.0.0.1:9002 -etcd 127.0.0.1:2379
```

---

## Docker

Build the image:

```bash
docker build -t lancex2214/lru_cache:latest .
```

---

## Kubernetes Deployment

Apply etcd and cache manifests:

```bash
kubectl apply -f k8s/etcd.yaml
kubectl apply -f k8s/cache.yaml
```

Optional HPA (requires metrics-server):

```bash
kubectl apply -f k8s/cache-hpa.yaml
```

**Notes**

- `k8s/etcd.yaml` provisions a PVC (`etcd-data`) mounted at `/etcd-data`
- Update the image name in `k8s/cache.yaml` if using a different registry

---

## Configuration

| Flag | Description |
|--------|-------------|
| `-peers` | Static peer list for consistent hashing |
| `-etcd` | Enable etcd-based service discovery |
| `-expire-ms` | Cache entry expiration in milliseconds |

---

## Future Work

- Cross-node replication  
- Metrics and tracing  
- Lock contention benchmarking  
- Fault-injection / chaos testing  
- Performance comparison between LRU and LRU-2  

---

## Why This Project Exists

This project started as an effort to better understand the internal mechanics of distributed caching systems.

Rather than hiding complexity behind frameworks, it exposes core ideas directly:

- How requests are routed  
- How cache misses propagate  
- How duplicate work is avoided  
- How nodes discover each other  

The system is intentionally small so these behaviors remain visible.
