package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mailgun/groupcache/v2"
	"github.com/spf13/cast"
)

const (
	CACHE_SIZE = 1 << 20 // 1 MB
	CACHE_NAME = "group_cache"
)

var (
	TIME_TO_LIVE time.Duration = cast.ToDuration("2m") // duration
)

type RateLimit struct {
	Key      string
	Count    int
	EpiredAt time.Time
}

func initiateRateLimit(key string) *RateLimit {
	return &RateLimit{Key: key, Count: 0, EpiredAt: time.Now().Local().Add(TIME_TO_LIVE)}
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
			// if cache missed, then initiate new rate limit
			data := initiateRateLimit(key)
			bs, _ := json.Marshal(data)
			return dest.SetBytes(bs, data.EpiredAt)
		}))

	return cache, pool
}
