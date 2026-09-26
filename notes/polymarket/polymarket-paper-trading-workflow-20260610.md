# Polymarket paper-trading workflow idea

Date: 2026-06-10

This note turns the current idea into an executable research and build plan:
discover live Polymarket crypto markets, run a virtual paper-trading bot, and only later consider real order placement.

## Current understanding

Polymarket has two API surfaces that matter for this project:

- Gamma API: market and event discovery. This is the right place to search for active BTC/ETH markets, event titles, market questions, token IDs, liquidity, min order size, tick size, and whether a market is accepting orders.
- CLOB API: order book, price, and real order placement. For paper trading, only read-only endpoints are needed at first. Real trading later requires wallet/private-key setup, Polymarket authentication, signed orders, allowance setup, and risk controls.

For the first version, do not place real orders. Treat Polymarket as a live market-data source and simulate all fills locally.

## What markets exist right now

I queried Polymarket public search on 2026-06-10.

Crypto markets found:

- Daily threshold markets, for example `Bitcoin above ___ on June 10?`.
  - Example active markets:
    - `Will the price of Bitcoin be above $58,000 on June 10?`
    - `Will the price of Bitcoin be above $60,000 on June 10?`
    - `Will the price of Bitcoin be above $62,000 on June 10?`
  - These resolve using Binance BTC/USDT one-minute candle close at a specified ET time.
- Similar Ethereum threshold markets, for example `Ethereum above ___ on June 10?`.
  - Example active market:
    - `Will the price of Ethereum be above $1,300 on June 10?`
  - I verified one CLOB price request for its YES token returned `{"price":"0.999"}`.
- Historical short-horizon markets exist with names like:
  - `Bitcoin Up or Down - April 2, 9:50PM-9:55PM ET`
  - `Ethereum Up or Down - March 16, 6:10PM-6:15PM ET`
  - `btc-updown-5m-...`
  - `eth-updown-5m-...`

Important implication: the 5-minute BTC/ETH "Up or Down" market family exists historically, but I did not confirm an active 5-minute market in the quick scan. The bot should discover active markets dynamically instead of hard-coding a slug.

## Minimal architecture

The first working system should be:

1. Market discovery
   - Search Polymarket for BTC/ETH crypto markets.
   - Keep only markets where:
     - `active == true`
     - `closed == false`
     - `acceptingOrders == true`
     - `enableOrderBook == true`
     - `clobTokenIds` exists
   - Rank by liquidity, 24h volume, spread, and time to resolution.

2. Market data snapshot
   - Use Gamma fields for metadata:
     - question
     - slug
     - end date
     - best bid / best ask
     - tick size
     - minimum order size
     - token IDs
   - Use CLOB read-only endpoints for executable-looking prices or order book.

3. Strategy interface
   - Input:
     - market metadata
     - latest Polymarket bid/ask
     - external reference price from Binance or another crypto data source
     - time to expiry
   - Output:
     - `BUY_YES`, `BUY_NO`, `SELL_YES`, `SELL_NO`, or `HOLD`
     - target size
     - max acceptable price
     - reason string

4. Paper execution engine
   - Simulate fills against observed bid/ask.
   - Conservative first rule:
     - market buy fills at best ask
     - market sell fills at best bid
     - reject if spread is too wide
     - reject if simulated order size is below Polymarket minimum order size
   - Record every simulated order and fill in local storage.

5. Portfolio and risk
   - Start with virtual cash, for example 1,000 USDC.
   - Hard limits:
     - max 1 percent of paper bankroll per trade
     - max 5 percent exposure per market group
     - no trading inside final N seconds until the settlement logic is well understood
     - no real private key in the first version

6. Settlement and PnL
   - For threshold markets, settlement can be estimated from Binance candles after the resolve time.
   - For pure paper trading, mark positions to bid/ask before official resolution.
   - Keep separate:
     - mark-to-market PnL
     - realized trading PnL
     - final resolved PnL

## Why paper trading first

Prediction-market microstructure is not the same as trading BTC spot or perpetuals.

Risks to model before real money:

- Market liquidity can disappear near expiry.
- Bid/ask spread can dominate expected edge.
- Some markets are near-certain and priced at 0.999 / 0.001, where fees and slippage matter more than direction.
- The market question and resolution source matter. Many crypto markets resolve from Binance one-minute candles at an exact ET timestamp, not from a generic "BTC price".
- Real order placement requires signing and operational security.

## First strategy ideas

Simple strategies worth paper-testing:

1. Threshold arbitrage-style model
   - For `BTC above X at time T`, compare current Binance BTC/USDT price, recent volatility, and time to T.
   - Estimate probability of close above X.
   - Buy YES only when estimated probability is meaningfully above Polymarket ask after slippage and fees.

2. Very short-horizon up/down model
   - Only if active 5-minute markets are found.
   - Use Binance one-minute momentum, short-term realized volatility, and order-book imbalance.
   - Avoid if spread is wide or time to close is too short.

3. Market-maker / spread capture paper model
   - Simulate posting bid/ask around fair value.
   - This should stay paper-only until real CLOB maker-order behavior is understood.

## Ready-to-use prompt for the next phase

Use this when you want Codex to build the first local prototype:

```text
目标：在 /Users/yitwah/Documents/ai4trading 中实现一个 Polymarket crypto paper-trading 原型，不接真实钱包、不下真实订单。

背景：我想研究 BTC/ETH 的短周期 Polymarket 市场，尤其是 5 分钟 up/down 或每日阈值市场。系统必须实时发现市场，不能硬编码某个市场 slug。当前第一阶段只做 paper trade。

工作方式：
1. 阅读 notes/polymarket/polymarket-paper-trading-workflow-20260610.md。
2. 使用 Polymarket Gamma API 搜索 BTC/ETH 相关 active、not closed、acceptingOrders、enableOrderBook 的市场。
3. 使用 CLOB 只读接口读取 token price 或 order book。
4. 建立本地 paper-trading 数据结构：market snapshot、signal、simulated order、fill、position、PnL。
5. 先实现一个简单策略：
   - 对 threshold market：根据 Binance BTC/USDT 或 ETH/USDT 当前价格、阈值、剩余时间，生成 HOLD/BUY_YES/BUY_NO。
   - 如果无法稳定取得 Binance 数据，先用策略 stub，但市场发现和价格读取必须真实可用。
6. 保存运行结果到 notes/polymarket/outputs/，包括 JSONL 交易日志和一份中文运行报告。

不要做：
- 不要接入真实私钥。
- 不要调用真实下单 endpoint。
- 不要假设一定存在活跃 5 分钟市场；必须先发现。
- 不要只写建议，请直接做一个能运行的最小版本。

验证：
- 运行一次 market discovery，列出发现的候选市场。
- 对至少一个候选市场读取 CLOB price 或 order book。
- 执行至少一次 dry-run paper strategy，并写入本地日志。
- 最后报告实际发现了哪些市场、哪些接口成功、哪些接口被限制或失败。
```

## Implementation notes

Suggested local structure:

```text
notes/polymarket/
  polymarket-paper-trading-workflow-20260610.md
  outputs/
    market-scan-YYYYMMDD-HHMMSS.json
    paper-trades-YYYYMMDD-HHMMSS.jsonl
    run-report-YYYYMMDD-HHMMSS.md
  src/
    polymarket_client.py
    market_discovery.py
    paper_broker.py
    strategies.py
    run_paper_scan.py
```

If this evolves into real execution later, create a separate module and force a manual config flag such as `ENABLE_REAL_TRADING=true`. The default should always be paper-only.

