# Bitget API + Polymarket execution guide

Date: 2026-06-10

## Core point

Bitget API cannot directly trade on Polymarket.

There are three separate roles:

1. Bitget as data source: get BTC/ETH ticker, candles, trades, and order book.
2. Bitget as execution venue: place spot or futures orders on Bitget only.
3. Polymarket as execution venue: place YES/NO orders through Polymarket CLOB, not through Bitget.

For the five-minute Polymarket crypto up/down idea, the first real architecture should be:

```text
Bitget market data -> signal model -> Polymarket market discovery/orderbook -> paper decision
                                                            |
                                                            v
                                      optional later: Polymarket real order
                                      optional later: Bitget hedge order
```

## Recommended mode order

Do not start with real trading.

Use these modes in order:

1. `paper`: read Bitget and Polymarket, simulate all trades.
2. `shadow`: generate real orders but do not submit them; save the exact payload.
3. `bitget_real_small`: place tiny Bitget-only orders to verify Bitget signing, order status, cancellation, and fills.
4. `polymarket_real_small`: place tiny Polymarket CLOB orders only after wallet, allowances, fees, market rules, and regional eligibility are clear.
5. `hedged_real`: Polymarket order plus optional Bitget futures hedge.

## Bitget public data

For this strategy, Bitget public endpoints are enough for the signal side.

Useful endpoints:

```text
GET /api/v2/spot/market/tickers?symbol=BTCUSDT
GET /api/v2/spot/market/candles?symbol=BTCUSDT&granularity=1min&limit=100
GET /api/v2/spot/market/orderbook?symbol=BTCUSDT&type=step0&limit=50

GET /api/v2/mix/market/tickers?productType=USDT-FUTURES
GET /api/v2/mix/market/candles?symbol=BTCUSDT&productType=USDT-FUTURES&granularity=1m&limit=100
```

Minimal Python:

```python
import requests

BASE = "https://api.bitget.com"

def bitget_get(path, params=None):
    r = requests.get(BASE + path, params=params, timeout=10)
    r.raise_for_status()
    data = r.json()
    if data.get("code") != "00000":
        raise RuntimeError(data)
    return data["data"]

btc_ticker = bitget_get("/api/v2/spot/market/tickers", {"symbol": "BTCUSDT"})
btc_candles = bitget_get(
    "/api/v2/spot/market/candles",
    {"symbol": "BTCUSDT", "granularity": "1min", "limit": "100"},
)
```

## Bitget authenticated signing

Bitget private endpoints need these headers:

```text
ACCESS-KEY
ACCESS-SIGN
ACCESS-TIMESTAMP
ACCESS-PASSPHRASE
Content-Type: application/json
```

The signature message is:

```text
timestamp + METHOD + requestPath + optional("?" + queryString) + body
```

Then sign with HMAC-SHA256 using the API secret and base64 encode the digest.

Minimal Python signer:

```python
import base64
import hashlib
import hmac
import json
import os
import time
import uuid

import requests

BASE = "https://api.bitget.com"

API_KEY = os.environ["BITGET_API_KEY"]
API_SECRET = os.environ["BITGET_API_SECRET"]
API_PASSPHRASE = os.environ["BITGET_API_PASSPHRASE"]

def bitget_timestamp_ms():
    return str(int(time.time() * 1000))

def bitget_sign(timestamp, method, request_path, body="", query_string=""):
    method = method.upper()
    suffix = f"?{query_string}" if query_string else ""
    payload = f"{timestamp}{method}{request_path}{suffix}{body}"
    digest = hmac.new(
        API_SECRET.encode("utf-8"),
        payload.encode("utf-8"),
        hashlib.sha256,
    ).digest()
    return base64.b64encode(digest).decode("utf-8")

def bitget_private(method, path, body_dict=None, query_string=""):
    body = json.dumps(body_dict or {}, separators=(",", ":")) if body_dict else ""
    ts = bitget_timestamp_ms()
    headers = {
        "ACCESS-KEY": API_KEY,
        "ACCESS-SIGN": bitget_sign(ts, method, path, body, query_string),
        "ACCESS-TIMESTAMP": ts,
        "ACCESS-PASSPHRASE": API_PASSPHRASE,
        "Content-Type": "application/json",
        "locale": "en-US",
    }
    url = BASE + path + (f"?{query_string}" if query_string else "")
    r = requests.request(method, url, headers=headers, data=body, timeout=10)
    r.raise_for_status()
    data = r.json()
    if data.get("code") != "00000":
        raise RuntimeError(data)
    return data["data"]
```

## Trading on Bitget

This only trades Bitget spot or futures. It does not buy Polymarket YES/NO shares.

Spot limit order:

```python
def place_bitget_spot_limit(symbol, side, price, base_size):
    if os.getenv("ENABLE_BITGET_REAL_TRADING") != "true":
        raise RuntimeError("real Bitget trading disabled")

    return bitget_private(
        "POST",
        "/api/v2/spot/trade/place-order",
        {
            "symbol": symbol,
            "side": side,                 # "buy" or "sell"
            "orderType": "limit",
            "force": "gtc",               # "gtc", "post_only", "fok", "ioc"
            "price": str(price),
            "size": str(base_size),        # limit order size is base coin amount
            "clientOid": f"ai4trading-{uuid.uuid4().hex[:24]}",
        },
    )
```

Spot market-buy detail:

- For spot limit orders, `size` is base coin amount.
- For spot market sell, `size` is base coin amount.
- For spot market buy, `size` is quote coin amount.

Futures order:

```python
def place_bitget_futures_limit_open_short(symbol, price, base_size):
    if os.getenv("ENABLE_BITGET_REAL_TRADING") != "true":
        raise RuntimeError("real Bitget trading disabled")

    return bitget_private(
        "POST",
        "/api/v2/mix/order/place-order",
        {
            "symbol": symbol,
            "productType": "USDT-FUTURES",
            "marginMode": "isolated",
            "marginCoin": "USDT",
            "size": str(base_size),
            "price": str(price),
            "side": "sell",
            "tradeSide": "open",
            "orderType": "limit",
            "force": "gtc",
            "clientOid": f"ai4trading-{uuid.uuid4().hex[:24]}",
        },
    )
```

For the Polymarket idea, Bitget futures should only be considered later as a hedge. Example: if you buy a Polymarket "BTC up" exposure and want to reduce raw BTC beta, you may short a small BTCUSDT perpetual position on Bitget. This is not necessary for the first experiment.

## Trading on Polymarket

Polymarket trading uses Polymarket CLOB.

Important mechanics:

- Public reads: Gamma API, Data API, and CLOB orderbook/prices/spreads do not require auth.
- Trading endpoints require CLOB authentication.
- Orders are signed locally; the SDK handles EIP-712 signing and submission.
- Polymarket "market orders" are implemented as marketable limit orders.

Minimal conceptual flow:

```python
# Pseudocode. Keep this disabled until wallet, allowances, and eligibility are clear.
from py_clob_client_v2 import ClobClient, Side, OrderType

client = ClobClient(
    host="https://clob.polymarket.com",
    chain_id=137,
    key=os.environ["POLY_PRIVATE_KEY"],
)
creds = client.create_or_derive_api_key()
client.set_api_creds(creds)

order = client.create_and_post_order(
    {
        "tokenID": token_id,
        "price": 0.51,
        "size": 10,
        "side": Side.BUY,
    },
    {"tickSize": "0.01", "negRisk": False},
    OrderType.GTC,
)
```

Before enabling this, the code must verify:

- wallet/funder address;
- USDC balance;
- token allowances;
- `conditionID`, `tokenID`, `tickSize`, and `negRisk`;
- `feesEnabled` and market fee parameters;
- market is active, not closed, accepting orders, and has order book enabled;
- regional eligibility and platform terms.

## Strategy execution loop

The first bot loop should be paper-only:

```text
every 5-10 seconds:
  1. fetch Bitget BTC/ETH ticker and 1min candles
  2. fetch Polymarket active crypto up/down markets
  3. fetch CLOB book/midpoint/spread for candidate token IDs
  4. parse market rule and resolution source
  5. estimate fair probability from Bitget momentum/volatility
  6. subtract Polymarket taker fee and spread
  7. decide HOLD / BUY_YES / BUY_NO
  8. in paper mode, simulate fill and write JSONL
```

For real trading, require two explicit switches:

```text
ENABLE_POLYMARKET_REAL_TRADING=true
MAX_REAL_ORDER_USDC=5
```

The program should refuse to trade if either is missing.

## Practical recommendation

Build these files next:

```text
notes/polymarket/src/
  bitget_client.py
  polymarket_read_client.py
  fee_model.py
  paper_broker.py
  run_bitget_poly_paper.py
```

The first acceptance test should be:

1. Bitget ticker and one-minute candles succeed.
2. Polymarket discovery either succeeds or records the exact connectivity failure.
3. The strategy produces only paper decisions.
4. No code path can place a real order unless an explicit environment flag is set.

## Sources

- Bitget signature docs: https://www.bitget.com/api-doc/common/signature
- Bitget spot place order: https://www.bitget.com/api-doc/spot/trade/Place-Order
- Bitget futures place order: https://www.bitget.com/api-doc/contract/trade/Place-Order
- Bitget spot candles: https://www.bitget.com/api-doc/spot/market/Get-Candle-Data
- Polymarket authentication: https://docs.polymarket.com/api-reference/authentication
- Polymarket market data overview: https://docs.polymarket.com/market-data/overview
- Polymarket create order: https://docs.polymarket.com/trading/orders/create
- Polymarket fees: https://docs.polymarket.com/trading/fees
