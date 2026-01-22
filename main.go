package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"fast-trading-app/internal/config"
	"fast-trading-app/internal/logger"
	"fast-trading-app/internal/position"
	"fast-trading-app/internal/server"
	"fast-trading-app/internal/trader"

	kiteconnect "github.com/zerodha/gokiteconnect/v4"
	"golang.org/x/term"
)

var (
	cfg        *config.Config
	kc         *kiteconnect.Client
	posMgr     *position.Manager
	trdr       *trader.Trader
	srv        *server.Server
	lgr        *logger.Logger
	instrument config.InstrumentConfig
	
	// Command state machine
	cmdMutex      sync.Mutex
	cmdState      string // "" or "-"
	terminalState *term.State
)

func main() {
	// Ensure terminal is in cooked mode initially
	restoreTerminal()
	
	// Initialize logger first
	var err error
	lgr, err = logger.New("trading.log")
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer lgr.Close()

	lgr.Info("=== Fast Trading Application Starting ===")

	// Load configuration
	cfg, err = config.Load("config.json")
	if err != nil {
		lgr.Error("Failed to load config: %v", err)
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Kite Connect
	kc = kiteconnect.New(cfg.APIKey)
	kc.SetAccessToken(cfg.AccessToken)

	// Verify credentials
	profile, err := kc.GetUserProfile()
	if err != nil {
		lgr.Error("Failed to verify credentials: %v", err)
		log.Fatalf("Failed to verify credentials: %v\nPlease generate a new access token using: go run cmd/generate-token/main.go", err)
	}
	lgr.Info("Logged in as: %s (%s)", profile.UserName, profile.UserID)

	// Initialize managers
	posMgr = position.NewManager()

	// Select instrument
	selectInstrument()

	// Initialize trader and server
	trdr = trader.New(kc, posMgr, instrument, lgr)
	srv = server.New(posMgr, &instrument, lgr)

	// Start web server
	go func() {
		if err := srv.Start("8080"); err != nil {
			lgr.Error("Failed to start web server: %v", err)
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// Start position update loop
	go updatePositionLoop()

	// Setup graceful shutdown
	setupGracefulShutdown()

	// Give web server time to start
	time.Sleep(500 * time.Millisecond)

	lgr.Info("Application ready. Waiting for commands...")
	fmt.Println("\n=== Web GUI: http://localhost:8080 ===")
	fmt.Println("\nReady. Enter commands:")
	fmt.Print("> ")

	// Start command loop
	commandLoop()
}

func selectInstrument() {
	fmt.Println("\n=== Available Instruments ===")
	for i, inst := range cfg.Instruments {
		fmt.Printf("%d. %s (%s) - Lot Size: %d\n", i+1, inst.Symbol, inst.Exchange, inst.LotSize)
	}

	fmt.Print("\nSelect instrument number: ")
	
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	num, err := strconv.Atoi(input[:len(input)-1])
	
	if err != nil || num < 1 || num > len(cfg.Instruments) {
		fmt.Println("Invalid selection. Using first instrument.")
		num = 1
	}

	instrument = cfg.Instruments[num-1]
	lgr.Info("Selected instrument: %s (%s)", instrument.Symbol, instrument.Exchange)
}

func commandLoop() {
	// Put terminal in raw mode for single-character input
	var err error
	terminalState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		lgr.Error("Failed to set terminal to raw mode: %v", err)
		log.Fatalf("Failed to set terminal to raw mode: %v", err)
	}
	defer restoreTerminal()

	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		char := buf[0]

		// Handle Ctrl+C
		if char == 3 {
			lgr.Info("Received Ctrl+C, exiting...")
			cleanup()
			os.Exit(0)
		}

		// Handle Quit
		if char == 'Q' || char == 'q' {
			lgr.Info("Quit command received")
			cleanup()
			os.Exit(0)
		}

		// Handle Change Instrument
		if char == 'C' || char == 'c' {
			handleChangeInstrument()
			continue
		}

		// Handle numeric input and minus
		if (char >= '0' && char <= '9') || char == '-' {
			handleNumericInput(char)
		} else {
			// Invalid character - reset state
			cmdMutex.Lock()
			if cmdState == "-" {
				// Was waiting for digit after minus
				cmdState = ""
			}
			cmdMutex.Unlock()
		}
	}
}

func handleNumericInput(char byte) {
	cmdMutex.Lock()
	defer cmdMutex.Unlock()

	// State machine for command processing
	switch cmdState {
	case "":
		// Initial state
		if char == '-' {
			fmt.Print("-")  // Echo the minus
			cmdState = "-"
		} else if char >= '1' && char <= '9' {
			fmt.Print(string(char))  // Echo the digit
			num := int(char - '0')
			cmdState = ""
			go placeBuyOrder(num)
		} else if char == '0' {
			// Invalid command
			fmt.Print("0")  // Echo but ignore
			cmdState = ""
		}
		
	case "-":
		// Waiting for next character after minus
		if char == '-' {
			fmt.Print("-")  // Echo second minus
			cmdState = ""
			go closeAllPositions()
		} else if char >= '1' && char <= '9' {
			fmt.Print(string(char))  // Echo the digit
			num := int(char - '0')
			cmdState = ""
			go placeSellOrder(num)
		} else if char == '0' {
			// Invalid: -0
			fmt.Print("0")  // Echo but ignore
			cmdState = ""
		} else {
			// Invalid character after minus
			cmdState = ""
		}
	}
}

func handleChangeInstrument() {
	if posMgr.HasOpenPosition() {
		lgr.Warn("Cannot change instrument with open positions")
		fmt.Print("\n⚠️  Cannot change with open positions\n> ")
		srv.BroadcastUpdate()
		return
	}

	// Restore terminal temporarily
	restoreTerminal()
	
	posMgr.Reset()
	lgr.Info("Changing instrument...")
	fmt.Println()
	selectInstrument()
	trdr.UpdateInstrument(instrument)
	srv.UpdateInstrument(&instrument)
	srv.BroadcastUpdate()
	
	// Put terminal back in raw mode
	var err error
	terminalState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		lgr.Error("Failed to restore raw mode: %v", err)
	}
	
	fmt.Println("\nReady. Enter commands:")
	fmt.Print("> ")
}

func placeBuyOrder(lots int) {
	if err := trdr.PlaceOrder("BUY", lots); err != nil {
		lgr.Error("Buy order failed: %v", err)
		// Don't print anything on failure
	} else {
		fmt.Print(" ")  // Space on success
	}
	srv.BroadcastUpdate()
}

func placeSellOrder(lots int) {
	if err := trdr.PlaceOrder("SELL", lots); err != nil {
		lgr.Error("Sell order failed: %v", err)
		// Don't print anything on failure
	} else {
		fmt.Print(" ")  // Space on success
	}
	srv.BroadcastUpdate()
}

func closeAllPositions() {
	lots := posMgr.GetOpenLots()
	if lots <= 0 {
		lgr.Warn("No open positions to close")
		// Don't print anything
		srv.BroadcastUpdate()
		return
	}

	lgr.Info("CLOSE ALL command: closing %d lots", lots)
	if err := trdr.PlaceOrder("SELL", lots); err != nil {
		lgr.Error("Close all failed: %v", err)
		// Don't print anything on failure
	} else {
		fmt.Print(" ")  // Space on success
	}
	srv.BroadcastUpdate()
}

func updatePositionLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if posMgr.GetPosition().QtyUnits > 0 {
			price, err := trdr.FetchCurrentPrice()
			if err != nil {
				lgr.Warn("Failed to fetch current price: %v", err)
				continue
			}
			
			posMgr.UpdateCMP(price)
			srv.BroadcastUpdate()
		}
	}
}

func setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		lgr.Info("Received shutdown signal")
		cleanup()
		os.Exit(0)
	}()
}

func restoreTerminal() {
	if terminalState != nil {
		term.Restore(int(os.Stdin.Fd()), terminalState)
		terminalState = nil
	}
}

func cleanup() {
	lgr.Info("Cleaning up resources...")
	
	// Restore terminal
	restoreTerminal()
	
	// Close logger
	if lgr != nil {
		lgr.Close()
	}
	
	fmt.Println("\nGoodbye!")
}