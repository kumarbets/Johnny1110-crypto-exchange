package utils

import "sync/atomic"

// Scale is a tiny tolerance for floating-point "dust". Prices/sizes are float64,
// and float math leaves specks (0.30000000000000004). Anything ≤ Scale counts as zero.
const Scale = 0.00000001

// Process-wide counters of orders and trades across ALL users. atomic.AddInt64 is a
// lock-free, thread-safe increment (many goroutines can bump it at once safely).
var ordersPlacedTotal int64
var tradesTotal int64

func IncOrdersPlaced()       { atomic.AddInt64(&ordersPlacedTotal, 1) }
func AddTrades(n int64)      { atomic.AddInt64(&tradesTotal, n) }
func GetOrdersPlaced() int64 { return atomic.LoadInt64(&ordersPlacedTotal) }
func GetTradesTotal() int64  { return atomic.LoadInt64(&tradesTotal) }
