package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mailgun/groupcache/v2"
)

const (
	CACHE_SIZE   = 1 << 20 // 1 MB
	CACHE_NAME   = "group_cache"
	TIME_TO_LIVE = 2 * time.Minute
)

type RateLimit struct {
	Key       string
	Count     int
	StartTime int64
}

func fetchFromDatabase(key string) *RateLimit {
	return &RateLimit{Key: key, Count: 0, StartTime: time.Now().Unix()}
}

var (
	SERVERS = []string{"http://localhost:8081", "http://localhost:8082", "http://localhost:8083", "http://localhost:8084"}
)

func InitializeCache(self string) (*groupcache.Group, *groupcache.HTTPPool) {
	// initialize the peers list with the addresses of all servers
	pool := groupcache.NewHTTPPool(self)
	pool.Set(SERVERS...) // set other peer servers

	cache := groupcache.NewGroup(CACHE_NAME, CACHE_SIZE, groupcache.GetterFunc(
		func(ctx context.Context, key string, dest groupcache.Sink) error {
			data := fetchFromDatabase(key)
			bs, _ := json.Marshal(data)
			return dest.SetBytes(bs, time.Now().Add(TIME_TO_LIVE))
		}))

	return cache, pool
}
