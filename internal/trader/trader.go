package trader

import (
	"fmt"
	"math"
	"time"

	"fast-trading-app/internal/config"
	"fast-trading-app/internal/logger"
	"fast-trading-app/internal/position"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
)

type Trader struct {
	kc         *kiteconnect.Client
	posMgr     *position.Manager
	instrument config.InstrumentConfig
	logger     *logger.Logger
	lastOrder  time.Time
}

const (
	maxRetries       = 3
	retryDelay       = time.Second
	orderRateLimit   = 100 * time.Millisecond // 10 orders/second max
	freezeQtyNFO     = 1800 // NFO freeze quantity (varies by instrument)
)

func New(kc *kiteconnect.Client, posMgr *position.Manager, instrument config.InstrumentConfig, logger *logger.Logger) *Trader {
	return &Trader{
		kc:         kc,
		posMgr:     posMgr,
		instrument: instrument,
		logger:     logger,
		lastOrder:  time.Now(),
	}
}

func (t *Trader) PlaceOrder(txnType string, lots int) error {
	// Validate market hours (optional - remove if trading outside regular hours)
	if !t.isMarketHours() {
		t.logger.Warn("Order attempted outside market hours")
		return fmt.Errorf("market is closed")
	}

	// Rate limiting
	if time.Since(t.lastOrder) < orderRateLimit {
		time.Sleep(orderRateLimit - time.Since(t.lastOrder))
	}
	t.lastOrder = time.Now()

	quantity := lots * t.instrument.LotSize

	// Validate freeze quantity
	if quantity > freezeQtyNFO && t.instrument.Exchange == "NFO" {
		t.logger.Warn("Quantity %d exceeds freeze limit %d", quantity, freezeQtyNFO)
		return fmt.Errorf("quantity exceeds freeze limit")
	}

	t.logger.Info("Placing %s order for %d lots (%d units) of %s",
		txnType, lots, quantity, t.instrument.Symbol)

	// Place order with retry logic
	orderResp, err := t.placeOrderWithRetry(txnType, quantity)
	if err != nil {
		return err
	}

	t.logger.Info("Order placed successfully! Order ID: %s", orderResp.OrderID)

	// Fetch and update order details
	if err := t.updateOrderDetails(orderResp.OrderID, txnType, lots); err != nil {
		t.logger.Warn("Could not fetch order details: %v", err)
		// Don't return error - order was placed successfully
	}

	return nil
}

func (t *Trader) placeOrderWithRetry(txnType string, quantity int) (kiteconnect.OrderResponse, error) {
	var orderResp kiteconnect.OrderResponse
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		orderParams := kiteconnect.OrderParams{
			Exchange:        t.instrument.Exchange,
			Tradingsymbol:   t.instrument.Symbol,
			TransactionType: txnType,
			Quantity:        quantity,
			Product:         kiteconnect.ProductMIS,
			OrderType:       kiteconnect.OrderTypeMarket,
			Validity:        kiteconnect.ValidityDay,
		}

		resp, err := t.kc.PlaceOrder(kiteconnect.VarietyRegular, orderParams)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		t.logger.Warn("Order attempt %d/%d failed: %v", attempt, maxRetries, err)
		
		if attempt < maxRetries {
			// Exponential backoff
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * retryDelay
			time.Sleep(backoff)
		}
	}

	return orderResp, fmt.Errorf("order failed after %d attempts: %w", maxRetries, lastErr)
}

func (t *Trader) updateOrderDetails(orderID, txnType string, lots int) error {
	// Wait for order to process
	time.Sleep(500 * time.Millisecond)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		orders, err := t.kc.GetOrderHistory(orderID)
		if err != nil {
			t.logger.Debug("Fetch order attempt %d failed: %v", attempt, err)
			time.Sleep(retryDelay)
			continue
		}

		if len(orders) == 0 {
			continue
		}

		lastOrder := orders[len(orders)-1]

		// Verify order status
		if lastOrder.Status == kiteconnect.OrderStatusRejected {
			t.logger.Error("Order rejected: %s", lastOrder.StatusMessage)
			return fmt.Errorf("order rejected: %s", lastOrder.StatusMessage)
		}

		if lastOrder.Status == kiteconnect.OrderStatusCancelled {
			t.logger.Error("Order cancelled: %s", lastOrder.StatusMessage)
			return fmt.Errorf("order cancelled")
		}

		// Only update position if order is complete or partially filled
		if lastOrder.FilledQuantity > 0 {
			avgPrice := lastOrder.AveragePrice
			filledLots := int(lastOrder.FilledQuantity) / t.instrument.LotSize

			order := position.Order{
				Timestamp: time.Now(),
				Type:      txnType,
				Quantity:  filledLots,
				Price:     avgPrice,
				OrderID:   orderID,
				Status:    lastOrder.Status,
			}

			t.posMgr.AddOrder(order)
			t.posMgr.UpdatePosition(txnType, filledLots, avgPrice, t.instrument.LotSize)

			t.logger.Info("Order executed: %d lots @ ₹%.2f (Status: %s)",
				filledLots, avgPrice, lastOrder.Status)
			return nil
		}

		if lastOrder.Status == kiteconnect.OrderStatusComplete {
			return nil
		}

		time.Sleep(retryDelay)
	}

	return fmt.Errorf("could not verify order execution")
}

func (t *Trader) FetchCurrentPrice() (float64, error) {
	key := fmt.Sprintf("%s:%s", t.instrument.Exchange, t.instrument.Symbol)

	quote, err := t.kc.GetLTP(key)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch LTP: %w", err)
	}

	if data, ok := quote[key]; ok {
		return data.LastPrice, nil
	}

	return 0, fmt.Errorf("no price data available for %s", key)
}

func (t *Trader) UpdateInstrument(instrument config.InstrumentConfig) {
	t.instrument = instrument
	t.logger.Info("Instrument updated to: %s (%s)", instrument.Symbol, instrument.Exchange)
}

func (t *Trader) isMarketHours() bool {
	now := time.Now()
	
	// Check if weekend
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return false
	}

	// Market hours: 9:15 AM to 3:30 PM IST
	loc, _ := time.LoadLocation("Asia/Kolkata")
	nowIST := now.In(loc)
	
	hour := nowIST.Hour()
	minute := nowIST.Minute()
	
	// Before 9:15 AM
	if hour < 9 || (hour == 9 && minute < 15) {
		return false
	}
	
	// After 3:30 PM
	if hour > 15 || (hour == 15 && minute > 30) {
		return false
	}
	
	return true
}