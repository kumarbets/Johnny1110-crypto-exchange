package utils

import "sync/atomic"

// Process-wide monotonic counters for orders and trades across ALL users. They let
// the UI show real system-level totals and per-second rates (sample each second,
// take the delta). Orders/trades are never deleted, so these only grow. They live
// in the leaf utils package so the order service (writes) and the ws feeder (reads)
// can share them without an import cycle. Seeded from the DB at startup so the
// displayed totals reflect the whole history, not just this process's uptime.
var (
	ordersPlacedTotal int64
	tradesTotal       int64
)

func IncOrdersPlaced()      { atomic.AddInt64(&ordersPlacedTotal, 1) }
func AddTrades(n int64)     { atomic.AddInt64(&tradesTotal, n) }
func GetOrdersPlaced() int64 { return atomic.LoadInt64(&ordersPlacedTotal) }
func GetTradesTotal() int64  { return atomic.LoadInt64(&tradesTotal) }

// SetOrdersPlaced / SetTradesTotal seed the counters from the DB at startup.
func SetOrdersPlaced(n int64) { atomic.StoreInt64(&ordersPlacedTotal, n) }
func SetTradesTotal(n int64)  { atomic.StoreInt64(&tradesTotal, n) }
