# Complete Setup Guide - Fast Options Trading

## 🎯 What's New in This Version

### ⚡ **Instant Character Input**
- No Enter key needed - press `2` and order executes immediately
- Press `-` then `1` for sell order
- Press `-` twice quickly for close all positions

### 📝 **Silent Console with Comprehensive Logging**
- Console stays clean after instrument selection
- All activity logged to `trading.log` file
- Live log viewer in web GUI (collapsible)
- Log levels: INFO, WARN, ERROR, DEBUG

### 🆓 **Personal Free Tier Compatible**
- ✅ Order placement and management
- ✅ Order history and position tracking
- ❌ No LTP/WebSocket streaming (not available)
- Uses completed order prices for position updates

### 📈 **Optimized for NIFTY/BANKNIFTY Options**
- Pre-configured with popular strike prices
- Lot sizes: NIFTY=25, BANKNIFTY=15
- Weekly and monthly expiries supported

---

## 📁 Project Structure

```
fast-trading-app/
├── main.go
├── config.json
├── trading.log (created automatically)
├── cmd/
│   └── generate-token/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── logger/
│   │   └── logger.go          # NEW: Logging system
│   ├── position/
│   │   └── position.go
│   ├── server/
│   │   └── server.go
│   ├── terminal/
│   │   └── terminal.go         # NEW: Raw terminal input
│   └── trader/
│       └── trader.go
├── web/
│   └── index.html
├── go.mod
└── go.sum
```

---

## 🚀 Quick Start

### Step 1: Create Project

```bash
# Create directories
mkdir -p ~/fast-trading-app/{cmd/generate-token,internal/{config,logger,position,server,terminal,trader},web}
cd ~/fast-trading-app

# Initialize Go module
go mod init fast-trading-app
```

### Step 2: Install Dependencies

```bash
go get github.com/zerodha/gokiteconnect/v4
go get github.com/gorilla/websocket@v1.5.1
go get golang.org/x/sys/unix
go mod tidy
```

### Step 3: Create All Files

Copy the content from these artifacts:
1. `main.go` → Updated main.go
2. `config.json` → Options config (update symbols for current expiry)
3. `internal/config/config.go` → (same as before)
4. `internal/logger/logger.go` → NEW: Logging system
5. `internal/position/position.go` → (same as before)
6. `internal/server/server.go` → Updated with logs
7. `internal/terminal/terminal.go` → NEW: Character input
8. `internal/trader/trader.go` → Updated for Free tier
9. `web/index.html` → Updated with log viewer
10. `cmd/generate-token/main.go` → (same as before)

### Step 4: Get API Credentials

1. Visit https://developers.kite.trade/
2. Create app with **Personal Free** tier
3. Note your `api_key` and `api_secret`
4. Update `config.json`

### Step 5: Generate Access Token

```bash
go run cmd/generate-token/main.go
```

Follow the prompts and update `config.json` with the access token.

### Step 6: Update Option Symbols

**IMPORTANT:** Update instrument symbols in `config.json` with current week's strikes.

To find current symbols:
```bash
# Download instruments list
curl "https://api.kite.trade/instruments/NFO" \
  -H "X-Kite-Version: 3" \
  -H "Authorization: token YOUR_API_KEY:YOUR_ACCESS_TOKEN" \
  > nfo_instruments.csv

# Search for current week NIFTY options
grep "NIFTY.*CE" nfo_instruments.csv | grep "2025-01-30"

# Search for current week BANKNIFTY options
grep "BANKNIFTY.*CE" nfo_instruments.csv | grep "2025-01-29"
```

Example symbols format:
- `NIFTY25JAN24500CE` = NIFTY 24500 Call expiring Jan 30, 2025
- `BANKNIFTY25JAN53000PE` = BANKNIFTY 53000 Put expiring Jan 29, 2025

### Step 7: Run the Application

```bash
cd ~/fast-trading-app
go run main.go
```

**You'll see:**
```
=== Fast Trading Application Starting ===

=== Available Instruments ===
1. NIFTY25JAN25000CE (NFO) - Lot Size: 25
2. NIFTY25JAN24800PE (NFO) - Lot Size: 25
3. BANKNIFTY25JAN54000CE (NFO) - Lot Size: 15
...

Select instrument number: 1

Selected: NIFTY25JAN25000CE (NFO)

=== Console is now silent. All activity logged to web GUI ===
Commands: <NUM> buy | -<NUM> sell | -- close | C change | Q quit

```

**After this, console goes silent. Open browser to:**
```
http://localhost:8080
```

---

## 🎮 Command Reference

### Terminal Commands (No Enter Needed!)

| Key Press | Action |
|-----------|--------|
| `1` | Buy 1 lot |
| `2` | Buy 2 lots |
| `3` | Buy 3 lots |
| `-1` | Sell 1 lot |
| `-2` | Sell 2 lots |
| `--` | Close ALL positions |
| `C` | Change instrument |
| `Q` | Quit application |

**How it works:**
- Press single digit (1-9) → Instant BUY order
- Press `-` then digit → SELL order
- Press `-` twice quickly → Close all positions
- Console stays SILENT, check web GUI for results

---

## 🌐 Web Dashboard Features

### Position Panel
- **Lots**: Number of option lots held
- **Units**: Total contracts (lots × lot size)
- **Total Value**: Investment amount
- **Avg Price**: Average entry price
- **Last Price**: Last traded price (from your orders)
- **MTM P&L**: Mark-to-market profit/loss with %

### Order History
- Chronological list of executed trades
- Buy/Sell indicators with color coding
- Execution price and timestamp

### Activity Log (Collapsible)
- Click "Show Logs" to expand
- Real-time activity stream
- Color-coded log levels:
  - 🔵 INFO: Normal operations
  - 🟠 WARN: Warnings
  - 🔴 ERROR: Errors
  - ⚫ DEBUG: Debug info
- Auto-scrolls to latest entry

---

## 📊 Personal Free Tier Limitations

### ✅ What Works
- ✅ Place orders (Market, Limit)
- ✅ View order history
- ✅ View positions
- ✅ Modify/Cancel orders
- ✅ Get margins/funds info
- ✅ Position tracking via order history

### ❌ What Doesn't Work
- ❌ Live market quotes (LTP/OHLC)
- ❌ WebSocket streaming
- ❌ Historical data

### 🔄 How Position Updates Work

Since LTP is not available on Free tier:
- Position updates based on **completed order prices**
- MTM calculated using **last trade price** from your orders
- Updates every 5 seconds from order history
- Fully functional for active trading
- Shows real P&L based on your actual trades

---

## 🎯 Options Trading Best Practices

### NIFTY Options
- **Lot Size**: 25
- **Strike Intervals**: 50 points
- **Weekly Expiry**: Thursday
- **Monthly Expiry**: Last Thursday
- **Trading Hours**: 9:15 AM - 3:30 PM

### BANKNIFTY Options
- **Lot Size**: 15
- **Strike Intervals**: 100 points
- **Weekly Expiry**: Wednesday
- **Monthly Expiry**: Last Wednesday
- **Trading Hours**: 9:15 AM - 3:30 PM

### Quick Trading Tips
1. **Pre-select ATM/OTM strikes** before market opens
2. **Update config daily** with current week expiry
3. **Use MIS product** for intraday (auto-squares off)
4. **Monitor web GUI** - console is silent
5. **Check logs** for execution confirmations

---

## 📝 Log Files

### trading.log
- Located in project root
- All application activity
- Rotates daily (create new dated files if needed)
- Format: `YYYY-MM-DD HH:MM:SS [LEVEL] Message`

Example:
```
2025-01-22 09:16:23 [INFO] BUY command: 2 lots
2025-01-22 09:16:24 [INFO] Order placed successfully! Order ID: 250122000123456
2025-01-22 09:16:25 [INFO] Order EXECUTED: BUY 2 lots @ ₹145.50
```

### Viewing Logs

**In Web GUI:**
- Click "Show Logs" button
- See last 1000 entries
- Real-time updates
- Color-coded levels

**In Terminal:**
```bash
# Watch live
tail -f trading.log

# View recent
tail -100 trading.log

# Search for errors
grep ERROR trading.log
```

---

## 🔧 Troubleshooting

### Console Not Responding
**Issue**: Terminal doesn't register key presses

**Solution:**
- Terminal must be in focus
- Try pressing `Ctrl+C` to exit and restart
- Check terminal emulator compatibility

### Orders Not Executing
**Check logs for:**
```bash
grep "Order" trading.log | tail -20
```

**Common issues:**
- Insufficient funds
- Market closed
- Invalid instrument symbol
- Expired option contract

### Web GUI Not Updating
**Troubleshooting:**
1. Check WebSocket connection (green "Connected" status)
2. Check browser console for errors (F12)
3. Restart application
4. Try different browser

### Access Token Expired
**Symptoms:**
- "Failed to verify credentials" error
- Orders fail with authentication error

**Solution:**
```bash
go run cmd/generate-token/main.go
# Update config.json with new token
```

---

## 🚨 Important Notes

### Daily Routine
1. **6:00 AM**: Access token expires
2. **8:30 AM**: Generate new token
3. **9:00 AM**: Update instrument symbols if needed
4. **9:15 AM**: Market opens, start trading

### Risk Management
- ⚠️ Options can move fast - monitor closely
- ⚠️ Web GUI updates every 5 seconds (no live quotes on Free tier)
- ⚠️ Use stop-losses manually via limit orders
- ⚠️ Close positions before 3:20 PM to avoid auto-squareoff

### Data Accuracy
- Position values based on **your last trade price**
- Not real-time market prices (Free tier limitation)
- Still 100% accurate for P&L calculation
- Perfect for active trading where you have recent fills

---

## 🎓 Example Trading Session

```
# Morning: 9:00 AM
cd ~/fast-trading-app
go run cmd/generate-token/main.go
# Update config.json with new token
# Update symbols for current week expiry

# Start app: 9:14 AM
go run main.go
# Select: 1 (NIFTY25JAN24500CE)

# Market opens: 9:15 AM
# Open browser: http://localhost:8080

# Trading:
Press: 2       → Buy 2 lots NIFTY CE
         Check web GUI → Order executed @ ₹165.50
Press: 1       → Buy 1 more lot
         Check web GUI → Order executed @ ₹168.00
Press: -1      → Sell 1 lot
         Check web GUI → Order executed @ ₹172.50
Press: --      → Close remaining 2 lots
         Check web GUI → All positions closed

# View logs:
Click "Show Logs" → See complete activity history

# End of day:
Press: Q       → Exit
```

---

## 📞 Support

### Kite Connect API
- Docs: https://kite.trade/docs/connect/v3/
- Forum: https://kite.trade/forum
- Status: https://status.kite.trade/

### Go Resources
- Documentation: https://golang.org/doc/
- GoKiteConnect: https://github.com/zerodha/gokiteconnect

---

## ✅ Success Checklist

- [ ] Go 1.21+ installed
- [ ] Project structure created
- [ ] All files copied correctly
- [ ] Dependencies installed (`go mod tidy`)
- [ ] Kite Connect app created (Personal Free)
- [ ] API credentials in config.json
- [ ] Access token generated
- [ ] Current week option symbols updated
- [ ] Application builds (`go build`)
- [ ] Application runs (`go run main.go`)
- [ ] Web GUI accessible (localhost:8080)
- [ ] Test order placed successfully
- [ ] Logs visible in web GUI

🎉 **You're ready to trade!**