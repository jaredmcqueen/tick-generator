package util

import (
	"math/rand"
	"strings"
	"time"
)

const (
	alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomPrice(oldPrice float64) float64 {
	volatility := float64(0.02)
	changePercent := 2 * volatility * rand.Float64()
	if changePercent > volatility {
		changePercent -= (2 * volatility)
	}
	changeAmount := oldPrice * changePercent
	return oldPrice + changeAmount
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
