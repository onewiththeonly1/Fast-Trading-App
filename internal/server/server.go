package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"fast-trading-app/internal/config"
	"fast-trading-app/internal/logger"
	"fast-trading-app/internal/position"

	"github.com/gorilla/websocket"
)

type Server struct {
	posMgr     *position.Manager
	instrument *config.InstrumentConfig
	logger     *logger.Logger
	wsClients  map[*websocket.Conn]bool
	wsMutex    sync.RWMutex
	upgrader   websocket.Upgrader
}

func New(posMgr *position.Manager, instrument *config.InstrumentConfig, lgr *logger.Logger) *Server {
	return &Server{
		posMgr:     posMgr,
		instrument: instrument,
		logger:     lgr,
		wsClients:  make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin:     func(r *http.Request) bool { return true },
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (s *Server) Start(port string) error {
	http.HandleFunc("/", s.serveHome)
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/api/state", s.handleGetState)
	http.HandleFunc("/api/logs", s.handleGetLogs)

	s.logger.Info("Web server starting on port %s", port)
	fmt.Printf("\n=== Web GUI: http://localhost:%s ===\n", port)

	return http.ListenAndServe(":"+port, nil)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile("web/index.html")
	if err != nil {
		http.Error(w, "Could not load page", http.StatusInternalServerError)
		s.logger.Error("Failed to load index.html: %v", err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(content)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade error: %v", err)
		return
	}

	s.wsMutex.Lock()
	s.wsClients[conn] = true
	clientCount := len(s.wsClients)
	s.wsMutex.Unlock()

	s.logger.Debug("New WebSocket client connected (total: %d)", clientCount)
	
	// Send initial state
	s.sendStateToClient(conn)

	// Start ping/pong handler
	go s.handlePingPong(conn)

	// Read messages (mainly for detecting disconnection)
	go func() {
		defer s.removeClient(conn)
		
		// Set read deadline
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func (s *Server) handlePingPong(conn *websocket.Conn) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		s.wsMutex.RLock()
		active := s.wsClients[conn]
		s.wsMutex.RUnlock()
		
		if !active {
			return
		}
		
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			s.removeClient(conn)
			return
		}
	}
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	
	if s.wsClients[conn] {
		delete(s.wsClients, conn)
		conn.Close()
		s.logger.Debug("WebSocket client disconnected (remaining: %d)", len(s.wsClients))
	}
}

func (s *Server) handleGetState(w http.ResponseWriter, r *http.Request) {
	data := s.getStateData()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Error encoding state: %v", err)
	}
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	logs := s.logger.GetEntries()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs": logs,
	})
}

func (s *Server) getStateData() map[string]interface{} {
	return map[string]interface{}{
		"instrument":   s.instrument,
		"orderHistory": s.posMgr.GetOrderHistory(),
		"position":     s.posMgr.GetPosition(),
		"logs":         s.logger.GetEntries(),
		"lastUpdate":   time.Now(),
	}
}

func (s *Server) BroadcastUpdate() {
	data := s.getStateData()
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Error marshaling update: %v", err)
		return
	}

	s.wsMutex.RLock()
	clients := make([]*websocket.Conn, 0, len(s.wsClients))
	for client := range s.wsClients {
		clients = append(clients, client)
	}
	s.wsMutex.RUnlock()

	// Send to all clients
	for _, client := range clients {
		// Set write deadline
		client.SetWriteDeadline(time.Now().Add(10 * time.Second))
		
		if err := client.WriteMessage(websocket.TextMessage, jsonData); err != nil {
			s.logger.Warn("Error writing to WebSocket: %v", err)
			s.removeClient(client)
		}
	}
}

func (s *Server) sendStateToClient(conn *websocket.Conn) {
	data := s.getStateData()
	jsonData, err := json.Marshal(data)
	if err != nil {
		s.logger.Error("Error marshaling state: %v", err)
		return
	}
	
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		s.logger.Error("Error sending initial state: %v", err)
	}
}

func (s *Server) UpdateInstrument(instrument *config.InstrumentConfig) {
	s.instrument = instrument
	s.logger.Info("Server instrument updated to: %s", instrument.Symbol)
}

// Close all WebSocket connections gracefully
func (s *Server) Shutdown() {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	
	for client := range s.wsClients {
		client.Close()
	}
	s.wsClients = make(map[*websocket.Conn]bool)
	s.logger.Info("Server shutdown complete")
}