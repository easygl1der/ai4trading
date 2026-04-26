# Quickstart - Polymarket Documentation

Get up and running with the Polymarket API in minutes — fetch market data and place your first order.

## 1. Fetch a Market

All data endpoints are public — no API key or authentication needed. Use the markets endpoint to find a market and get its token IDs:

```python
import requests
response = requests.get(
    "https://gamma-api.polymarket.com/markets",
    params={"active": "true", "closed": "false", "limit": 1}
)
markets = response.json()
market = markets[0]
print(market["question"])
print(market["clobTokenIds"])
# ["123456...", "789012..."] — [Yes token ID, No token ID]
```

Save a token ID from `clobTokenIds` — you'll need it to place an order. The first ID is the Yes token, the second is the No token. See [Fetching Markets](https://docs.polymarket.com/market-data/fetching-markets) for more strategies like fetching by slug, tag, or event.

## 2. Install the SDK

```bash
pip install py_clob_client
```

## 3. Set Up Your Client

Derive API credentials and initialize the trading client:

```python
from py_clob_client.client import ClobClient
import os

host = "https://clob.polymarket.com"
chain = 137 # Polygon mainnet
private_key = os.getenv("PRIVATE_KEY")

# Derive API credentials (L1 → L2 auth)
temp_client = ClobClient(host, key=private_key, chain=chain)
api_creds = temp_client.create_or_derive_api_creds()

# Initialize trading client
client = ClobClient(
    host,
    key=private_key,
    chain=chain,
    creds=api_creds,
    signature_type=0, # Signature type: 0 = EOA
    funder="YOUR_WALLET_ADDRESS", # Funder address
)
```

> This example uses an EOA wallet (signature type 0) — your wallet pays its own gas. Proxy wallet users (types 1 and 2) can use Polymarket's gasless relayer instead. See [Authentication](https://docs.polymarket.com/trading/authentication) for details on signature types.
>
> Before trading, your funder address needs pUSD (for buying outcome tokens) and POL (for gas, if using EOA type 0).

## 4. Place an Order

Use the token_id from Step 1 to place a limit order:

```python
from py_clob_client.clob_types import OrderArgs, OrderType
from py_clob_client.order_builder.constants import BUY

# Fetch market details to get tick size and neg risk
market = client.get_market("YOUR_CONDITION_ID")
tick_size = str(market["minimum_tick_size"]) # e.g., "0.01"
neg_risk = market["neg_risk"] # e.g., False

response = client.create_and_post_order(
    OrderArgs(
        token_id="YOUR_TOKEN_ID", # From Step 1
        price=0.50,
        size=10,
        side=BUY,
        order_type=OrderType.GTC,
    ),
    options={
        "tick_size": tick_size,
        "neg_risk": neg_risk,
    },
)

print("Order ID:", response["orderID"])
print("Status:", response["status"])
```

## Next Steps

- [Authentication](https://docs.polymarket.com/trading/authentication) - Understand L1/L2 auth, signature types, and API credentials.
- [Trading Quickstart](https://docs.polymarket.com/trading/quickstart) - Detailed trading guide with order management and troubleshooting.
- [Fetching Markets](https://docs.polymarket.com/market-data/fetching-markets) - Strategies for discovering markets by slug, tag, or category.
- [Core Concepts](https://docs.polymarket.com/getting-started/core-concepts) - Understand markets, events, prices, and positions.
