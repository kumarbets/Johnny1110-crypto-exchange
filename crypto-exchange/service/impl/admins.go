package serviceImpl

import (
	"context"
	"database/sql"
	"errors"
	"github.com/johnny1110/crypto-exchange/dto"
	"github.com/johnny1110/crypto-exchange/engine-v2/core"
	"github.com/johnny1110/crypto-exchange/engine-v2/model"
	"github.com/johnny1110/crypto-exchange/repository"
	"github.com/johnny1110/crypto-exchange/service"
	"github.com/johnny1110/crypto-exchange/utils"
)

type adminService struct {
	db           *sql.DB
	userRepo     repository.IUserRepository
	balanceRepo  repository.IBalanceRepository
	orderService service.IOrderService
	engine       *core.MatchingEngine
}

func NewIAdminService(db *sql.DB,
	userRepo repository.IUserRepository,
	balanceRepo repository.IBalanceRepository,
	orderService service.IOrderService,
	engine *core.MatchingEngine) service.IAdminService {
	return &adminService{
		db:           db,
		userRepo:     userRepo,
		balanceRepo:  balanceRepo,
		orderService: orderService,
		engine:       engine,
	}
}

// ResetExchange wipes all trading state (orders, trades, ohlcv), zeroes balances,
// re-funds the demo users, resets the in-memory engine + counters. Ready for a
// fresh Start. Live: no restart, so tokens and WebSocket connections stay valid.
func (as adminService) ResetExchange(ctx context.Context) error {
	err := WithTx(ctx, as.db, func(tx *sql.Tx) error {
		for _, stmt := range []string{
			`DELETE FROM orders`,
			`DELETE FROM trades`,
			`UPDATE balances SET available = 0, locked = 0`,
			`UPDATE balances SET available = 1000000      WHERE asset = 'BTC'  AND user_id IN (SELECT id FROM users WHERE username LIKE 'user%@gmail.com')`,
			`UPDATE balances SET available = 100000000000 WHERE asset = 'USDT' AND user_id IN (SELECT id FROM users WHERE username LIKE 'user%@gmail.com')`,
		} {
			if _, err := tx.Exec(stmt); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	// best-effort clear candlestick tables (names vary; ignore missing ones)
	for _, t := range []string{"ohlcv_1min", "ohlcv_15min", "ohlcv_1h", "ohlcv_1d", "ohlcv_1w", "ohlcv_realtime", "ohlcv_statistics"} {
		as.db.Exec("DELETE FROM " + t)
	}
	as.engine.Reset()        // empty the in-memory order books
	utils.SetOrdersPlaced(0)  // reset system counters
	utils.SetTradesTotal(0)
	return nil
}

func (as adminService) Settlement(ctx context.Context, req dto.SettlementReq) error {
	err := WithTx(ctx, as.db, func(tx *sql.Tx) error {
		user, err := as.userRepo.GetUserByUsername(ctx, tx, req.Username)
		if err != nil {
			return err
		}
		if user == nil {
			return errors.New("user not found by username")
		}

		err = as.balanceRepo.ModifyAvailableByUserIdAndAsset(ctx, tx, user.ID, req.Asset, true, req.Amount)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (as adminService) TestAutoMakeMarket(ctx context.Context) error {
	// make some testing maker
	user := &dto.User{
		Username: "market_maker",
		ID:       "1",
		VipLevel: 7,
		MakerFee: 0.0001,
		TakerFee: 0.002,
	}
	market := "ETH-USDT"
	// make 5 bid orders
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3000,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2900,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2800,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2700,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.BID,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     2600,
		Size:      10,
	})

	// make 5 ask orders
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3500,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3400,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3300,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3200,
		Size:      10,
	})
	_, _ = as.orderService.PlaceOrder(ctx, market, user, &dto.OrderReq{
		Side:      model.ASK,
		OrderType: model.LIMIT,
		Mode:      model.MAKER,
		Price:     3100,
		Size:      10,
	})

	return nil
}
