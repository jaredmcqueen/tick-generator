package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/shopspring/decimal"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())
}

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	log.Println("starting tick generator")

	flag.PrintDefaults()
	natsEndpoint := flag.String("natsEndpoint", "localhost:4222", "nats endpoint")
	symbolCount := flag.Int("symbolCount", 1, "number of symbols to generate")
	startingPrice := flag.Float64("startingPrice", 100.0, "number of ticks to generate")
	volatility := flag.Float64("volatility", 0.05, "price volatility as a float64")
	tickCount := flag.Int("tickCount", 100, "number of ticks to generate.  0 = unlimited")
	tickDelay := flag.Int("tickDelay", 10, "milliseconds between ticks")
	natsSubject := flag.String("natsSubject", "tickgen", "nats subject name")
	natsStream := flag.String("natsStream", "tickgen", "nats stream name")
	flag.Parse()

	flag.VisitAll(func(f *flag.Flag) {
		log.Printf("%s: %s (%s)", f.Name, f.Value, f.DefValue)
	})

	nc, _ := nats.Connect(*natsEndpoint)
	js, _ := nc.JetStream()

	// js.DeleteStream("foo")
	_, err := js.AddStream(&nats.StreamConfig{
		Name:     *natsStream,
		Subjects: []string{*natsSubject},
	})
	if err != nil {
		if err == nats.ErrStreamNameAlreadyInUse {
			log.Println("stream already exists")
		} else {
			log.Fatal(err)
		}
	}

	m := NewMarket(*symbolCount, *startingPrice)
	counter := 0
	sendTicks := true

	start := time.Now()

	log.Println("sending ticks")
	ticker := time.NewTicker(time.Duration(*tickDelay) * time.Millisecond)
	for sendTicks {
		select {
		case <-signalChan:
			return
		case t := <-ticker.C:
			for symbol, tick := range m.SymbolSet {
				m.SymbolSet[symbol] = tick.PerturbPrice(t.UnixNano(), *volatility)
				payload, err := encodeJSON(tick)
				if err != nil {
					log.Fatal(err)
				}
				_, err = js.PublishAsync(*natsSubject, payload)
				if err != nil {
					log.Panic(err)
				}
				counter++
				if *tickCount > 0 && counter >= *tickCount {
					sendTicks = false
				}
			}
		}
	}

	// exit
	ticker.Stop()
	close(signalChan)
	epsSeconds := int(time.Since(start).Seconds())
	if epsSeconds == 0 {
		epsSeconds += 1
	}
	eps := counter / epsSeconds

	log.Printf("sent %d ticks in %v at %v eps", counter, time.Since(start).Round(time.Second), eps)
}

func RandomSymbol() string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < 3; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}
	return sb.String()
}

type tick struct {
	T int64
	S string
	C decimal.Decimal
}

func (t *tick) PerturbPrice(timestamp int64, v float64) *tick {
	change := 2 * v * rand.Float64()
	if change > v {
		change -= (2 * v)
	}
	t.C = t.C.Add(t.C.Mul(decimal.NewFromFloat(change))).Round(4)
	t.T = timestamp
	return t
}

type market struct {
	SymbolSet map[string]*tick
}

func NewMarket(symbolCount int, startingPrice float64) *market {
	ss := make(map[string]*tick, symbolCount)

	for len(ss) < symbolCount {
		symbol := RandomSymbol()
		ss[symbol] = &tick{
			T: time.Now().UnixNano(),
			S: symbol,
			C: decimal.NewFromFloat(startingPrice),
		}
	}

	return &market{
		SymbolSet: ss,
	}
}

func encodeJSON(v *tick) ([]byte, error) {
	var b bytes.Buffer
	err := json.NewEncoder(&b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
