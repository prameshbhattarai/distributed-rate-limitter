package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mailgun/groupcache/v2"
)

const (
	PORT                = "8081"
	BASE_URL            = "http://localhost:" + PORT
	WINDOW        int64 = 60 // in second
	ALLOWED_LIMIT int   = 4  // number of allowed call
)

func routes(e *echo.Echo, cache *groupcache.Group, pool *groupcache.HTTPPool) {
	e.GET("/", func(c echo.Context) error {
		key := c.QueryParam("key")
		if key == "" {
			return c.String(http.StatusOK, "Server 1 :: Query param not provided")
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
		defer cancel()

		var data []byte
		err := cache.Get(ctx, key, groupcache.AllocatingByteSliceSink(&data))
		if err != nil {
			response := fmt.Sprintf("Server 1 :: Error getting data from cache, %v", err)
			return c.String(http.StatusInternalServerError, response)
		}

		var rateLimit RateLimit
		_ = json.Unmarshal(data, &rateLimit)

		timeWindow := (time.Now().Unix() - rateLimit.StartTime) // time window
		fmt.Printf("Server 1 :: timeWindow %v \n", timeWindow)

		// is limit crossed
		if rateLimit.Count > ALLOWED_LIMIT && timeWindow <= int64(WINDOW) {
			return c.String(http.StatusForbidden, "Server 1 :: Too many request")
		}

		// is timeWindow expired, then reset
		if timeWindow > WINDOW {
			fmt.Printf("Server 1 :: expired %v \n", rateLimit)
			err := cache.Remove(ctx, key)
			if err != nil {
				fmt.Printf("Server 1 :: error while removing cache %v \n", err)
			}
		} else {

			fmt.Printf("Server 1 :: before increment %v \n", rateLimit)

			// increment the counter, and update
			rateLimit.Count += 1
			bs, _ := json.Marshal(rateLimit)

			fmt.Printf("Server 1 :: increment and set %v \n", rateLimit)
			err := cache.Set(ctx, key, bs, time.Now().Add(TIME_TO_LIVE), true)
			if err != nil {
				fmt.Printf("Server 1 :: error while updating cache %v \n", err)
			}
		}
		response := fmt.Sprintf("Server 1 :: Response %v", rateLimit)
		return c.String(http.StatusOK, response)
	})

	// group cache use following path to communicate with other peers
	// so add '/_groupcache/' path in our server
	e.GET("/_groupcache/*path", func(c echo.Context) error {
		pool.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.PUT("/_groupcache/*path", func(c echo.Context) error {
		pool.ServeHTTP(c.Response(), c.Request())
		return nil
	})

	e.DELETE("/_groupcache/*path", func(c echo.Context) error {
		pool.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

func main() {
	cache, pool := InitializeCache(BASE_URL)

	e := echo.New()
	routes(e, cache, pool)

	if err := e.Start(":" + PORT); err != nil {
		log.Fatalf("Error starting Server 1: %v", err)
	}
}
