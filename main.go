package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jaredmcqueen/sherpa/generator/util"
)

// generateTicks uses TS.MADD to send random ticks to redis
func generateTicks() {
	ctx := context.Background()
	fmt.Printf("connecting to %v\n", util.Config.RedisEndpoint)
	rdb := redis.NewClient(&redis.Options{
		Addr: util.Config.RedisEndpoint,
	})
	r, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("%v successfully connected\n", r)

	fmt.Println("clearing out DB")
	rdb.FlushAll(ctx)

	pipe := rdb.Pipeline()

	symbolCount := util.Config.SymbolCount
	symbolSet := make(map[string]float64)

	// symbols are randomly generated
	// all symbol values start at $100.00
	for len(symbolSet) < symbolCount {
		symbolSet[util.RandomSymbol()] = float64(100)
	}

	log.Println("starting tick generator")
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case t := <-ticker.C:
			for k, v := range symbolSet {
				values := map[string]interface{}{
					"t": t.UnixMilli(),
					"T": "quotes",
					"S": k,
					"h": fmt.Sprintf("%.2f", v),
					"l": fmt.Sprintf("%.2f", v),
				}
				pipe.XAdd(ctx, &redis.XAddArgs{
					Stream: "quotes",
					ID:     "*",
					Values: values,
				})

				// calculate NEXT tick ahead of time
				symbolSet[k] = util.RandomPrice(v)
			}
			pipe.Exec(ctx)
		}
	}
}

func main() {
	// catch control+c
	sigsChan := make(chan os.Signal, 1)
	signal.Notify(sigsChan, syscall.SIGINT, syscall.SIGTERM)

	go generateTicks()

	<-sigsChan

	fmt.Print("received termination signal")
	os.Exit(0)
}
