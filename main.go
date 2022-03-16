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

var ctx = context.Background()
var rdb *redis.Client

type tick struct {
	Time   time.Time `json:"time"`
	Symbol string    `json:"symbol"`
	Close  float64   `json:"close"`
}

var counter int32

// generateTicks uses TS.MADD to send random ticks to redis
func generateTicks() {
	symbolCount := 1000
	symbolSet := make(map[string]float64)

	for len(symbolSet) < symbolCount {
		symbolSet[util.RandomSymbol()] = float64(100)
	}

	log.Println("starting tick generator")
	// ticker := time.NewTicker(1 * time.Second)

	for {
		// select {
		// case t := <-ticker.C:

		// create a redis pipeline to reduce network traffic
		// pipe := rdb.Pipeline()
		// var values map[string]string

		// iterate through all the values in the symbolSet
		for k, v := range symbolSet {
			// enqueue an XADD operation on the pipeline

			// log.Println(k, v, t)
			// values["t"] = t
			// values["symbol"] = k
			// values["close"] = strconv.FormatFloat(v, 'f', -1, 64)
			values := map[string]string{
				"uno":    "dos",
				"symbol": k,
				"close":  strconv.FormatFloat(v, 'f', -1, 64),
			}

			// log.Println(k, v)
			_ = values

			// pipe.XAdd(ctx, &redis.XAddArgs{
			// 	Stream: "ticks",
			// 	ID:     "*",
			// 	Values: values,
			// })
			symbolSet[k] = util.RandomPrice(v)
			atomic.AddInt32(&counter, 1)
		}

		// log.Println((t.UnixMilli()))
		// start := time.Now()
		// pipe.Exec(ctx)
		// log.Printf("XADD pipeline for %v ticks took %v\n", len(symbolSet), time.Since(start))
	}

}

func main() {
	// load config
	// config, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Fatal("cannot load configuration", err)
	// }

	// catch control+c
	sigsChan := make(chan os.Signal, 1)
	signal.Notify(sigsChan, syscall.SIGINT, syscall.SIGTERM)

	// log.Println("connecting to redis endpoint", config.RedisEndpoint)
	// rdb = redis.NewClient(&redis.Options{
	// 	Addr: config.RedisEndpoint,
	// })

	// test redis connection
	// _, err = rdb.Ping(ctx).Result()
	// if err != nil {
	// 	log.Fatal("error", err)
	// }
	// log.Println("successfully connected to", config.RedisEndpoint)

	// clear out the db
	// TODO: make this an envar
	// rdb.FlushAll(ctx)

	// fully populate N random symbols
	// 100 is the starting value

	go func() {
		for {
			time.Sleep(1 * time.Second)
			log.Println(counter)
			counter = int32(0)
		}
	}()

	// start the generator
	for i := 0; i < 1; i++ {
		go generateTicks()
	}

	<-sigsChan

	fmt.Print("received termination signal")
	os.Exit(0)
}
