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

var ctx = context.Background()
var rdb *redis.Client

type tick struct {
	Time   time.Time `json:"time"`
	Symbol string    `json:"symbol"`
	Close  float64   `json:"close"`
}

// generateTicks uses TS.MADD to send random ticks to redis
func generateTicks(symbolSet map[string]float64) {
	log.Println("starting tick generator")
	ticker := time.NewTicker(1 * time.Second)

	// TS.MADD requires timeseries keys before use
	// redis commands look like TS.CREATE ticks:abc:close
	for s := range symbolSet {
		rdb.Do(
			ctx,
			"TS.CREATE",
			fmt.Sprintf("ticks:%s:close", s),
			"LABELS",
			"type",
			"tick",
		)
	}

	for {
		select {
		case t := <-ticker.C:
			// send pre-computed values to nats immediately
			var cmds []interface{}
			cmds = append(cmds, "TS.MADD")
			for k, v := range symbolSet {
				tt := tick{
					Time:   t,
					Symbol: k,
					Close:  v,
				}

				cmds = append(cmds,
					fmt.Sprintf("ticks:%s:close", tt.Symbol),
					fmt.Sprintf("%d", tt.Time.UnixMilli()),
					fmt.Sprintf("%f", tt.Close),
				)
			}
			log.Println("sent", len(symbolSet), "ticks")
			rdb.Do(ctx, cmds...)

			// compute the next batch of ticks
			for k, v := range symbolSet {
				symbolSet[k] = util.RandomPrice(v)
			}
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

	log.Println("connecting to redis endpoint", config.RedisEndpoint)
	rdb = redis.NewClient(&redis.Options{
		Addr: config.RedisEndpoint,
	})

	// test redis connection
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatal("error", err)
	}
	log.Println("successfully connected to", config.RedisEndpoint)

	// clear out the db
	// TODO: make this an envar
	rdb.FlushAll(ctx)

	// symbolCount is defined in config
	symbolCount := config.SymbolCount
	symbolSet := make(map[string]float64)

	// fully populate N random symbols
	// 100 is the starting value
	for len(symbolSet) < symbolCount {
		symbolSet[util.RandomSymbol()] = float64(100)
	}

	// start the generator
	go generateTicks(symbolSet)

	<-sigsChan

	fmt.Print("received termination signal")
	os.Exit(0)

}
