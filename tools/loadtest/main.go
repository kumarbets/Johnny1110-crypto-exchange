package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
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

	// Shared drifting mid: every generator polls the live traded price and anchors its
	// orders to it. Because trades move the price and everyone follows it, the whole book
	// walks up/down together -> heavy, visible movement (not a book pinned to a fixed mid).
	var curMid int64
	atomic.StoreInt64(&curMid, int64(*mid))
	go func() {
		cl := &http.Client{Timeout: 3 * time.Second}
		snapURL := *base + "/api/v1/orderbooks/" + *market + "/snapshot"
		for {
			time.Sleep(300 * time.Millisecond)
			resp, err := cl.Get(snapURL)
			if err != nil {
				continue
			}
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			s := string(b)
			if i := strings.Index(s, `"latest_price":`); i >= 0 {
				var lp float64
				fmt.Sscanf(s[i+len(`"latest_price":`):], "%f", &lp)
				if lp > 1000 {
					atomic.StoreInt64(&curMid, int64(lp))
				}
			}
		}
	}()

	worker := func() {
		defer wg.Done()
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		for range jobs {
			if *delayms > 0 {
				time.Sleep(time.Duration(*delayms) * time.Millisecond)
			}
			side := rng.Intn(2) // 0 buy, 1 sell
			anchor := float64(atomic.LoadInt64(&curMid))
			if anchor < 1000 {
				anchor = *mid // fallback (e.g. right after a reset, before any trade)
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
				// LIMIT order across a band: outer prices rest (depth = many levels),
				// near-mid crossing prices trade.
				off := float64(rng.Intn(2*(*band)+1) - *band)
				mode := 0
				if (side == 0 && off >= 1) || (side == 1 && off <= -1) {
					mode = 1 // aggressive -> taker
				}
				body = fmt.Sprintf(`{"side":%d,"order_type":0,"mode":%d,"price":%.0f,"size":0.001}`, side, mode, anchor+off)
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
