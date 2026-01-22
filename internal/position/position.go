package position

import (
	"sync"
	"time"
)

type Order struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
}

type Position struct {
	QtyLots          int     `json:"qty_lots"`
	QtyUnits         int     `json:"qty_units"`
	TotalValue       float64 `json:"total_value"`
	AvgPrice         float64 `json:"avg_price"`
	CMP              float64 `json:"cmp"`
	MTM              float64 `json:"mtm"`
	MTMChangePercent float64 `json:"mtm_change_percent"`
}

type Manager struct {
	mu           sync.RWMutex
	position     Position
	orderHistory []Order
	
	// For tracking cost basis
	totalBuyCost   float64
	totalBuyUnits  int
}

func NewManager() *Manager {
	return &Manager{
		orderHistory: make([]Order, 0),
	}
}

func (m *Manager) AddOrder(order Order) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orderHistory = append(m.orderHistory, order)
}

func (m *Manager) UpdatePosition(txnType string, lots int, price float64, lotSize int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	units := lots * lotSize
	value := float64(units) * price

	if txnType == "BUY" {
		// Track cumulative buy cost
		m.totalBuyCost += value
		m.totalBuyUnits += units
		
		// Update position
		m.position.QtyUnits += units
		m.position.QtyLots += lots
		
		// Recalculate average price based on total cost
		if m.totalBuyUnits > 0 {
			m.position.AvgPrice = m.totalBuyCost / float64(m.totalBuyUnits)
		}
		
		// Update total value
		m.position.TotalValue = float64(m.position.QtyUnits) * m.position.AvgPrice
		
	} else { // SELL
		m.position.QtyUnits -= units
		m.position.QtyLots -= lots
		
		// Reduce tracked buy cost proportionally
		if m.totalBuyUnits > 0 {
			costReduction := (float64(units) / float64(m.totalBuyUnits)) * m.totalBuyCost
			m.totalBuyCost -= costReduction
			m.totalBuyUnits -= units
		}
		
		// Reset if position is fully closed
		if m.position.QtyUnits <= 0 {
			m.position.QtyUnits = 0
			m.position.QtyLots = 0
			m.position.TotalValue = 0
			m.position.AvgPrice = 0
			m.position.MTM = 0
			m.position.MTMChangePercent = 0
			m.totalBuyCost = 0
			m.totalBuyUnits = 0
		} else {
			// Recalculate for remaining position
			m.position.TotalValue = float64(m.position.QtyUnits) * m.position.AvgPrice
		}
	}
	
	// Recalculate MTM with current CMP if available
	if m.position.CMP > 0 && m.position.QtyUnits > 0 {
		m.calculateMTM()
	}
}

func (m *Manager) UpdateCMP(price float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.position.CMP = price
	
	if m.position.QtyUnits > 0 {
		m.calculateMTM()
	}
}

// Internal helper - must be called with lock held
func (m *Manager) calculateMTM() {
	if m.position.QtyUnits <= 0 {
		m.position.MTM = 0
		m.position.MTMChangePercent = 0
		return
	}
	
	currentValue := float64(m.position.QtyUnits) * m.position.CMP
	m.position.MTM = currentValue - m.position.TotalValue

	if m.position.TotalValue > 0 {
		m.position.MTMChangePercent = (m.position.MTM / m.position.TotalValue) * 100
	} else {
		m.position.MTMChangePercent = 0
	}
}

func (m *Manager) GetPosition() Position {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy
	return Position{
		QtyLots:          m.position.QtyLots,
		QtyUnits:         m.position.QtyUnits,
		TotalValue:       m.position.TotalValue,
		AvgPrice:         m.position.AvgPrice,
		CMP:              m.position.CMP,
		MTM:              m.position.MTM,
		MTMChangePercent: m.position.MTMChangePercent,
	}
}

func (m *Manager) GetOrderHistory() []Order {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Return a copy
	orders := make([]Order, len(m.orderHistory))
	copy(orders, m.orderHistory)
	return orders
}

func (m *Manager) GetOpenLots() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.position.QtyLots
}

func (m *Manager) HasOpenPosition() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.position.QtyUnits != 0
}

func (m *Manager) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.position = Position{}
	m.orderHistory = make([]Order, 0)
	m.totalBuyCost = 0
	m.totalBuyUnits = 0
}

// Get realized P&L (for closed positions)
func (m *Manager) GetRealizedPnL() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var realizedPnL float64
	var buyValue, sellValue float64
	
	for _, order := range m.orderHistory {
		value := float64(order.Quantity) * order.Price
		if order.Type == "BUY" {
			buyValue += value
		} else {
			sellValue += value
		}
	}
	
	// Only count as realized if position is closed
	if m.position.QtyUnits == 0 {
		realizedPnL = sellValue - buyValue
	}
	
	return realizedPnL
}