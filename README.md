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

## Notes
- Set `-etcd` to enable service discovery via etcd.
- Cache expiration can be set with `-expire-ms`.
