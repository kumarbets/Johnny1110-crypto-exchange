package utils

import "sync/atomic"

// Process-wide monotonic counters for orders and trades across ALL users. They let
// the UI show real system-level totals and per-second rates (sample each second,
// take the delta). Orders/trades are never deleted, so these only grow. They live
// in the leaf utils package so the order service (writes) and the ws feeder (reads)
// can share them without an import cycle. Seeded from the DB at startup so the
// displayed totals reflect the whole history, not just this process's uptime

var (
	ordersPlacedTotal int64
	tradesTotal       int64
)

func IncOrdersPlaced()       { atomic.AddInt64(&ordersPlacedTotal, 1) }
func AddTrades(n int64)      { atomic.AddInt64(&tradesTotal, n) }
func GetOrdersPlaced() int64 { return atomic.LoadInt64(&ordersPlacedTotal) }
func GetTradesTotal() int64  { return atomic.LoadInt64(&tradesTotal) }

// SetOrdersPlaced / SetTradesTotal seed the counters from the DB at startup.
func SetOrdersPlaced(n int64) { atomic.StoreInt64(&ordersPlacedTotal, n) }
func SetTradesTotal(n int64)  { atomic.StoreInt64(&tradesTotal, n) }

// --- Simulation duration (broadcast over the sysstats WS channel) ---
// "Running" is inferred from live order flow: if orders were placed since the last
// 1-second tick, the sim is running and we accumulate a second. This needs no
// start/stop signalling from the load controller and naturally freezes when the
// generators are stopped. Only SimReset() zeroes it.
var (
	simDurationSecs int64
	simLastOrders   int64
	simRunningFlag  int32
)

// SimTick is called once per second by the ws feeder.
func SimTick() {
	cur := atomic.LoadInt64(&ordersPlacedTotal)
	if cur > atomic.LoadInt64(&simLastOrders) {
		atomic.AddInt64(&simDurationSecs, 1)
		atomic.StoreInt32(&simRunningFlag, 1)
	} else {
		atomic.StoreInt32(&simRunningFlag, 0)
	}
	atomic.StoreInt64(&simLastOrders, cur)
}

func SimDuration() int64 { return atomic.LoadInt64(&simDurationSecs) }
func SimRunning() bool   { return atomic.LoadInt32(&simRunningFlag) == 1 }
func SimReset() {
	atomic.StoreInt64(&simDurationSecs, 0)
	atomic.StoreInt64(&simLastOrders, atomic.LoadInt64(&ordersPlacedTotal))
	atomic.StoreInt32(&simRunningFlag, 0)
}
