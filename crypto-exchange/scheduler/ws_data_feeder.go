package scheduler

import (
	"context"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/ohlcv"
	"github.com/johnny1110/crypto-exchange/security"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/utils"
	"github.com/johnny1110/crypto-exchange/ws"
	"github.com/labstack/gommon/log"
	"time"
)

type WSDataFeederJob struct {
	wsHub             *ws.Hub
	ohlcvAgg          *ohlcv.OHLCVAggregator
	orderbookService  service.IOrderBookService
	marketDataService service.IMarketDataService
	credentialCache   *security.CredentialCache
	orderService      service.IOrderService
	balanceService    service.IBalanceService
}

func NewWSDataFeederJob(
	wsHub *ws.Hub,
	ohlcvAgg *ohlcv.OHLCVAggregator,
	orderbookService service.IOrderBookService,
	marketDataService service.IMarketDataService,
	credentialCache *security.CredentialCache,
	orderService service.IOrderService,
	balanceService service.IBalanceService) Scheduler {
	return &WSDataFeederJob{
		wsHub:             wsHub,
		ohlcvAgg:          ohlcvAgg,
		orderbookService:  orderbookService,
		marketDataService: marketDataService,
		credentialCache:   credentialCache,
		orderService:      orderService,
		balanceService:    balanceService,
	}
}

func (W *WSDataFeederJob) Start() error {
	log.Infof("[WSDataFeederJob] start to feeding data to ws.")
	go func() {

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				keys := W.wsHub.GetSubscriptionKeys()
				for _, key := range keys {
					go W.collectAndSend(key)
				}
			}
		}

	}()

	return nil
}

func (W *WSDataFeederJob) collectAndSend(key ws.SubscriptionKey) {
	ctx := context.Background()

	switch key.Channel {
	case ws.MARKETS:
		data, err := W.marketDataService.GetAllMarketData()
		if err != nil {
			log.Warnf("[WSDataFeederJob] get all markets error: %v", err)
			return
		}
		W.wsHub.BroadcastToSubscribers(key, data)

	case ws.OHLCV:
		ohlcvWsParam := key.Params.(ws.OHLCVReqParams)
		symbol := ohlcvWsParam.Symbol
		interval := ohlcv.OHLCV_INTERVAL(ohlcvWsParam.Interval)

		realtimeOHLCV, err := W.ohlcvAgg.GetRealtimeOHLCVData(ctx, symbol, interval)
		if err != nil {
			log.Warnf("[WSDataFeederJob] get realtime ohlcv error: %v", err)
			return
		}
		W.wsHub.BroadcastToSubscribers(key, realtimeOHLCV)
		return
	case ws.ORDERBOOK:
		obWsParam := key.Params.(ws.OrderBookReqParams)
		market := obWsParam.Market
		snapshot, err := W.orderbookService.GetSnapshot(ctx, market)
		if err != nil {
			log.Warnf("[WSDataFeederJob] get orderbook snapshot error: %v", err)
			return
		}
		W.wsHub.BroadcastToSubscribers(key, snapshot)
		return
	case ws.USER_DATA:
		p, ok := key.Params.(ws.PrivateReqParams)
		if !ok {
			return
		}
		user, err := W.credentialCache.Get(p.Token)
		if err != nil || user == nil {
			return // not logged in / expired token
		}
		openResp, err := W.orderService.PaginationQuery(ctx, &dto.GetOrdersQueryReq{
			UserID: user.ID, Market: p.Market, Type: dto.OPENING_ORDER, PageSize: 10, CurrentPage: 1,
		})
		if err != nil {
			return
		}
		closedResp, err := W.orderService.PaginationQuery(ctx, &dto.GetOrdersQueryReq{
			UserID: user.ID, Market: p.Market, Type: dto.CLOSED_ORDER, PageSize: 10, CurrentPage: 1,
		})
		if err != nil {
			return
		}
		balances, _ := W.balanceService.GetBalances(ctx, user.ID)
		W.wsHub.BroadcastToSubscribers(key, map[string]interface{}{
			"open_orders":         openResp.Result,
			"open_total":          openResp.Total,
			"closed_orders":       closedResp.Result,
			"closed_total":        closedResp.Total,
			"balances":            balances,
			"system_orders_total": utils.GetOrdersPlaced(), // monotonic, all users -> live orders/sec + total
			"system_trades_total": utils.GetTradesTotal(),  // monotonic, all users -> live trades/sec + total
		})
		return
	default:
		return
	}
}

func (W *WSDataFeederJob) Stop() error {
	//TODO implement me
	panic("implement me")
}

func (W *WSDataFeederJob) Name() string {
	return "WSDataFeederJob"
}

func (W *WSDataFeederJob) RunTimes() int64 {
	return 0
}
