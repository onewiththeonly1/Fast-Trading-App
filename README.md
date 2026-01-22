# 🚀 Fast Options Trading Application

**Ultra-Low Latency | Single-Keystroke Trading | Web Dashboard | Free-Tier Compatible**

---

## 🎯 What This Application Offers

### ⚡ **Instant Single-Keystroke Trading**

* No `Enter` key — every action is **immediate**
* Buy, sell, close, change instruments, quit — all with **single keys**
* Raw terminal input for zero latency

### ⚡ **Instant Instrument Selection (NEW)**

* Select instruments using **1–9, A–Z**
* Supports **up to 35 instruments**
* Case-insensitive (`a` = `A`)
* Press `Q` anytime to exit

### 🧼 **Silent Console, Full Visibility**

* Console goes silent after selection
* **Everything logged** to `trading.log`
* Real-time **log viewer in Web UI**
* Clean, distraction-free trading

### 🆓 **Zerodha Kite – Personal Free Tier Optimized**

* No WebSocket or live LTP required
* Position tracking via **executed order prices**
* Fully usable for **real trading**

### 📈 **Optimized for Options & Equities**

* NIFTY / BANKNIFTY / Equity supported
* MIS / NRML / CNC products
* Intraday + positional strategies

---

## 📁 Project Structure

```
fast-trading-app/
├── main.go
├── config.json
├── trading.log
├── cmd/
│   └── generate-token/
│       └── main.go
├── internal/
│   ├── config/
│   ├── logger/        # Logging system
│   ├── position/
│   ├── server/
│   ├── terminal/      # Raw terminal input
│   └── trader/
├── web/
│   └── index.html
├── go.mod
└── go.sum
```

---

## 🚀 Quick Start

### 1️⃣ Create Project

```bash
mkdir -p ~/fast-trading-app/{cmd/generate-token,internal/{config,logger,position,server,terminal,trader},web}
cd ~/fast-trading-app
go mod init fast-trading-app
```

### 2️⃣ Install Dependencies

```bash
go get github.com/zerodha/gokiteconnect/v4
go get github.com/gorilla/websocket@v1.5.1
go get golang.org/x/sys/unix
go mod tidy
```

### 3️⃣ Copy Files

Copy all project files exactly as provided.

---

## 🔐 Kite API Setup

1. Visit [https://developers.kite.trade/](https://developers.kite.trade/)
2. Create **Personal Free** app
3. Get `api_key`, `api_secret`
4. Update `config.json`

### Generate Access Token

```bash
go run cmd/generate-token/main.go
```

---

## ⚡ Instrument Selection – Single Keystroke

### Key Mapping

| Instrument # | Key     |
| ------------ | ------- |
| 1–9          | `1`–`9` |
| 10–35        | `A`–`Z` |
| Quit         | `Q`     |

**Maximum instruments:** 35

---

### Instrument Selection Screen

```
=== Available Instruments ===
1. NIFTY26JAN25100CE (NFO) [MIS] - Lot Size: 65
2. BANKNIFTY26JAN58700CE (NFO) [NRML] - Lot Size: 30
3. RELIANCE (NSE) [CNC] - Lot Size: 1
A. AXISBANK (NSE) [MIS] - Lot Size: 1

Select instrument (1-9, A-Z) or Q to quit:
```

* **No Enter required**
* Selection is instant
* Logged to `trading.log`

---

## 🎮 Trading Commands (No Enter Ever)

| Key       | Action              |
| --------- | ------------------- |
| `1`–`9`   | Buy lots            |
| `-1`–`-9` | Sell lots           |
| `--`      | Close all positions |
| `C`       | Change instrument   |
| `Q`       | Quit app            |

**Console stays silent — use Web UI**

---

## 🌐 Web Dashboard

### 📊 Positions Panel

* Lots, Units
* Avg Price
* Last Trade Price
* MTM P&L (₹ and %)

### 📜 Order History

* Time-sorted trades
* Buy/Sell color coding
* Execution prices

### 📝 Activity Logs (Live)

* INFO / WARN / ERROR / DEBUG
* Auto-scroll
* Collapsible

Open:

```
http://localhost:8080
```

---

## 🆓 Free Tier Behavior (Important)

### ✅ Supported

* Order placement
* Positions & orders
* Modify / cancel
* Funds & margins

### ❌ Not Available

* Live LTP
* WebSocket streaming
* Historical candles

### 🔄 How P&L Works

* Uses **last executed trade price**
* Updates every 5 seconds
* Accurate for real trading

---

## 🔄 Change Instrument During Trading

Press `C` (only when no open positions):

```
=== Available Instruments ===
1. NIFTY26JAN25100CE (NFO) [MIS]
2. BANKNIFTY26JAN58700CE (NFO) [NRML]

Select instrument (1-9, A-Z) or Q to quit: 2
```

---

## 🧠 Best Practices

### Organize Instruments

* 1–3 → Intraday options
* 4–6 → Positional
* 7–9 → Equity

### Daily Routine

1. Generate token (8:30 AM)
2. Update weekly expiry symbols
3. Start app before market open

---

## 📝 Logs

### trading.log

```
2025-01-22 09:16:24 [INFO] BUY 2 lots @ 145.50
```

```bash
tail -f trading.log
grep ERROR trading.log
```

---

## 🧯 Error Handling

* Invalid keys ignored during selection
* Terminal auto-restores state
* Fallback to first instrument if raw mode fails
* Graceful exit on `Q`

---

## 🎓 Example Trading Session

```
go run main.go
Press: 1 → Buy 1 lot
Press: 2 → Buy 2 lots
Press: -1 → Sell 1 lot
Press: -- → Close all
Press: Q → Exit
```

---

## ✅ Final Checklist

* [ ] Go 1.21+
* [ ] Kite app created
* [ ] Token generated
* [ ] Weekly symbols updated
* [ ] Web UI opens
* [ ] Test order works

---

🎉 **You now have a lightning-fast professional trading terminal built in Go.**
