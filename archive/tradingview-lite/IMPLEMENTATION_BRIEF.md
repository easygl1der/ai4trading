# TradingView-Lite · Codex Implementation Brief

> **目标：** 在我的 VPS 上从零部署一个单用户、自托管的 TradingView-lite Web App。  
> **语言栈：** **Go（后端）+ TypeScript/React（前端）**  
> **核心数据源：** Yahoo Finance（通过已有的 Yahoo Finance MCP 或直接调 Yahoo Finance 非官方 API）  
> **发送方：** Yihua YUE，HKUST，Hong Kong VPS  

---

## 0. 任务总览

| 层 | 技术 | 职责 |
|---|---|---|
| 反向代理 | Caddy 2 | HTTPS 自动证书、路由 |
| API 服务 | Go + Fiber v2 | REST + WebSocket，核心逻辑 |
| 数据库 | PostgreSQL 16 | 持久化：OHLCV、alerts、layouts、watchlists |
| 缓存/队列 | Redis 7 | 热数据缓存、alert 状态机、任务队列 |
| 后台调度 | Go goroutine + ticker | 轮询行情、评估 alerts、发通知 |
| 前端 | React 18 + Vite + TypeScript | SPA，图表层使用 TradingView lightweight-charts v5 |
| 图表引擎 | lightweight-charts v5 | K 线、指标叠加、多图联看 |
| 指标计算 | cinar/indicator v2（Go） | SMA/EMA/RSI/MACD/BBANDS/ATR 等 80+ 指标 |
| 回测 | 自研 bar-based engine（Go） | 基于 cinar/indicator v2 自带 backtest 框架 |
| 迁移工具 | Goose | SQL 文件版本化迁移 |
| 部署 | Docker Compose | 一键启停，systemd 守护 |

---

## 1. 目录结构

```
tradingview-lite/
├── docker-compose.yml
├── Caddyfile
├── .env.example
├── Makefile
│
├── backend/                          # Go 后端
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/         config.go
│   │   ├── db/             postgres.go, redis.go
│   │   ├── model/          ohlcv.go, alert.go, layout.go, watchlist.go, strategy.go, backtest.go
│   │   ├── repository/     ohlcv_repo.go, alert_repo.go, layout_repo.go, watchlist_repo.go
│   │   ├── service/
│   │   │   ├── market/     fetcher.go, cache.go
│   │   │   ├── indicator/  engine.go       # 调 cinar/indicator v2
│   │   │   ├── alert/      engine.go       # 状态机
│   │   │   ├── backtest/   engine.go
│   │   │   └── screener/   engine.go
│   │   ├── handler/        bars.go, indicators.go, alerts.go, backtests.go, layouts.go, watchlists.go, ws.go
│   │   ├── scheduler/      scheduler.go
│   │   └── notifier/       webhook.go, telegram.go
│   ├── migrations/         # Goose SQL 文件
│   │   ├── 001_init.sql
│   │   ├── 002_alerts.sql
│   │   └── ...
│   ├── go.mod
│   └── Dockerfile
│
├── frontend/                         # React + Vite + TS
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── api/            client.ts, bars.ts, indicators.ts, alerts.ts, backtests.ts
│   │   ├── components/
│   │   │   ├── ChartPane.tsx         # 单张图封装
│   │   │   ├── ChartLayout.tsx       # 多图联看布局（2/4图）
│   │   │   ├── IndicatorSidebar.tsx
│   │   │   ├── AlertPanel.tsx
│   │   │   ├── WatchlistSidebar.tsx
│   │   │   └── BacktestPanel.tsx
│   │   ├── hooks/          useChart.ts, useAlerts.ts, useBacktest.ts
│   │   └── store/          chartStore.ts (Zustand)
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
│
└── README.md
```

---

## 2. 数据库 Schema（Goose SQL 迁移）

### 001_init.sql

```sql
-- +goose Up

CREATE TABLE symbols (
  id          BIGSERIAL PRIMARY KEY,
  symbol      TEXT NOT NULL UNIQUE,
  asset_type  TEXT NOT NULL DEFAULT 'stock',  -- stock|etf|crypto|forex
  exchange    TEXT,
  timezone    TEXT NOT NULL DEFAULT 'America/New_York',
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ohlcv_bars (
  id         BIGSERIAL PRIMARY KEY,
  symbol     TEXT NOT NULL,
  timeframe  TEXT NOT NULL,              -- 1m|5m|15m|1h|4h|1d|1w
  ts         TIMESTAMPTZ NOT NULL,
  open       DOUBLE PRECISION NOT NULL,
  high       DOUBLE PRECISION NOT NULL,
  low        DOUBLE PRECISION NOT NULL,
  close      DOUBLE PRECISION NOT NULL,
  volume     DOUBLE PRECISION NOT NULL,
  UNIQUE(symbol, timeframe, ts)
);
CREATE INDEX idx_ohlcv_symbol_tf_ts ON ohlcv_bars(symbol, timeframe, ts DESC);

CREATE TABLE watchlists (
  id         BIGSERIAL PRIMARY KEY,
  name       TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE watchlist_items (
  id           BIGSERIAL PRIMARY KEY,
  watchlist_id BIGINT NOT NULL REFERENCES watchlists(id) ON DELETE CASCADE,
  symbol       TEXT NOT NULL,
  added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(watchlist_id, symbol)
);

CREATE TABLE chart_layouts (
  id          BIGSERIAL PRIMARY KEY,
  name        TEXT NOT NULL,
  layout_json JSONB NOT NULL,            -- {tabs:[{charts:[{symbol,tf,indicators:[]}]}]}
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS chart_layouts, watchlist_items, watchlists, ohlcv_bars, symbols;
```

### 002_alerts.sql

```sql
-- +goose Up

CREATE TYPE alert_type  AS ENUM ('price', 'technical');
CREATE TYPE alert_status AS ENUM ('active', 'triggered', 'disabled');
CREATE TYPE trigger_mode AS ENUM ('intrabar_best_effort', 'once_per_bar_close', 'once_until_reset');

CREATE TABLE alerts (
  id             BIGSERIAL PRIMARY KEY,
  type           alert_type   NOT NULL,
  symbol         TEXT NOT NULL,
  timeframe      TEXT NOT NULL DEFAULT '1d',
  rule_json      JSONB NOT NULL,
  -- price alert rule_json example:
  -- {"operator":"cross_above","value":150.0}
  -- technical alert rule_json example:
  -- {"indicator":"rsi","params":{"length":14},"condition":{"operator":"cross_above","left":"rsi","right":70}}
  trigger_mode   trigger_mode NOT NULL DEFAULT 'once_per_bar_close',
  status         alert_status NOT NULL DEFAULT 'active',
  cooldown_sec   INT NOT NULL DEFAULT 3600,
  last_triggered TIMESTAMPTZ,
  notify_webhook TEXT,
  notify_telegram TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at     TIMESTAMPTZ
);
CREATE INDEX idx_alerts_symbol_status ON alerts(symbol, status);

CREATE TABLE alert_events (
  id           BIGSERIAL PRIMARY KEY,
  alert_id     BIGINT NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
  triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  payload_json JSONB
);

-- +goose Down
DROP TABLE IF EXISTS alert_events, alerts;
DROP TYPE IF EXISTS trigger_mode, alert_status, alert_type;
```

### 003_strategies.sql

```sql
-- +goose Up

CREATE TYPE strategy_kind AS ENUM ('indicator_rule', 'script');

CREATE TABLE strategies (
  id              BIGSERIAL PRIMARY KEY,
  name            TEXT NOT NULL,
  kind            strategy_kind NOT NULL DEFAULT 'indicator_rule',
  definition_json JSONB NOT NULL,
  -- indicator_rule example:
  -- {
  --   "entry": {"indicator":"rsi","params":{"length":14},"condition":{"operator":"cross_below","left":"rsi","right":30}},
  --   "exit":  {"indicator":"rsi","params":{"length":14},"condition":{"operator":"cross_above","left":"rsi","right":70}},
  --   "direction": "long"
  -- }
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE backtest_runs (
  id           BIGSERIAL PRIMARY KEY,
  strategy_id  BIGINT NOT NULL REFERENCES strategies(id) ON DELETE CASCADE,
  symbol       TEXT NOT NULL,
  timeframe    TEXT NOT NULL,
  from_ts      TIMESTAMPTZ NOT NULL,
  to_ts        TIMESTAMPTZ NOT NULL,
  params_json  JSONB,
  result_json  JSONB,   -- {total_return, cagr, max_drawdown, win_rate, sharpe, trade_count, equity_curve:[{ts,equity}], trades:[...]}
  ran_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS backtest_runs, strategies;
DROP TYPE IF EXISTS strategy_kind;
```

---

## 3. Go 后端核心模块

### 3.1 依赖（go.mod 关键项）

```
github.com/gofiber/fiber/v2               v2.52+     # HTTP 框架
github.com/gofiber/websocket/v2            v2.3+      # WebSocket
github.com/cinar/indicator/v2              v2.*       # TA 指标 + 内置回测框架（80+ 指标，零依赖）
github.com/markcheno/go-talib              latest     # 可选，TA-Lib pure Go port（备用）
github.com/jackc/pgx/v5                   v5.*       # Postgres 驱动
github.com/redis/go-redis/v9              v9.*       # Redis 客户端
github.com/pressly/goose/v3               v3.*       # DB 迁移
github.com/z-Wind/yahoofinance            latest     # Yahoo Finance Go 客户端
github.com/spf13/viper                    v1.*       # 配置
go.uber.org/zap                           v1.*       # 结构化日志
```

> **关于指标库选型：** 优先用 `cinar/indicator v2`，它是零依赖纯 Go 实现，自带 80+ 指标和事件驱动回测框架；`markcheno/go-talib` 作为备选（需要系统安装 TA-Lib C 库，容器化时增加复杂度）。

### 3.2 数据获取层（`internal/service/market/fetcher.go`）

```go
package market

// FetchBars 通过 Yahoo Finance 拉取 OHLCV
// symbol: "AAPL", timeframe: "1d"|"1h"|"5m" 等
// 返回 []model.Bar，内部先查 Redis 缓存，再查 Postgres，最后才远程拉取
type Fetcher interface {
    FetchBars(ctx context.Context, symbol, timeframe string, from, to time.Time) ([]model.Bar, error)
    FetchQuote(ctx context.Context, symbol string) (*model.Quote, error)
}

// 实现要点：
// 1. 远程拉取：使用 github.com/z-Wind/yahoofinance 的 History.Period() 或 History.Range()
// 2. timeframe 映射：
//    "1m" -> interval="1m", "5m" -> "5m", "1h" -> "1h", "1d" -> "1d", "1w" -> "1wk", "1mo" -> "1mo"
// 3. 拉回来后存入 ohlcv_bars（upsert on conflict do nothing）
// 4. Redis 缓存：key = "bars:{symbol}:{timeframe}:{date_from}:{date_to}"，TTL 5 分钟（日线），1 分钟（分钟线）
// 5. 注意 Yahoo Finance 分钟数据只能回溯约 30 天；日线可以回溯数年
```

### 3.3 指标引擎（`internal/service/indicator/engine.go`）

```go
package indicator

import (
    "github.com/cinar/indicator/v2/trend"
    "github.com/cinar/indicator/v2/momentum"
    "github.com/cinar/indicator/v2/volatility"
    // 等子包
)

// ComputeRequest 前端发来的指标计算请求
type ComputeRequest struct {
    Symbol    string          `json:"symbol"`
    Timeframe string          `json:"timeframe"`
    From      time.Time       `json:"from"`
    To        time.Time       `json:"to"`
    Indicators []IndicatorSpec `json:"indicators"`
}

type IndicatorSpec struct {
    Name   string            `json:"name"`   // "sma"|"ema"|"rsi"|"macd"|"bbands"|"atr"|"stoch"|"adx"|...
    Params map[string]any    `json:"params"` // {"period":14} 等
}

type ComputeResponse struct {
    Bars   []BarData             `json:"bars"`
    Series map[string][]float64  `json:"series"` // key = "sma_20", "rsi_14" 等
}

// 支持的指标（最低要求，cinar/indicator v2 可扩展到 80+）：
// SMA, EMA, DEMA, TEMA, WMA
// RSI, Stochastic, Williams %R
// MACD (line, signal, histogram)
// Bollinger Bands (upper, middle, lower)
// ATR, True Range
// ADX, +DI, -DI
// CCI, CMO
// Volume (OBV, VWAP)
```

### 3.4 Alert 引擎（`internal/service/alert/engine.go`）

```go
package alert

// AlertState 每个 alert 的运行时状态，存 Redis
type AlertState struct {
    AlertID        int64     `json:"alert_id"`
    LastBarTS      time.Time `json:"last_bar_ts"`      // 上一个已评估的 bar 时间
    PrevValue      float64   `json:"prev_value"`        // 上一 bar 的指标值（用于 crossover 检测）
    Fired          bool      `json:"fired"`             // once_until_reset 模式下是否已触发
    CooldownUntil  time.Time `json:"cooldown_until"`
}

// EvalAlert 核心逻辑：
// 1. 拉最近 N 根 bar（N = max indicator lookback + 2）
// 2. 对 technical alert：调 indicator engine 计算当前值
// 3. 根据 trigger_mode 决定是否触发：
//    - once_per_bar_close：只在 bar 收盘后评估，且同一根 bar 只触发一次
//    - intrabar_best_effort：每次轮询都评估
//    - once_until_reset：触发后不再触发直到手动重置
// 4. crossover/crossunder 检测：需要 prev_value（来自 AlertState）
// 5. 触发后：写 alert_events，调 Notifier，更新 last_triggered + CooldownUntil
// 6. 所有状态保存在 Redis（key = "alert_state:{alert_id}"）避免重复触发

// Notifier 接口
type Notifier interface {
    Send(ctx context.Context, alert *model.Alert, event *model.AlertEvent) error
}
// 实现：WebhookNotifier（POST JSON）、TelegramNotifier（Bot API）
```

### 3.5 Backtest 引擎（`internal/service/backtest/engine.go`）

```go
package backtest

// 使用 cinar/indicator/v2 自带的 backtest 包，或自实现简单 bar-based loop
// 参考：https://github.com/cinar/indicator/v2/backtest

type RunRequest struct {
    StrategyID int64       `json:"strategy_id"`
    Symbol     string      `json:"symbol"`
    Timeframe  string      `json:"timeframe"`
    From       time.Time   `json:"from"`
    To         time.Time   `json:"to"`
    InitCash   float64     `json:"init_cash"`  // default 100000
    FeeRate    float64     `json:"fee_rate"`   // default 0.001
    Slippage   float64     `json:"slippage"`   // default 0.0
    Direction  string      `json:"direction"`  // "long"|"short"|"both"
}

type RunResult struct {
    TotalReturn   float64         `json:"total_return"`
    CAGR          float64         `json:"cagr"`
    MaxDrawdown   float64         `json:"max_drawdown"`
    SharpeRatio   float64         `json:"sharpe_ratio"`
    WinRate       float64         `json:"win_rate"`
    TradeCount    int             `json:"trade_count"`
    EquityCurve   []EquityPoint   `json:"equity_curve"`  // [{ts, equity}]
    Trades        []TradeRecord   `json:"trades"`        // [{entry_ts, exit_ts, entry_price, exit_price, pnl}]
}
```

### 3.6 HTTP API（Go Fiber 路由）

```go
// cmd/server/main.go 路由注册示意
api := app.Group("/api/v1")

// === 行情数据 ===
api.Get("/symbols/search", handler.SearchSymbols)          // ?q=AAPL
api.Get("/bars",           handler.GetBars)                // ?symbol=AAPL&timeframe=1d&from=2024-01-01&to=2025-01-01&limit=500
api.Get("/quote/:symbol",  handler.GetQuote)               // 最新报价

// === 指标计算 ===
api.Post("/indicators/compute", handler.ComputeIndicators)

// === Alerts ===
api.Get("/alerts",         handler.ListAlerts)
api.Post("/alerts",        handler.CreateAlert)
api.Put("/alerts/:id",     handler.UpdateAlert)
api.Delete("/alerts/:id",  handler.DeleteAlert)
api.Post("/alerts/:id/test", handler.TestAlert)           // 立即测试触发一次
api.Post("/alerts/:id/reset", handler.ResetAlert)         // 重置 once_until_reset

// === 策略 ===
api.Get("/strategies",     handler.ListStrategies)
api.Post("/strategies",    handler.CreateStrategy)
api.Put("/strategies/:id", handler.UpdateStrategy)
api.Delete("/strategies/:id", handler.DeleteStrategy)

// === 回测 ===
api.Post("/backtests/run", handler.RunBacktest)
api.Get("/backtests/:id",  handler.GetBacktestResult)

// === Layout ===
api.Get("/layouts",        handler.ListLayouts)
api.Post("/layouts",       handler.SaveLayout)
api.Put("/layouts/:id",    handler.UpdateLayout)
api.Delete("/layouts/:id", handler.DeleteLayout)

// === Watchlist ===
api.Get("/watchlists",               handler.ListWatchlists)
api.Post("/watchlists",              handler.CreateWatchlist)
api.Post("/watchlists/:id/items",    handler.AddWatchlistItem)
api.Delete("/watchlists/:id/items/:symbol", handler.RemoveWatchlistItem)

// === WebSocket（实时行情推送）===
app.Get("/ws/quotes", websocket.New(handler.WsQuotes))

// === 健康检查 ===
app.Get("/health", handler.Health)
```

### 3.7 Scheduler（`internal/scheduler/scheduler.go`）

```go
// 两个核心任务：
// Task 1: RefreshQuotes（每 30 秒）
//   - 遍历所有 active watchlist symbols
//   - 调 fetcher.FetchQuote() 更新 Redis 热数据
//   - 通过 WebSocket 广播最新报价给已连接的前端

// Task 2: EvaluateAlerts（每 1 分钟，或按最小 alert timeframe 动态调整）
//   - 读取所有 status='active' 的 alerts
//   - 批量按 symbol+timeframe 分组，每组拉 bars 一次（避免重复拉取）
//   - 调 alert.EvalAlert() 对每个 alert 评估
//   - 触发的 alert 调 notifier 发出通知

// 实现要点：
// - 用 time.NewTicker + select + context.WithCancel 做优雅关闭
// - 每次 EvaluateAlerts 加分布式锁（Redis SETNX），防止并发重复执行
// - 分钟线 alert 每分钟评估，日线 alert 可以降低到每 5 分钟或 bar close 时间点
```

---

## 4. Alert Rule JSON Schema

### 4.1 Price Alert

```json
{
  "type": "price",
  "symbol": "AAPL",
  "timeframe": "1d",
  "trigger_mode": "once_per_bar_close",
  "cooldown_sec": 86400,
  "rule_json": {
    "operator": "cross_above",
    "value": 200.0
  }
}
```

支持的 operator：`cross_above`、`cross_below`、`greater_than`、`less_than`、`equal`

### 4.2 Technical Alert

```json
{
  "type": "technical",
  "symbol": "AAPL",
  "timeframe": "1h",
  "trigger_mode": "once_per_bar_close",
  "cooldown_sec": 3600,
  "rule_json": {
    "indicator": "rsi",
    "params": {"length": 14},
    "condition": {
      "operator": "cross_above",
      "left": "rsi",
      "right": 70
    }
  }
}
```

多条件 AND（可扩展）：
```json
{
  "rule_json": {
    "logic": "AND",
    "conditions": [
      {"indicator":"rsi","params":{"length":14},"condition":{"operator":"cross_above","left":"rsi","right":70}},
      {"indicator":"macd","params":{"fast":12,"slow":26,"signal":9},"condition":{"operator":"greater_than","left":"macd_histogram","right":0}}
    ]
  }
}
```

---

## 5. 前端核心模块（React + Vite + TypeScript）

### 5.1 依赖（package.json 关键项）

```json
{
  "dependencies": {
    "lightweight-charts": "^5.0",
    "react": "^18",
    "react-dom": "^18",
    "zustand": "^4",
    "axios": "^1",
    "dayjs": "^1"
  },
  "devDependencies": {
    "@types/react": "^18",
    "typescript": "^5",
    "vite": "^5"
  }
}
```

### 5.2 ChartPane.tsx（单图封装）

```typescript
// 每个 ChartPane 封装一个 lightweight-charts createChart 实例
// Props:
interface ChartPaneProps {
  symbol: string;
  timeframe: string;
  indicators: IndicatorSpec[];    // 最多 5 个（Essential 限制），前端强制校验
  syncTimeRange?: (range: LogicalRange) => void;     // 多图同步回调
  sharedTimeRange?: LogicalRange | null;             // 来自兄弟图的 range
}

// 关键实现点：
// 1. useEffect 初始化 chart + candlestickSeries + overlayPanes（RSI/MACD 等 oscillator 用 panes）
// 2. 向 /api/v1/bars?symbol=...&timeframe=... 拉历史数据，setData()
// 3. indicator 计算：向 /api/v1/indicators/compute 发请求，拿到 series 后 addLineSeries().setData()
// 4. subscribeVisibleLogicalRangeChange → 调 syncTimeRange 与兄弟图同步
// 5. WebSocket 连 /ws/quotes，收到最新 bar 时调 candlestickSeries.update()
// 6. cleanup: chart.remove() in return
```

### 5.3 多图同步（参考 lightweight-charts 官方文档）

```typescript
// ChartLayout.tsx 中用 Zustand 管理共享 LogicalRange
// 任意一图的 subscribeVisibleLogicalRangeChange 触发时：
//   store.setSharedRange(range)
// 每个 ChartPane 监听 store.sharedRange：
//   chart.timeScale().setVisibleLogicalRange(sharedRange)
// 注意：避免循环触发，加一个 isSyncing ref
```

### 5.4 前端页面布局

```
┌──────────────────────────────────────────────────────────┐
│  顶部导航：Symbol 搜索 | Timeframe 选择 | Layout 保存/加载  │
├──────────────┬───────────────────────────────────────────┤
│  Watchlist   │                                           │
│  Sidebar     │        ChartLayout（1图/2图/4图 切换）      │
│  - 股票列表  │                                           │
│  - 最新报价  │                                           │
│  - 涨跌幅    │                                           │
├──────────────┼───────────────────────────────────────────┤
│  Alerts      │        底部面板（Alerts / Backtest 切换）   │
│  Panel       │                                           │
└──────────────┴───────────────────────────────────────────┘
```

---

## 6. Docker Compose（`docker-compose.yml`）

```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${POSTGRES_DB}
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER}"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 10s
      retries: 5

  backend:
    build: ./backend
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable
      REDIS_ADDR: redis:6379
      REDIS_PASSWORD: ${REDIS_PASSWORD}
      APP_PORT: 8080
      APP_ENV: production
    volumes:
      - ./backend/migrations:/app/migrations

  frontend:
    build: ./frontend
    restart: unless-stopped
    depends_on:
      - backend

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - backend
      - frontend

volumes:
  postgres_data:
  redis_data:
  caddy_data:
  caddy_config:
```

### Caddyfile

```
chart.yourdomain.com {
    # 前端静态文件
    handle /api/* {
        reverse_proxy backend:8080
    }
    handle /ws/* {
        reverse_proxy backend:8080 {
            header_up Host {host}
            header_up Upgrade {>Upgrade}
            header_up Connection {>Connection}
        }
    }
    handle {
        reverse_proxy frontend:3000
    }
}
```

> **注意：** 部署前请把 `chart.yourdomain.com` 替换为你实际的子域名，确保 DNS A 记录指向 VPS IP，Caddy 会自动申请 Let's Encrypt 证书。

---

## 7. `.env.example`

```bash
# PostgreSQL
POSTGRES_DB=tradingview_lite
POSTGRES_USER=tvlite
POSTGRES_PASSWORD=CHANGE_ME_STRONG_PASSWORD

# Redis
REDIS_PASSWORD=CHANGE_ME_REDIS_PASSWORD

# App
APP_PORT=8080
APP_ENV=production
JWT_SECRET=CHANGE_ME_JWT_SECRET

# Notifications (可选)
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
```

---

## 8. 后端 Dockerfile

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /server .
COPY migrations/ ./migrations/
EXPOSE 8080
CMD ["./server"]
```

> **注意：** cinar/indicator v2 是纯 Go 零依赖，`CGO_ENABLED=0` 完全没问题。如果日后需要用 go-talib（CGO），需改为 `CGO_ENABLED=1` 并安装 TA-Lib C 库。

---

## 9. 前端 Dockerfile（多阶段构建）

```dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:22-alpine
RUN npm install -g serve
WORKDIR /app
COPY --from=builder /app/dist ./dist
EXPOSE 3000
CMD ["serve", "-s", "dist", "-l", "3000"]
```

---

## 10. Makefile

```makefile
.PHONY: up down logs migrate build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

build:
	docker compose build --no-cache

migrate:
	docker compose exec backend ./server migrate up

psql:
	docker compose exec postgres psql -U $$POSTGRES_USER -d $$POSTGRES_DB
```

---

## 11. 分期里程碑与验收标准

### MVP（第一阶段）

| 验收项 | 描述 |
|---|---|
| V1 | 输入 AAPL，能加载并显示日线 K 线图 |
| V2 | 单 Tab 能同时显示 2 张图（symbol/timeframe 各自独立） |
| V3 | 每张图能叠加至少 5 个指标（SMA/EMA/RSI/MACD/BBANDS） |
| V4 | 能浏览至少 10,000 根历史 bars（向左拖动查看） |
| V5 | 能创建 20 个价格提醒并在触发时收到 webhook 通知 |
| V6 | 能创建 20 个技术提醒（RSI cross、MA cross），bar close 模式下不重复触发 |
| V7 | 能保存一个 layout（symbol + timeframe + indicators），关闭浏览器后重新打开能恢复 |
| V8 | 能运行一个 RSI 策略的回测，返回 equity curve 和交易列表 |
| V9 | VPS 重启后服务自动恢复（systemd 或 Docker restart policy） |

### v1 扩展（第二阶段，MVP 完成后）

- 4 图联看布局切换
- 更多指标：Stochastic、ATR、ADX、CCI、OBV、VWAP
- Alert 多条件 AND/OR 逻辑
- 邮件/Telegram 通知
- 简单 screener（按 RSI/MA 条件过滤 watchlist 中的股票）
- 回测参数调整 UI（期初资金、费率、滑点）
- Preset layouts 快速切换

### v2 未来扩展（不在 MVP 范围内）

- 类 Pine Script DSL / JSON 规则语言
- 绘图工具（趋势线、矩形、斐波那契）
- Alert 与绘图联动（价格穿越趋势线触发）
- 多用户/权限系统
- 策略 JSON Webhook alert payload

---

## 12. 关键技术注意事项

### 12.1 Yahoo Finance 数据限制

- **日线数据：** 可回溯数年，基本无问题。
- **小时数据：** 通常可回溯 730 天左右。
- **分钟数据（1m/2m/5m）：** 只能回溯约 **7-30 天**，这是 Yahoo Finance 的接口限制，不是 bug。如果你需要分钟历史回测，需要配合其他数据源（如 Polygon.io、Alpaca、Alpha Vantage）。
- **速率限制：** 非官方 API 无明确限制，但建议每次批量请求之间加 200ms-500ms sleep 避免被封。
- **实时行情：** Yahoo Finance 数据有约 15 分钟延迟（非官方免费接口），不是真正 Level 1 实时行情。

### 12.2 cinar/indicator v2 使用要点

- 所有指标输入/输出均为 `<-chan float64`（channel-based streaming），不是 slice，这和传统 TA-Lib 接口不同。
- 如果你的 bars 是静态 slice，需要先用辅助函数转为 channel，或者用 v2 提供的 `helper.SliceToChan()`。
- 回测框架在 `github.com/cinar/indicator/v2/backtest` 包，是事件驱动的，支持自定义 Strategy 接口。
- 80+ 指标分布在多个子包：`trend/`、`momentum/`、`volatility/`、`volume/`、`oscillator/` 等，import 时需要按需引入。

### 12.3 lightweight-charts v5 关键 API

- **多 panes：** v5 正式支持多 pane（RSI、MACD 等 oscillator 放到独立 pane），在 `addSeries()` 时指定 `paneIndex`。
- **双图同步：** 用 `subscribeVisibleLogicalRangeChange` + `setVisibleLogicalRange` 实现，见前端章节。
- **Crosshair 同步：** 用 `subscribeCrosshairMove` 跨两个 chart 实例同步十字线位置。
- **WebSocket 更新：** 实时 bar 更新只需调 `series.update(bar)`，不需要重新 setData。

### 12.4 Alert 状态机语义

- **once_per_bar_close：** 每根 bar 关闭后评估一次；同一根 bar 内无论轮询多少次都不重复触发。实现方式：在 `AlertState.LastBarTS` 记录上一次触发的 bar 时间，每次评估前比较当前 bar 的 ts 是否大于 `LastBarTS`。
- **intrabar_best_effort：** 每次轮询都评估，cooldown_sec 控制最小间隔。
- **once_until_reset：** 触发一次后设 `AlertState.Fired = true`，直到用户手动调 `/api/v1/alerts/:id/reset`。
- **crossover 检测：** 需要前后两个值；`AlertState.PrevValue` 存上一 bar 的指标值，当前值从 Go channel 里取最后一个值（因为 cinar/indicator 是流式的，需要消费完整个 channel 才能拿到最后值）。

### 12.5 数据库时序表优化

- `ohlcv_bars` 表要加复合索引 `(symbol, timeframe, ts DESC)`，查询历史 bars 时走索引范围扫描。
- 不需要一开始就上 TimescaleDB；普通 Postgres 对于单用户研究级别（每日数据、几十个标的）完全够用。数据量超过 1 亿行再考虑迁移。
- 定期运行 `VACUUM ANALYZE ohlcv_bars` 保持查询性能。

---

## 13. 启动与日常运维

```bash
# 0. 复制环境变量
cp .env.example .env
# 编辑 .env，填入真实密码、域名等

# 1. 首次构建并启动
make build
make up

# 2. 运行数据库迁移
make migrate

# 3. 查看日志
make logs

# 4. 数据库进入交互
make psql

# 5. 停止服务
make down

# 6. 备份（在 VPS 上执行）
docker compose exec postgres pg_dump -U $POSTGRES_USER $POSTGRES_DB > backup_$(date +%Y%m%d).sql

# 7. 更新部署
git pull
make build
make up
```

---

## 14. 给 Codex 的执行顺序建议

1. **先初始化 Go module 和目录结构**，确认 `go mod init` 正常。
2. **实现 DB 层**：goose 迁移脚本 + 基础 repository，先跑通 `make migrate`。
3. **实现 fetcher**：用 `z-Wind/yahoofinance` 拉一次 AAPL 日线，存入 DB，确认数据形状正确。
4. **实现 `/api/v1/bars` endpoint**，用 curl 验证能返回 JSON。
5. **实现指标引擎 + `/api/v1/indicators/compute`**，先只实现 SMA、EMA、RSI 三个。
6. **搭前端脚手架**：Vite + React + lightweight-charts，先画一张静态 K 线。
7. **前后端联通**：前端从 `/api/v1/bars` 拉数据，渲染真实 K 线。
8. **实现多图联看**：2 图 + 时间轴同步。
9. **实现 Alert 引擎 + Scheduler**：先只做 price alert，通过 `/api/v1/alerts/:id/test` 验证。
10. **实现技术 alert**：RSI cross 触发，once_per_bar_close 模式。
11. **实现 Backtest**：先只支持 RSI mean-reversion 策略，验证 equity curve 合理。
12. **Docker Compose 整合**：所有服务一键 `make up` 跑通。
13. **Caddy HTTPS**：配置域名，验证公网可访问。
14. **MVP 验收**：逐项检查第 11 节的 V1-V9。

---

*文档版本：2026-07-07 | 作者：Yihua YUE | 目标 VPS：Hong Kong*
