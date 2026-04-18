// Package position manages trading position tracking and P&L calculations
package position

import (
	"sync"
	"time"
)

// Order represents a single trading order
type Order struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"` // "BUY" or "SELL"
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	OrderID   string    `json:"order_id"`
	Status    string    `json:"status"`
}

// Position represents the current trading position
type Position struct {
	QtyLots          int     `json:"qty_lots"`          // Number of lots held
	QtyUnits         int     `json:"qty_units"`         // Total units (lots * lot_size)
	TotalValue       float64 `json:"total_value"`       // Total value at average price
	AvgPrice         float64 `json:"avg_price"`         // Average price of position
	CMP              float64 `json:"cmp"`               // Current market price
	MTM              float64 `json:"mtm"`               // Mark-to-market P&L
	MTMChangePercent float64 `json:"mtm_change_percent"` // MTM as percentage
}

// Manager handles position tracking and order history
type Manager struct {
	mu           sync.RWMutex
	position     Position
	orderHistory []Order
	
	// For tracking cost basis (needed for accurate average price calculation)
	totalBuyCost   float64 // Total cost of all buy orders
	totalBuyUnits  int     // Total units from buy orders
	realizedPnL    float64 // Realized profit/loss from closed positions
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

// UpdatePosition updates the position based on a trade execution
func (m *Manager) UpdatePosition(txnType string, lots int, price float64, lotSize int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	units := lots * lotSize
	value := float64(units) * price

	if txnType == "BUY" {
		// Track cumulative buy cost for accurate average price calculation
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
		
		// Calculate realized P&L for this sell
		if m.totalBuyUnits > 0 {
			// Cost basis per unit for the sold shares
			costBasisPerUnit := m.totalBuyCost / float64(m.totalBuyUnits)
			costBasisForSale := costBasisPerUnit * float64(units)
			sellValue := float64(units) * price
			
			// Realized P&L = sell proceeds - cost basis
			m.realizedPnL += sellValue - costBasisForSale
			
			// Reduce tracked buy cost proportionally
			m.totalBuyCost -= costBasisForSale
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
	m.realizedPnL = 0
}

// GetRealizedPnL returns the realized profit/loss from closed positions
func (m *Manager) GetRealizedPnL() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.realizedPnL
}