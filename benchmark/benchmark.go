package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

const (
	addr        = "localhost:9736" // Redis 地址
	addrGoRedis = "localhost:6380" // Go 实现的 Redis 地址
	clients     = 1000             // 并发客户端数量
	requests    = 10000            // 每个客户端的请求数量
)

func benchmark(addr string) {
	var wg sync.WaitGroup
	ctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	start := time.Now()

	for i := 0; i < clients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requests; j++ {
				client.Set(ctx, fmt.Sprintf("key%d", j), "value", 0)
				client.Get(ctx, fmt.Sprintf("key%d", j))
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("Benchmark for %s took %v\n", addr, duration)
}

func main() {
	fmt.Println("Benchmarking Redis")
	benchmark(addr)

	//fmt.Println("Benchmarking Go Redis")
	//benchmark(addrGoRedis)
}
