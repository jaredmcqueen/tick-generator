package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jaredmcqueen/sherpa/generator/util"
)

var counter int32

// generateTicks uses TS.MADD to send random ticks to redis
func generateTicks(config util.Config) {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: config.RedisEndpoint,
	})
	rdb.FlushAll(ctx)

	// rdb := redis.NewClusterClient(&redis.ClusterOptions{
	// 	Addrs: []string{":4000", ":4001", ":4002", ":4003", ":4004", ":4005"},
	// 	// Username: "bitnami",
	// 	Password: "bitnami",
	// 	// RouteByLatency: true,
	// 	// RouteRandomly:  true,
	// })
	pipe := rdb.Pipeline()

	symbolCount := config.SymbolCount
	symbolSet := make(map[string]float64)

	for len(symbolSet) < symbolCount {
		symbolSet[util.RandomSymbol()] = float64(100)
	}

	log.Println("starting tick generator")
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			start := time.Now()

			// iterate through all the values in the symbolSet
			totalCounter := 0
			myCounter := 0
			for i := 0; i < config.Multiplyer; i++ {
				for k, v := range symbolSet {
					values := map[string]string{
						"symbol": k,
						"close":  strconv.FormatFloat(v, 'f', -1, 64),
					}

					pipe.XAdd(ctx, &redis.XAddArgs{
						Stream: k,
						ID:     "*",
						Values: values,
					})

					// calculate NEXT tick ahead of time
					symbolSet[k] = util.RandomPrice(v)
					atomic.AddInt32(&counter, 1)
					totalCounter++
					myCounter++

					if myCounter > config.PayloadSize-1 {
						pipe.Exec(ctx)
						myCounter = 0
					}
				}
			}
			pipe.Exec(ctx)
			fmt.Printf("%v inserts complete in %v ms\n", totalCounter, time.Since(start).Milliseconds())
			totalCounter = 0

		}
	}

}

func main() {
	// load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load configuration", err)
	}

	// catch control+c
	sigsChan := make(chan os.Signal, 1)
	signal.Notify(sigsChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			log.Println(counter)
			counter = int32(0)
		}
	}()

	for i := 0; i < config.Workers; i++ {
		go generateTicks(config)
	}

	<-sigsChan

	fmt.Print("received termination signal")
	os.Exit(0)
}
