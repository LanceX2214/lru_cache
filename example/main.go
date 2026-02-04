package main

import (
	"context"
	"flag"
	"log"
	"strings"
	"time"

	"lru_cache"
	"lru_cache/registry"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	var (
		listen_address    = flag.String("addr", "127.0.0.1:9000", "gRPC listen address")
		service_name      = flag.String("svc", "lcache", "service name")
		peer_addresses    = flag.String("peers", "", "comma-separated peer addresses")
		etcd_endpoints    = flag.String("etcd", "", "comma-separated etcd endpoints")
		get_key           = flag.String("get", "", "optional key to fetch")
		cache_megabytes   = flag.Int64("cache-mb", 64, "cache size in MB")
		expiration_millis = flag.Int64("expire-ms", 0, "default expiration in ms (0 = no expiration)")
	)
	flag.Parse()

	data := map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	getter := lru_cache.GetterFunc(func(key string) ([]byte, error) {
		if value, ok := data[key]; ok {
			log.Printf("[getter] load key=%s", key)
			return []byte(value), nil
		}
		return nil, lru_cache.ErrNotFound
	})

	group := lru_cache.NewGroup(
		"scores",
		(*cache_megabytes)<<20,
		getter,
		lru_cache.WithExpiration(time.Duration(*expiration_millis)*time.Millisecond),
	)

	picker := lru_cache.NewClientPicker(*listen_address)
	group.RegisterPeers(picker)

	if *peer_addresses != "" {
		peers := strings.Split(*peer_addresses, ",")
		picker.SetPeers(peers...)
		log.Printf("[peers] %v", peers)
	} else {
		log.Printf("[peers] none (single node)")
	}

	if *etcd_endpoints != "" {
		endpoints := strings.Split(*etcd_endpoints, ",")
		client, error_value := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 3 * time.Second})
		if error_value != nil {
			log.Fatalf("etcd connect failed: %v", error_value)
		}
		log.Printf("[etcd] endpoints=%v svc=%s", endpoints, *service_name)
		registry_client := registry.New(client)
		request_context, cancel := context.WithCancel(context.Background())
		defer cancel()
		watch_channel := registry_client.Watch(request_context, *service_name)
		go func() {
			for addresses := range watch_channel {
				picker.SetPeers(addresses...)
				log.Printf("[etcd] peers updated: %v", addresses)
			}
		}()
	}

	server := lru_cache.NewServer(*listen_address, *service_name)
	if error_value := server.Start(); error_value != nil {
		log.Fatalf("server start failed: %v", error_value)
	}
	log.Printf("[server] started addr=%s svc=%s", *listen_address, *service_name)

	if *etcd_endpoints != "" {
		if error_value := server.RegisterEtcd(context.Background(), strings.Split(*etcd_endpoints, ","), 10*time.Second); error_value != nil {
			log.Fatalf("etcd register failed: %v", error_value)
		}
		log.Printf("[etcd] registered addr=%s ttl=10s", *listen_address)
	}

	if *get_key != "" {
		value, error_value := group.Get(*get_key)
		if error_value != nil {
			log.Printf("get %s error: %v", *get_key, error_value)
		} else {
			log.Printf("get %s => %s", *get_key, value.String())
		}
	}

	select {}
}
