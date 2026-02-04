# lru_cache

A lightweight distributed cache in Go with local LRU/LRU-2 storage, consistent hashing, singleflight, and gRPC peer communication.

## Features
- Local cache with LRU / LRU-2 eviction
- Group namespace and getter-based loading
- Consistent hashing for peer selection
- Singleflight to suppress duplicate loads
- gRPC peer API (JSON codec; no protoc required)
- Optional etcd registry for discovery

## Quick start
```bash
go run ./example -addr 127.0.0.1:9001 -peers 127.0.0.1:9001,127.0.0.1:9002
```

## Run etcd locally (Docker)
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

Use etcd for discovery:
```bash
go run ./example -addr 127.0.0.1:9001 -etcd 127.0.0.1:2379
go run ./example -addr 127.0.0.1:9002 -etcd 127.0.0.1:2379
```

## Build Docker image
```bash
docker build -t lancex2214/lru_cache:latest .
```

## K8s Deploy
Apply etcd and cache:
```bash
kubectl apply -f k8s/etcd.yaml
kubectl apply -f k8s/cache.yaml
```

Optional HPA (requires metrics-server):
```bash
kubectl apply -f k8s/cache-hpa.yaml
```

## Notes
- Set `-etcd` to enable service discovery via etcd.
- Cache expiration can be set with `-expire-ms`.
