package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	base := flag.String("base", "http://localhost:8080", "")
	market := flag.String("market", "BTC-USDT", "")
	token := flag.String("token", "", "")
	n := flag.Int("n", 5000, "orders")
	conc := flag.Int("c", 32, "concurrency")
	mid := flag.Float64("mid", 65000, "mid price")
	delayms := flag.Int("delayms", 0, "sleep between orders per worker")
	mktPct := flag.Int("mktpct", 25, "percent MARKET orders (rest are LIMIT)")
	band := flag.Int("band", 25, "limit price half-band around mid (wider => more resting price levels)")
	flag.Parse()

	url := *base + "/api/v1/orders/" + *market
	var ok, fail, filled int64
	jobs := make(chan int, *n)
	client := &http.Client{Timeout: 10 * time.Second}
	var wg sync.WaitGroup

	// Price is driven by a shared time-based wave (every generator computes the same value
	// from the wall clock). Critically, the order side is biased toward the wave's SLOPE:
	// when the wave rises we favor aggressive buys, when it falls we favor sells. That
	// directional pressure (a) trends the price via real trades and (b) consumes the
	// lagging side so no stale maker orders are left behind to cross the book.

	worker := func() {
		defer wg.Done()
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for range jobs {
			if *delayms > 0 {
				time.Sleep(time.Duration(*delayms) * time.Millisecond)
			}
			now := float64(time.Now().UnixNano()) / 1e9
			anchor := *mid + 250*math.Sin(now/18.0) + 90*math.Sin(now/5.0)         // the wave (price target)
			slope := (250.0/18.0)*math.Cos(now/18.0) + (90.0/5.0)*math.Cos(now/5.0) // its direction
			// bias side toward the trend: rising => ~62% buys, falling => ~62% sells
			side := 0 // buy
			if slope >= 0 {
				if rng.Float64() >= 0.62 {
					side = 1
				}
			} else {
				if rng.Float64() < 0.62 {
					side = 1
				}
			}
			var body string
			if rng.Intn(100) < *mktPct {
				// MARKET order: buy consumes quote_amount, sell consumes size
				if side == 0 {
					body = fmt.Sprintf(`{"side":0,"order_type":1,"mode":1,"quote_amount":%.2f}`, anchor*0.001)
				} else {
					body = `{"side":1,"order_type":1,"mode":1,"size":0.001}`
				}
			} else {
				// LIMIT order across a band, ALWAYS taker (mode=1). The engine matches a
				// taker against the opposite side BEFORE resting the remainder, so an order
				// priced through the book consumes the crossing side instead of leaving the
				// book crossed; a non-crossing price simply rests -> depth. This is what keeps
				// the book valid (best bid < best ask) while the price trends.
				off := float64(rng.Intn(2*(*band)+1) - *band)
				body = fmt.Sprintf(`{"side":%d,"order_type":0,"mode":1,"price":%.0f,"size":0.001}`, side, anchor+off)
			}
			req, _ := http.NewRequest("POST", url, bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", *token)
			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt64(&fail, 1)
				continue
			}
			buf := make([]byte, 4096)
			m, _ := resp.Body.Read(buf)
			resp.Body.Close()
			s := string(buf[:m])
			if resp.StatusCode == 200 && bytes.Contains([]byte(s), []byte(`"0000000"`)) {
				atomic.AddInt64(&ok, 1)
				if bytes.Contains([]byte(s), []byte(`"matches":[{`)) {
					atomic.AddInt64(&filled, 1)
				}
			} else {
				atomic.AddInt64(&fail, 1)
			}
		}
	}
	for i := 0; i < *conc; i++ {
		wg.Add(1)
		go worker()
	}
	start := time.Now()
	for i := 0; i < *n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	el := time.Since(start).Seconds()
	fmt.Printf("\n===== END-TO-END TRADING THROUGHPUT (%s) =====\n", *market)
	fmt.Printf("orders sent      : %d  (concurrency %d, market%%=%d)\n", *n, *conc, *mktPct)
	fmt.Printf("accepted (2xx)   : %d\n", ok)
	fmt.Printf("failed           : %d\n", fail)
	fmt.Printf("orders that traded: %d\n", filled)
	fmt.Printf("elapsed          : %.2fs\n", el)
	fmt.Printf(">> THROUGHPUT    : %.0f orders/sec\n", float64(ok)/el)
}
